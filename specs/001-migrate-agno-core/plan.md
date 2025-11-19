# 实现计划：Agno 核心模块 Go 迁移

**分支**：`001-migrate-agno-core` | **日期**： 2025-11-19 | **规格**： /home/rex/codes/agno-Go/specs/001-migrate-agno-core/spec.md
**输入**：来自 `/specs/001-migrate-agno-core/spec.md` 的功能规格

**说明**：该模板由 `/speckit.plan` 命令填充。执行流程见 `.specify/templates/commands/plan.md`。

## 摘要

本迭代将把 Python 版本的 agno 核心运行时（Agent、Workflow、Session、Telemetry 与扩展点）迁移为纯 Go 实现，确保在运行期完全脱离 Python 依赖，同时保持与 `agno/libs/agno/agno/` 提供的行为严格一致。关键路径包括：为 Agent/Workflow/Session 建立可序列化的数据模型与包结构、补齐工具链/知识/Guardrail 接口、实现跨语言对照测试与性能基线、并同步 README/Cookbook/Quickstart 以指导用户在 Go 环境落地。Deliverables 覆盖 parity 测试脚本、Go API 设计、docs 与性能衡量方法，使后续实现工作可以直接按 plan 展开。

## 技术背景

<!--
  必须：将此处内容替换为该项目的技术细节。
  结构仅为建议，可根据迭代需要调整。
-->

**语言/版本**：Go 1.25.1（实现语言）+ Python ≥3.11（行为基线，与脚本要求一致）  
**主要依赖**：Go 标准库 + 现有 `go/internal/errors`、`go/internal/telemetry`、`go/providers` 等包；对照脚本将复用 `agno/libs/agno` 中的 asyncio/pydantic 依赖作为参考实现  
**存储**：沿用 Python 侧的数据库/缓存约定（Sqlite、Postgres、Redis）；Go 侧通过接口封装 SessionState 与 MemoryManager，不在本迭代内绑定具体驱动  
**测试**：Go 侧使用 `go test ./...`（含 parity/benchmark 套件、≥85% 覆盖率目标），Python 侧继续使用 pytest/内置单测；`scripts/ci/cross-language-parity.sh` 将 orchestrate 双端运行  
**目标平台**：Linux/macOS 服务器与本地开发环境；Go 模块需可在容器化环境中运行，无 UI 依赖  
**项目类型**：多模块库（library）；Go 包向 SDK/API 用户暴露 Agent、Workflow、Session、Tools、Providers 等抽象  
**性能目标**：实现规格中的成功标准：Go 运行时需在关键场景中实现 ≥99% 输出一致率、p95 延迟 ≤ Python 的 70%、100 并发下 CPU <75%、平均 RSS 至少降低 25%，并记录 tokens/sec；`scripts/benchmarks` + `go/agent/us1_basic_coordination_bench_test.go` 将联动触发 100 并发负载、采集 CPU/内存/tokens/sec，并在偏离 Python >10% 时输出可行的调优建议
**约束条件**：运行期禁止调用 Python；所有配置/Telemetry 与 Python 兼容；跨语言行为差异需记录；docs/Quickstart 必须覆盖 Go Runtime；需要 runtime=go 标签；Go 代码覆盖率 ≥85%  
**规模/范围**：覆盖 cookbook 中单 Agent / 团队 / Workflow 代表场景，面向 2+ 内部团队的生产部署与社区 Quickstart；Go 代码当前约几十个文件，目标构建完整核心运行时骨架

## 宪章检查

*Gate：Phase 0 调研前必须通过；Phase 1 设计后需再次检查。*

- [x] **Python 参考实现与行为对齐**：本功能以 `/home/rex/codes/agno-Go/agno/libs/agno/agno/agent`, `workflow`, `session`, `run`, `tools`, `knowledge`, `memory` 及其 cookbook 示例为唯一行为来源。Go 设计需映射 Agent.run、Workflow.run、SessionManager、RunMessages、Telemetry events，所有允许差异（Go 错误类型、context 超时）将在需求/文档中记录并配对测试。
- [x] **Go 原生实现与运行时独立**：Go 交付物位于 `/home/rex/codes/agno-Go/go/{agent,workflow,session,providers,internal}`，通过 Go modules 构建；运行期不得 `os/exec` Python。唯一允许的 Python 依赖位于 `scripts/ci/cross-language-parity.sh` 与 specs/contracts，用于对照与 fixture capture，并要求在 CI 中可独立清理。
- [x] **Go API 设计与包结构**：目标包包括 `go/agent`（Agent 配置+执行器）、`go/workflow`（步骤/并发控制）、`go/session`（持久化）、`go/providers`（LLM/tool registry）、`go/internal`（errors/telemetry），API 遵循 `context.Context`、显式 error、不可变配置结构体，避免 Python 风格的动态 kwargs。
- [x] **跨语言测试纪律与 85%+ 覆盖率**：需要扩展 `go/agent/us1_parity_config_test.go`、`go/providers/us3_custom_provider_parity_test.go` 等 parity 测试；新增对话 fixture（YAML/JSON）在 `scripts/ci/cross-language-parity.sh` 中运行 Python & Go，比较 RunOutput、Metrics、Events；所有新增 Go 代码需在 `go test ./... -cover` 下达到 ≥85%，并提供基准测试（`go/agent/us1_basic_coordination_bench_test.go` 扩展）。
- [x] **性能与资源使用基线**：在 `scripts/benchmarks/`（新建）记录 Python baseline 数据（延迟、RSS、CPU）；Go 版本需运行相同 workloads（单 Agent / 团队 / workflow）并满足 spec 中 p95 延迟 ≤70% Python、RSS 下降 ≥25% 的目标，偏离需在计划/文档中列出 mitigation。
- [x] **安全、配置与 Telemetry**：涉及的配置包括模型 provider keys、数据库连接、Telemetry sinks；Go 侧延续 Python 的 env key 约定，并在 `go/internal/telemetry` 中确保脱敏与 runtime=go 标签。所有 secrets 通过调用方注入，不写入 repo；错误分类需与 Python 的 `ModelProviderError/InputCheckError` 映射。
- [x] **文档与示例对齐**：需更新 `/home/rex/codes/agno-Go/agno/README.md`、`agno/cookbook/*`、`go/examples/us1_basic_coordination` 及新建 quickstart，列出 Go Runtime 支持矩阵、限制与迁移指南；docs 中必须指向 issue 模板反馈缺失特性。

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
```text
/home/rex/codes/agno-Go/
├── agno/
│   ├── libs/agno/agno/                # Python 行为基线（agent/workflow/session/...）
│   └── cookbook/                      # 官方示例，作为 parity 场景来源
├── go/
│   ├── agent/                         # 当前 Go Agent API + US1 示例与测试
│   ├── workflow/                      # Workflow 核心与 not_migrated 占位
│   ├── session/                       # Session 数据结构与 US1 demo
│   ├── providers/                     # Provider 注册、告警与 parity tests
│   ├── internal/{errors,telemetry,testutil}
│   └── examples/us1_basic_coordination # Go Quickstart 示例起点
├── scripts/
│   └── ci/                            # 将存放 cross-language parity & benchmark 脚本
├── specs/
│   ├── 001-migrate-agno-agents/       # 既有规格
│   └── 001-migrate-agno-core/         # 本迭代的 spec/plan/research 等 artefacts
└── .specify/                          # speckit 配置与模板脚本
```

**结构决策**：Go 代码集中在 `/home/rex/codes/agno-Go/go` 子树，遵循模块化包结构；文档与规划统一放在 `/home/rex/codes/agno-Go/specs/001-migrate-agno-core`。本计划将扩展 `scripts/ci`（对照测试）、`go/examples`（Quickstart）以及现有包中的数据模型与契约文件，确保未来实现落地时文件归属清晰、依赖方向保持内聚。

## 性能采集与并发验证策略

- `go/agent/us1_basic_coordination_bench_test.go` 将扩展为可配置并发（含 100 并发）并输出 latency/RSS/CPU/tokens/sec，结合 `scripts/benchmarks/collect_runtime_baselines.sh` 写入 `scripts/benchmarks/data/*.json`，作为 SC-002/FR-007 的唯一真源。
- `scripts/go-ci.sh` 在 Final Phase 中会串联上述基准，若 CPU ≥75%、tokens/sec 偏差 >10% 或未生成并发样本，则直接失败并提示调优步骤/issue 链接。
- 偏离 Python 基线时，计划要求同步更新 `specs/001-migrate-agno-core/checklists/perf.md`，记录调优实验、结果和 workaround，保证调优建议对齐宪章的性能基线原则。

## Pilot Deployment & Support（SC-003）

- 选择两支内部团队（AgentOS Core、Solutions Automation）作为无 Python 运行的试点，在计划阶段准备 `specs/001-migrate-agno-core/checklists/pilot.md`，列出环境约束、CI 链路和联络人。
- 每支团队需完成一次端到端迁移演练（含 cookbook 场景）并记录所需时间，要求 1 小时内走通；演练日志存档在 `specs/001-migrate-agno-core/logs/pilot-*.md`。
- 支持票统计通过新增的 `scripts/support-metrics/export_support_stats.sh`（任务中定义）拉取工单量，若任意周 >2 起则阻断发布并在 tasks 中标记 BLOCKED，直到补救完成。

## 复杂度追踪

> **仅当宪章检查存在违规且必须说明时填写**

| 违规项 | 必要原因 | 更简单方案被拒绝的理由 |
|--------|----------|--------------------------|
| （暂无） | 当前方案未出现违背宪章的强制需求 | 如未来为性能或外部依赖引入例外，将在此记录理由与替代方案 |
