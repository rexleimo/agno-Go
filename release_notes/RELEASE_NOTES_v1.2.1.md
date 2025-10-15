# Release Notes - Agno-Go v1.2.1

Release Date: 2025-10-15
Type: Documentation + Minor Features

## Highlights

- Documentation reorganization: `website/` hosts implemented docs; `docs/` hosts designs/WIP/migration/internal
- Added `docs/README.md` (policy), `CONTRIBUTING.md` (contributor onboarding)
- Website Release Notes and API pages updated

## Implemented in 1.2.1

- SSE event filtering (A2A streaming)
  - `POST /api/v1/agents/:id/run/stream?types=token,complete`
  - Standard SSE output; filter by `types`; supports context cancel

- Content extraction middleware (AgentOS)
  - JSON/Form → context injection of `content/metadata/user_id/session_id`
  - Request size guard via `MaxRequestSize`; skip paths supported

- Google Sheets toolkit
  - Functions: `read_range`, `write_range`, `append_rows`
  - Credentials via JSON string or file path

- Minimal knowledge ingestion endpoint
  - `POST /api/v1/knowledge/content` with `text/plain` or `application/json`

## Enterprise Validation

Step-by-step acceptance checks are documented in `docs/ENTERPRISE_MIGRATION_PLAN.md` (Knowledge API, SSE filtering, middleware, Sheets, ingestion).

## Behavior Changes

- None; code changes are additive. Focus is on docs alignment and small feature completions.

