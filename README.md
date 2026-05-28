# Azure DevOps PR Governor

Policy-driven governance platform for Azure DevOps pull requests. Define automation policies that create PRs on a schedule or in response to events, with conditions, constraints, and full audit logging.

## Architecture

```
┌─────────────────────┐     ┌──────────────────────┐     ┌─────────────┐
│  Azure DevOps       │     │  PocketBase Backend  │     │  SQLite DB  │
│  Extension (Phase3) │────▶│  (Go, single binary) │────▶│  (embedded) │
└─────────────────────┘     └──────────┬───────────┘     └─────────────┘
                                       │
                                       ▼
                            ┌─────────────────────┐
                            │  Azure DevOps API   │
                            │  (PR create, refs)  │
                            └─────────────────────┘
```

- **PocketBase** — backend framework providing REST API, admin UI, cron scheduling, record hooks, and embedded SQLite
- **Azure DevOps API** — external API for creating pull requests and reading repository state
- **Extension** — Azure DevOps private extension for policy management UI (Phase 3)

## Prerequisites

- Go 1.23+
- Azure DevOps organization with a Personal Access Token (PAT) with Git/PR permissions

## Quick Start

```bash
# Clone and build
cd backend
go build -o governor .

# Configure (first run creates pb_data/ automatically)
export AZURE_DEVOPS_ORG_URL="https://dev.azure.com/your-org"
export AZURE_DEVOPS_PAT="your-pat"

# Start
./governor serve

# Admin UI: http://localhost:8090/_/
# API:      http://localhost:8090/api/
```

## Repository Layout

```
backend/
├── main.go                    # App entry point, wires hooks/routes/cron
├── migrations/                # PocketBase collection definitions
├── internal/
│   ├── hooks/                 # Record lifecycle hooks (validation)
│   ├── routes/                # Custom API routes (simulate, execute, webhook)
│   ├── services/              # Business logic (policy, schedule, PR, audit)
│   ├── azuredevops/           # Azure DevOps REST API client
│   └── models/                # Go types for policies, runs, audit events
├── tests/                     # Unit and integration tests
schemas/                       # JSON Schema for policy validation
docs/                          # Architecture, security, runbook
extension/                     # Phase 3: Azure DevOps extension UI
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/pr-governor/simulate` | Dry-run a policy without creating PRs |
| POST | `/api/pr-governor/execute` | Manually trigger a policy execution |
| POST | `/api/pr-governor/webhook` | Receive Azure DevOps Service Hook events |

PocketBase also provides standard CRUD endpoints under `/api/collections/{collection}/records`.

## Development

```bash
# Run with hot reload (requires air)
go run . serve

# Run tests
go test ./...

# Run with verbose logging
go run . serve --debug
```

## License

Internal use only.
