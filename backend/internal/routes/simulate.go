package routes

import (
	"encoding/json"
	"net/http"

	"github.com/pangarabbit/azure-devops-pr-governor/internal/models"
	"github.com/pangarabbit/azure-devops-pr-governor/internal/services"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type simulateRequest struct {
	PolicyID string          `json:"policy_id,omitempty"`
	Policy   json.RawMessage `json:"policy,omitempty"`
}

type simulateResponse struct {
	PolicyID       string                 `json:"policy_id"`
	DryRun         bool                   `json:"dry_run"`
	Conditions     []models.ConditionResult `json:"conditions"`
	Actions        []models.ActionResult  `json:"actions"`
	WouldCreatePR  bool                   `json:"would_create_pr"`
}

// RegisterSimulateRoute registers the POST /api/pr-governor/simulate route.
func RegisterSimulateRoute(app core.App) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.POST("/api/pr-governor/simulate", func(e *core.RequestEvent) error {
			var req simulateRequest
			if err := e.BindBody(&req); err != nil {
				return e.JSON(http.StatusBadRequest, map[string]string{
					"error": "invalid request body",
				})
			}

			var policy *models.Policy
			var err error

			if req.PolicyID != "" {
				ps := services.NewPolicyService(app)
				policy, err = ps.GetPolicyByID(e.Request.Context(), req.PolicyID)
				if err != nil {
					return e.JSON(http.StatusNotFound, map[string]string{
						"error": "policy not found: " + req.PolicyID,
					})
				}
			} else if len(req.Policy) > 0 {
				policy = &models.Policy{}
				if err := json.Unmarshal(req.Policy, policy); err != nil {
					return e.JSON(http.StatusBadRequest, map[string]string{
						"error": "invalid policy JSON: " + err.Error(),
					})
				}
			} else {
				return e.JSON(http.StatusBadRequest, map[string]string{
					"error": "provide either policy_id or policy",
				})
			}

			adoClient := newADOClientFromEnv()
			audit := services.NewAuditService(app)
			prService := services.NewPullRequestService(adoClient, audit)

			condResults, actionResults, err := prService.SimulatePolicy(e.Request.Context(), policy)
			if err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]string{
					"error": "simulation failed: " + err.Error(),
				})
			}

			wouldCreatePR := false
			for _, ar := range actionResults {
				if ar.Success && ar.Type == "create_pr" {
					wouldCreatePR = true
					break
				}
			}

			return e.JSON(http.StatusOK, simulateResponse{
				PolicyID:      policy.ID,
				DryRun:        true,
				Conditions:    condResults,
				Actions:       actionResults,
				WouldCreatePR: wouldCreatePR,
			})
		}).Bind(apis.RequireSuperuserAuth())

		return se.Next()
	})
}
