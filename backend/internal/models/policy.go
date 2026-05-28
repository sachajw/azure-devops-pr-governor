package models

import "encoding/json"

// Policy represents a PR governance policy.
type Policy struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Version        string          `json:"version"`
	Enabled        bool            `json:"enabled"`
	Scope          PolicyScope     `json:"scope"`
	Schedule       PolicySchedule  `json:"schedule"`
	Conditions     []PolicyCondition `json:"conditions"`
	Actions        []PolicyAction  `json:"actions"`
	Constraints    *PolicyConstraints `json:"constraints,omitempty"`
	Tags           []string        `json:"tags,omitempty"`
	Created        string          `json:"created"`
	Updated        string          `json:"updated"`
}

// PolicyScope defines the Azure DevOps resources a policy targets.
type PolicyScope struct {
	Organization string `json:"organization"`
	Project      string `json:"project"`
	Repository   string `json:"repository,omitempty"`
}

// PolicySchedule defines when a policy should be evaluated.
type PolicySchedule struct {
	Cron     string `json:"cron"`
	Timezone string `json:"timezone"`
	Enabled  bool   `json:"enabled"`
}

// PolicyCondition is a discriminated union for preconditions.
type PolicyCondition struct {
	Type     string          `json:"type"`
	Operator string          `json:"operator"`
	Value    json.RawMessage `json:"value"`
}

// PolicyAction is a discriminated union for actions to perform.
type PolicyAction struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

// PolicyConstraints defines limits on PR creation.
type PolicyConstraints struct {
	MaxActivePRs       *int   `json:"max_active_prs,omitempty"`
	AutoComplete       bool   `json:"auto_complete,omitempty"`
	MergeStrategy      string `json:"merge_strategy,omitempty"`
	RequireMinReviewers *int  `json:"require_min_reviewers,omitempty"`
}

// CreatePRParams extracts parameters for a create_pr action.
func (a PolicyAction) CreatePRParams() (sourceRef, targetRef, title, description string, isDraft bool) {
	if v, ok := a.Parameters["sourceRefName"].(string); ok {
		sourceRef = v
	}
	if v, ok := a.Parameters["targetRefName"].(string); ok {
		targetRef = v
	}
	if v, ok := a.Parameters["title"].(string); ok {
		title = v
	}
	if v, ok := a.Parameters["description"].(string); ok {
		description = v
	}
	if v, ok := a.Parameters["isDraft"].(bool); ok {
		isDraft = v
	}
	return
}

// PolicyFromRecordMap converts a PocketBase record map to a Policy.
func PolicyFromRecordMap(m map[string]interface{}) (*Policy, error) {
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var p Policy
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
