## Why
The new EvoLink documentation site (offline Sora-2 HTML under `/home/rex/下载/Sora-2 视频生成 - EvoLink.html` plus the GPT-4O image page at `https://docs.evolink.ai/en/api-manual/image-series/gpt-4o/gpt-4o-image-generation`) exposes concrete requirements for video, image, and narrative agents. Instead of only providing an example CLI, we want EvoLink to be a first-class provider inside the framework so that any agent or workflow can consume EvoLink video/image/text capabilities.

## What Changes
- Add a `pkg/agno/providers/evolink` (exact path TBD) that encapsulates API clients for Sora-2 video, GPT-4O image, and GPT-4O text (chat completions) following the latest doc constraints.
- Expose typed configs, option validation, and async task polling utilities (GET `/v1/tasks/{id}`) as reusable components that other packages (agents, workflows, examples) can import.
- Provide an example agent wiring (could still add a CLI wrapper later) plus documentation in `website/examples/evolink-media-agents.md` / zh mirror explaining how to configure and invoke the new provider from any Agno-Go integration.

## Impact
- **Affected specs:** rename to `media-agents-provider` and `media-agents-docs`.
- **Affected code:** `pkg/agno/**` (new EvoLink provider + models), potential shared polling utilities, and website example pages (now referencing the provider instead of a standalone CLI).
