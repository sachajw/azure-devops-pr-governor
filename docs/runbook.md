# Runbook

## Deployment

### Local Development

```bash
cd backend
go run . serve
```

PocketBase creates `pb_data/` automatically on first run with the SQLite database and default superuser setup.

### Production

```bash
# Build
go build -o governor .

# Run with environment variables
export AZURE_DEVOPS_ORG_URL="https://dev.azure.com/your-org"
export AZURE_DEVOPS_PAT="your-pat"
./governor serve --http=0.0.0.0:8090
```

For Azure deployment, containerize with Docker:

```dockerfile
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY backend/ .
RUN go build -o governor .

FROM alpine:3.20
COPY --from=build /app/governor /governor
EXPOSE 8090
CMD ["/governor", "serve", "--http=0.0.0.0:8090"]
```

## Monitoring

### Health Check

```bash
curl http://localhost:8090/api/health
```

### Checking Recent Runs

```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8090/api/collections/runs/records?sort=-started_at&page=1&perPage=10"
```

### Checking Failed Runs

```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8090/api/collections/runs/records?filter=(status='failed')"
```

## PAT Rotation

1. Generate a new PAT in Azure DevOps (User Settings → Personal Access Tokens)
2. Grant **Code (read & write)** and **Pull Requests (read & write)**
3. Update the environment variable: `export AZURE_DEVOPS_PAT="new-pat"`
4. Restart the backend: `kill -HUP <pid>` or restart the container
5. Verify by running a simulation: `POST /api/pr-governor/simulate`

## Troubleshooting

### Schedule runner not firing

- Check the cron is registered: logs should show `Schedule runner triggered` every 5 minutes
- Check policies are enabled: `GET /api/collections/policies/records?filter=(enabled=true)`
- Check the policy's `schedule_cron` expression is valid
- Check `schedule_enabled` is `true` on the policy

### PR creation failing

- Check the PAT is valid and has the required scopes
- Check the Azure DevOps organization URL is correct
- Check the repository ID and branch names are correct
- Review audit events for the run: `GET /api/collections/audit_events/records?filter=(run='<run-id>')`

### Database issues

- PocketBase stores data in `pb_data/data.db`
- For corruption: stop the server, backup `pb_data/`, then restart
- For migrations: `./governor migrate` runs pending migrations

## Backup

The entire state is in `pb_data/`:

```bash
# Stop the server first
cp -r pb_data/ pb_data_backup_$(date +%Y%m%d)/
```
