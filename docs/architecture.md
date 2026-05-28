# Architecture

## System Overview

```mermaid
graph LR
    A[Azure DevOps Extension] -->|REST API| B[PocketBase Backend]
    B -->|Embedded| C[SQLite Database]
    B -->|REST API| D[Azure DevOps]
    E[Cron Scheduler] -->|Every 5 min| B
    F[Service Hooks] -->|Webhook| B
```

## Components

### PocketBase Backend

The backend is a Go application using PocketBase as a framework. It provides:

- **REST API** — PocketBase auto-generates CRUD endpoints for all collections
- **Admin UI** — Built-in dashboard at `/_/` for managing collections, records, and settings
- **Cron Scheduler** — Built-in `app.Cron()` runs the schedule evaluator every 5 minutes
- **Record Hooks** — `OnRecordCreate`/`OnRecordUpdate` for policy validation
- **Custom Routes** — `/api/pr-governor/*` for simulation, execution, and webhook handling

### Data Flow

```mermaid
sequenceDiagram
    participant Cron as Cron (5 min)
    participant SS as ScheduleService
    participant PS as PolicyService
    participant PRS as PullRequestService
    participant ADO as Azure DevOps API
    participant DB as SQLite

    Cron->>SS: evaluateAndExecute()
    SS->>PS: getEnabledPolicies()
    PS->>DB: SELECT policies WHERE enabled=true
    SS->>SS: isCronDue(policy.schedule)
    SS->>DB: INSERT run (status=running)
    SS->>PRS: executePolicyActions(policy)
    PRS->>ADO: GET refs (branch check)
    PRS->>ADO: POST pullrequests
    PRS->>DB: INSERT audit_events
    SS->>DB: UPDATE run (status=succeeded)
```

### Collections

| Collection | Purpose |
|---|---|
| `policies` | Policy definitions with scope, schedule, conditions, actions |
| `runs` | Execution history for each policy invocation |
| `audit_events` | Fine-grained event log within each run |

### Services

| Service | Responsibility |
|---|---|
| `policyService` | Load, validate, and resolve effective policies |
| `scheduleService` | Orchestrate cron evaluation and policy execution |
| `pullRequestService` | Evaluate conditions, check constraints, create PRs |
| `auditService` | Record structured events for every action |

### Azure DevOps Integration

The backend communicates with Azure DevOps via the REST API (v7.1):

- **Create Pull Request** — `POST /_apis/git/repositories/{id}/pullrequests`
- **Get Refs** — `GET /_apis/git/repositories/{id}/refs` (for branch_exists conditions)

Authentication uses a PAT passed as a Basic auth header (`:` + PAT base64-encoded).
