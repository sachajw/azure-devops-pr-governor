package services

import (
	"encoding/json"
	"testing"

	"github.com/sachajw/azure-devops-pr-scheduler/internal/models"
)

func TestPolicyModel(t *testing.T) {
	policy := &models.Policy{
		Name:    "test",
		Version: "1.0.0",
		Enabled: true,
		Scope: models.PolicyScope{
			Organization: "myorg",
			Project:      "myproject",
			Repository:   "myrepo",
		},
		Schedule: models.PolicySchedule{
			Cron:     "0 1 * * *",
			Timezone: "Africa/Johannesburg",
			Enabled:  true,
		},
		Actions: []models.PolicyAction{
			{
				Type: "create_pr",
				Parameters: map[string]interface{}{
					"sourceRefName": "refs/heads/dev",
					"targetRefName": "refs/heads/qa",
					"title":         "Auto PR: dev -> qa",
				},
			},
		},
	}

	if policy.Name != "test" {
		t.Errorf("expected name 'test', got %s", policy.Name)
	}
	if policy.Scope.Organization != "myorg" {
		t.Errorf("expected org 'myorg', got %s", policy.Scope.Organization)
	}
	if len(policy.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(policy.Actions))
	}
}

func TestPolicyJSONRoundtrip(t *testing.T) {
	original := models.Policy{
		Name:        "nightly-dev-to-qa",
		Description: "Automated nightly promotion",
		Version:     "1.0.0",
		Enabled:     true,
		Scope: models.PolicyScope{
			Organization: "pollinate",
			Project:      "cloudops",
			Repository:   "infrastructure",
		},
		Schedule: models.PolicySchedule{
			Cron:     "0 1 * * *",
			Timezone: "Africa/Johannesburg",
			Enabled:  true,
		},
		Conditions: []models.PolicyCondition{},
		Actions: []models.PolicyAction{
			{
				Type: "create_pr",
				Parameters: map[string]interface{}{
					"sourceRefName": "refs/heads/dev",
					"targetRefName": "refs/heads/qa",
					"title":         "Auto PR: dev -> qa",
				},
			},
		},
		Constraints: &models.PolicyConstraints{
			MaxActivePRs:  intPtr(3),
			MergeStrategy: "squash",
		},
		Tags: []string{"nightly", "promotion"},
	}

	raw, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded models.Policy
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("name mismatch: got %s", decoded.Name)
	}
	if decoded.Scope.Organization != original.Scope.Organization {
		t.Errorf("scope.org mismatch: got %s", decoded.Scope.Organization)
	}
	if decoded.Schedule.Cron != original.Schedule.Cron {
		t.Errorf("schedule.cron mismatch: got %s", decoded.Schedule.Cron)
	}
	if len(decoded.Actions) != 1 {
		t.Errorf("actions count mismatch: got %d", len(decoded.Actions))
	}
	if decoded.Constraints == nil || *decoded.Constraints.MaxActivePRs != 3 {
		t.Errorf("constraints.max_active_prs mismatch")
	}
}

func TestPolicyValidation(t *testing.T) {
	svc := &PolicyService{}

	tests := []struct {
		name    string
		policy  map[string]interface{}
		wantErr bool
	}{
		{
			name: "missing name",
			policy: map[string]interface{}{
				"version": "1.0.0",
				"scope":   map[string]interface{}{"organization": "org", "project": "proj"},
				"schedule": map[string]interface{}{"cron": "0 1 * * *"},
				"actions": []interface{}{},
			},
			wantErr: true,
		},
		{
			name: "missing version",
			policy: map[string]interface{}{
				"name":    "test",
				"scope":   map[string]interface{}{"organization": "org", "project": "proj"},
				"schedule": map[string]interface{}{"cron": "0 1 * * *"},
				"actions": []interface{}{},
			},
			wantErr: true,
		},
		{
			name: "missing scope",
			policy: map[string]interface{}{
				"name":     "test",
				"version":  "1.0.0",
				"schedule": map[string]interface{}{"cron": "0 1 * * *"},
				"actions":  []interface{}{},
			},
			wantErr: true,
		},
		{
			name: "missing schedule",
			policy: map[string]interface{}{
				"name":    "test",
				"version": "1.0.0",
				"scope":   map[string]interface{}{"organization": "org", "project": "proj"},
				"actions": []interface{}{},
			},
			wantErr: true,
		},
		{
			name: "missing actions",
			policy: map[string]interface{}{
				"name":     "test",
				"version":  "1.0.0",
				"scope":    map[string]interface{}{"organization": "org", "project": "proj"},
				"schedule": map[string]interface{}{"cron": "0 1 * * *"},
			},
			wantErr: true,
		},
		{
			name: "non-object payload",
			policy: map[string]interface{}{
				"just": "a string",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, _ := json.Marshal(tt.policy)
			err := svc.ValidatePolicyJSON(raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePolicyJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRecordToPolicyConversion(t *testing.T) {
	// Test that the record conversion works with the models.Policy struct
	raw := `{
		"name": "test-policy",
		"version": "1.0.0",
		"scope": {
			"organization": "myorg",
			"project": "myproject",
			"repository": "myrepo"
		},
		"schedule": {
			"cron": "0 1 * * *",
			"timezone": "UTC",
			"enabled": true
		},
		"actions": [
			{
				"type": "create_pr",
				"parameters": {
					"sourceRefName": "refs/heads/dev",
					"targetRefName": "refs/heads/qa",
					"title": "Auto PR"
				}
			}
		]
	}`

	var p models.Policy
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if p.Name != "test-policy" {
		t.Errorf("expected name 'test-policy', got %s", p.Name)
	}
	if p.Scope.Organization != "myorg" {
		t.Errorf("expected org 'myorg', got %s", p.Scope.Organization)
	}
	if p.Schedule.Cron != "0 1 * * *" {
		t.Errorf("expected cron '0 1 * * *', got %s", p.Schedule.Cron)
	}
	if len(p.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(p.Actions))
	}
}

func intPtr(i int) *int {
	return &i
}
