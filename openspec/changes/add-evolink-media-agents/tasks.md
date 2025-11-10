## 1. Implementation
- [x] 1.1 Create `pkg/agno/providers/evolink` with shared HTTP client, config, and task polling utilities aligned with `/home/rex/下载/Sora-2 视频生成 - EvoLink.html` and `https://docs.evolink.ai/en/api-manual/image-series/gpt-4o/gpt-4o-image-generation`.
- [x] 1.2 Add `pkg/agno/models/evolink/video`, `image`, and `text` packages that expose typed constructors (`New`) hooking into the provider and enforcing parameter validation (aspect ratios, durations, sizes, counts, mask/reference limits, HTTPS callbacks).
- [ ] 1.3 Integrate the models into at least one agent pipeline (e.g., update an existing workflow or add a lightweight orchestrator) to prove they can be composed like other providers.
- [x] 1.4 Update `website/examples/evolink-media-agents.md` and the zh mirror to document how to configure the provider, including env vars, sample Go snippets, and compliance callouts.

## 2. Validation
- [x] 2.1 `go test ./pkg/agno/providers/evolink/... ./pkg/agno/models/evolink/...` covering success/failure flows with httptest servers.
- [ ] 2.2 Run or simulate a workflow using the new models (unit/integration test or `go run` snippet) that chains text → image → video to confirm compatibility.
- [ ] 2.3 `npm run docs:build` (inside `website/`) to ensure the documentation updates compile.
