// Policy types matching Go models in backend/internal/models/

export interface Policy {
  id: string;
  created: string;
  updated: string;
  name: string;
  description: string;
  version: string;
  enabled: boolean;
  scope_org: string;
  scope_project: string;
  scope_repo: string;
  schedule_cron: string;
  schedule_timezone: string;
  schedule_enabled: boolean;
  conditions: PolicyCondition[];
  actions: PolicyAction[];
  constraints: PolicyConstraints | null;
  tags: string[];
}

export interface PolicyCondition {
  type: string;
  parameters: Record<string, unknown>;
}

export interface PolicyAction {
  type: string;
  parameters: Record<string, unknown>;
}

export interface PolicyConstraints {
  max_concurrent_runs?: number;
  prevent_duplicate_prs?: boolean;
  require_approval?: boolean;
  [key: string]: unknown;
}

export type RunStatus =
  | "pending"
  | "running"
  | "succeeded"
  | "failed"
  | "skipped"
  | "dry_run";

export type TriggeredBy = "schedule" | "simulation" | "webhook" | "manual";

export interface Run {
  id: string;
  created: string;
  updated: string;
  policy: string;
  status: RunStatus;
  triggered_by: TriggeredBy;
  dry_run: boolean;
  cron_eval_time: string;
  started_at: string;
  completed_at: string;
  error: string;
  result_summary: string;
}

export type AuditEventType =
  | "policy_evaluated"
  | "pr_created"
  | "condition_met"
  | "condition_failed"
  | "action_executed"
  | "error";

export interface AuditEvent {
  id: string;
  created: string;
  run: string;
  event_type: AuditEventType;
  actor: string;
  detail: Record<string, unknown>;
  ado_org: string;
  ado_project: string;
  ado_repo: string;
  ado_pr_id: number;
}

export interface ConditionResult {
  type: string;
  met: boolean;
  detail: string;
}

export interface ActionResult {
  type: string;
  success: boolean;
  detail: string;
}

export interface SimulateResult {
  policy_id: string;
  conditions: ConditionResult[];
  actions: ActionResult[];
}

export interface ExecuteResult {
  run_id: string;
  policy_id: string;
  dry_run: boolean;
}

export interface PocketBaseListResult<T> {
  page: number;
  perPage: number;
  totalItems: number;
  totalPages: number;
  items: T[];
}
