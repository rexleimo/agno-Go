# 任务清单：Agno 核心模块 Go 迁移

**输入**：`/specs/001-migrate-agno-core/`（plan.md、spec.md、research.md、data-model.md、contracts/、quickstart.md）  
**前置**：完成 `.specify/scripts/bash/check-prerequisites.sh` 检查并获取 Python 参考实现 `agno/libs/agno/agno/`  
**测试**：Go 端运行 `go test ./... -cover`（含 parity/benchmark 套件）；Python 端运行 `python -m agno.tests.contracts.run`；跨语言脚本使用 `scripts/ci/cross-language-parity.sh`  
**覆盖率**：所有新增 Go 代码需在 `go test ./... -cover` 下达到 ≥85%，并通过 `scripts/go-ci.sh` 的 gate  
**栈约束**：实现语言为 Go 1.25.1，Python ≥3.11 仅作对照；禁止在 Go runtime 中调用 Python  
**组织方式**：Phase 1-2 为共享基建，Phase 3+ 分别对应 spec.md 的 US1(P1)/US2(P2)/US3(P3)，最终阶段用于 Polish

- [X] T001 更新 `.specify/scripts/bash/check-prerequisites.sh`，新增 Go 1.25.1 与 Python ≥3.11 版本校验以及缺失时报错输出，确保 parity 与 benchmark 环境一致。（已确认脚本当前实现包含版本检测并输出错误信息）
- [X] T002 [P] 新建 `specs/001-migrate-agno-core/fixtures/README.md`，按照 `data-model.md` 的 ParityFixture 字段定义目录结构与示例 JSON/YAML 片段。
- [X] T003 [P] 脚手架 `scripts/ci/cross-language-parity.sh`，实现 `--fixture/--python/--go` 参数解析与基础 JSON diff 占位输出，为后续接入 Go/Python CLI 做准备。
- [X] T004 [P] 创建 `scripts/benchmarks/README.md` 与 `scripts/benchmarks/data/.gitkeep`，说明 Python baseline/Go 运行结果的产物格式与收集命令。

## Phase 2: Foundational（阻塞性前置）

- [X] T005 在 `go/agent/runtime_manifest.go` 定义 `AgentRuntime` 配置结构与构造校验逻辑，对齐 `data-model.md` AgentRuntime 字段并校验 memory/session/toolkits 互斥条件。
- [X] T006 [P] 在 `go/workflow/run_state.go` 建模 `WorkflowRun`、Step 状态与 routing metadata，确保可序列化字段与 `data-model.md` WorkflowRun 匹配。
- [X] T007 [P] 在 `go/session/store.go` 暴露 `Store` 接口、内存实现与 Sqlite/Postgres/Redis 钩子，落实 `research.md` 的驱动可插拔决策。
- [X] T008 [P] 新建 `go/internal/testutil/parity/fixtures.go`，实现 ParityFixture 加载、随机种子注入与 diff 辅助函数，供 Go 端 parity 测试复用。
- [X] T009 [P] 在 `go/providers/manifest_loader.go` 解析 Toolkit/Knowledge/Memory Provider manifest（YAML/JSON），填充 `Provider`/`Capability` 数据并暴露 Guardrail 接口占位。

## Phase 3: 用户故事 1 - 平台工程师在 Go Runtime 中复用既有 Agent（优先级：P1）

**Story goal**：在纯 Go 环境注册/运行 Agent 与 Workflow，实现与 `agno.libs.agno.agno.agent.Agent` 行为一致的 runtime。  
**Independent testing**：对 cookbook 的单 Agent、团队、多步骤 workflow 三个场景分别运行 `scripts/ci/cross-language-parity.sh --fixture specs/001-migrate-agno-core/fixtures/us1_basic_coordination.yaml`，比较 RunOutput/Session state/Metrics。  
**Tests**：`go test ./go/agent -run TestParity`、`python -m agno.tests.contracts.run --scenario us1_basic_coordination`

- [X] T010 [US1] 在 `go/agent/runtime_service.go` 实现 AgentRuntime 注册 API，将 `agno/libs/agno/agno/agent/agent.py` 的配置映射到 `AgentRuntime` 并将 Session 策略写入 `go/session/store.go`。
- [X] T011 [P] [US1] 扩充 `go/workflow/runner.go`，实现顺序/团队协作 orchestration，与 `agno/libs/agno/agno/workflow/workflow.py` 的步骤/路由保持一致。
- [X] T012 [US1] 新建 `go/session/session_record_mapper.go`，将 Go 侧 `SessionRecord` 读写映射到 `agno/libs/agno/agno/session/agent.py` & `session/workflow.py` 的字段，包括 history/summary/metrics。
- [X] T013 [P] [US1] 在 `go/providers/toolchain_registry.go` 定义 Toolkit/Knowledge/Guardrail 接口，实现从 manifest 到 Go runtime 的注入，并对 `not_migrated` provider 抛出 `internal/errors.CodeNotMigrated` 的同时返回推荐 fallback/provider 与迁移文档链接。
- [X] T014 [P] [US1] 编写 `specs/001-migrate-agno-core/fixtures/us1_basic_coordination.yaml`，覆盖单 Agent、团队、workflow 三种输入/工具响应，并引用 cookbook 的 provider/agent ID。
- [X] T015 [P] [US1] 扩展 `go/agent/us1_parity_config_test.go`，通过 `go/internal/testutil/parity` 加载 fixture，调用 Go API，比较结果与 Python CLI 输出。
- [X] T016 [US1] 完成 `scripts/ci/cross-language-parity.sh` 正式逻辑：调用 `python -m agno.tests.contracts.run` 与 `go test ./go/... -run TestParity`，生成 `scripts/ci/.cache/parity_results.json` 及 diff。
- [X] T017 [US1] 新建 `go/agent/runtime_service_contract_test.go`，依据 `specs/001-migrate-agno-core/contracts/runtime-openapi.yaml` 校验 `/v1/runtime/go/agents` 与 `/v1/runtime/go/workflows/{workflowId}:run` 请求/响应结构。

**Parallel example（US1）**：T013（toolchain registry）与 T014（fixture authoring）互不依赖，可并行推进以加速 parity 验证准备。

## Phase 4: 用户故事 2 - 可靠性工程师验证性能与 Telemetry（优先级：P2）

**Story goal**：提供与 Python 等价的 Telemetry 事件、性能指标与监控入口，满足 P95 延迟和 RSS 目标。  
**Independent testing**：运行 `scripts/benchmarks/collect_runtime_baselines.sh --workflow us1_basic_coordination` 及 `curl /v1/runtime/go/telemetry/events`，比较 Python/Go 事件序列与性能曲线。  
**Tests**：`go test ./go/internal/telemetry -run TestTelemetryParity`、`go test ./go/agent -run TestUS1Bench -bench .`

- [X] T018 [US2] 扩展 `go/internal/telemetry/telemetry.go`，新增 RunStarted/ReasoningStep/ToolCall/SessionSummary/RunCompleted 常量与 runtime=go 标签校验，覆盖 `TelemetryEnvelope` 字段。
- [X] T019 [P] [US2] 在 `go/internal/telemetry/recorder.go` 实现事件 Recorder，将 payload 写入结构化存储并附带 tokens/latency 统计，同时对未知事件类型回退为 `unknown_event`（记录 Runtime 版本、不打断对话），供 Session/Workflow 注入。
- [X] T020 [US2] 新建 `go/internal/telemetry/http_handlers.go`，实现 `/v1/runtime/go/telemetry/events` 过滤（sessionId/workflowRunId/runtime）并与 `contracts/runtime-openapi.yaml` schema 对齐。
- [X] T021 [P] [US2] 更新 `go/agent/us1_basic_coordination_bench_test.go`，将基准扩展为可配置并发（含 100 并发）并记录 latency/RSS/CPU/tokens/sec，对比 Python baseline 时断言 p95 延迟 ≤70%、平均 RSS 改善 ≥25%、100 并发 CPU <75%、tokens/sec 偏差 ≤10%。
- [X] T022 [P] [US2] 实现 `scripts/benchmarks/collect_runtime_baselines.sh`，串联 `python agno/tests/benchmarks/run.py` 与 `GOCPU=75 go test ./go/agent -run TestUS1Bench -bench .`，抓取系统指标写入 `scripts/benchmarks/data/us1_basic_coordination.json`（含 latency/RSS/CPU/tokens/sec），并在偏离 >10% 时生成调优指引记录于 `specs/001-migrate-agno-core/checklists/perf.md`。
- [X] T023 [US2] 新建 `go/internal/telemetry/telemetry_parity_test.go`，读取 Python 日志（`scripts/parity_stats.sh` 输出）并验证事件序列与 payload 键值一致，同时注入未知事件样本以覆盖 `unknown_event` fallback 行为。

**Parallel example（US2）**：T021（Go benchmark）与 T022（baseline脚本）分别操作 Go 测试与 shell 脚本，可在不同人员下同时进行。

## Phase 5: 用户故事 3 - 生态开发者理解双语栈差异（优先级：P3）

**Story goal**：在 README/Cookbook/示例中提供全面的 Go Runtime Quickstart、支持矩阵与迁移 FAQ，帮助生态开发者自助上手。  
**Independent testing**：审阅 `agno/README.md`、`agno/cookbook/README.md` 与 `go/examples/us1_basic_coordination/README.md`，确认 10 分钟内可完成 Go Quickstart 并识别未迁移模块。  
**Tests**：文档校对（无自动化），review checklist 记录在 `specs/001-migrate-agno-core/checklists/docs.md`

- [X] T024 [US3] 更新 `agno/README.md`，新增 “Go Runtime Quickstart / 迁移指南” 章节并链接 `scripts/ci/cross-language-parity.sh`、`go/examples/us1_basic_coordination`。
- [X] T025 [P] [US3] 更新 `agno/cookbook/README.md` 与相关场景小节，标记 Python-only/Go-supported 功能，并引用 SupportMatrix。
- [X] T026 [P] [US3] 为 `go/examples/us1_basic_coordination/README.md` 添加运行命令、配置示例、Telemetry/benchmark 阅读指南。
- [X] T027 [US3] 新建 `agno/docs/go-runtime-support-matrix.md`，以 `contracts/runtime-openapi.yaml` 的 SupportMatrix schema 填充特性可用性、限制、planned 说明，并在 README 中链接。

**Parallel example（US3）**：T025（Cookbook 更新）与 T026（示例 README）修改不同路径，可同时完成，再由 T024 汇总入口。

## Phase 6: Pilot Deployment & Support（SC-003）

**Story goal**：在 Go-only 环境中让两支内部团队完成端到端部署，记录 ≤1 小时迁移流程并把周度支持票控制在 ≤2 起，使 SC-003 可验证。  
**Independent testing**：审阅 `specs/001-migrate-agno-core/logs/pilot-*.md` 与 `scripts/support-metrics/export_support_stats.sh` 输出，确认各团队迁移时间、无 Python 依赖与支持票数量。  
**Tests**：`scripts/support-metrics/export_support_stats.sh --since 14d`、pilot checklist 评审

- [X] T031 [Pilot] 在 `specs/001-migrate-agno-core/checklists/pilot.md` 定义试点流程（涉及团队、环境、CI 链路、验收步骤）并附上用于计时/记录的模板。
- [X] T032 [P] [Pilot] 协调 AgentOS Core 与 Solutions Automation 执行无 Python 运行的迁移演练，生成 `specs/001-migrate-agno-core/logs/pilot-{team}.md`，记录完成时间与遗留问题。
- [X] T033 [P] [Pilot] 实现 `scripts/support-metrics/export_support_stats.sh`，按周拉取/汇总支持票并在 >2 起时生成告警（写入 `specs/001-migrate-agno-core/logs/support-alerts.md`），作为发布 gate。

## Final Phase: Polish & Cross-Cutting Concerns

- [X] T028 扩展 `scripts/go-ci.sh`，串联 `go test ./... -coverprofile=/tmp/coverage.out` 与 `scripts/ci/cross-language-parity.sh`, 并在覆盖率 <85% 或 parity diff 时失败。
- [X] T029 [P] 更新 `scripts/parity_stats.sh`，解析 `scripts/ci/.cache/parity_results.json` 并生成按 fixture 聚合的通过率，供发布前审查。
- [X] T030 [P] 在 `specs/001-migrate-agno-core/checklists/release.md` 记录 Telemetry/性能/文档签核步骤与已知 not_migrated 列表，确保交付透明。

## Dependencies

### 用户故事顺序图
```
Phase1 Setup ─▶ Phase2 Foundational ─▶ US1 (P1) ─▶ US2 (P2) ─▶ US3 (P3) ─▶ Pilot ─▶ Polish
```

- US1 依赖 Setup/Foundational 完成 AgentRuntime/Session/fixture 基座。
- US2 依赖 US1 的运行路径与 parity harness 才能填充 Telemetry 与 benchmark。
- US3 依赖 US1+US2 的行为结论来撰写文档与支持矩阵。
- Pilot 阶段依赖 US3 的文档与工具链来指导内测团队，验证无 Python 部署与支持票指标。
- Polish 阶段依赖全部用户故事与 Pilot 完成，以便统一执行 CI gate 与发布清单。

## Implementation strategy（先 MVP，分步交付）

1. **MVP（US1）**：完成 Phase 1→Phase 2→Phase 3，验证 Go runtime 可运行并通过 parity（交付最小可用版本）。
2. **强韧化（US2）**：在 MVP 稳定后接入 Telemetry/性能脚本，确保监控指标达到 spec 的 P95/RSS/CPU/tokens/sec 目标。
3. **生态扩展（US3）**：补齐 README/Cookbook/Support Matrix，让外部开发者可自助迁移。
4. **试点部署（Pilot）**：驱动两支内部团队在无 Python 环境完成部署，记录 1 小时迁移流程与支持票指标，满足 SC-003。
5. **收尾**：执行 Final Phase 任务，将 CI gate、parity 报告、性能/支持门槛固化，形成可重复交付流程。
