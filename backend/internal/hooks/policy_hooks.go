package hooks

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
)

// RegisterPolicyHooks attaches validation hooks to the policies collection.
func RegisterPolicyHooks(app core.App) {
	app.OnRecordCreate("policies").Bind(&hook.Handler[*core.RecordEvent]{
		Id: "policy-create-validate",
		Func: func(e *core.RecordEvent) error {
			if err := validatePolicyRecord(e.Record); err != nil {
				return fmt.Errorf("policy validation failed: %w", err)
			}
			return e.Next()
		},
	})

	app.OnRecordUpdate("policies").Bind(&hook.Handler[*core.RecordEvent]{
		Id: "policy-update-validate",
		Func: func(e *core.RecordEvent) error {
			if err := validatePolicyRecord(e.Record); err != nil {
				return fmt.Errorf("policy validation failed: %w", err)
			}
			return e.Next()
		},
	})
}

// validatePolicyRecord validates a policy record before it is saved.
func validatePolicyRecord(r *core.Record) error {
	name := r.GetString("name")
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if len(name) > 200 {
		return fmt.Errorf("name must be 200 characters or less")
	}

	version := r.GetString("version")
	if version == "" {
		return fmt.Errorf("version is required")
	}

	scopeOrg := r.GetString("scope_org")
	if scopeOrg == "" {
		return fmt.Errorf("scope_org is required")
	}

	scopeProject := r.GetString("scope_project")
	if scopeProject == "" {
		return fmt.Errorf("scope_project is required")
	}

	cron := r.GetString("schedule_cron")
	if cron == "" {
		return fmt.Errorf("schedule_cron is required")
	}

	actionsRaw := r.Get("actions")
	if actionsRaw == nil {
		return fmt.Errorf("actions is required")
	}
	if actionsStr, ok := actionsRaw.(string); ok {
		if actionsStr == "" {
			return fmt.Errorf("actions is required")
		}
		var actions []interface{}
		if err := json.Unmarshal([]byte(actionsStr), &actions); err != nil {
			return fmt.Errorf("actions must be valid JSON array")
		}
		if len(actions) == 0 {
			return fmt.Errorf("actions must contain at least one action")
		}
	}

	conditionsRaw := r.Get("conditions")
	if conditionsRaw != nil {
		if conditionsStr, ok := conditionsRaw.(string); ok && conditionsStr != "" {
			var conditions []interface{}
			if err := json.Unmarshal([]byte(conditionsStr), &conditions); err != nil {
				return fmt.Errorf("conditions must be valid JSON array")
			}
		}
	}

	log.Printf("policy record validated: %s (%s)", name, r.Id)
	return nil
}
