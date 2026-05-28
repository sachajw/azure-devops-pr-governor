# Security

## Authentication

### Azure DevOps PAT

The backend authenticates to Azure DevOps using a Personal Access Token (PAT):

- Stored as an environment variable (`AZURE_DEVOPS_PAT`)
- Passed as a Basic auth header (`Authorization: Basic base64(:PAT)`)
- PAT must have **Code (read & write)** and **Pull Requests (read & write)** scopes

For production deployment:
- Store the PAT in Azure Key Vault
- Use managed identity to retrieve it at startup
- Rotate on a regular schedule (documented in runbook)

### PocketBase Admin

PocketBase provides a built-in superuser admin account:

- Created on first startup via the admin UI (`/_/`)
- All custom routes under `/api/pr-governor/*` require superuser auth
- PocketBase CRUD endpoints use collection-level access rules

## Authorization

- Only superusers can create, update, or delete policies
- Custom API routes bind `apis.RequireSuperuserAuth()`
- The schedule runner runs in the server context (no user auth needed)

## Input Validation

- All policy payloads are validated against the JSON Schema (`schemas/policy.schema.json`) before persistence
- Record hooks (`OnRecordCreate`, `OnRecordUpdate`) enforce validation at the database layer
- Invalid payloads are rejected with structured error responses

## Audit Trail

Every policy execution produces an immutable audit trail:

- **Policy changes** — all create/update/delete operations are logged
- **Run outcomes** — each run records status, timing, and error details
- **Fine-grained events** — individual condition evaluations, PR creation attempts, and constraint checks
- Events are append-only and cannot be modified after creation

## Network

- The backend makes outbound HTTPS requests to `dev.azure.com` / `visualstudio.com`
- No inbound connections are needed from Azure DevOps (webhooks are opt-in, Phase 2)
- For production, deploy behind a reverse proxy with TLS termination

## Data Protection

- SQLite database file (`pb_data/data.db`) contains all policy and audit data
- The PAT is never stored in the database — only in environment variables
- Back up `pb_data/` regularly
