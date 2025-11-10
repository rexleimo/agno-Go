## ADDED Requirements

### Requirement: EvoLink docs align with provider usage
The website SHALL offer English + zh example pages demonstrating how to use the new EvoLink provider inside Agno workflows. Each page must include:
- Environment setup (`EVO_API_KEY`, `EVO_BASE_URL`).
- Go snippets for video, image, and text agents referencing `pkg/agno/models/evolink`.
- Parameter tables summarizing aspect ratios, durations, sizes, `n` values, reference/mask limits, webhook rules, and 24 h asset expiry.
- Links to `/home/rex/下载/Sora-2 视频生成 - EvoLink.html` and `https://docs.evolink.ai/en/api-manual/image-series/gpt-4o/gpt-4o-image-generation`.

#### Scenario: zh page parity
- **WHEN** the zh reviewer opens `website/zh/examples/evolink-media-agents.md`  
- **THEN** it mirrors the English content (translated) and references the same EvoLink docs, constraints, and warnings.
