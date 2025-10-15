# Documentation Layout

Policy: implemented, user-facing docs live in `website/`; design drafts, WIP, internal notes live in `docs/`.

- `website/`: VitePress site (Guides, API, Advanced) for features that are implemented.
- `docs/`: Design proposals, migration plans, tasks, and developer/internal docs.

Keep a single source of truth. Do not duplicate implemented docs into `docs/`.

Current categories in `docs/`:
- Design/WIP: `ENTERPRISE_MIGRATION_PLAN.md`, `M1_SQLITE_SESSION_DESIGN.md`, `task/`
- Developer: `DEVELOPMENT.md`, `VITEPRESS.md`

Website entry points:
- Guides: `website/guide/`
- API: `website/api/`
- Advanced: `website/advanced/`

