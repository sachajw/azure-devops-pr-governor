package hooks

import (
	"context"
	"log"

	"github.com/pollinate/azure-devops-pr-governor/internal/services"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterRunHooks attaches hooks to the runs collection.
func RegisterRunHooks(app core.App) {
	app.OnRecordCreate("runs").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		audit := services.NewAuditService(app)
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
	})
}
