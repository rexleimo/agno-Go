<!--
版本变更：0.0.0 → 1.0.0
被修改的原则：
- [PRINCIPLE_1_NAME] → Go + DDD 内核与热插拔服务
- [PRINCIPLE_2_NAME] → Compose-First 可部署性
- [PRINCIPLE_3_NAME] → GORM 数据治理与迁移
- [PRINCIPLE_4_NAME] → Makefile 驱动的自动化流水线
- [PRINCIPLE_5_NAME] → Remix + React Router V7 体验标准
新增章节：
- 全栈技术栈要求
- 开发流程与质量门
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
- 仓库必须提供单条 `docker compose up` 流程，将 Go 服务、Remix 前端、数据库及额外依赖（队列、缓存等）全部启动，新增依赖必须同步更新 Compose 文件及 `.env.example`。
- Compose 服务需定义健康检查、卷与网络隔离，且默认镜像均源自本仓库构建的镜像或已验证的公开镜像；临时脚本或手动步骤不得作为部署前提。
- 任何基础设施变更都必须在 PR 中更新 Compose 文件、文档以及对应的 `make compose-*` 目标，确保 CI/CD 与本地体验一致。

### GORM 数据治理与迁移
- 数据访问层 MUST 使用 GORM；仅允许在性能关键路径通过仓储对象执行原生 SQL，且必须附带基准与回退方案。
- 所有 schema 变更需通过受版本控制的迁移（`/db/migrations/<timestamp>_<name>.sql|.go`），迁移必须包含 up/down、变更说明及回滚验证。
- `make migrate` 和 `make rollback` 需要串联 GORM/迁移工具（如 gormigrate 或 golang-migrate），并在 CI 中自动运行；直接修改数据库或导出 dump 被禁止。
- 每次发布必须记录当前迁移版本号，并在 Compose 启动与健康检查中验证迁移已执行。

### Makefile 驱动的自动化流水线
- Makefile 是唯一入口，涵盖初始化、依赖安装、构建、测试、代码生成、Compose 管理与部署打包；文档中禁止调用裸命令而绕过 Make。
- 新增工作流（如 `pnpm install`, `go generate`, `docker build`）必须对应新的 `make` 目标或扩展现有目标，且目标需通过 `help`/注释描述用途与依赖顺序。
- 所有 CI 任务必须复用相同的 `make` 目标，确保本地与流水线行为一致；绕过 Make 的脚本将被拒绝。

### Remix + React Router V7 体验标准
- 前端唯一合法栈为 Remix + React Router V7，项目需通过 pnpm workspace 管理 `apps/*` 与 `packages/*`；新增应用或包必须注册在 `pnpm-workspace.yaml` 中。
- UI 组件基于 shadcn/ui 原子组件构建，并通过 Apple Human Interface Guidelines 与 Microsoft Fluent 设计语言定义的 tokens（色彩、动效、间距）驱动；提交中需说明引用的设计规范。
- 路由、数据加载、action 必须使用 Remix/React Router 提供的 data APIs（loader/action/defer）；任何自定义路由实现需经架构评审。
- 组件库与页面布局需保持主题化和按需分发，支持在不破坏 API 的情况下替换视觉层，以匹配“热插拔”理念。

## 全栈技术栈要求
- **语言与运行时**：后端锁定 Go 1.23+（升级需验证所有模块兼容），前端锁定 TypeScript 5.x + Remix 2.x；跨语言组件通过 gRPC/REST 契约通信。
- **项目结构**：`backend/` 容纳 Go 服务（按限界上下文划分子模块），`frontend/` 由 pnpm workspace 管理 Remix 应用与 shadcn 基础库，`deploy/compose/` 储存环境特定 Compose 文件，`db/migrations/` 保存迁移。
- **依赖服务**：默认数据库为 PostgreSQL 15+，缓存优先选用 Redis 7+；引入其他服务需同时提供 Compose 配置与运行文档。
- **可观察性**：Compose 环境必须暴露 Prometheus/OpenTelemetry collector，Makefile 中需提供 `make observe` 以启动/查看指标。
- **安全性**：所有 secrets 通过 `.env` 模板 + SOPS/密管注入，禁止硬编码；CI 必须在密文缺失时失败。

## 开发流程与质量门
1. **设计前置**：所有 speckit 规格需声明所涉限界上下文、需要热插拔的模块、需要新增的 Compose 服务、数据库迁移及 Make 目标。
2. **计划 Gate**：Plan 中的 “宪章检查” 必须逐条验证五大原则（Go+DDD、Compose、GORM+migration、Makefile、Remix+React Router/shadcn/pnpm）。
3. **实现 Gate**：
   - 新增后端功能前先生成/修改迁移与契约测试，并通过 `make test`、`make lint`、`make compose-test`。
   - 前端变更需在 pnpm workspace 中新增/更新包并运行 `make ui-test`，且 PR 描述需链接所采用的设计语言章节。
   - 任意 PR 都必须附带 Compose 与 Makefile 变更验证结果（日志或截图）。
4. **发布 Gate**：`make release` 需打包 Go 二进制、前端构建产物与 Compose 清单，并记录迁移版本号；发布说明中必须包含热插拔模块的启用/禁用步骤。

## 治理
- 本宪章优于其他工程规范；任何冲突由架构小组裁决，未列明主题默认遵循 Go/Remix 社区最佳实践。
- 修订流程：提出 RFC → 在 speckit 文档中描述影响 → 更新宪章与相关模板 → 依据变更范围更新版本号（MAJOR：原则重写或删除；MINOR：新增原则/章节或大幅扩写；PATCH：措辞澄清）。
- 监督机制：每次 PR 审查必须引用 Plan 的宪章检查项；CI 需运行 `make constitution-check`（规则校验脚本，可调用 lint/test/migration/ds 验证）并在失败时阻塞。
- 合规审查：季度审查要回顾 Compose 可用性、迁移版本差异、Make 目标覆盖率与前端设计一致性；发现违规需在两周内补齐或提交宪章修订提案。

**Version**: 1.0.0 | **Ratified**: 2025-11-18 | **Last Amended**: 2025-11-18
