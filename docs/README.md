# Documentation Index

Agno-Go maintains two documentation surfaces:

- **`website/`** – the VitePress site that ships with the project (Guides, API reference, Advanced topics, Examples, translations). This is the canonical reference for user-facing docs.
- **`docs/`** – internal notes, design proposals, migration plans, and contributor-oriented material. Anything that graduates to a user-facing feature should be promoted to `website/`.

Please avoid duplicating the same content across both locations. Use the tables below to navigate the current internal docs.

## Developer & Process Notes

| File | Description |
| --- | --- |
| `DEVELOPMENT.md` | Local development workflow, tooling, testing standards. |
| `VITEPRESS.md` | How to build and preview the documentation site locally. |
| `task/` | Engineering task documentation and completed implementation notes. |

## Designs & Proposals

| File | Description |
| --- | --- |
| `ENTERPRISE_MIGRATION_PLAN.md` | Enterprise feature migration roadmap. |
| `_archive/M1_SQLITE_SESSION_DESIGN.md` | Session storage design (SQLite focus). |

## Release Notes & Playbooks

| File | Description |
| --- | --- |
| `release/v1.2.6-announcement.md` | Launch article and social copy for the v1.2.6 release. |
| `release/logfire_observability.md` | How to wire Agno-Go telemetry into Logfire using OpenTelemetry. |
| `_archive/migration_plan.md` | High-level v2.1.5 migration tracker. |

## Website Quick Links

| Section | Path |
| --- | --- |
| Guides | `website/guide/` |
| API Reference | `website/api/` |
| Advanced Topics | `website/advanced/` |
| Examples | `website/examples/` |
| Translations | `website/ja/`, `website/ko/`, `website/zh/` |

## Contributing to Docs

1. Add or edit design notes in `docs/`.
2. For ready features, create/update the corresponding page in `website/`.
3. Run `pnpm docs:dev` (see `VITEPRESS.md`) to preview changes.
4. Keep this index up to date when new documents are added.
