Here is a copy-ready **comprehensive plan** for building `azure-devops-pr-governor` as a private Azure DevOps extension plus backend policy engine. This plan is based on Azure DevOps extension manifests and contribution points, Azure Functions timer triggers for scheduled execution, and Azure DevOps Service Hooks for event-driven integration. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/develop/manifest?view=azure-devops)

## Product goal

Build an internal tool called **Azure DevOps PR Governor** that gives teams a native Azure DevOps UI for defining pull request automation policies, while a backend service evaluates those policies and creates PRs on a schedule or in response to events. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/service-hooks/overview?view=azure-devops)

The tool should support more than simple PR scheduling. It should be designed as a policy-driven governance platform that can expand into approvals, blackout windows, compliance checks, reviewer enforcement, change windows, and auditability without needing a major rewrite. [learn.microsoft](https://learn.microsoft.com/en-us/azure/azure-functions/functions-bindings-timer)

## Scope

### In scope for v1
- Private Azure DevOps extension with a configuration hub. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/reference/targets/overview?view=azure-devops)
- Backend service for policy evaluation and PR creation through the Azure DevOps Git Pull Requests API. [learn.microsoft](https://learn.microsoft.com/en-us/rest/api/azure/devops/git/pull-requests/create?view=azure-devops-rest-7.1)
- Timer-triggered schedule runner in Azure Functions. [learn.microsoft](https://learn.microsoft.com/en-us/azure/azure-functions/functions-create-scheduled-function)
- JSON-schema-based policy definition. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/develop/manifest?view=azure-devops)
- Basic audit logging.
- Dry-run simulation mode.
- Example policy for scheduled PR creation.

### In scope for v2
- Service Hook ingestion for PR/repo events. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/service-hooks/events?view=azure-devops)
- Policy inheritance by organization, project, and repository.
- Reviewer policies and approval routing.
- Blackout windows and maintenance windows.
- Integration with Teams or other notification targets.

### Out of scope for v1
- Full enterprise reporting dashboard.
- Multi-org tenancy.
- Rich RBAC model beyond admin/operator basics.
- Public Marketplace publication, because private sharing is sufficient for internal rollout and Azure DevOps supports private sharing of extensions to specific organizations. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/publish/overview?view=azure-devops)

## Architecture

### High-level design
- **Azure DevOps extension**: the user-facing layer for policy configuration, simulation, and audit views. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/reference/targets/overview?view=azure-devops)
- **Backend API/service**: the system of record for policy evaluation, PR orchestration, and audit decisions.
- **Azure Functions**: timer-triggered execution for scheduled policies. [learn.microsoft](https://learn.microsoft.com/en-us/azure/azure-functions/functions-create-scheduled-function)
- **Azure DevOps Service Hooks**: event-driven triggers for future enhancements. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/service-hooks/overview?view=azure-devops)
- **Storage**: policy definitions, run history, audit events, and exceptions.
- **Secrets/config**: Azure-hosted secure configuration for API access and runtime settings.

### Why this architecture
Azure DevOps extensions are best for contributing UI and experience into Azure DevOps, but they are not the right place to run scheduled backend jobs.  Azure Functions timer triggers are designed for scheduled execution, and Service Hooks are designed to react to Azure DevOps events, so the clean architecture is extension UI + backend control plane + scheduled/event-driven runtime. [learn.microsoft](https://learn.microsoft.com/en-us/azure/azure-functions/functions-bindings-timer)

## Repository structure

Use this structure:

```text
azure-devops-pr-governor/
├── README.md
├── docs/
│   ├── architecture.md
│   ├── policy-schema.md
│   ├── security.md
│   ├── roadmap.md
│   └── runbook.md
├── extension/
│   ├── vss-extension.json
│   ├── package.json
│   ├── tsconfig.json
│   ├── src/
│   │   ├── admin/
│   │   ├── simulation/
│   │   ├── audit/
│   │   ├── shared/
│   │   └── main.ts
│   ├── public/
│   │   └── icon.png
│   └── dist/
├── backend/
│   ├── host.json
│   ├── package.json
│   ├── tsconfig.json
│   ├── local.settings.json.example
│   ├── src/
│   │   ├── functions/
│   │   │   ├── scheduleRunner.ts
│   │   │   ├── simulatePolicy.ts
│   │   │   └── webhookReceiver.ts
│   │   ├── services/
│   │   │   ├── policyService.ts
│   │   │   ├── pullRequestService.ts
│   │   │   ├── auditService.ts
│   │   │   └── scheduleService.ts
│   │   ├── models/
│   │   │   └── policy.ts
│   │   ├── repositories/
│   │   └── utils/
│   └── tests/
├── schemas/
│   └── policy.schema.json
└── .github/
    └── workflows/
        ├── ci.yml
        └── release-extension.yml
```

This structure keeps frontend, backend, schemas, and operational documentation separated, which supports maintainability as the product grows. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/develop/manifest?view=azure-devops)

## Extension plan

### Objective
Create a private Azure DevOps extension that contributes a hub into Azure Repos or another suitable Azure DevOps area for managing policies. Azure DevOps extensions are defined by `vss-extension.json`, and the manifest controls identity, files, contribution points, and discovery metadata. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/reference/targets/overview?view=azure-devops)

### Extension deliverables
- `vss-extension.json` in the extension root. [devkimchi](https://devkimchi.com/2019/07/17/building-azure-devops-extension-on-azure-devops-4/)
- A contributed hub for **PR Governor**. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/reference/targets/overview?view=azure-devops)
- An admin page for creating and editing policies.
- A simulation page for previewing what a policy would do.
- An audit page for viewing past runs and decisions.
- Shared UI components for forms, validation, and tables.

### Manifest requirements
- `publisher` must match your Marketplace publisher ID when packaging and uploading the extension. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/publish/overview?view=azure-devops)
- Keep the extension private initially and share it only with your Azure DevOps organization. Azure DevOps allows private sharing before installation. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/publish/overview?view=azure-devops)
- Add `baseUri` later for faster local debugging, because Microsoft documents it as a way to speed development without redeploying for every source change. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/publish/overview?view=azure-devops)

### UI requirements
- Policy list page.
- Policy create/edit form.
- Policy simulation button.
- Audit run table.
- Validation error display.
- Effective policy viewer.

### UI design rules
- Keep the UI small and administrative, not flashy.
- Make forms explicit and schema-driven.
- Prefer server-side validation results over only client-side assumptions.
- Treat the extension as a control plane, not the engine.

## Backend plan

### Objective
Build a backend service that owns all privileged actions and runtime execution. This service creates PRs through the Azure DevOps Git API and evaluates policy rules outside the extension. [learn.microsoft](https://learn.microsoft.com/en-us/rest/api/azure/devops/git/pull-requests?view=azure-devops-rest-7.1)

### Backend responsibilities
- Policy CRUD.
- Policy validation.
- Policy simulation.
- Schedule expansion and execution.
- Pull request creation.
- Audit logging.
- Event ingestion for Service Hooks. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/service-hooks/overview?view=azure-devops)

### Core services
- `policyService`: load, validate, merge, and resolve effective policy.
- `scheduleService`: interpret cron and eligible execution windows.
- `pullRequestService`: call Azure DevOps PR Create API. [learn.microsoft](https://learn.microsoft.com/en-us/rest/api/azure/devops/git/pull-requests/create?view=azure-devops-rest-7.1)
- `auditService`: record runs, decisions, failures, and exceptions.
- `webhookService`: process Azure DevOps events later through Service Hooks. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/service-hooks/overview?view=azure-devops)

### Suggested API surface
- `GET /policies`
- `POST /policies`
- `PUT /policies/{id}`
- `POST /policies/{id}/validate`
- `POST /policies/{id}/simulate`
- `POST /runs/execute`
- `POST /webhooks/azure-devops`
- `GET /audit`

## Scheduling plan

Use Azure Functions with timer triggers for scheduled execution because timer triggers are explicitly designed to run functions on schedules.  The initial schedule runner should scan active policies, determine which are due, simulate the outcome, and create PRs when conditions pass. [learn.microsoft](https://learn.microsoft.com/en-us/azure/azure-functions/functions-create-scheduled-function)

### Schedule runner algorithm
1. Load active policies.
2. Filter to due schedules for the current time.
3. Resolve effective policy by scope.
4. Check freeze windows or blackout rules.
5. Validate branch/source-target rules.
6. Optionally verify preconditions such as required work items or build status.
7. Build PR title and description from templates.
8. Call Azure DevOps PR Create API. [learn.microsoft](https://learn.microsoft.com/en-us/rest/api/azure/devops/git/pull-requests/create?view=azure-devops-rest-7.1)
9. Record audit result.
10. Emit notification or event.

### Local testing
Timer-triggered functions can be tested locally, and there are documented ways to execute them during local development.  Keep an HTTP-triggered simulation endpoint for easier iterative testing in early development. [stackoverflow](https://stackoverflow.com/questions/46556621/what-is-the-simplest-way-to-run-a-timer-triggered-azure-function-locally-once)

## Event-driven plan

Use Azure DevOps Service Hooks in v2 so the system can react to Azure DevOps events such as PR activity, repository changes, or other supported triggers. Azure DevOps documents Service Hooks as the way to run tasks on other services when events happen in a project. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/service-hooks/overview?view=azure-devops)

### Service Hook setup
- Go to project settings.
- Open Service Hooks.
- Create a subscription.
- Select the event and target.
- Test the subscription before finalizing. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/service-hooks/overview?view=azure-devops)

### Event use cases
- Reevaluate policy after PR creation.
- Notify external systems after PR completion.
- Trigger audit or exception handling.
- Sync execution history.

## Policy model

Make the policy model versioned and schema-driven from day one so corporate requirements become new configuration fields instead of code forks. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/develop/manifest?view=azure-devops)

### Required policy concepts
- Policy ID and version.
- Scope: organization, project, repository, branch selectors.
- Schedule: cron expression, time zone, active state. Timer schedules are a natural fit because Azure Functions timer triggers use cron-style schedules. [learn.microsoft](https://learn.microsoft.com/en-us/azure/azure-functions/functions-bindings-timer)
- Conditions: preconditions that must be met before a PR is created.
- Actions: create PR, set title/description, reviewers, labels, draft mode.
- Constraints: blackout windows, restricted branches, naming rules.
- Notifications.
- Audit metadata.

### Example v1 policy
```json
{
  "policyId": "nightly-dev-to-qa",
  "version": 1,
  "scope": {
    "organization": "your-org",
    "project": "your-project",
    "repository": "your-repo"
  },
  "schedule": {
    "type": "cron",
    "expression": "0 0 1 * * *",
    "timeZone": "Africa/Johannesburg",
    "enabled": true
  },
  "conditions": {
    "allowExistingOpenPr": false
  },
  "actions": {
    "createPullRequest": true,
    "sourceRefName": "refs/heads/dev",
    "targetRefName": "refs/heads/qa",
    "titleTemplate": "Auto PR: dev -> qa"
  }
}
```

## Security plan

### Authentication and permissions
- Keep Git/PR operations in the backend, not the extension browser app. This reduces exposure and simplifies future security review. [learn.microsoft](https://learn.microsoft.com/en-us/rest/api/azure/devops/git/pull-requests/create?view=azure-devops-rest-7.1)
- Use secure application configuration and secret storage.
- Limit API permissions to the minimum needed to read repo state and create PRs.

### Auditability
- Log every policy change.
- Log every run, decision, and PR creation attempt.
- Distinguish success, skip, failure, and blocked outcomes.
- Preserve correlation IDs for traceability.

### Safe rollout
- Start private only. Azure DevOps supports private extension sharing to specific organizations. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/publish/overview?view=azure-devops)
- Use a dev extension ID and a production extension ID later if needed.
- Validate extension updates before installing broadly. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/publish/overview?view=azure-devops)

## Packaging and publishing plan

### Packaging
Azure DevOps extensions are packaged as VSIX files, and the extension manifest must be in the root of the extension package structure.  Microsoft also documents the private upload and sharing flow through Marketplace. [devkimchi](https://devkimchi.com/2019/07/17/building-azure-devops-extension-on-azure-devops-4/)

### Initial publishing plan
1. Create or use a Marketplace publisher. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/publish/overview?view=azure-devops)
2. Set `publisher` in `vss-extension.json` to match that publisher. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/publish/overview?view=azure-devops)
3. Package the extension as VSIX.
4. Upload it as a private extension. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/publish/overview?view=azure-devops)
5. Share it with your Azure DevOps organization. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/publish/overview?view=azure-devops)
6. Install it into the organization. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/publish/overview?view=azure-devops)

### CI/CD later
- Build extension bundle.
- Validate schema.
- Run backend tests.
- Produce VSIX.
- Optionally automate extension publishing.

## Delivery phases

### Phase 1: Repository foundation
- Create repo structure.
- Add README, docs, schema, and stubs.
- Define policy schema.
- Add CI placeholders.

### Phase 2: Working backend MVP
- Implement policy loading and validation.
- Implement timer-triggered schedule runner. [learn.microsoft](https://learn.microsoft.com/en-us/azure/azure-functions/functions-bindings-timer)
- Implement Azure DevOps PR create integration. [learn.microsoft](https://learn.microsoft.com/en-us/rest/api/azure/devops/git/pull-requests/create?view=azure-devops-rest-7.1)
- Add basic audit persistence.

### Phase 3: Extension MVP
- Create extension manifest and hub contribution. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/develop/manifest?view=azure-devops)
- Build policy list and create/edit forms.
- Add simulate action and validation display.

### Phase 4: Pilot rollout
- Package private extension. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/publish/overview?view=azure-devops)
- Share with organization and install it. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/publish/overview?view=azure-devops)
- Test against one project/repository.
- Verify audit and failure handling.

### Phase 5: Enterprise hardening
- Add Service Hooks. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/service-hooks/overview?view=azure-devops)
- Add approval workflows.
- Add notification adapters.
- Add policy inheritance and exception handling.
- Add release automation for extension/package updates.

## Definition of done

A first usable version is complete when all of the following are true:

- A private extension is installed in the target Azure DevOps organization. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/publish/overview?view=azure-devops)
- A user can create a policy in the extension UI. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/develop/manifest?view=azure-devops)
- The backend can validate and store the policy.
- A timer-triggered function can execute the policy on schedule. [learn.microsoft](https://learn.microsoft.com/en-us/azure/azure-functions/functions-bindings-timer)
- The backend can create a pull request through the Azure DevOps API. [learn.microsoft](https://learn.microsoft.com/en-us/rest/api/azure/devops/git/pull-requests/create?view=azure-devops-rest-7.1)
- Every run is logged with an audit outcome.
- A simulation view explains what would happen before enabling a policy.
- The repo includes documentation for setup, packaging, and operations.

## Immediate next actions

1. Push the scaffold into `azure-devops-pr-governor`.
2. Strengthen the repo into a real TypeScript monorepo.
3. Implement the policy schema and validation first.
4. Implement the Azure Functions timer runner second. [learn.microsoft](https://learn.microsoft.com/en-us/azure/azure-functions/functions-bindings-timer)
5. Implement the PR creation service third. [learn.microsoft](https://learn.microsoft.com/en-us/rest/api/azure/devops/git/pull-requests/create?view=azure-devops-rest-7.1)
6. Build the extension UI once the backend contract is stable. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/develop/manifest?view=azure-devops)
7. Package and privately install the extension for pilot testing. [learn.microsoft](https://learn.microsoft.com/en-us/azure/devops/extend/publish/overview?view=azure-devops)

## Copy-ready summary

You can paste this as your working implementation direction:

```md
Build Azure DevOps PR Governor as a private Azure DevOps extension plus backend policy engine.

Architecture:
- Azure DevOps extension for configuration, simulation, and audit UI.
- Backend service for policy evaluation, PR creation, and audit logging.
- Azure Functions timer triggers for schedule execution.
- Azure DevOps Service Hooks for future event-driven behavior.
- JSON-schema-based policy model with versioning.

Implementation order:
1. Repo scaffold and docs.
2. Policy schema and validation.
3. Backend services and Azure Functions timer runner.
4. Azure DevOps PR create integration.
5. Extension manifest and admin hub UI.
6. Private packaging, sharing, and pilot installation.
7. Enterprise hardening with hooks, approvals, notifications, and inheritance.
```

If you want, I can turn this next into a **copy-ready markdown project brief** or a **task backlog with epics and stories**.
