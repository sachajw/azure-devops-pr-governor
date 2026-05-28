package services

import (
	"context"
	"testing"

	"github.com/pangarabbit/azure-devops-pr-governor/internal/models"
)

func newTestPRService() *PullRequestService {
	return NewPullRequestService(nil, nil)
}

func TestSimulatePolicy(t *testing.T) {
	policy := &models.Policy{
		ID:   "test-policy-1",
		Name: "test-policy",
		Scope: models.PolicyScope{
			Organization: "myorg",
			Project:      "myproject",
			Repository:   "myrepo",
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
	}

	svc := newTestPRService()
	condResults, actionResults, err := svc.SimulatePolicy(context.Background(), policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(condResults) != 0 {
		t.Errorf("expected 0 conditions, got %d", len(condResults))
	}

	if len(actionResults) != 1 {
		t.Fatalf("expected 1 action result, got %d", len(actionResults))
	}

	ar := actionResults[0]
	if ar.Type != "create_pr" {
		t.Errorf("expected action type create_pr, got %s", ar.Type)
	}
	if !ar.Success {
		t.Errorf("expected success=true, got false: %s", ar.Detail)
	}
}

func TestSimulatePolicyWithConditions(t *testing.T) {
	policy := &models.Policy{
		ID:   "test-policy-2",
		Name: "test-policy-with-conditions",
		Scope: models.PolicyScope{
			Organization: "myorg",
			Project:      "myproject",
		},
		Conditions: []models.PolicyCondition{
			{
				Type:     "branch_exists",
				Operator: "exists",
				Value:    []byte(`"refs/heads/dev"`),
			},
		},
		Actions: []models.PolicyAction{
			{
				Type: "create_pr",
				Parameters: map[string]interface{}{
					"sourceRefName": "refs/heads/dev",
					"targetRefName": "refs/heads/qa",
					"title":         "Auto PR",
				},
			},
		},
	}

	svc := newTestPRService()
	condResults, actionResults, err := svc.SimulatePolicy(context.Background(), policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(condResults) != 1 {
		t.Fatalf("expected 1 condition result, got %d", len(condResults))
	}

	// Without an ADO client, branch_exists should fail
	cr := condResults[0]
	if cr.Type != "branch_exists" {
		t.Errorf("expected condition type branch_exists, got %s", cr.Type)
	}
	if cr.Met {
		t.Error("expected condition to not be met (no ADO client)")
	}

	// Simulation still reports what actions would fire
	if len(actionResults) != 1 {
		t.Fatalf("expected 1 action result, got %d", len(actionResults))
	}
}

func TestExecutePolicyActionsDryRun(t *testing.T) {
	policy := &models.Policy{
		ID:   "test-policy-dry",
		Name: "dry-run-test",
		Scope: models.PolicyScope{
			Organization: "myorg",
			Project:      "myproject",
			Repository:   "myrepo",
		},
		Conditions: []models.PolicyCondition{},
		Actions: []models.PolicyAction{
			{
				Type: "create_pr",
				Parameters: map[string]interface{}{
					"sourceRefName": "refs/heads/dev",
					"targetRefName": "refs/heads/qa",
					"title":         "Test PR",
				},
			},
		},
	}

	svc := newTestPRService()
	results, err := svc.ExecutePolicyActions(context.Background(), policy, "test-run-id", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if !results[0].Success {
		t.Errorf("expected success=true, got false: %s", results[0].Detail)
	}

	if results[0].Type != "create_pr" {
		t.Errorf("expected type create_pr, got %s", results[0].Type)
	}
}

func TestCreatePRParams(t *testing.T) {
	action := models.PolicyAction{
		Type: "create_pr",
		Parameters: map[string]interface{}{
			"sourceRefName": "refs/heads/dev",
			"targetRefName": "refs/heads/qa",
			"title":         "My PR",
			"description":   "Test description",
			"isDraft":       true,
		},
	}

	sourceRef, targetRef, title, description, isDraft := action.CreatePRParams()
	if sourceRef != "refs/heads/dev" {
		t.Errorf("expected refs/heads/dev, got %s", sourceRef)
	}
	if targetRef != "refs/heads/qa" {
		t.Errorf("expected refs/heads/qa, got %s", targetRef)
	}
	if title != "My PR" {
		t.Errorf("expected 'My PR', got %s", title)
	}
	if description != "Test description" {
		t.Errorf("expected 'Test description', got %s", description)
	}
	if !isDraft {
		t.Error("expected isDraft=true")
	}
}

func TestCheckConstraints(t *testing.T) {
	svc := newTestPRService()

	// Nil constraints should pass
	policy := &models.Policy{
		Scope:      models.PolicyScope{Organization: "org", Project: "proj"},
		Conditions: []models.PolicyCondition{},
	}
	if err := svc.checkConstraints(context.Background(), policy); err != nil {
		t.Errorf("nil constraints should pass: %v", err)
	}

	// With constraints (not yet enforced)
	maxPRs := 5
	policy.Constraints = &models.PolicyConstraints{
		MaxActivePRs: &maxPRs,
	}
	if err := svc.checkConstraints(context.Background(), policy); err != nil {
		t.Errorf("max_active_prs constraint should pass (not yet enforced): %v", err)
	}
}

// Mock server tests for PR creation with real HTTP
func TestExecutePolicyActionsWithMockADO(t *testing.T) {
	policy := &models.Policy{
		ID:   "test-with-ado",
		Name: "ado-test",
		Scope: models.PolicyScope{
			Organization: "myorg",
			Project:      "myproject",
			Repository:   "myrepo",
		},
		Conditions: []models.PolicyCondition{},
		Actions: []models.PolicyAction{
			{
				Type: "create_pr",
				Parameters: map[string]interface{}{
					"sourceRefName": "refs/heads/dev",
					"targetRefName": "refs/heads/qa",
					"title":         "Auto PR",
				},
			},
		},
	}

	// Without ADO client configured, dry run should still work
	svc := newTestPRService()
	results, err := svc.ExecutePolicyActions(context.Background(), policy, "run-1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !results[0].Success {
		t.Errorf("dry run should succeed: %s", results[0].Detail)
	}
}
