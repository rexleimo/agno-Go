## ADDED Requirements

### Requirement: Provide EvoLink provider inside pkg/agno
The framework SHALL expose EvoLink support under `pkg/agno/providers/evolink` that covers Sora-2 video, GPT-4O image, and GPT-4O text endpoints. The provider must read `EVO_API_KEY` (required) and `EVO_BASE_URL` (default `https://api.evolink.ai`), offer typed option structs matching the EvoLink doc constraints (aspect ratios 16:9/9:16, durations 10/15, video reference ≤1 image, image sizes {1:1,2:3,3:2,1024x1024,1024x1536,1536x1024}, `n ∈ {1,2,4}`, up to 5 references, mask limited to single PNG, HTTPS-only callbacks), and expose an async task polling helper for `/v1/tasks/{id}`.

#### Scenario: Creating a Sora-2 video task from agents
- **WHEN** an agent invokes `evolink/video.New(...).Invoke(ctx, prompt, options)` with `aspect_ratio="9:16"`, `duration=15`, `reference=https://.../frame.png`, and `remove_watermark=false`  
- **THEN** the provider MUST send `POST /v1/videos/generations` with those fields, enforce the documented limits, poll `GET /v1/tasks/{task_id}` until `completed|failed|cancelled`, and return the final task payload (including URLs) to the caller or an error if the status is not `completed`.

#### Scenario: Generating GPT-4O images with references + mask
- **WHEN** a workflow calls `evolink/image.New(...).Invoke(ctx, prompt, WithReferences([...]), WithMaskURL(...), WithCount(4), WithSize("1:1"))`  
- **THEN** the provider MUST validate the `n=4` constraint, ensure exactly one reference is present when using `mask_url`, send `POST /v1/images/generations`, poll the task, and surface the resulting image URLs (noting they expire in 24 h) through the model response.

### Requirement: Expose EvoLink models via pkg/agno/models
`pkg/agno/models/evolink/*` MUST implement Agno’s model interfaces so sessions/agents can configure EvoLink via Go constructors without touching HTTP details. Constructors need to accept config structs, support dependency injection (custom HTTP client, logger), and provide helpful errors if `EVO_API_KEY` is missing or the EvoLink API rejects parameters.

#### Scenario: Wiring EvoLink text into a workflow
- **WHEN** a developer instantiates `models/evolink/text.New(models.EvolinkConfig{APIKey: os.Getenv("EVO_API_KEY")})` and executes a storyboard request (`shots=5`, `temperature=0.6`, `topic="Underwater city"`),  
- **THEN** the model MUST call `POST /v1/chat/completions`, return the assistant content, and respect the configured timeout.

### Requirement: Document provider usage
`website/examples/evolink-media-agents.md` (English + zh) SHALL explain how to configure and invoke the new provider, referencing the official EvoLink docs, required env vars, supported parameters, async behavior, and compliance considerations (moderation, HTTPS callbacks, 24 h asset expiry).

#### Scenario: Reviewer reads the English page
- **WHEN** someone opens `website/examples/evolink-media-agents.md`  
- **THEN** they find Go snippets using the new provider, guidance on `EVO_API_KEY` / `EVO_BASE_URL`, tables for video/image/text options, links to `/home/rex/下载/Sora-2 视频生成 - EvoLink.html` and the GPT-4O Mintlify page, plus warnings about strict moderation and asset expiry. The zh page must mirror the same content in Chinese.
