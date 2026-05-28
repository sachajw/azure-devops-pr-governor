package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sachajw/azure-devops-pr-scheduler/internal/models"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/dbx"
)

// PolicyService handles policy CRUD, validation, and scope resolution.
type PolicyService struct {
	app core.App
}

// NewPolicyService creates a new PolicyService.
func NewPolicyService(app core.App) *PolicyService {
	return &PolicyService{app: app}
}

// GetEnabledPolicies returns all policies that are enabled and have scheduling enabled.
func (s *PolicyService) GetEnabledPolicies(ctx context.Context) ([]*models.Policy, error) {
	collection, err := s.app.FindCollectionByNameOrId("policies")
	if err != nil {
		return nil, fmt.Errorf("find policies collection: %w", err)
	}

	records, err := s.app.FindRecordsByFilter(
		collection.Id,
		"enabled = true && schedule_enabled = true",
		"-created",
		0,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("query enabled policies: %w", err)
	}

	policies := make([]*models.Policy, 0, len(records))
	for _, r := range records {
		p, err := recordToPolicy(r)
		if err != nil {
			continue
		}
		policies = append(policies, p)
	}

	return policies, nil
}

// GetPoliciesByScope returns policies matching the given scope.
func (s *PolicyService) GetPoliciesByScope(ctx context.Context, org, project, repo string) ([]*models.Policy, error) {
	collection, err := s.app.FindCollectionByNameOrId("policies")
	if err != nil {
		return nil, fmt.Errorf("find policies collection: %w", err)
	}

	filter := "scope_org = {:org} && scope_project = {:project}"
	params := dbx.Params{"org": org, "project": project}

	if repo != "" {
		filter += " && (scope_repo = '' || scope_repo = {:repo})"
		params["repo"] = repo
	}

	records, err := s.app.FindRecordsByFilter(
		collection.Id,
		filter,
		"-created",
		0,
		0,
		params,
	)
	if err != nil {
		return nil, fmt.Errorf("query policies by scope: %w", err)
	}

	policies := make([]*models.Policy, 0, len(records))
	for _, r := range records {
		p, err := recordToPolicy(r)
		if err != nil {
			continue
		}
		policies = append(policies, p)
	}

	return policies, nil
}

// GetPolicyByID returns a single policy by ID.
func (s *PolicyService) GetPolicyByID(ctx context.Context, id string) (*models.Policy, error) {
	record, err := s.app.FindRecordById("policies", id)
	if err != nil {
		return nil, fmt.Errorf("find policy %s: %w", id, err)
	}
	return recordToPolicy(record)
}

// ValidatePolicyJSON validates a raw JSON payload against the policy schema.
func (s *PolicyService) ValidatePolicyJSON(raw json.RawMessage) error {
	schemaPath := os.Getenv("POLICY_SCHEMA_PATH")
	if schemaPath == "" {
		schemaPath = "../schemas/policy.schema.json"
	}

	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read policy schema: %w", err)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		return fmt.Errorf("parse policy schema: %w", err)
	}

	var payload interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("parse policy payload: %w", err)
	}

	return validateAgainstSchema(payload, schema)
}

// validateAgainstSchema performs basic required-field validation.
func validateAgainstSchema(payload, schema interface{}) error {
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid schema")
	}

	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return fmt.Errorf("policy must be a JSON object")
	}

	required, _ := schemaMap["required"].([]interface{})
	for _, r := range required {
		field, _ := r.(string)
		if _, exists := payloadMap[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	return nil
}

// recordToPolicy converts a PocketBase record to a Policy.
func recordToPolicy(r *core.Record) (*models.Policy, error) {
	p := &models.Policy{
		ID:          r.Id,
		Name:        r.GetString("name"),
		Description: r.GetString("description"),
		Version:     r.GetString("version"),
		Enabled:     r.GetBool("enabled"),
		Scope: models.PolicyScope{
			Organization: r.GetString("scope_org"),
			Project:      r.GetString("scope_project"),
			Repository:   r.GetString("scope_repo"),
		},
		Schedule: models.PolicySchedule{
			Cron:     r.GetString("schedule_cron"),
			Timezone: r.GetString("schedule_timezone"),
			Enabled:  r.GetBool("schedule_enabled"),
		},
		Created: r.GetDateTime("created").String(),
		Updated: r.GetDateTime("updated").String(),
	}

	p.Conditions = parseJSONField[[]models.PolicyCondition](r.Get("conditions"))
	p.Actions = parseJSONField[[]models.PolicyAction](r.Get("actions"))
	p.Constraints = parseJSONField[*models.PolicyConstraints](r.Get("constraints"))
	p.Tags = parseJSONField[[]string](r.Get("tags"))

	return p, nil
}

// parseJSONField handles PocketBase JSON fields that may be strings or already-parsed values.
func parseJSONField[T any](v interface{}) T {
	var zero T
	if v == nil {
		return zero
	}

	// PocketBase may return the value as a string (raw JSON) or as a parsed Go value
	if raw, ok := v.(string); ok && raw != "" {
		var result T
		if json.Unmarshal([]byte(raw), &result) == nil {
			return result
		}
	}

	// Already parsed by PocketBase — re-marshal to get into the target type
	reRaw, err := json.Marshal(v)
	if err != nil {
		return zero
	}

	var result T
	if json.Unmarshal(reRaw, &result) == nil {
		return result
	}

	return zero
}
