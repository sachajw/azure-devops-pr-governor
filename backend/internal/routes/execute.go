package routes

import (
	"net/http"
	"time"

	"github.com/sachajw/azure-devops-pr-scheduler/internal/models"
	"github.com/sachajw/azure-devops-pr-scheduler/internal/services"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type executeRequest struct {
	PolicyID string `json:"policy_id"`
	DryRun   bool   `json:"dry_run,omitempty"`
}

// RegisterExecuteRoute registers the POST /api/pr-scheduler/execute route.
func RegisterExecuteRoute(app core.App) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.POST("/api/pr-scheduler/execute", func(e *core.RequestEvent) error {
			var req executeRequest
			if err := e.BindBody(&req); err != nil {
				return e.JSON(http.StatusBadRequest, map[string]string{
					"error": "invalid request body",
				})
			}

			if req.PolicyID == "" {
				return e.JSON(http.StatusBadRequest, map[string]string{
					"error": "policy_id is required",
				})
			}

			policyService := services.NewPolicyService(app)
			policy, err := policyService.GetPolicyByID(e.Request.Context(), req.PolicyID)
			if err != nil {
				return e.JSON(http.StatusNotFound, map[string]string{
					"error": "policy not found: " + req.PolicyID,
				})
			}

			adoClient := newADOClientFromEnv()
			auditService := services.NewAuditService(app)
			prService := services.NewPullRequestService(adoClient, auditService)

			triggeredBy := models.TriggeredByManual
			if req.DryRun {
				triggeredBy = models.TriggeredBySimulation
			}

			// Create run record
			scheduleService := services.NewScheduleService(app, policyService, prService, auditService)
			runID, err := scheduleService.CreateRun(policy.ID, triggeredBy, req.DryRun)
			if err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]string{
					"error": "failed to create run: " + err.Error(),
				})
			}

			// Execute
			results, err := prService.ExecutePolicyActions(e.Request.Context(), policy, runID, req.DryRun)
			if err != nil {
				scheduleService.FinalizeRun(e.Request.Context(), runID, string(models.RunStatusFailed), err.Error(), nil)
				return e.JSON(http.StatusInternalServerError, map[string]string{
					"error":   "execution failed: " + err.Error(),
					"run_id":  runID,
				})
			}

			allSuccess := true
			for _, r := range results {
				if !r.Success {
					allSuccess = false
					break
				}
			}

			status := models.RunStatusSucceeded
			if !allSuccess {
				status = models.RunStatusFailed
			}
			if req.DryRun {
				status = models.RunStatusDryRun
			}
			scheduleService.FinalizeRun(e.Request.Context(), runID, string(status), "", results)

			return e.JSON(http.StatusOK, map[string]interface{}{
				"run_id":   runID,
				"policy_id": policy.ID,
				"dry_run":  req.DryRun,
				"results":  results,
				"executed_at": time.Now().UTC().Format(time.RFC3339),
			})
		}).Bind(apis.RequireSuperuserAuth())

		return se.Next()
	})
}
