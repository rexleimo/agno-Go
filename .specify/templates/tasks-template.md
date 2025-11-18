---

description: "功能实施任务模板"
---

# 任务清单：[FEATURE NAME]

**输入**：`/specs/[###-feature-name]/` 中的设计文档
**前置**：plan.md（必需）、spec.md（用户故事必需）、research.md、data-model.md、contracts/

**测试**：示例中包含测试任务。测试是可选项——仅当功能规格明确要求或用户主动请求时才加入。

**覆盖率**：所有功能必须交付匹配的单元测试，`make test`、`make ui-test`、`make docs-test`、`make data-matrix`、`make coverage` 必须可运行并产出 ≥85% 的综合覆盖率；任务描述需明确测试文件与命令。

**栈约束**：所有任务必须指向真实路径，覆盖 `backend/internal/<context>/`（Go + DDD + 热插拔）、`backend/internal/<context>/infra/datastore/<provider>/`（多存储适配器）、`db/migrations/`（GORM 迁移）、`frontend/apps/*` 与 `frontend/packages/*`（Remix + React Router V7 + shadcn）、`docs/vitepress/`（文档）、`deploy/compose/`（Docker Compose）、`.github/workflows/docs.yml`（自动化部署）以及 `Makefile`（自动化入口）。

**组织方式**：任务按用户故事分组，确保每个故事都能独立实现与测试。

## 格式：`[ID] [P?] [Story] 描述`

- **[P]**：可并行执行（不同文件，且无依赖）
- **[Story]**：所属用户故事（如 US1、US2、US3）
- 描述中必须包含精确文件路径

## 路径约定

- **后端（Go + DDD）**：`backend/cmd/`（入口），`backend/internal/<context>/`（聚合/服务/ports），`backend/pkg/`（公共适配器），`backend/tests/`（contract/integration/unit）。
- **数据层/多存储**：`db/migrations/`（受版本控制的 GORM 迁移），`backend/internal/<context>/infra/datastore/<provider>/`（SQLite/MySQL/PostgreSQL/MongoDB/Redis/DynamoDB/Firestore 适配器），`configs/datastores/`（可选矩阵配置），`db/seed/`（可选种子数据）。
- **前端（Remix + React Router V7）**：`frontend/apps/web/app/`（routes, loaders, actions），`frontend/packages/ui/`（shadcn 主题），`frontend/packages/<feature>/`（热插拔组件）。
- **文档与自动化**：`docs/vitepress/`（内容 + `.vitepress/config.ts`）、`.github/workflows/docs.yml`（文档部署流水线）、`pnpm-workspace.yaml`（包含 docs 包）、`Makefile`（dev/test/build/release/docs/observe）。
- **基础设施**：`deploy/compose/*.yml`（Compose 环境）、`.env*` 模板、`scripts/`（如需的 helper）。
- 下文示例均以该结构为例，交付时请替换为真实路径。

<!-- 
  ============================================================================
  重要：以下任务仅作示例。
  /speckit.tasks 必须替换为基于以下信息生成的真实任务：
  - spec.md 中带优先级的用户故事（P1、P2、P3...）
  - plan.md 描述的功能需求
  - data-model.md 中的实体
  - contracts/ 中的端点
  任务必须按用户故事组织，以便：
  - 独立实现
  - 独立测试
  - 按 MVP 增量交付
  不要在最终 tasks.md 中保留这些示例。
  ============================================================================
-->

## Phase 1: Setup（共享基础设施）

**目的**：搭建 Go + Remix + Vitepress 双栈骨架、Compose 环境与 Makefile/Workflow 入口

- [ ] T001 在 `backend/` 初始化 Go module（`go.mod`）、`cmd/api/main.go` 与基础限界上下文目录结构
- [ ] T002 在 `frontend/` 通过 pnpm 初始化 workspace、创建 Remix 应用（`apps/web`）、shadcn/ui 基础包（`packages/ui`）以及 `docs/vitepress` workspace（含 `.vitepress/config.ts`）
- [ ] T003 [P] 在 `deploy/compose/docker-compose.local.yml` 中声明 API、前端、SQLite、MySQL、PostgreSQL、MongoDB、Redis、LocalStack（DynamoDB/Firestore）与 Vitepress 预览服务，并生成 `.env.example`
- [ ] T004 [P] 扩展根 `Makefile`：`dev`（Go/Remix/Vitepress）、`test`、`lint`、`compose-up/down`、`migrate/rollback`、`ui-test`、`data-matrix`、`docs-build`、`docs-deploy`、`coverage`
- [ ] T005 配置统一的 lint/格式化/测试/覆盖率工具与 PNPM scripts（`golangci-lint`, `gofumpt`, `go test ./... -coverprofile`, `pnpm lint`, `pnpm test --coverage`, `pnpm docs:lint`, `stylelint` 等），并在 CI 中通过 `make lint`、`make coverage`
- [ ] T006 [P] 创建/更新 `.github/workflows/ci.yml` 与 `.github/workflows/docs.yml`，确保调用 `make test`, `make data-matrix`, `make docs-build`, `make docs-deploy`
- [ ] T007 [P] 生成 `configs/datastores/*.yaml`（或等价配置）用于声明多存储矩阵及默认驱动，供 speckit 与运行时引用

---

## Phase 2: Foundational（阻塞性前置）

**目的**：准备限界上下文、GORM 迁移、Compose 服务与可观察性，解除后续故事阻塞

**⚠️ 关键**：在此阶段完成前，禁止开始用户故事任务。

示例（按项目调整）：

- [ ] T008 在 `backend/internal/<context>/` 中实现聚合、仓储接口以及 `service` 层，确保符合热插拔接口
- [ ] T009 [P] 在 `db/migrations/` 中新增 `<timestamp>_<feature>.sql|.go`，包含 up/down，并通过 `make migrate`、`make rollback` 验证
- [ ] T010 [P] 在 `deploy/compose/` 中为新增依赖（DB/Redis/队列等）添加服务、健康检查与 `.env` 条目
- [ ] T011 [P] 在 `Makefile` 中追加 `observe`, `constitution-check`, `feature-toggle` 等目标，并在 CI workflow 中引用
- [ ] T012 在 `backend/tests/contract/` 与 `backend/tests/integration/` 中补充测试骨架，覆盖限界上下文 API/事件
- [ ] T013 在 `frontend/packages/ui/` 定义 shadcn 主题 tokens（Apple Human Interface + Microsoft Fluent 配色/间距）
- [ ] T014 在 `frontend/apps/web/app/routes/` 中创建基础路由/loader/action，并将 UI 包以 pnpm workspace 链接

**检查点**：基础完备，可开始并行处理用户故事

---

## Phase 3: 用户故事 1 - [标题]（优先级：P1）🎯 MVP

**目标**：[该故事交付的能力]

**独立测试**：[如何单独验证该故事]

### 用户故事 1 的测试（可选）⚠️

> **注意：若包含测试，先编写并确保失败，再实现功能。**

- [ ] T015 [P] [US1] 在 `backend/tests/contract/<context>_contract_test.go` 中编写契约测试，并在 `backend/internal/<context>/domain/<aggregate>_test.go`/`infra/<adapter>_test.go` 中补充单元测试
- [ ] T016 [P] [US1] 在 `backend/tests/integration/<journey>_integration_test.go` 与 `frontend/apps/web/tests/<journey>.spec.ts`、`frontend/packages/ui/components/__tests__/<component>.spec.tsx` 中编写端到端 + UI 单元测试

### 用户故事 1 的实现

- [ ] T017 [P] [US1] 在 `backend/internal/<context>/domain/<aggregate>.go` 中创建聚合/值对象
- [ ] T018 [P] [US1] 在 `backend/internal/<context>/infra/<adapter>.go` 中实现 GORM 仓储/外部适配器
- [ ] T019 [US1] 在 `backend/internal/<context>/app/service.go` 中实现应用服务并注册热插拔接口
- [ ] T020 [US1] 在 `backend/cmd/api/handlers/<route>.go` 或 `frontend/apps/web/app/routes/<route>.tsx` 中暴露 HTTP/API/页面
- [ ] T021 [US1] 更新 `db/migrations/<timestamp>_<name>.sql|.go` 及 `deploy/compose/` 中相关服务配置
- [ ] T022 [US1] 在 `frontend/packages/ui/components/` 中构建/扩展 shadcn 组件并对齐 Apple/Microsoft 设计语言
- [ ] T023 [US1] 更新 `Makefile` 功能开关/自测目标并追加可观察性日志
- [ ] T024 [US1] 在 `docs/vitepress/<context>/<story>.md` 中记录接口、Compose/Make 命令与设计规范，并在 `.github/workflows/docs.yml` 中验证部署

**检查点**：故事 1 应可独立运行并测试

---

## Phase 4: 用户故事 2 - [标题]（优先级：P2）

**目标**：[该故事交付的能力]

**独立测试**：[如何单独验证]

### 用户故事 2 的测试（可选）⚠️

- [ ] T025 [P] [US2] 在 `backend/tests/contract/<context>_contract_test.go` 中补充契约测试，并在 `backend/internal/<context>/domain/<aggregate>_test.go`/`infra/<adapter>_test.go` 中覆盖单元测试
- [ ] T026 [P] [US2] 在 `frontend/apps/web/tests/<journey>.spec.ts` 与 `frontend/packages/ui/components/__tests__/<component>.spec.tsx` 中实现用户旅程集成 + UI 单元测试

### 用户故事 2 的实现

- [ ] T027 [P] [US2] 在 `backend/internal/<context>/domain/<aggregate>.go` 中建模实体
- [ ] T028 [US2] 在 `backend/internal/<context>/app/<usecase>.go` 中实现用例逻辑并通过接口注册
- [ ] T029 [US2] 在 `frontend/apps/web/app/routes/<route>.tsx` 中创建 Remix 路由/loader/action，与 React Router V7 data APIs 对齐
- [ ] T030 [US2] 在 `frontend/packages/ui/components/<component>.tsx` 中扩展 shadcn 组件或主题 token
- [ ] T031 [US2] 在 `deploy/compose/docker-compose.local.yml` 中为该故事依赖的外部服务添加配置并更新 `.env`
- [ ] T032 [US2] （如需）与用户故事 1 的热插拔模块集成，同时保持可独立启停
- [ ] T033 [US2] 在 `docs/vitepress/<context>/<story>.md` 中补充 API、UI、数据库驱动及 `make docs-*`、`.github/workflows/docs.yml` 的验证结果

**检查点**：故事 1 与 2 均可独立运行

---

## Phase 5: 用户故事 3 - [标题]（优先级：P3）

**目标**：[该故事交付的能力]

**独立测试**：[如何单独验证]

### 用户故事 3 的测试（可选）⚠️

- [ ] T034 [P] [US3] 在 `backend/tests/contract/<context>_contract_test.go` 中编写契约测试，并在 `backend/internal/<context>/domain/<aggregate>_test.go`/`infra/<adapter>_test.go` 中补充单元测试
- [ ] T035 [P] [US3] 在 `frontend/apps/web/tests/<journey>.spec.ts` 与 `frontend/packages/ui/components/__tests__/<component>.spec.tsx` 中编写集成 + UI 单元测试

### 用户故事 3 的实现

- [ ] T036 [P] [US3] 在 `backend/internal/<context>/domain/<aggregate>.go` 中建模实体
- [ ] T037 [US3] 在 `backend/internal/<context>/app/<workflow>.go` 中实现服务，并在 `backend/cmd/api/handlers/` 或 gRPC/事件层暴露
- [ ] T038 [US3] 在 `frontend/apps/web/app/routes/<route>.tsx` 中实现页面/loader/action，并在 `frontend/packages/ui/` 中扩展组件
- [ ] T039 [US3] 更新 `db/migrations/`、`Makefile`、`deploy/compose/` 中的增量配置，确保该模块可单独启停
- [ ] T040 [US3] 在 `docs/vitepress/<context>/<story>.md` 中更新发布说明、数据库切换指南及 GitHub Workflow 验证结果

**检查点**：所有用户故事可独立运行

---

[如需更多故事，按同样模式扩展]

---

## Phase N: 抛光与跨领域事项

**目的**：影响多个用户故事的改进

- [ ] TXXX [P] 更新 docs/ 文档
- [ ] TXXX 代码清理与重构
- [ ] TXXX 全局性能优化
- [ ] TXXX [P] 如需新增的单元测试（tests/unit/）
- [ ] TXXX 安全加固
- [ ] TXXX 验证 quickstart.md 场景

---

## 覆盖率 Gate（所有故事完成后执行）

- [ ] T041 运行 `make test`, `make ui-test`, `make docs-test`, `make data-matrix`, `make coverage`，确保综合覆盖率 ≥85%，并上传覆盖率报告/构建工件
- [ ] T042 在 `docs/vitepress/status/coverage.md`、README badge 或 CI 注释中更新覆盖率、测试结果及 `make coverage` 命令输出链接

---

## 依赖与执行顺序

### 阶段依赖

- **Setup（Phase 1）**：无依赖，可立即开始
- **Foundational（Phase 2）**：依赖 Setup，阻塞所有用户故事
- **用户故事（Phase 3+）**：均依赖 Foundational 完成
  - 若人手充足，可并行
  - 否则按优先级顺序（P1→P2→P3）串行
- **Polish（最终阶段）**：依赖所有目标用户故事完成

### 用户故事依赖

- **US1 (P1)**：Foundational 后即可开始，无其他故事依赖
- **US2 (P2)**：Foundational 后即可开始，若需与 US1 集成也应保持可独立测试
- **US3 (P3)**：同上

### 单个故事内的顺序

- 若有测试，必须“先写测试并看到失败”
- 模型 → 服务 → 端点
- 核心实现完成后再做整合
- 完成一个故事后再进入下一优先级

### 可并行机会

- 所有标记 [P] 的 Setup/Foundational 任务
- Foundational 完成后，各用户故事可并行
- 同一故事内标记 [P] 的测试/模型可并行
- 不同用户故事可由不同成员并行推进

---

## 并行示例：用户故事 1

```bash
# 若需要测试，可同时启动：
Task: "在 backend/tests/contract/<context>_contract_test.go 编写契约测试"
Task: "在 backend/tests/integration/<journey>_integration_test.go 或 frontend/apps/web/tests/<journey>.spec.ts 编写集成测试"

# 可同时创建的模型：
Task: "在 backend/internal/<context>/domain/<aggregate>.go 定义聚合"
Task: "在 backend/internal/<context>/infra/<adapter>.go 实现 GORM 仓储"
```

---

## 实施策略

### MVP 优先（仅交付故事 1）

1. 完成 Phase 1
2. 完成 Phase 2（⚠️ 阻塞）
3. 完成 Phase 3（US1）
4. **暂停并验证**：独立测试 US1
5. 若准备好，可部署/演示

### 渐进式交付

1. Setup + Foundational → 基础完成
2. 加入 US1 → 独立测试 → 部署/演示（MVP）
3. 加入 US2 → 独立测试 → 部署/演示
4. 加入 US3 → 独立测试 → 部署/演示
5. 每个故事都能带来增量价值且不破坏前面成果

### 并行团队策略

1. 团队协作完成 Setup + Foundational
2. 基础完成后：
   - 开发者 A：US1
   - 开发者 B：US2
   - 开发者 C：US3
3. 各故事独立完成并集成

---

## 备注

- 标记 [P] 表示不同文件、无依赖，可并行
- [Story] 标签便于溯源到具体用户故事
- 每个故事都应可独立完成与测试
- 若写测试，必须先失败后实现
- 建议每个任务或逻辑组完成后提交
- 任意检查点都可暂停做独立验证
- 避免模糊任务、同文件冲突、跨故事依赖打破独立性
