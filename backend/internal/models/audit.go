package models

import "time"

// RunStatus represents the status of a policy run.
type RunStatus string

const (
	RunStatusPending   RunStatus = "pending"
	RunStatusRunning   RunStatus = "running"
	RunStatusSucceeded RunStatus = "succeeded"
	RunStatusFailed    RunStatus = "failed"
	RunStatusSkipped   RunStatus = "skipped"
	RunStatusDryRun    RunStatus = "dry_run"
)

// TriggeredBy indicates what triggered a policy run.
type TriggeredBy string

const (
	TriggeredBySchedule   TriggeredBy = "schedule"
	TriggeredBySimulation TriggeredBy = "simulation"
	TriggeredByWebhook    TriggeredBy = "webhook"
	TriggeredByManual     TriggeredBy = "manual"
)

// AuditEventType represents the type of an audit event.
type AuditEventType string

const (
	EventTypePolicyEvaluated AuditEventType = "policy_evaluated"
	EventTypePRCreated       AuditEventType = "pr_created"
	EventTypePRSkipped       AuditEventType = "pr_skipped"
	EventTypeConditionMet    AuditEventType = "condition_met"
	EventTypeConditionFailed AuditEventType = "condition_failed"
	EventTypeActionExecuted  AuditEventType = "action_executed"
	EventTypeError           AuditEventType = "error"
)

// PolicyRun represents a single policy execution.
type PolicyRun struct {
	ID            string       `json:"id"`
	PolicyID      string       `json:"policy"`
	Status        RunStatus    `json:"status"`
	TriggeredBy   TriggeredBy  `json:"triggered_by"`
	DryRun        bool         `json:"dry_run"`
	CronEvalTime  time.Time    `json:"cron_eval_time"`
	StartedAt     time.Time    `json:"started_at"`
	CompletedAt   *time.Time   `json:"completed_at,omitempty"`
	Error         string       `json:"error,omitempty"`
	ResultSummary string       `json:"result_summary,omitempty"`
}

// AuditEvent represents a fine-grained event within a policy run.
type AuditEvent struct {
	ID         string        `json:"id"`
	RunID      string        `json:"run"`
	EventType  AuditEventType `json:"event_type"`
	Actor      string        `json:"actor,omitempty"`
	Detail     string        `json:"detail"`
	AdoOrg     string        `json:"ado_org,omitempty"`
	AdoProject string        `json:"ado_project,omitempty"`
	AdoRepo    string        `json:"ado_repo,omitempty"`
	AdoPrID    *int          `json:"ado_pr_id,omitempty"`
	Created    time.Time     `json:"created"`
}

// ExecutionSummary is returned after evaluating a batch of policies.
type ExecutionSummary struct {
	TotalEvaluated int `json:"total_evaluated"`
	TotalMatched   int `json:"total_matched"`
	TotalSucceeded int `json:"total_succeeded"`
	TotalFailed    int `json:"total_failed"`
	TotalSkipped   int `json:"total_skipped"`
}

// ConditionResult describes the outcome of evaluating a single condition.
type ConditionResult struct {
	Type    string `json:"type"`
	Met     bool   `json:"met"`
	Reason  string `json:"reason,omitempty"`
}

// ActionResult describes the outcome of executing a single action.
type ActionResult struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Detail  string `json:"detail,omitempty"`
	PRID    *int   `json:"pr_id,omitempty"`
}
