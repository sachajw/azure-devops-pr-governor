# Roadmap

## v1 — Backend MVP (Current)

- [x] Repository scaffold
- [x] Policy JSON schema and Go models
- [x] PocketBase collections (policies, runs, audit_events)
- [x] Azure DevOps REST API client
- [x] Service layer (policy, schedule, PR, audit)
- [x] Record hooks for validation
- [x] Custom routes (simulate, execute, webhook stub)
- [x] Cron-based schedule runner
- [x] Tests
- [ ] CI pipeline

## v2 — Event-Driven

- [ ] Service Hook ingestion for PR/repo events
- [ ] Policy inheritance (org → project → repo)
- [ ] Reviewer policies and approval routing
- [ ] Blackout windows and maintenance windows
- [ ] Teams/Slack notification adapters

## v3 — Extension UI

- [ ] Azure DevOps extension manifest and hub contribution
- [ ] Policy list, create/edit forms
- [ ] Simulation view
- [ ] Audit run table
- [ ] Effective policy viewer
- [ ] Private extension packaging and marketplace sharing

## v4 — Enterprise Hardening

- [ ] Full reporting dashboard
- [ ] Rich RBAC model
- [ ] Policy templates library
- [ ] Multi-org support
- [ ] Compliance reporting exports
