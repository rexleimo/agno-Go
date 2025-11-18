<!--
版本变更：1.1.0 → 1.2.0
被修改的原则：
- GORM 数据治理与迁移（SQL 驱动矩阵） → GORM 数据治理与迁移（SQL 驱动矩阵 + 覆盖率校验）
- 多存储适配与数据服务矩阵 → 多存储适配与数据服务矩阵（含测试矩阵）
- Remix + React Router V7 体验标准 → Remix + React Router V7 体验标准（前端单测约束）
- Vitepress 文档与 GitHub 自动化发布 → Vitepress 文档与 GitHub 自动化发布（文档测试门）
新增原则：
- 全栈测试纪律与 85%+ 覆盖率
新增章节：无（既有章节扩写）
移除章节：无
模板同步：
- ✅ .specify/templates/plan-template.md
- ✅ .specify/templates/spec-template.md
- ✅ .specify/templates/tasks-template.md
- ⚠️ .specify/templates/commands/（目录不存在，无文件可同步）
未完成的 TODO：无
-->
# Agno Go 宪章

## 核心原则

### Go + DDD 内核与热插拔服务
- 所有后端应用 MUST 使用 Go，并以领域驱动设计（DDD）的限界上下文、聚合、仓储、应用服务模式组织代码，禁止临时脚本或其他运行时语言直接进入生产路径。
- 每个能力必须实现稳定的 Go 接口（`/internal/<context>/ports.go` 等），并通过依赖注入或插件注册方式提供实现，以便按需装载/卸载，确保“所有功能可热插拔”。
- 领域逻辑与适配器需完全解耦；变更不得跨越上下文耦合数据库模型或 HTTP handler，热插拔模块只能依赖接口和领域事件。
- 针对每个上下文，必须提供契约级别的单元/集成测试，并在禁用该模块时保持其它上下文可运行，以验证热插拔能力。

### Compose-First 可部署性
- 仓库必须提供单条 `docker compose up` 流程，将 Go 服务、Remix 前端、数据库、队列、缓存、Vitepress 及其它依赖全部启动，新增依赖必须同步更新 Compose 文件及 `.env.example`。
- Compose 服务需定义健康检查、卷与网络隔离，且默认镜像均源自本仓库构建的镜像或已验证的公开镜像；临时脚本或手动步骤不得作为部署前提。
- 任何基础设施或依赖服务变更都必须在 PR 中更新 Compose 文件、文档以及对应的 `make compose-*` 目标，确保 CI/CD 与本地体验一致。

### GORM 数据治理与迁移（SQL 驱动矩阵）
- SQL 数据访问层 MUST 使用 GORM；SQLite、MySQL、PostgreSQL 的仓储均需共享领域接口及统一迁移定义，并禁止绕过 GORM 直接拼接 SQL。
- 所有 schema 变更需通过受版本控制的迁移（`/db/migrations/<timestamp>_<name>.sql|.go`），迁移必须包含 up/down、变更说明及回滚验证，且覆盖三种 SQL 驱动。
- `make migrate`、`make rollback`、`make sqldb-test` 需要串联 gormigrate/golang-migrate，并在 CI 中针对 SQLite/MySQL/PostgreSQL 并行运行；直接修改数据库或导出 dump 被禁止。
- 每次发布必须记录当前迁移版本号，并在 Compose 启动、健康检查与 `make coverage` 报告中验证三种 SQL 数据库迁移与测试均通过。

### 多存储适配与数据服务矩阵
- 项目内必须提供 SQLite、MySQL、PostgreSQL、MongoDB、Redis、DynamoDB、Firestore 的独立适配器，并通过 DDD 仓储接口热插拔；NoSQL/缓存驱动需位于 `backend/internal/<context>/infra/datastore/<provider>/` 并以依赖注入启用。
- `deploy/compose/` 必须包含相应服务或本地模拟器（MongoDB/Redis 原生镜像，DynamoDB/Firestore 可使用 LocalStack/Firestore emulator），并提供 `make compose-data` 快速启停。
- `make data-matrix` 必须运行跨存储契约/单元/集成测试，确保任一驱动被禁用或替换时其余驱动仍可运行，且测试结果纳入覆盖率统计。
- 规格与实现需要说明某个功能依赖哪种存储，并提供切换说明，禁止引入未列出的数据库类型。

### Makefile 驱动的自动化流水线
- Makefile 是唯一入口，涵盖初始化、依赖安装、构建、测试、代码生成、Compose 管理、迁移、Vitepress、部署打包；文档中禁止调用裸命令而绕过 Make。
- 新增工作流（如 `pnpm install`, `go generate`, `docker build`, `pnpm docs:build`, `pnpm test --coverage`）必须对应 `make` 目标或扩展现有目标，且目标需通过 `help`/注释描述用途与依赖顺序。
- 所有 CI 任务必须复用相同的 `make` 目标，确保本地与流水线行为一致；绕过 Make 的脚本将被拒绝。

### Remix + React Router V7 体验标准
- 前端唯一合法栈为 Remix + React Router V7，项目需通过 pnpm workspace 管理 `apps/*` 与 `packages/*`；新增应用或包必须注册在 `pnpm-workspace.yaml` 中，并与 Vitepress workspace 共存。
- UI 组件基于 shadcn/ui 原子组件构建，并通过 Apple Human Interface Guidelines 与 Microsoft Fluent 设计语言定义的 tokens（色彩、动效、间距）驱动；提交中需说明引用的设计规范。
- 路由、数据加载、action 必须使用 Remix/React Router 提供的 data APIs（loader/action/defer）；任何自定义路由实现需经架构评审。
- 组件库与页面布局需保持主题化和按需分发，支持在不破坏 API 的情况下替换视觉层，以匹配“热插拔”理念。

### Vitepress 文档与 GitHub 自动化发布
- `docs/vitepress` 是唯一权威文档站点，必须纳入 pnpm workspace，提供 `pnpm docs:dev`、`pnpm docs:build`、`pnpm docs:preview`，并在 Makefile 中映射 `make docs-*` 目标。
- 每个功能、数据库适配器或 UI 组件在交付前必须更新 Vitepress 内容，新增或修改条目需包含设计语言引用、API/CLI 示例以及 `docker compose`/`make` 操作指南。
- `.github/workflows/docs.yml`（或等价文件）必须在 docs 目录变更时自动构建并部署 Vitepress（GitHub Pages/静态存储）；禁用该流程需提交架构批准与替代方案。
- 文档构建必须在 CI 中运行（`make docs-test`），并阻止破坏性变更；缺失文档更新的 PR 不得合并。

### 全栈测试纪律与 85%+ 覆盖率
- 所有 Go 包、Remix/shadcn 组件、数据适配器与 Vitepress 自定义逻辑必须具备单元测试文件（`*_test.go`、`*.spec.ts(x)`、`docs/vitepress/tests/*.ts` 等），并在 PR 中随实现同时提交。
- `make test`（Go + 多数据库驱动）、`make ui-test`（Remix/shadcn）、`make docs-test`（Vitepress）与 `make data-matrix` 必须在 CI 中执行，任何跳过都需架构批准并附补救计划。
- `make coverage` 需聚合 Go 覆盖率（`go test ./... -coverpkg=./...`）、pnpm/Remix 覆盖率（`pnpm test --coverage`）、Vitepress 自测，综合覆盖率 MUST ≥85%，并通过构建日志、badge 或 PR 机器人公布；低于阈值的 PR 必须补测或缩小改动范围。
- 临时跳过测试（`t.Skip`, `test.skip`, `describe.skip` 等）必须在 PR 描述中列出并登记补测 issue，禁止在生产分支长期存在。

## 全栈技术栈要求
- **语言与运行时**：后端锁定 Go 1.23+（升级需验证所有模块兼容），前端锁定 TypeScript 5.x + Remix 2.x；跨语言组件通过 gRPC/REST 契约通信。
- **项目结构**：`backend/` 容纳 Go 服务（按限界上下文划分子模块），`frontend/` 由 pnpm workspace 管理 Remix 应用与 shadcn 基础库，`docs/vitepress/` 保存文档站点，`deploy/compose/` 储存环境特定 Compose 文件，`db/migrations/` 保存迁移。
- **数据与缓存矩阵**：SQLite 使用本地文件卷，MySQL/PostgreSQL 通过独立容器，MongoDB/Redis 使用官方镜像，DynamoDB/Firestore 通过 LocalStack 或官方 emulator；全量驱动需在 `configs/datastores/*.yaml` 或等价目录声明。
- **可观察性**：Compose 环境必须暴露 Prometheus/OpenTelemetry collector，Makefile 中需提供 `make observe` 以启动/查看指标，并在多存储模式下产出独立指标命名空间。
- **文档与知识**：Vitepress 主题必须实现 Apple/Microsoft 设计语言 tokens，`docs/vitepress/.vitepress/config.ts` 需记录版本信息、导航与多存储矩阵；GitHub Workflow 的部署密钥必须通过 SOPS/secret manager 注入。
- **测试与覆盖**：必须配置 `make test`、`make ui-test`、`make docs-test`、`make data-matrix`、`make coverage`，并在 CI 中强制运行；覆盖率报告需上传至构建工件或 README/Docs badge。
- **安全性**：所有 secrets 通过 `.env` 模板 + SOPS/密管注入，禁止硬编码；CI 必须在密文缺失时失败，且多数据库凭据需分离。

## 开发流程与质量门
1. **设计前置**：所有 speckit 规格需声明所涉限界上下文、需要热插拔的模块、依赖的数据库/缓存驱动、需要新增的 Compose 服务、数据库迁移、Vitepress 页面、必须补齐的测试类型（unit/contract/integration/docs）以及必需的 Make/coverage 目标。
2. **计划 Gate**：Plan 中的 “宪章检查” 必须逐条验证 Go+DDD、Compose、GORM+迁移、多存储矩阵、Makefile 自动化、Remix+pnpm+shadcn、Vitepress+GitHub Workflow、测试与 85% 覆盖率以及平台运行约束。
3. **实现 Gate**：
   - 新增后端功能前先生成/修改迁移、SQL/NoSQL 驱动以及契约/单元测试，并通过 `make test`、`make lint`、`make compose-test`、`make data-matrix`；未覆盖的新代码不得合并。
   - 前端变更需在 pnpm workspace 中新增/更新包并运行 `make ui-test`、`pnpm test --coverage`，且 PR 描述需链接所采用的设计语言章节与 UI 测试路径。
   - 文档更新必须包含 Vitepress 页面/导航、示例代码与部署步骤，并通过 `make docs-build`、`make docs-test`；PR 需附带 GitHub workflow 运行记录。
   - 任意 PR 都必须附带 Compose、Makefile、Docs 变更与 `make coverage` 报告的验证结果（日志或截图）。
4. **发布 Gate**：`make release` 必须打包 Go 二进制、前端构建产物、Vitepress 静态资源与 Compose 清单，记录 SQL 迁移版本号、多存储驱动版本、最新覆盖率及 Vitepress 构建号；发布说明中必须包含热插拔模块、数据库驱动的启用/禁用、测试与覆盖率报告及文档链接。

## 治理
- 本宪章优于其他工程规范；任何冲突由架构小组裁决，未列明主题默认遵循 Go/Remix/Vitepress 社区最佳实践。
- 修订流程：提出 RFC → 在 speckit 文档中描述影响（涵盖多存储矩阵、Vitepress 部署与测试覆盖）→ 更新宪章与相关模板 → 依据变更范围更新版本号（MAJOR：原则重写或删除；MINOR：新增原则/章节或大幅扩写；PATCH：措辞澄清）。
- 监督机制：每次 PR 审查必须引用 Plan 的宪章检查项；CI 需运行 `make constitution-check`（调用 lint/test/migration/data/compose/docs/coverage 验证）并在失败时阻塞。
- 合规审查：季度审查要回顾 Compose 可用性、多数据库驱动一致性、迁移版本差异、Make 目标覆盖率、前端体验一致性、Vitepress/GitHub Workflow 成功率与覆盖率指标；发现违规需在两周内补齐或提交宪章修订提案。

**Version**: 1.2.0 | **Ratified**: 2025-11-18 | **Last Amended**: 2025-11-18
