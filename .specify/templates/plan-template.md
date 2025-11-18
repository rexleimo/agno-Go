# 实现计划：[FEATURE]

**分支**：`[###-feature-name]` | **日期**： [DATE] | **规格**： [链接]
**输入**：来自 `/specs/[###-feature-name]/spec.md` 的功能规格

**说明**：该模板由 `/speckit.plan` 命令填充。执行流程见 `.specify/templates/commands/plan.md`。

## 摘要

[摘自功能规格：关键需求 + 调研后的技术路径]

## 技术背景

<!--
  必须：将此处内容替换为该项目的技术细节。
  结构仅为建议，可根据迭代需要调整。
-->

**语言/版本**： [例如 Python 3.11、Swift 5.9、Rust 1.75 或 NEEDS CLARIFICATION]  
**主要依赖**： [例如 FastAPI、UIKit、LLVM 或 NEEDS CLARIFICATION]  
**存储**： [如适用，例如 PostgreSQL、CoreData、文件，或 N/A]  
**测试**： [例如 pytest、XCTest、cargo test 或 NEEDS CLARIFICATION]  
**目标平台**： [例如 Linux server、iOS 15+、WASM 或 NEEDS CLARIFICATION]  
**项目类型**： [single/web/mobile —— 决定源码结构]  
**性能目标**： [领域相关，例如 1000 req/s、10k lines/sec、60 fps 或 NEEDS CLARIFICATION]  
**约束条件**： [如 <200ms p95、<100MB 内存、离线可用，或 NEEDS CLARIFICATION]  
**规模/范围**： [例如 1 万用户、100 万 LOC、50 个页面，或 NEEDS CLARIFICATION]

## 宪章检查

*Gate：Phase 0 调研前必须通过；Phase 1 设计后需再次检查。*

- [ ] **Go + DDD / 热插拔**：明确受影响的限界上下文、接口契约与可插拔模块，阐述停用该模块时系统如何保持可运行。
- [ ] **Compose-First 可部署性**：列出需要新增/更新的 Compose 服务、健康检查与 `.env` 变量，并说明如何通过 `make compose-*` 目标验证。
- [ ] **GORM + 迁移版本控制**：说明本迭代涉及的表/实体、需要新增的迁移文件名及验证策略，确认 `make migrate` / `make rollback` 流程完整。
- [ ] **多存储适配矩阵**：列出需要支持/新增的 SQLite、MySQL、PostgreSQL、MongoDB、Redis、DynamoDB、Firestore 驱动，说明所在适配器路径、`make data-matrix`/`make compose-data` 验证方式以及禁用某驱动时的替代策略。
- [ ] **Makefile 自动化**：标记需要新增或调整的 `make` 目标，并说明 CI 复用方式，避免出现与文档不一致的裸命令。
- [ ] **Remix + React Router V7 + pnpm + shadcn**：枚举要交付的前端 workspace、路由、shadcn 组件与 Apple/Microsoft 设计引用章节，说明如何保证热插拔的 UI 配置。
- [ ] **Vitepress + GitHub Docs Workflow**：说明本功能涉及的文档章节、`docs/vitepress` 目录增量、`make docs-*`/`pnpm docs:*` 操作以及 `.github/workflows/docs.yml` 的部署策略。
- [ ] **测试纪律 + 85% 覆盖率**：列出需要新增/更新的 Go/Remix/文档单元测试、契约/集成测试、`make test|ui-test|docs-test|data-matrix|coverage` 的执行方式以及若覆盖率下降的补救方案。
- [ ] **平台运行约束**：确认 Compose 环境的可观察性、安全（Secrets/SOPS）和默认数据库/缓存策略未被破坏，如有例外需附上补偿控制。

## 项目结构

### 文档（当前功能）

```text
specs/[###-feature]/
├── plan.md              # 本文件（/speckit.plan 输出）
├── research.md          # Phase 0 输出（/speckit.plan）
├── data-model.md        # Phase 1 输出（/speckit.plan）
├── quickstart.md        # Phase 1 输出（/speckit.plan）
├── contracts/           # Phase 1 输出（/speckit.plan）
└── tasks.md             # Phase 2 输出（/speckit.tasks，非 /speckit.plan 创建）
```

### 源码（仓库根目录）
<!--
  必须：将占位树替换为真实结构，删除未用选项，补上真实路径（如 apps/admin、packages/foo）。交付的计划中不得保留 Option 标签。
-->

```text
backend/
├── cmd/
│   └── api/                  # Go 入口（组合限界上下文）
├── internal/
│   ├── <contextA>/           # 限界上下文：entity、aggregate、service、ports
│   └── <contextB>/
├── pkg/                      # 可复用适配器/工具
└── tests/
    ├── contract/
    ├── integration/
    └── unit/

db/
└── migrations/               # `<timestamp>_<name>.sql|.go`

frontend/
├── apps/web/                 # Remix + React Router V7 应用
├── packages/ui/              # 基于 shadcn/ui 的设计系统
└── packages/<feature>/       # 可热插拔组件/模块

deploy/compose/
├── docker-compose.local.yml
└── docker-compose.ci.yml

docs/vitepress/
├── .vitepress/config.ts      # 文档导航、部署配置
└── content/                  # 组件/数据库/部署文档

configs/datastores/           # 多存储矩阵配置（可选）

.github/workflows/
└── docs.yml                  # Vitepress 自动化部署

Makefile                      # 单一入口（dev/test/build/release）
```

**结构决策**： [记录所选结构，并引用上方列出的真实目录]

## 复杂度追踪

> **仅当宪章检查存在违规且必须说明时填写**

| 违规项 | 必要原因 | 更简单方案被拒绝的理由 |
|--------|----------|--------------------------|
| [例：第 4 个项目] | [当前需求] | [为何 3 个项目不够] |
| [例：Repository 模式] | [具体问题] | [为何直接 DB 访问不够] |
