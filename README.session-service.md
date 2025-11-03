# Go Session Service

The Go session service re-implements the Python AgentOS `/sessions` REST API and
shares the same Postgres schema. It exposes endpoints for listing, creating,
renaming, deleting, and inspecting session runs.

## Features

- Postgres-backed session CRUD operations with multi-database (`db_id`) support
- HTTP API served via Chi with health checks and structured errors
- Contract test suite that compares responses against Python fixtures
- Dockerfile, Compose stack, and Helm chart for repeatable deployments

## Quick Start

```bash
export AGNO_PG_DSN="postgres://user:pass@localhost:5432/agentos?sslmode=disable"
go run ./cmd/agentos-session
```

Run `./scripts/test-session-api.sh http://localhost:8080` to verify the primary
endpoints or `make contract-test` for parity checks.

Deployment guidance, including Docker Compose and Helm instructions, is located
in `docs/session-service-deployment.md`.
