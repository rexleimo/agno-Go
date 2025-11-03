# Release Notes - Agno-Go v1.2.7

Date: 2025-11-03

## Highlights
- Go-native session service that mirrors the Python AgentOS `/sessions` REST API with Postgres-backed CRUD, Chi routing, health checks, and typed DTOs.
- Deployment assets for every stage: dedicated Dockerfile, Docker Compose stack with Postgres, and a Helm chart for cluster rollouts.
- Updated documentation and scripts, including `README.session-service.md`, deployment guide, and `test-session-api.sh` helper for endpoint validation.

## Fixes & Improvements
- Postgres store implementation with transaction-safe operations and contract enforcement against existing Python fixtures.
- Comprehensive contract tests verifying parity for list/create/read/rename/delete workflows plus DSN-driven configuration helpers.

## Compatibility
- Backward compatible; the Go session runtime is optional and coexists with the existing Python deployment.
