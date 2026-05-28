package services_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/sachajw/azure-devops-pr-scheduler/internal/models"
	"github.com/sachajw/azure-devops-pr-scheduler/internal/services"
	"github.com/sachajw/azure-devops-pr-scheduler/internal/testhelpers"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/dbx"
)

func TestScheduleService_CreateRun(t *testing.T) {
	t.Run("should create run and audit event atomically", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		policyID := testhelpers.CreatePolicyRecord(t, app)

		policyService := services.NewPolicyService(app)
		auditService := services.NewAuditService(app)
		prService := services.NewPullRequestService(nil, auditService)
		scheduleService := services.NewScheduleService(app, policyService, prService, auditService)

		runID, err := scheduleService.CreateRun(policyID, models.TriggeredByManual, true)
		if err != nil {
			t.Fatalf("CreateRun failed: %v", err)
		}
		if runID == "" {
			t.Fatal("expected non-empty run ID")
		}

		// Verify run record exists
		run, err := app.FindRecordById("runs", runID)
		if err != nil {
			t.Fatalf("failed to find run: %v", err)
		}

		if run.GetString("status") != "pending" {
			t.Errorf("expected status 'pending', got %q", run.GetString("status"))
		}
		if run.GetString("triggered_by") != "manual" {
			t.Errorf("expected triggered_by 'manual', got %q", run.GetString("triggered_by"))
		}
		if !run.GetBool("dry_run") {
			t.Error("expected dry_run to be true")
		}

		// Verify audit event was created within the same transaction
		auditRecords, err := app.FindRecordsByFilter(
			"audit_events",
			"run = {:runId}",
			"-created",
			10, 0,
			dbx.Params{"runId": runID},
		)
		if err != nil {
			t.Fatalf("failed to query audit events: %v", err)
		}
		if len(auditRecords) == 0 {
			t.Fatal("expected audit event from CreateRun transaction, got 0")
		}
	})
}

func TestScheduleService_FinalizeRun(t *testing.T) {
	t.Run("should update run status and create audit atomically", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		policyID := testhelpers.CreatePolicyRecord(t, app)

		auditService := services.NewAuditService(app)
		policyService := services.NewPolicyService(app)
		prService := services.NewPullRequestService(nil, auditService)
		scheduleService := services.NewScheduleService(app, policyService, prService, auditService)

		runID, err := scheduleService.CreateRun(policyID, models.TriggeredByManual, false)
		if err != nil {
			t.Fatalf("CreateRun failed: %v", err)
		}

		results := []models.ActionResult{
			{Type: "create_pr", Success: true, Detail: "dry run: would create PR"},
		}

		scheduleService.FinalizeRun(context.Background(), runID, string(models.RunStatusSucceeded), "", results)

		// Verify run was updated
		run, err := app.FindRecordById("runs", runID)
		if err != nil {
			t.Fatalf("failed to find run: %v", err)
		}

		if run.GetString("status") != "succeeded" {
			t.Errorf("expected status 'succeeded', got %q", run.GetString("status"))
		}
		if run.GetString("completed_at") == "" {
			t.Error("expected completed_at to be set")
		}
		if run.GetString("result_summary") == "" {
			t.Error("expected result_summary to be set")
		}
	})

	t.Run("should store error message on failed run", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		policyID := testhelpers.CreatePolicyRecord(t, app)

		auditService := services.NewAuditService(app)
		policyService := services.NewPolicyService(app)
		prService := services.NewPullRequestService(nil, auditService)
		scheduleService := services.NewScheduleService(app, policyService, prService, auditService)

		runID, err := scheduleService.CreateRun(policyID, models.TriggeredByManual, false)
		if err != nil {
			t.Fatalf("CreateRun failed: %v", err)
		}

		scheduleService.FinalizeRun(context.Background(), runID, string(models.RunStatusFailed), "something went wrong", nil)

		run, err := app.FindRecordById("runs", runID)
		if err != nil {
			t.Fatalf("failed to find run: %v", err)
		}

		if run.GetString("status") != "failed" {
			t.Errorf("expected status 'failed', got %q", run.GetString("status"))
		}
		if run.GetString("error") != "something went wrong" {
			t.Errorf("expected error 'something went wrong', got %q", run.GetString("error"))
		}
	})
}

func TestPolicyService_GetPoliciesByScope(t *testing.T) {
	t.Run("should return policies matching scope with parameterized query", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		// Create policies with different scopes
		testhelpers.CreatePolicyRecord(t, app, func(r *core.Record) {
			r.Set("name", "policy-a")
			r.Set("scope_org", "org1")
			r.Set("scope_project", "proj1")
			r.Set("scope_repo", "repo1")
		})
		testhelpers.CreatePolicyRecord(t, app, func(r *core.Record) {
			r.Set("name", "policy-b")
			r.Set("scope_org", "org1")
			r.Set("scope_project", "proj1")
			r.Set("scope_repo", "")
		})
		testhelpers.CreatePolicyRecord(t, app, func(r *core.Record) {
			r.Set("name", "policy-c")
			r.Set("scope_org", "org2")
			r.Set("scope_project", "proj2")
			r.Set("scope_repo", "repo2")
		})

		svc := services.NewPolicyService(app)

		// Should match org1/proj1/repo1 (exact match + empty repo wildcard)
		policies, err := svc.GetPoliciesByScope(context.Background(), "org1", "proj1", "repo1")
		if err != nil {
			t.Fatalf("GetPoliciesByScope failed: %v", err)
		}
		if len(policies) != 2 {
			t.Errorf("expected 2 policies for org1/proj1/repo1, got %d", len(policies))
		}

		// Should only match org2/proj2
		policies, err = svc.GetPoliciesByScope(context.Background(), "org2", "proj2", "repo2")
		if err != nil {
			t.Fatalf("GetPoliciesByScope failed: %v", err)
		}
		if len(policies) != 1 {
			t.Errorf("expected 1 policy for org2/proj2/repo2, got %d", len(policies))
		}

		// Should match nothing
		policies, err = svc.GetPoliciesByScope(context.Background(), "org3", "proj3", "repo3")
		if err != nil {
			t.Fatalf("GetPoliciesByScope failed: %v", err)
		}
		if len(policies) != 0 {
			t.Errorf("expected 0 policies for unknown scope, got %d", len(policies))
		}
	})
}

func TestAuditService(t *testing.T) {
	t.Run("should log and retrieve audit events", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		policyID := testhelpers.CreatePolicyRecord(t, app)

		auditService := services.NewAuditService(app)

		// Create a run first
		runsCol, _ := app.FindCollectionByNameOrId("runs")
		run := core.NewRecord(runsCol)
		run.Set("policy", policyID)
		run.Set("status", "running")
		run.Set("triggered_by", "manual")
		run.Set("dry_run", false)
		if err := app.Save(run); err != nil {
			t.Fatal(err)
		}

		// Log events
		err := auditService.LogEvent(context.Background(), run.Id, "condition_met", map[string]any{
			"condition_type": "branch_exists",
			"met":            true,
		})
		if err != nil {
			t.Fatalf("LogEvent failed: %v", err)
		}

		err = auditService.LogEvent(context.Background(), run.Id, "pr_created", map[string]any{
			"pr_id": 42,
		})
		if err != nil {
			t.Fatalf("LogEvent failed: %v", err)
		}

		// Retrieve audit trail
		events, err := auditService.GetRunAuditTrail(context.Background(), run.Id)
		if err != nil {
			t.Fatalf("GetRunAuditTrail failed: %v", err)
		}

		// Hook creates 1 "policy_evaluated" event + 2 manual events = 3 total
			if len(events) != 3 {
				t.Fatalf("expected 3 audit events, got %d", len(events))
			}

			foundCondition := false
			foundPR := false
			for _, e := range events {
				if e.EventType == models.EventTypeConditionMet {
					foundCondition = true
				}
				if e.EventType == models.EventTypePRCreated {
					foundPR = true
				}
			}
			if !foundCondition {
				t.Error("expected a condition_met event")
			}
			if !foundPR {
				t.Error("expected a pr_created event")
			}
	})

	t.Run("should log errors with context", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		policyID := testhelpers.CreatePolicyRecord(t, app)

		auditService := services.NewAuditService(app)

		runsCol, _ := app.FindCollectionByNameOrId("runs")
		run := core.NewRecord(runsCol)
		run.Set("policy", policyID)
		run.Set("status", "failed")
		run.Set("triggered_by", "schedule")
		run.Set("dry_run", false)
		if err := app.Save(run); err != nil {
			t.Fatal(err)
		}

		err := auditService.LogError(context.Background(), run.Id, fmt.Errorf("ADO API timeout"), map[string]any{
			"retry_count": 3,
		})
		if err != nil {
			t.Fatalf("LogError failed: %v", err)
		}

		events, err := auditService.GetRunAuditTrail(context.Background(), run.Id)
		if err != nil {
			t.Fatalf("GetRunAuditTrail failed: %v", err)
		}

		// Hook creates 1 "policy_evaluated" event + 1 manual error event = 2 total
			if len(events) != 2 {
				t.Fatalf("expected 2 audit events, got %d", len(events))
			}
			foundError := false
			for _, e := range events {
				if e.EventType == models.EventTypeError {
					foundError = true
				}
			}
			if !foundError {
				t.Error("expected an error event")
			}
	})
}
