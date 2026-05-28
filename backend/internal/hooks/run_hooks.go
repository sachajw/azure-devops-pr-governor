package hooks

import (
	"context"
	"log"

	"github.com/sachajw/azure-devops-pr-scheduler/internal/services"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
)

// RegisterRunHooks attaches hooks to the runs collection.
func RegisterRunHooks(app core.App) {
	app.OnRecordCreate("runs").Bind(&hook.Handler[*core.RecordEvent]{
		Id: "run-create-audit",
		Func: func(e *core.RecordEvent) error {
			if err := e.Next(); err != nil {
				return err
			}

			audit := services.NewAuditService(e.App)
			runID := e.Record.Id
			policyID := e.Record.GetString("policy")
			triggeredBy := e.Record.GetString("triggered_by")

			ctx := context.Background()
			if err := audit.LogEvent(ctx, runID, "policy_evaluated", map[string]interface{}{
				"policy_id":    policyID,
				"triggered_by": triggeredBy,
				"dry_run":      e.Record.GetBool("dry_run"),
			}); err != nil {
				log.Printf("failed to log audit event for run %s: %v", runID, err)
			}

			return nil
		},
	})
}
