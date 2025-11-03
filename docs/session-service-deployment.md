# Go Session Service Deployment Guide

This document describes how to run and deploy the Go-based session runtime that
mirrors the Python AgentOS `/sessions` API.

## Prerequisites

- Go 1.23+ (required for building from source)
- Docker 24+ (for container images and local Postgres)
- Kubernetes 1.27+ (for Helm chart deployment)
- Access to a PostgreSQL database populated with the `agno_sessions` table

## Local Development

1. Export the Postgres DSN and optionally override the listening port:

   ```bash
   export AGNO_PG_DSN="postgres://user:pass@localhost:5432/agentos?sslmode=disable"
   export AGNO_SERVICE_PORT=8080
   ```

2. Run the contract tests to ensure behaviour matches the Python fixtures:

   ```bash
   make contract-test
   ```

3. Start the service:

   ```bash
   go run ./cmd/agentos-session
   ```

4. Use the bundled script to hit the primary endpoints:

   ```bash
   ./scripts/test-session-api.sh http://localhost:8080
   ```

## Docker Compose

The repository includes `docker-compose.session.yml` for a fully local stack. It
provisions Postgres and the Go session runtime with a shared network.

```bash
docker compose -f docker-compose.session.yml up --build
```

Once the containers are healthy the service listens on `http://localhost:8080`.

## Kubernetes

The Helm chart under `deploy/helm/agno-session/` deploys the runtime into a
cluster. Provide the DSN and secrets via values overrides:

```bash
helm upgrade --install agno-session ./deploy/helm/agno-session \
  --set image.repository=ghcr.io/<org>/agno-session \
  --set image.tag=v1.0.0 \
  --set config.dsn="postgres://user:pass@postgres:5432/agentos?sslmode=disable"
```

Refer to `values.yaml` for configurable parameters such as replica counts,
resource limits, probes, and service exposure.

## Migration Checklist

- Run `make contract-test` against the Go service.
- Mirror a subset of production traffic and compare responses with Python.
- Verify dashboards and alerts using the included Prometheus metrics endpoints.
- Coordinate the DNS or load balancer cutover once parity is confirmed.
- Keep the Python stack on standby until the Go deployment is stable.
