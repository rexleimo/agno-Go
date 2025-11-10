## Overview
EvoLink support becomes a reusable provider under `pkg/agno`. The implementation mirrors existing providers (OpenAI, Gemini, etc.) so any agent can call EvoLink through a single config surface. The provider exposes:

- Config structs that read from env/constructor args (`EVO_API_KEY`, `EVO_BASE_URL`, timeouts).
- Validated request builders for Sora-2 video, GPT-4O image, and GPT-4O text endpoints.
- An async polling helper for task-based endpoints that clients can reuse.

## Components
1. **Provider Package (`pkg/agno/providers/evolink`)**  
   - Subpackages for `video`, `image`, `text`, each exporting typed `Options`, `Client`, and `Response` structs.  
   - Shared `client.go` handles base URL, auth header, JSON encode/decode, error wrapping, and `GetTask` polling.

2. **Model Interfaces**  
   - Implement `models.Model` (or the relevant interface) so that `pkg/agno/session`, workflows, and examples can instantiate EvoLink agents just like other providers.  
   - Provide constructors such as `models/evolink/video.New(...)`, `models/evolink/image.New(...)`, `models/evolink/text.New(...)`.

3. **Documentation & Samples**  
   - Update website example pages to demonstrate using the provider within Agno workflows (e.g., `agents.New(...)` referencing EvoLink models) rather than a bespoke CLI.

## Validations
- Unit tests under `pkg/agno/providers/evolink/...` covering payload validation, polling, and error paths using httptest servers.
- Example integration test (or doc snippet) showing a workflow that chains EvoLink text → image → video.
- Docs build check (`npm run docs:build`) to ensure the updated example renders.
