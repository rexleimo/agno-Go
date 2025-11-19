# 功能规格：Agno 核心模块 Go 迁移

**功能分支**：`001-migrate-agno-core`  
**创建时间**：2025-11-19  
**状态**：草稿  
**输入**：用户描述：“实现迁移 python 实现的agno 核心模块,纯golang语言实现”

## 用户场景与测试 *(必填)*

### 用户故事 1 - 平台工程师在 Go Runtime 中复用既有 Agent（优先级：P1）

作为负责 AgentOS 部署的平台注册工程师，我希望不需要安装 Python 运行时即可在 Go 语言环境中实例化 `agno.agent.Agent` 级别的能力，复用现有工作流配置与工具链。

**优先级原因**：直接影响核心用户对 Agno 的采纳速度，也是所有后续迁移与商业交付的前置条件。

**独立测试**：针对 cookbook 中三个示例（单 Agent、团队、多步骤 workflow）分别运行 Python 与 Go 版本，比较 RunOutput、Session state 和 Metrics 是否一致。

**验收场景**：

1. **Given** 现有 `agno/libs/agno/agno/agent/agent.py` 配置文件与工具集，**When** 使用 `go/agent` 包提供的 API 初始化代理并执行一次对话，**Then** 输出消息顺序、工具调用次数与 Python 版本一致且无需 Python 环境。
2. **Given** `agno/libs/agno/agno/workflow` 中定义的多步骤 workflow，**When** 通过 `go/workflow` orchestration API 运行，**Then** 每个步骤的上下文在 Session 存储中保持与 Python 模块相同的键名与值语义。

---

### 用户故事 2 - 可靠性工程师验证性能与 Telemetry（优先级：P2）

作为可靠性工程师，我需要在 Go 版本中看到与 Python 版本等价的运行事件、指标与告警阈值，并确保团队能够在 95% 的流量下获得更低的延迟与资源占用。

**优先级原因**：迁移若缺乏可观测性与性能优势，将无法上线生产流量。

**独立测试**：运行 `go/agent/us1_basic_coordination_bench_test.go` 和新增的对照测试，收集事件并对比 `agno/libs/agno/agno/run` 中 Python 事件序列，验证指标与告警映射。

**验收场景**：

1. **Given** 统一的 Telemetry schema（RunStarted、ToolCall、SessionSummary），**When** 在 Go runtime 中触发 100 次对话，**Then** 事件被写入与 Python 相同的 sink，并附带新的 runtime 标签用于区分来源。
2. **Given** Python 版本基准内存占用记录，**When** 在相同数据量下运行 Go 版本，**Then** 平均 RSS 降低至少 25%，并能通过监控仪表板查看趋势。

---

### 用户故事 3 - 生态开发者理解双语栈差异（优先级：P3）

作为基于 Agno 开发插件与示例的生态开发者，我需要在 README / Cookbook 中看到清晰的 Go 支持矩阵、迁移指南与限制提示，以便快速判断如何贡献代码。

**优先级原因**：虽然不阻塞初始交付，但缺少文档会导致社区无法复用与贡献。

**独立测试**：审阅更新后的 `agno/README.md` 与 `go/examples` 目录，确认是否提供 Go 版本的 Quickstart、差异清单与 issue 上报指引。

**验收场景**：

1. **Given** 旧版文档仅描述 Python 栈，**When** 查看新版 README、Cookbook 以及 `go/examples/us1_basic_coordination`，**Then** 可以在 10 分钟内完成基础 Go Agent 的运行并理解与 Python 的差异。
2. **Given** 开发者查看迁移 FAQ，**When** 搜索特性（例如 Guardrails、Knowledge graph）时，**Then** 可以了解到 Go 端可用性、限制与预计上线时间。

---

### 边界场景

- 当用户启用尚未迁移的 Python-only 模块（如特定第三方集成）时，Go 运行时需要在初始化阶段显式报错并给出 fallback 建议，而不是静默失败。
- 系统在 run 过程中遇到 Python 参考实现中未定义的新事件类型时，必须将其记录为 “unknown event” Telemetry 并不中断对话，同时在日志中标注版本。
- 迁移范围不覆盖前端 UI、AgentOS FastAPI 服务端 nor SaaS 控制面板，这些保持现状，仅替换内部 execution runtime。

## 需求 *(必填)*

### 功能需求

- **FR-001**：系统必须在 `go/agent` 包中提供与 `agno.libs.agno.agno.agent.Agent` 等价的构造与配置能力（模型、工具、记忆、会话参数），并保证所有必填字段有默认值以支撑最小示例。
- **FR-002**：系统必须在 `go/workflow`、`go/session` 包中提供 run orchestration、SessionState、Summary 与历史检索 API，使多步骤 workflow 与团队模式的行为与 `agno/libs/agno/agno/workflow`、`agno/libs/agno/agno/team` 对齐。
- **FR-003**：运行期必须写入/读取兼容 `agno.libs.agno.agno.session` 约定的数据结构，包括 SessionState、RunMessages、Metrics，确保数据库 schema 与事件 key 名完全一致。
- **FR-004**：系统必须在 Go 中实现工具链、Guardrails、知识检索等扩展点的接口，使 Python 侧的 Toolkit、KnowledgeFilter、MemoryManager 可通过配置映射到 Go 版本，迁移文档需指明差异。
- **FR-005**：所有 Go Runtime 的 RunEvent、Telemetry、日志必须覆盖 RunStarted、ReasoningStep、ToolCall、SessionSummary、RunCompleted 五大事件，并提供 runtime=go 标签帮助平台分流监控。
- **FR-006**：需要提供自动化对照测试套件，能够以 YAML/JSON 形式描述输入，分别调用 Python 与 Go 版本运行后比较 RunOutput 差异，作为 `scripts/ci/cross-language-parity.sh` 的一部分。
- **FR-007**：Go 版本需要提供资源与性能基准（CPU、内存、tokens/sec），并在偏离 Python 参考基线 10% 以上时给出调优指引或 explicit limitation 记录在规格与文档中。
- **FR-008**：文档与示例（`agno/README.md`、`cookbook`、`go/examples`）必须新增 “Go Runtime” 章节，涵盖安装、API 摘要、迁移步骤、限制与 issue 模板链接。

### 关键实体 *(若涉及数据则必填)*

- **AgentRuntime**：抽象单个 Agent 的配置、工具、模型引用与运行时回调。关键属性包含 Name、ModelConfig、Toolkits、SessionPolicy，与 SessionRecord 形成 1:N 关系。
- **WorkflowRun**：描述多步骤或团队运行实例，追踪每个 Step 的输入、输出与 Reasoning metadata，关联多个 RunEvents 与 TelemetryEnvelopes。
- **SessionRecord**：存储跨运行共享的数据（State、Summary、History、Metrics），既可以来自 Python 旧表也可以由 Go runtime 写入，需保持键值兼容。
- **TelemetryEnvelope**：包装 RunEvent、性能计数与安全告警，字段包含 EventType、Timestamp、PayloadHash、RuntimeSource，用于跨语言仪表盘。

## 假设

- Python 对照版本以 `agno/libs/agno` 主分支最新 commit 为准，迁移期间不考虑尚未发布的破坏性改动。
- 目标数据库与队列与现有 Python runtime 相同（SQLite / Postgres / Redis 任选一），因此可沿用既有连接配置。
- 所有用户仍通过 AgentOS 或 CLI 触发运行，本次迭代不新增新的触发渠道。

## 成功标准 *(必填)*

### 可度量结果

- **SC-001**：在三类代表性工作流中（单 Agent、团队、Workflow），Go 运行时的 RunOutput 与 Python 结果一致率达到 ≥99%（按消息文本与工具结果对比），任何差异均记录为 issue。
- **SC-002**：95% 的对话步骤在 Go runtime 中的端到端延迟 ≤ Python 版本的 70%，并且在 100 并发下保持 CPU 使用率 < 75%。
- **SC-003**：至少两家内部团队在生产环境中无需安装 Python 即可完成部署，且迁移指导流程在 1 小时内可走通，支持票数量 ≤ 2 起/周。
- **SC-004**：社区文档更新后，随机抽样 10 名开发者可以在 15 分钟内完成 Go Quickstart，自评满意度 ≥ 4/5。

## 宪章对齐 *(必填)*

- **Python 参考实现与行为对齐**：需覆盖 `agno/libs/agno/agno/agent/agent.py`、`agno/libs/agno/agno/workflow/*`, `agno/libs/agno/agno/session/*`, `agno/libs/agno/agno/run/*` 的核心函数，包括 Agent.run、Workflow.run、SessionManager 与 RunMessages。允许的差异仅限于 Go 语言特定的错误类型与并发模型，需在文档列举。
- **Go 原生实现与运行时独立**：Go 交付物位于 `go/agent`, `go/workflow`, `go/session`, `go/internal/...`，编译与运行均使用 `go build` / `go test`，运行期不得调用 Python。唯一可接受的 Python 依赖是对照测试脚本（位于 `scripts/ci/cross-language-parity.sh`）用于捕获参考输出，并在 CI 中与 Go 结果比较。
- **Go API 设计与包结构**：导出 API 包括 `agent.Agent`, `workflow.Sequential`, `session.Store`, `tools.Registry`，与 Python 抽象一一映射；若语义不同（例如上下文取消与错误处理），需提供迁移指南，指明 Go 版本的 context.Context 要求与超时策略。
- **跨语言测试纪律与 85% 覆盖率**：所有新增 Go 代码需在 `go test ./... -cover` 下达到 ≥85% 语句覆盖；需新增 parity tests（`go/agent/us1_parity_config_test.go` 扩展）与基准测试（`go/agent/us1_basic_coordination_bench_test.go`），并在 CI 中与 Python 结果对比。
- **性能与资源使用基线**：以 Python 运行三类 workload 的平均延迟/内存为基线，在 `scripts/benchmarks` 中记录。Go 需提交新的基准数据并保持 ≥25% 延迟改善与 ≥20% 内存下降；当未达标时需提供调优指南与已知问题列表。
- **安全、配置与 Telemetry**：必须复刻 Python 的配置键（API Keys、DB、Telemetry endpoints)，所有 Secrets 仍通过环境变量注入。Go 版本需要在日志中屏蔽 Secrets，Telemetry 中附带 runtime 标签与版本号，确保安全审计一致。
- **文档与示例对齐**：更新 `agno/README.md`、`cookbook/README.md`、`go/examples/*`，新增 “Go Runtime Quickstart”、“特性支持矩阵”、“迁移常见问题”，并指向 issue 模板帮助用户报告缺失特性。
