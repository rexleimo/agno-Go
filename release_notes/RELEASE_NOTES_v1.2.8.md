# Release Notes — v1.2.8 (2025-11-10)

## ✨ Highlights
- Run Context flows through execution (hooks → tools → telemetry) with `run_context_id` in streaming events for end-to-end correlation.
- Session state persists `AGUI` substate; `GET /sessions/{id}` returns UI state within `session_state`.
- Vector indexing enhancements:
  - Pluggable VectorDB providers (Chroma by default; Redis optional, not a hard dependency).
  - VectorDB migration CLI (`migrate up/down`) for idempotent collection/index setup.
- Embeddings: VLLM provider (local/remote) implementing the common `EmbeddingFunction` interface.
- MCPTools: optional `tool_name_prefix` to register prefixed tool names without behavior changes.

## 🔧 Improvements
- Redis decoupled from default vector DB dependency. No impact when not configured; enable via configuration to register provider.
- Team model inheritance now propagates only the primary model. Auxiliary parameters require explicit agent-level opt-in.

## 🐛 Fixes
- Bound model responses correctly to the active step to avoid unbound/zero-value results in histories.
- Team tool determination aligns with OS schema, preserving member agent tools.
- Async DB-backed knowledge filters honor composite predicates and context timeouts without goroutine leaks.
- Toolkit import resolution returns structured errors for missing modules instead of panicking.
- AgentOS error handling path standardized for cleaner contract assertions.

## 🧪 Tests
- Coverage added for Run Context propagation, AGUI persistence, team primary-model inheritance, MCP prefixing, and VLLM embeddings.
- Optional Redis test jobs are gated and skipped when the dependency is absent.

## ✅ Compatibility
- Additive release. Optional capabilities are disabled by default and do not change public APIs.

## 📚 Links
- Website Release Notes: /release-notes#version-128-2025-11-10
- Changelog: `CHANGELOG.md`
