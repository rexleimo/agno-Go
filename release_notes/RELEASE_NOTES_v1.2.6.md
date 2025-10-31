# Release Notes - Agno-Go v1.2.6

Date: 2025-10-31

## Highlights
- Session runtime parity with new reuse endpoint, sync/async summaries, history filters, and enriched run metadata (cache hits, cancellation reasons, timestamps).
- Response caching for agents and teams via in-memory LRU store plus pluggable summary manager configuration.
- Media attachment pipeline for agents, teams, and workflows, including validation helpers and workflow `WithMediaPayload` option.
- New storage adapters for MongoDB and SQLite alongside existing Postgres implementation, delivering identical JSON contracts.
- Expanded toolkits: Tavily Reader/Search, Claude Agent Skills, Gmail mark-as-read, Jira worklogs, ElevenLabs speech synthesis, and upgraded file utilities.
- Culture knowledge manager for curating organisational knowledge with async operations and tag filtering.

## Fixes & Improvements
- Workflow engine now persists cancellation reasons, supports resume-from checkpoints, and can execute media-only payloads.
- AgentOS session handlers expose summary endpoints, reuse semantics, and pagination for history queries.
- MCP client forwards media attachments and caches capability manifests for reduced round-trips.

## Compatibility
- Backward compatible (additive changes).
