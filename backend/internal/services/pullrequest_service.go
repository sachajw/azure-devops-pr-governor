package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pollinate/azure-devops-pr-governor/internal/azuredevops"
	"github.com/pollinate/azure-devops-pr-governor/internal/models"
)

// PullRequestService evaluates conditions and creates PRs via the Azure DevOps API.
type PullRequestService struct {
	adoClient *azuredevops.Client
	audit     *AuditService
}

// NewPullRequestService creates a new PullRequestService.
func NewPullRequestService(adoClient *azuredevops.Client, audit *AuditService) *PullRequestService {
	return &PullRequestService{
		adoClient: adoClient,
		audit:     audit,
	}
}

// logEvent safely logs an audit event, skipping if audit is nil.
func (s *PullRequestService) logEvent(ctx context.Context, runID, eventType string, detail map[string]interface{}) {
	if s.audit == nil {
		return
	}
	s.logEvent(ctx, runID, eventType, detail)
}

// ExecutePolicyActions evaluates conditions and executes actions for a policy.
func (s *PullRequestService) ExecutePolicyActions(ctx context.Context, policy *models.Policy, runID string, dryRun bool) ([]models.ActionResult, error) {
	// Evaluate conditions
	condResults := s.evaluateConditions(ctx, policy)
	allConditionsMet := true
	for _, cr := range condResults {
		eventType := models.EventTypeConditionMet
		if !cr.Met {
			eventType = models.EventTypeConditionFailed
			allConditionsMet = false
		}
		s.logEvent(ctx, runID, string(eventType), map[string]interface{}{
			"condition_type": cr.Type,
			"met":            cr.Met,
			"reason":         cr.Reason,
		})
	}

	if !allConditionsMet {
		return []models.ActionResult{{
			Type:    "conditions",
			Success: false,
			Detail:  "not all conditions met",
		}}, nil
	}

	// Check constraints
	if err := s.checkConstraints(ctx, policy); err != nil {
		s.logEvent(ctx, runID, string(models.EventTypePRSkipped), map[string]interface{}{
			"reason": err.Error(),
		})
		return []models.ActionResult{{
			Type:    "constraints",
			Success: false,
			Detail:  err.Error(),
		}}, nil
	}

	// Execute actions
	var results []models.ActionResult
	for _, action := range policy.Actions {
		result := s.executeAction(ctx, policy, runID, action, dryRun)
		results = append(results, result)

		s.logEvent(ctx, runID, string(models.EventTypeActionExecuted), map[string]interface{}{
			"action_type": action.Type,
			"success":     result.Success,
			"detail":      result.Detail,
		})
	}

	return results, nil
}

// evaluateConditions checks all conditions for a policy.
func (s *PullRequestService) evaluateConditions(ctx context.Context, policy *models.Policy) []models.ConditionResult {
	results := make([]models.ConditionResult, 0, len(policy.Conditions))

	for _, cond := range policy.Conditions {
		var result models.ConditionResult

		switch cond.Type {
		case "branch_exists":
			result = s.evaluateBranchExists(ctx, policy, cond)
		default:
			result = models.ConditionResult{
				Type:   cond.Type,
				Met:    true,
				Reason: "condition type not yet implemented, passing by default",
			}
		}

		results = append(results, result)
	}

	return results
}

// evaluateBranchExists checks if a branch exists in the repository.
func (s *PullRequestService) evaluateBranchExists(ctx context.Context, policy *models.Policy, cond models.PolicyCondition) models.ConditionResult {
	var branchName string
	_ = json.Unmarshal(cond.Value, &branchName)

	if branchName == "" {
		return models.ConditionResult{
			Type:   cond.Type,
			Met:    false,
			Reason: "no branch name provided in condition value",
		}
	}

	if s.adoClient == nil {
		return models.ConditionResult{
			Type:   cond.Type,
			Met:    false,
			Reason: "azure devops client not configured",
		}
	}

	repo := policy.Scope.Repository
	if repo == "" {
		return models.ConditionResult{
			Type:   cond.Type,
			Met:    false,
			Reason: "no repository specified in policy scope",
		}
	}

	exists, err := s.adoClient.BranchExists(ctx, policy.Scope.Project, repo, branchName)
	if err != nil {
		return models.ConditionResult{
			Type:   cond.Type,
			Met:    false,
			Reason: fmt.Sprintf("error checking branch: %v", err),
		}
	}

	met := exists
	if cond.Operator == "not_exists" {
		met = !exists
	}

	return models.ConditionResult{
		Type:   cond.Type,
		Met:    met,
		Reason: fmt.Sprintf("branch %s exists=%v", branchName, exists),
	}
}

// checkConstraints verifies constraints are met before creating a PR.
func (s *PullRequestService) checkConstraints(ctx context.Context, policy *models.Policy) error {
	if policy.Constraints == nil {
		return nil
	}

	if policy.Constraints.MaxActivePRs != nil && *policy.Constraints.MaxActivePRs > 0 {
		// In a full implementation, query Azure DevOps for active PRs
		// for the target branch and compare against the limit.
		log.Printf("max_active_prs constraint set to %d (not yet enforced)", *policy.Constraints.MaxActivePRs)
	}

	return nil
}

// executeAction executes a single policy action.
func (s *PullRequestService) executeAction(ctx context.Context, policy *models.Policy, runID string, action models.PolicyAction, dryRun bool) models.ActionResult {
	switch action.Type {
	case "create_pr":
		return s.createPullRequest(ctx, policy, runID, action, dryRun)
	default:
		return models.ActionResult{
			Type:    action.Type,
			Success: false,
			Detail:  fmt.Sprintf("action type %q not yet implemented", action.Type),
		}
	}
}

// createPullRequest creates a PR via the Azure DevOps API.
func (s *PullRequestService) createPullRequest(ctx context.Context, policy *models.Policy, runID string, action models.PolicyAction, dryRun bool) models.ActionResult {
	sourceRef, targetRef, title, description, isDraft := action.CreatePRParams()

	if dryRun {
		return models.ActionResult{
			Type:    "create_pr",
			Success: true,
			Detail:  fmt.Sprintf("dry run: would create PR %s -> %s with title %q", sourceRef, targetRef, title),
		}
	}

	if s.adoClient == nil {
		return models.ActionResult{
			Type:    "create_pr",
			Success: false,
			Detail:  "azure devops client not configured",
		}
	}

	repo := policy.Scope.Repository
	if repo == "" {
		return models.ActionResult{
			Type:    "create_pr",
			Success: false,
			Detail:  "no repository specified in policy scope",
		}
	}

	req := &azuredevops.CreatePRRequest{
		SourceRefName: sourceRef,
		TargetRefName: targetRef,
		Title:         title,
		Description:   description,
	}

	if isDraft {
		req.IsDraft = &isDraft
	}

	result, err := s.adoClient.CreatePullRequest(ctx, policy.Scope.Project, repo, req)
	if err != nil {
		return models.ActionResult{
			Type:    "create_pr",
			Success: false,
			Detail:  fmt.Sprintf("failed to create PR: %v", err),
		}
	}

	// Log the PR creation
	s.logEvent(ctx, runID, string(models.EventTypePRCreated), map[string]interface{}{
		"pr_id":         result.ID,
		"source_ref":    sourceRef,
		"target_ref":    targetRef,
		"ado_project":   policy.Scope.Project,
		"ado_repo":      repo,
	})

	prID := result.ID
	return models.ActionResult{
		Type:    "create_pr",
		Success: true,
		Detail:  fmt.Sprintf("created PR #%d: %s -> %s", result.ID, sourceRef, targetRef),
		PRID:    &prID,
	}
}

// SimulatePolicy runs a dry-run simulation and returns detailed results.
func (s *PullRequestService) SimulatePolicy(ctx context.Context, policy *models.Policy) ([]models.ConditionResult, []models.ActionResult, error) {
	condResults := s.evaluateConditions(ctx, policy)

	var actionResults []models.ActionResult
	for _, action := range policy.Actions {
		if action.Type == "create_pr" {
			sourceRef, targetRef, title, _, _ := action.CreatePRParams()
			actionResults = append(actionResults, models.ActionResult{
				Type:    "create_pr",
				Success: true,
				Detail:  fmt.Sprintf("would create PR %s -> %s with title %q", sourceRef, targetRef, title),
			})
		} else {
			actionResults = append(actionResults, models.ActionResult{
				Type:    action.Type,
				Success: false,
				Detail:  fmt.Sprintf("action type %q not yet implemented", action.Type),
			})
		}
	}

	return condResults, actionResults, nil
}

// marshalDetail converts a map to JSON string for audit storage.
func marshalDetail(m map[string]interface{}) string {
	raw, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(raw)
}
