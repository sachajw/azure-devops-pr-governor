package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// --- policies collection ---
		policies := core.NewBaseCollection("policies")

		policies.Fields.Add(&core.TextField{
			Name:     "name",
			Required: true,
			Max:      200,
		})
		policies.Fields.Add(&core.TextField{
			Name: "description",
		})
		policies.Fields.Add(&core.TextField{
			Name:     "version",
			Required: true,
		})
		policies.Fields.Add(&core.BoolField{
			Name: "enabled",
		})
		policies.Fields.Add(&core.TextField{
			Name:     "scope_org",
			Required: true,
		})
		policies.Fields.Add(&core.TextField{
			Name:     "scope_project",
			Required: true,
		})
		policies.Fields.Add(&core.TextField{
			Name: "scope_repo",
		})
		policies.Fields.Add(&core.TextField{
			Name:     "schedule_cron",
			Required: true,
		})
		policies.Fields.Add(&core.TextField{
			Name: "schedule_timezone",
		})
		policies.Fields.Add(&core.BoolField{
			Name: "schedule_enabled",
		})
		policies.Fields.Add(&core.JSONField{
			Name: "conditions",
		})
		policies.Fields.Add(&core.JSONField{
			Name:     "actions",
			Required: true,
		})
		policies.Fields.Add(&core.JSONField{
			Name: "constraints",
		})
		policies.Fields.Add(&core.JSONField{
			Name: "tags",
		})
		policies.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		})
		policies.Fields.Add(&core.AutodateField{
			Name:     "updated",
			OnCreate: true,
			OnUpdate: true,
		})

		policies.AddIndex("idx_policies_scope", false, "scope_org, scope_project, scope_repo", "")
		policies.AddIndex("idx_policies_enabled", false, "enabled, schedule_enabled", "")

		if err := app.Save(policies); err != nil {
			return err
		}

		// --- runs collection ---
		runs := core.NewBaseCollection("runs")

		runs.Fields.Add(&core.RelationField{
			Name:         "policy",
			Required:     true,
			MaxSelect:    1,
			CollectionId: policies.Id,
		})
		runs.Fields.Add(&core.SelectField{
			Name:     "status",
			Required: true,
			Values:   []string{"pending", "running", "succeeded", "failed", "skipped", "dry_run"},
		})
		runs.Fields.Add(&core.SelectField{
			Name:     "triggered_by",
			Required: true,
			Values:   []string{"schedule", "simulation", "webhook", "manual"},
		})
		runs.Fields.Add(&core.BoolField{
			Name: "dry_run",
		})
		runs.Fields.Add(&core.DateField{
			Name: "cron_eval_time",
		})
		runs.Fields.Add(&core.AutodateField{
			Name:     "started_at",
			OnCreate: true,
		})
		runs.Fields.Add(&core.DateField{
			Name: "completed_at",
		})
		runs.Fields.Add(&core.TextField{
			Name: "error",
		})
		runs.Fields.Add(&core.JSONField{
			Name: "result_summary",
		})

		runs.AddIndex("idx_runs_policy", false, "policy", "")
		runs.AddIndex("idx_runs_status", false, "status", "")

		if err := app.Save(runs); err != nil {
			return err
		}

		// --- audit_events collection ---
		auditEvents := core.NewBaseCollection("audit_events")

		auditEvents.Fields.Add(&core.RelationField{
			Name:         "run",
			Required:     true,
			MaxSelect:    1,
			CollectionId: runs.Id,
		})
		auditEvents.Fields.Add(&core.SelectField{
			Name:     "event_type",
			Required: true,
			Values: []string{
				"policy_evaluated",
				"pr_created",
				"pr_skipped",
				"condition_met",
				"condition_failed",
				"action_executed",
				"error",
			},
		})
		auditEvents.Fields.Add(&core.TextField{
			Name: "actor",
		})
		auditEvents.Fields.Add(&core.JSONField{
			Name:     "detail",
			Required: true,
		})
		auditEvents.Fields.Add(&core.TextField{
			Name: "ado_org",
		})
		auditEvents.Fields.Add(&core.TextField{
			Name: "ado_project",
		})
		auditEvents.Fields.Add(&core.TextField{
			Name: "ado_repo",
		})
		auditEvents.Fields.Add(&core.NumberField{
			Name: "ado_pr_id",
		})
		auditEvents.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		})

		auditEvents.AddIndex("idx_audit_run", false, "run", "")
		auditEvents.AddIndex("idx_audit_event_type", false, "event_type", "")

		return app.Save(auditEvents)
	}, func(app core.App) error {
		for _, name := range []string{"audit_events", "runs", "policies"} {
			col, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				continue
			}
			if err := app.Delete(col); err != nil {
				return err
			}
		}
		return nil
	})
}
