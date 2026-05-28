package hooks_test

import (
	"testing"

	"github.com/pangarabbit/azure-devops-pr-governor/internal/testhelpers"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/dbx"
)

func TestPolicyHooks(t *testing.T) {
	t.Run("should create a valid policy", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		testhelpers.CreatePolicyRecord(t, app)

		// Verify it was saved
		collection, err := app.FindCollectionByNameOrId("policies")
		if err != nil {
			t.Fatal(err)
		}

		records, err := app.FindRecordsByFilter(
			collection.Id,
			"name = {:name}",
			"-created",
			1, 0,
			dbx.Params{"name": "test-policy"},
		)
		if err != nil {
			t.Fatal(err)
		}
		if len(records) != 1 {
			t.Fatalf("expected 1 policy, got %d", len(records))
		}
	})

	t.Run("should reject policy without name", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		collection, err := app.FindCollectionByNameOrId("policies")
		if err != nil {
			t.Fatal(err)
		}

		record := core.NewRecord(collection)
		record.Set("name", "")
		record.Set("version", "1.0.0")
		record.Set("scope_org", "testorg")
		record.Set("scope_project", "testproject")
		record.Set("schedule_cron", "0 1 * * *")
		record.Set("actions", []any{})

		err = app.Save(record)
		if err == nil {
			t.Fatal("expected validation error for missing name")
		}
	})

	t.Run("should reject policy without version", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		collection, err := app.FindCollectionByNameOrId("policies")
		if err != nil {
			t.Fatal(err)
		}

		record := core.NewRecord(collection)
		record.Set("name", "test")
		record.Set("version", "")
		record.Set("scope_org", "testorg")
		record.Set("scope_project", "testproject")
		record.Set("schedule_cron", "0 1 * * *")
		record.Set("actions", []any{})

		err = app.Save(record)
		if err == nil {
			t.Fatal("expected validation error for missing version")
		}
	})

	t.Run("should reject policy without scope_org", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		collection, err := app.FindCollectionByNameOrId("policies")
		if err != nil {
			t.Fatal(err)
		}

		record := core.NewRecord(collection)
		record.Set("name", "test")
		record.Set("version", "1.0.0")
		record.Set("scope_org", "")
		record.Set("scope_project", "testproject")
		record.Set("schedule_cron", "0 1 * * *")
		record.Set("actions", []any{})

		err = app.Save(record)
		if err == nil {
			t.Fatal("expected validation error for missing scope_org")
		}
	})

	t.Run("should reject policy without schedule_cron", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		collection, err := app.FindCollectionByNameOrId("policies")
		if err != nil {
			t.Fatal(err)
		}

		record := core.NewRecord(collection)
		record.Set("name", "test")
		record.Set("version", "1.0.0")
		record.Set("scope_org", "testorg")
		record.Set("scope_project", "testproject")
		record.Set("schedule_cron", "")
		record.Set("actions", []any{})

		err = app.Save(record)
		if err == nil {
			t.Fatal("expected validation error for missing schedule_cron")
		}
	})
}

func TestRunHooks(t *testing.T) {
	t.Run("should create audit event on run creation", func(t *testing.T) {
		app := testhelpers.NewTestApp(t)
		defer app.Cleanup()

		policyID := testhelpers.CreatePolicyRecord(t, app)

		// Create a run record
		runsCol, err := app.FindCollectionByNameOrId("runs")
		if err != nil {
			t.Fatal(err)
		}

		run := core.NewRecord(runsCol)
		run.Set("policy", policyID)
		run.Set("status", "pending")
		run.Set("triggered_by", "manual")
		run.Set("dry_run", false)

		if err := app.Save(run); err != nil {
			t.Fatalf("failed to create run: %v", err)
		}

		// Verify audit event was created
		auditCol, err := app.FindCollectionByNameOrId("audit_events")
		if err != nil {
			t.Fatal(err)
		}

		auditRecords, err := app.FindRecordsByFilter(
			auditCol.Id,
			"run = {:runId}",
			"-created",
			10, 0,
			dbx.Params{"runId": run.Id},
		)
		if err != nil {
			t.Fatalf("failed to query audit events: %v", err)
		}

		if len(auditRecords) == 0 {
			t.Fatal("expected at least 1 audit event after run creation, got 0")
		}

		eventType := auditRecords[0].GetString("event_type")
		if eventType != "policy_evaluated" {
			t.Errorf("expected event_type 'policy_evaluated', got %q", eventType)
		}
	})
}
