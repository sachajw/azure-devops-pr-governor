package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/pollinate/azure-devops-pr-governor/internal/models"
	"github.com/robfig/cron/v3"
	"github.com/pocketbase/pocketbase/core"
)

// ScheduleService evaluates cron schedules and orchestrates policy execution.
type ScheduleService struct {
	app            core.App
	policyService  *PolicyService
	prService      *PullRequestService
	auditService   *AuditService
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
		if !s.isCronDue(policy.Schedule.Cron, policy.Schedule.Timezone) {
			continue
		}

		summary.TotalMatched++

		runID, err := s.CreateRun(policy.ID, models.TriggeredBySchedule, dryRun)
		if err != nil {
			log.Printf("failed to create run for policy %s: %v", policy.ID, err)
			summary.TotalFailed++
			continue
		}

		s.updateRunStatus(runID, string(models.RunStatusRunning))

		results, err := s.prService.ExecutePolicyActions(ctx, policy, runID, dryRun)
		if err != nil {
			s.updateRunStatus(runID, string(models.RunStatusFailed))
			s.auditService.LogError(ctx, runID, err, nil)
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

		s.updateRunStatus(runID, string(status))

		if allSuccess {
			summary.TotalSucceeded++
		} else {
			summary.TotalFailed++
		}
	}

	return summary, nil
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

	// Check if now falls within the current minute window
	return now.After(prev.Add(-time.Second)) && now.Before(next)
}

// CreateRun creates a new policy run record.
func (s *ScheduleService) CreateRun(policyID string, triggeredBy models.TriggeredBy, dryRun bool) (string, error) {
	collection, err := s.app.FindCollectionByNameOrId("runs")
	if err != nil {
		return "", err
	}

	record := core.NewRecord(collection)
	record.Set("policy", policyID)
	record.Set("status", string(models.RunStatusPending))
	record.Set("triggered_by", string(triggeredBy))
	record.Set("dry_run", dryRun)
	record.Set("cron_eval_time", time.Now().UTC().Format(time.RFC3339))

	if err := s.app.Save(record); err != nil {
		return "", err
	}

	return record.Id, nil
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

// updateRunError sets the error on a failed run.
func (s *ScheduleService) updateRunError(runID, errMsg string) {
	record, err := s.app.FindRecordById("runs", runID)
	if err != nil {
		return
	}

	record.Set("error", errMsg)
	record.Set("status", string(models.RunStatusFailed))
	record.Set("completed_at", time.Now().UTC().Format(time.RFC3339))

	if err := s.app.Save(record); err != nil {
		log.Printf("failed to update run %s error: %v", runID, err)
	}
}

// setRunResultSummary stores the result summary JSON.
func (s *ScheduleService) setRunResultSummary(runID string, results []models.ActionResult) {
	summary, err := json.Marshal(results)
	if err != nil {
		return
	}

	record, err := s.app.FindRecordById("runs", runID)
	if err != nil {
		return
	}

	record.Set("result_summary", string(summary))
	s.app.Save(record)
}
