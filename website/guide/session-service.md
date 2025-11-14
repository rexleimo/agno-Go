# Go Session Service

The Go session service mirrors the Python AgentOS `/sessions` REST API while
running as a standalone Go binary. Ship it alongside AgentOS to gain native,
performant session CRUD backed by PostgreSQL, production-ready HTTP middleware,
and deployment assets for local stacks or Kubernetes clusters.

## Feature Highlights

- **Parity endpoints** – `/sessions` list/create, `/sessions/{id}` detail,
  rename and delete, `/sessions/{id}/runs` history, plus `/healthz` liveness
  checks.
- **Postgres-backed storage** – Typed DTO layer with transaction-safe
  operations that match the existing AgentOS JSON contracts.
- **Multi-database routing** – Support multiple DSNs via
  `AGNO_SESSION_DSN_MAP` and select stores per request with the `db_id` query.
- **Operational middleware** – Chi router with request IDs, structured logging,
  real IP handling, panic recovery, and a 60s timeout guard.
- **Deployment assets** – Dedicated Dockerfile, Docker Compose stack with
  Postgres, Helm chart, and curl-based smoke test script.

## Quick Start (Local Binary)

```bash
export AGNO_PG_DSN="postgres://user:pass@localhost:5432/agentos?sslmode=disable"
export AGNO_SERVICE_PORT=8080
go run ./cmd/agentos-session
```

The service prints `Go session service listening on :8080`. Visit
`http://localhost:8080/healthz` for a JSON health response.

### Contract Test Parity

- `make contract-test` – Runs the end-to-end contract suite comparing Go and
  Python responses.
- `./scripts/test-session-api.sh http://localhost:8080` – Executes a basic CRUD
  walkthrough (list, create, rename, delete) using curl and jq.

## Configuration Reference

| Environment Variable   | Description                                                                                  | Default    |
|------------------------|----------------------------------------------------------------------------------------------|------------|
| `AGNO_PG_DSN`          | Primary Postgres DSN. Required when `AGNO_SESSION_DSN_MAP` is not provided.                 | –          |
| `DATABASE_URL`         | Fallback DSN key (Heroku-style). Used if `AGNO_PG_DSN` is unset.                             | –          |
| `AGNO_SESSION_DSN_MAP` | JSON map of `{"dbID":"dsn"}` for multi-database routing. Enables `db_id` query selection. | –          |
| `AGNO_DEFAULT_DB_ID`   | Optional default store identifier when using `AGNO_SESSION_DSN_MAP`.                        | first key  |
| `AGNO_SERVICE_PORT`    | HTTP port to bind.                                                                           | `8080`     |

When `AGNO_SESSION_DSN_MAP` is supplied, issue requests like
`/sessions?type=agent&db_id=analytics` to target a specific store.

## API Surface

| Endpoint                 | Method | Description                                                                                          |
|--------------------------|--------|------------------------------------------------------------------------------------------------------|
| `/healthz`               | GET    | Health probe returning `{"status":"ok"}`.                                                         |
| `/sessions`              | GET    | Paginated list with filters (`type`, `component_id`, `user_id`, `session_name`, `sort_by`, `db_id`). |
| `/sessions`              | POST   | Create session with state, metadata, and optional pre-seeded runs or summary payloads.              |
| `/sessions/{id}`         | GET    | Retrieve session detail by type/ID with optional `db_id`.                                           |
| `/sessions/{id}`         | DELETE | Remove session and history for the selected type.                                                   |
| `/sessions/{id}/rename`  | POST   | Update `session_name`.                                                                              |
| `/sessions/{id}/runs`    | GET    | Fetch stored run history (mirrors Python AgentOS response shape).                                   |

All mutating routes expect JSON payloads that align with the Python fixtures
and are validated during contract testing.

## Docker Compose Stack

Use the bundled `docker-compose.session.yml` to spin up Postgres and the Go
service together:

```bash
docker compose -f docker-compose.session.yml up --build
```

Once healthy, access the API at `http://localhost:8080` and run the test script.

## Helm Deployment

The Helm chart under `deploy/helm/agno-session/` deploys the runtime into a
cluster. Override the DSN and image with your registry:

```bash
helm upgrade --install agno-session ./deploy/helm/agno-session \
  --set image.repository=ghcr.io/<org>/agno-session \
  --set image.tag=v1.2.9 \
  --set config.dsn="postgres://user:pass@postgres:5432/agentos?sslmode=disable"
```

Refer to `values.yaml` for probe tuning, replica counts, and service exposure.

## Production Checklist

- Run `make contract-test` against staging to confirm parity.
- Mirror a subset of production traffic and compare JSON responses with the
  Python service.
- Monitor Postgres connections and latency; enable connection pooling if
  required.
- Configure dashboards and alerts off the included health endpoint and HTTP
  logs.
- Keep the Python runtime available until parity is proven in production.
