# Data Model：Agno 核心模块 Go 迁移

**Feature**：/home/rex/codes/agno-Go/specs/001-migrate-agno-core/spec.md  
**Plan**：/home/rex/codes/agno-Go/specs/001-migrate-agno-core/plan.md  
**Date**：2025-11-19

本数据模型描述迁移 Python Agno 核心模块到 Go 运行时时涉及的关键实体、字段与约束，保持技术无关，仅关注语义对齐与可测试性。

---

## Entity：AgentRuntime

代表 Go 运行时中可执行的单个 Agent 配置，需与 Python `agno.agent.Agent` 抽象一致。

- **Fields**
  - `id`：唯一标识（UUID/slug），用于 Session 与 Workflow 的引用。
  - `name`：人类可读名称，默认来自配置；用于日志与 Telemetry。
  - `model_ref`：引用模型/供应商标识（如 `openai:gpt-5-mini`）。
  - `toolkits`：工具集或 Function 列表，与 Python `ToolExecution` 对齐。
  - `memory_policy`：结构体，包含是否持久化、窗口大小、脱敏策略。
  - `session_policy`：会话 ID、State 合并策略、是否允许 agentic state。
  - `hooks`：前/后置钩子与过滤器列表，需保持顺序执行。
  - `timeouts`：运行与工具调用的时间上限。
  - `metadata`：额外标签，如团队、deployment、runtime version。
- **Relationships**
  - 一个 `AgentRuntime` 可由多个 `WorkflowRun` 引用。
  - `toolkits` 与 Provider registry 绑定，共享 Capability 定义。
  - `session_policy` 依赖 `SessionRecord` 的字段以进行状态同步。
- **Validation rules**
  - `id` 全局唯一，禁止与未迁移 agent 重名。
  - 未显式配置 `toolkits` 时需允许空列表，但 telemetry 中必须记录“无工具”。
  - `memory_policy.window_size` ≥0；若 `persist=true` 则需要提供 store 接口实现。
  - `session_policy.overwrite` 与 `enable_agentic_state` 不可同时为 true，否则可能丢失用户 state，需在构造时抛错。

---

## Entity：WorkflowRun

描述一次多步骤执行的状态，覆盖顺序/并行/团队模式。

- **Fields**
  - `id`：运行标识，通常等同 session run id。
  - `workflow_ref`：指向 Workflow 配置或 manifest（ID + 版本）。
  - `pattern_type`：`sequential` / `parallel` / `coordinator-worker` / `loop`。
  - `steps`：步骤数组，每项包含 `step_id`、关联 `agent_id`、输入/输出摘要、状态（pending/running/completed/failed/paused）。
  - `routing_rules_applied`：记录在运行中被触发的条件表达式。
  - `reasoning_trace`：结构化的 `ReasoningStep` 序列，并带分数/引用。
  - `background_tasks`：异步任务状态，用于工具并行执行。
  - `resources_used`：token 数、延迟、工具开销等指标。
- **Relationships**
  - `WorkflowRun` 依赖 `AgentRuntime` 集合定义步骤。
  - 通过 `SessionRecord` 获取/更新共享状态。
  - Telemetry 事件 `RunStarted/RunCompleted` 关联 `WorkflowRun.id`。
- **Validation rules**
  - 所有 `steps.agent_id` 必须在 `AgentRuntime` 列表中存在。
  - 并行步骤必须注明同步节点或终止条件，否则 plan 阶段即标记为错误。
  - `reasoning_trace` 需提供 append-only 语义，并在 parity 测试中可序列化为 JSON。
  - `resources_used` 必须至少包含 `prompt_tokens`, `completion_tokens`, `latency_ms`。

---

## Entity：SessionRecord

会话级持久化信息，用于跨运行共享用户上下文、历史与元数据。

- **Fields**
  - `session_id`：用户或任务实例的唯一标识。
  - `user_id` / `team_id`：可选，多租户标记。
  - `state_blob`：JSON 序列化的任意键值对，遵循 Python `session_state` 结构。
  - `history`：RunMessages 列表（角色、内容、引用、工具结果）及裁剪策略。
  - `summary`：最新 SessionSummary 文本与版本戳。
  - `metrics`：累计指标（成功率、平均 tokens、最近一次 run id 等）。
  - `cache_policy`：是否允许内存缓存、是否启用 search_session_history。
  - `created_at/updated_at`：时间戳，用于 TTL 与追踪。
- **Relationships**
  - `SessionRecord` 与多个 `WorkflowRun` (1:N) 关联。
  - `state_blob` 中的 `agent_state` 供 `AgentRuntime` 在运行前加载。
  - `metrics` 与 Telemetry 聚合结果保持一致。
- **Validation rules**
  - `state_blob` 必须可被 Python JSON schema 验证，避免键不一致。
  - 当 `history` 超过设定窗口时需自动裁剪并记录裁剪事件（Parity 重要）。
  - `summary` 版本发生变化时，应写入 `updated_at` 并触发 Telemetry 事件，便于 Go/Python 对齐。

---

## Entity：TelemetryEnvelope

封装跨语言共享的运行事件，驱动监控、审计与 parity 检测。

- **Fields**
  - `event_id`：唯一标识，用于幂等写入。
  - `timestamp`：UTC 时间。
  - `runtime`：固定枚举 `go` 或 `python`。
  - `event_type`：`run_started`, `run_completed`, `reasoning_started`, `reasoning_step`, `tool_call_started`, `tool_call_completed`, `session_summary_started`, `session_summary_completed`, `run_error` 等。
  - `session_id` / `workflow_run_id` / `agent_id` / `provider_id`（可选）：标记事件关联。
  - `payload`：结构化 map，包含非敏感字段（token 统计、工具名、故障分类、metadata）。
  - `attachments`：可选数组，用于 record scrubbed media/文件引用的元数据。
  - `correlation_ids`：跨系统关联（如 trace_id、thread_id）。
- **Relationships**
  - `TelemetryEnvelope` 引用 `SessionRecord` 与 `WorkflowRun`。
  - `payload.error_code` 必须映射到 `go/internal/errors` 中的 code。
- **Validation rules**
  - `runtime` 必须精确为 `go` 或 `python`，禁止空值。
  - `payload` 禁止包含 secrets；需要对 media/object id 进行白名单校验。
  - 事件序列需满足最小组合：每次运行至少有 `run_started` 与 `run_completed`，错误情况下追加 `run_error`。

---

## Supporting Artifact：ParityFixture

虽然规格把 Parity 测试归类到流程层，但需要一个共享描述；因此定义 `ParityFixture` 作为契约输入。

- **Fields**
  - `fixture_id`：与 cookbook/示例对齐的名称。
  - `description`：帮助调试的说明。
  - `workflow_template`：引用 Workflow/Agent 的静态快照（确保 Python/Go 同步）。
  - `user_inputs`：按时间排序的消息数组，包含角色、内容、媒体引用、随机种子。
  - `tool_responses`：当工具结果需要 mock 时的可复现数据。
  - `expected_assertions`：应包含 equality、contains、tolerance 三种断言类型。
- **Validation rules**
  - `workflow_template` 中的 Agent IDs 必须与 fixture 目标一致。
  - `expected_assertions` 必须指定 diff 方法（text equality、json diff、numeric tolerance）。
  - 所有 fixture JSON/YAML 必须可被 Python/Go 同时解析（UTF-8，无注释）。

---

本数据模型将直接驱动 Phase 1 contracts 与 Phase 2 tasks：AgentRuntime/WorkflowRun/SessionRecord 映射到 Go 包导出的 struct 与接口；TelemetryEnvelope/ParityFixture 则用于 cross-language 测试与文档示例。
