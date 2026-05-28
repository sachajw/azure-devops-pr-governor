package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/sachajw/azure-devops-pr-scheduler/internal/models"
	"github.com/robfig/cron/v3"
	"github.com/pocketbase/pocketbase/core"
)

// ScheduleService evaluates cron schedules and orchestrates policy execution.
type ScheduleService struct {
	app           core.App
	policyService *PolicyService
	prService     *PullRequestService
	auditService  *AuditService
}

// NewScheduleService creates a new ScheduleService.
func NewScheduleService(app core.App, ps *PolicyService, prs *PullRequestService, as *AuditService) *ScheduleService {
	return &ScheduleService{
		app:           app,
		policyService: ps,
		prService:     prs,
		auditService:  as,
	}
}

// EvaluateAndExecute loads enabled policies, checks which are due, and executes them.
func (s *ScheduleService) EvaluateAndExecute(ctx context.Context, dryRun bool) (*models.ExecutionSummary, error) {
	summary := &models.ExecutionSummary{}

	policies, err := s.policyService.GetEnabledPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("load enabled policies: %w", err)
	}

	summary.TotalEvaluated = len(policies)

	for _, policy := range policies {
		if ctx.Err() != nil {
			return summary, fmt.Errorf("evaluation cancelled: %w", ctx.Err())
		}

		if !s.isCronDue(policy.Schedule.Cron, policy.Schedule.Timezone) {
			continue
		}

		summary.TotalMatched++

		// Create run record (single save, no transaction needed)
		runID, err := s.CreateRun(policy.ID, models.TriggeredBySchedule, dryRun)
		if err != nil {
			log.Printf("failed to create run for policy %s: %v", policy.ID, err)
			summary.TotalFailed++
			continue
		}

		// Transition to running (single save)
		s.updateRunStatus(runID, string(models.RunStatusRunning))

		// Execute policy actions — external ADO API calls, must be outside transaction
		results, err := s.prService.ExecutePolicyActions(ctx, policy, runID, dryRun)
		if err != nil {
			s.FinalizeRun(ctx, runID, string(models.RunStatusFailed), err.Error(), nil)
			summary.TotalFailed++
			continue
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
		if dryRun {
			status = models.RunStatusDryRun
		}

		s.FinalizeRun(ctx, runID, string(status), "", results)

		if allSuccess {
			summary.TotalSucceeded++
		} else {
			summary.TotalFailed++
		}
	}

	return summary, nil
}

// FinalizeRun atomically updates run status, error, result summary, and logs a completion audit event.
func (s *ScheduleService) FinalizeRun(ctx context.Context, runID, status, errMsg string, results []models.ActionResult) {
	err := s.app.RunInTransaction(func(txApp core.App) error {
		record, err := txApp.FindRecordById("runs", runID)
		if err != nil {
			return err
		}

		record.Set("status", status)
		if status == string(models.RunStatusSucceeded) || status == string(models.RunStatusFailed) || status == string(models.RunStatusDryRun) {
			record.Set("completed_at", time.Now().UTC().Format(time.RFC3339))
		}

		if errMsg != "" {
			record.Set("error", errMsg)
		}

		if results != nil {
			summaryJSON, jsonErr := json.Marshal(results)
			if jsonErr == nil {
				record.Set("result_summary", string(summaryJSON))
			}
		}

		if err := txApp.Save(record); err != nil {
			return err
		}

		// Log completion audit event within the same transaction
		if s.auditService != nil {
			collection, err := txApp.FindCollectionByNameOrId("audit_events")
			if err != nil {
				return err
			}

			auditRecord := core.NewRecord(collection)
			auditRecord.Set("run", runID)
			auditRecord.Set("event_type", "action_executed")
			auditRecord.Set("detail", fmt.Sprintf(`{"status":"%s"}`, status))
			if err := txApp.Save(auditRecord); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Printf("failed to finalize run %s: %v", runID, err)
	}
}

// isCronDue checks if a cron expression matches the current time in the given timezone.
func (s *ScheduleService) isCronDue(cronExpr, timezone string) bool {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		log.Printf("invalid cron expression %q: %v", cronExpr, err)
		return false
	}

	loc := time.UTC
	if tz, err := time.LoadLocation(timezone); err == nil {
		loc = tz
	}

	now := time.Now().In(loc)
	prev := schedule.Next(now.Add(-time.Minute))
	next := schedule.Next(now)

	return now.After(prev.Add(-time.Second)) && now.Before(next)
}

// CreateRun creates a new policy run record.
func (s *ScheduleService) CreateRun(policyID string, triggeredBy models.TriggeredBy, dryRun bool) (string, error) {
	collection, err := s.app.FindCollectionByNameOrId("runs")
	if err != nil {
		return "", err
	}

	var runID string

	err = s.app.RunInTransaction(func(txApp core.App) error {
		record := core.NewRecord(collection)
		record.Set("policy", policyID)
		record.Set("status", string(models.RunStatusPending))
		record.Set("triggered_by", string(triggeredBy))
		record.Set("dry_run", dryRun)
		record.Set("cron_eval_time", time.Now().UTC().Format(time.RFC3339))

		if err := txApp.Save(record); err != nil {
			return err
		}

		// Log audit event for run creation within the same transaction
		if s.auditService != nil {
			auditCol, err := txApp.FindCollectionByNameOrId("audit_events")
			if err != nil {
				return err
			}
			auditRecord := core.NewRecord(auditCol)
			auditRecord.Set("run", record.Id)
			auditRecord.Set("event_type", string(models.EventTypePolicyEvaluated))
			auditRecord.Set("detail", fmt.Sprintf(`{"triggered_by":"%s","dry_run":%v}`, triggeredBy, dryRun))
			if err := txApp.Save(auditRecord); err != nil {
				return err
			}
		}

		runID = record.Id
		return nil
	})
	if err != nil {
		return "", err
	}

	return runID, nil
}

// updateRunStatus updates the status of a run.
func (s *ScheduleService) updateRunStatus(runID, status string) {
	record, err := s.app.FindRecordById("runs", runID)
	if err != nil {
		return
	}

	record.Set("status", status)
	if status == string(models.RunStatusSucceeded) || status == string(models.RunStatusFailed) || status == string(models.RunStatusDryRun) {
		record.Set("completed_at", time.Now().UTC().Format(time.RFC3339))
	}

	if err := s.app.Save(record); err != nil {
		log.Printf("failed to update run %s status: %v", runID, err)
	}
}
