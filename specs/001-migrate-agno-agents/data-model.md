# Data Model: agno 核心 agents 能力迁移

**Feature**: /Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/spec.md  
**Plan**: /Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/plan.md  
**Date**: 2025-11-19

本数据模型用于描述在 Go 中迁移 agno 核心 agents 能力时涉及的主要领域实体、字段与关系，为后续 Go 包设计（`go/agent`、`go/workflow`、`go/providers` 等）提供抽象基础。模型保持技术栈无关，仅描述语义与约束。

---

## Entity: Agent

代表一个具备明确角色和能力的执行单元。

- **Fields**
  - `id`: 全局唯一标识，用于在 Workflow/Session 中引用。
  - `name`: 人类可读名称，用于日志与调试。
  - `role`: 文本描述，体现 Agent 在协作中的职责（如 coordinator、worker、reviewer）。
  - `description`: 详细说明 Agent 的行为边界与使用场景。
  - `allowed_providers`: Provider ID 列表，约束该 Agent 可使用哪些供应商能力。
  - `allowed_tools`: 工具或操作的标识列表，与 Provider 的工具能力组合使用。
  - `input_schema`: 对 Agent 接收的输入结构的约定（例如字段名称和必填性）。
  - `output_schema`: 对 Agent 产生输出的结构约定，用于其他 Agent 或外部系统消费。
  - `memory_policy`: 会话记忆策略（如是否持久化、保留窗口大小、敏感信息过滤规则）。
- **Relationships**
  - 多个 Agent 可以被同一个 Workflow 引用。
  - Agent 通过 Provider 与外部能力交互，通过 Workflow/Session 环节参与协作。
- **Validation rules**
  - `id` 必须全局唯一。
  - `allowed_providers` 中的每个 Provider 必须存在且支持 Agent 所需能力类型。
  - `input_schema` 与 `output_schema` 在 Workflow 中必须可组合（下游 Agent 能消费上游输出）。

---

## Entity: Provider (Agents 供应商)

代表底层能力提供方（模型服务、检索服务、业务系统接口等）。

- **Fields**
  - `id`: 唯一标识，对应 cookbook 中的供应商配置名。
  - `type`: 能力类型，例如 `llm`, `retriever`, `tool-executor`, `business-api` 等。
  - `display_name`: 面向用户的名称，用于文档与 Telemetry。
  - `config`: 配置结构（如模型名称、端点、限流参数、重试策略），抽象为键值字段集合。
  - `capabilities`: 支持的操作集合（如 `generate`, `embed`, `search`, `invoke_tool`）。
  - `error_semantics`: 统一的错误分类与代码（如 `timeout`, `rate_limit`, `unauthorized`, `internal`）。
  - `telemetry_tags`: 在日志与指标中标注供应商相关信息的标签集合。
- **Relationships**
  - 一个 Provider 可以被多个 Agent 引用。
  - Provider 可与 Knowledge/Memory 等实体协同工作（例如将检索结果交给 Agent）。
- **Validation rules**
  - `config` 中必须包含能力类型要求的最小字段集合（例如 LLM 要求模型标识、超时等）。
  - `capabilities` 必须与 `type` 一致（如 `retriever` 不应提供 `generate` 能力）。
  - 错误语义必须映射到统一的 `error_semantics` 分类，便于跨运行时对照测试。

---

## Entity: Workflow / Collaboration Pattern

描述多个 Agent 之间的分工、顺序与决策规则。

- **Fields**
  - `id`: 唯一标识。
  - `name`: 人类可读名称。
  - `steps`: 有序步骤列表，每个步骤对应一个 Agent 调用或条件分支。
  - `pattern_type`: 协作模式类型（如 `sequential`, `parallel`, `coordinator-worker`, `loop`）。
  - `entry_points`: 可以触发 Workflow 的起始节点集合。
  - `termination_condition`: 终止条件（如达到目标、达到最大迭代轮数、出现不可恢复错误）。
  - `routing_rules`: 在不同步骤之间流转的条件（例如基于 Agent 输出或 Telemetry 作决策）。
- **Relationships**
  - Workflow 引用多个 Agent。
  - 一个 Workflow 可以被多个 Session/Task 实例化。
- **Validation rules**
  - `steps` 中引用的所有 Agent 必须存在。
  - `routing_rules` 不得形成无限循环，除非有明确的终止条件约束。
  - 对于标记为“严格等价”的场景，Workflow 的 `pattern_type` 与步骤结构应与 Python 侧配置一致。

---

## Entity: Session / Task

代表一次完整的用户任务或多轮对话。

- **Fields**
  - `id`: 会话或任务的唯一标识。
  - `workflow_id`: 使用的 Workflow 标识。
  - `user_context`: 初始用户输入与元数据（如用户标识、渠道、语言偏好）。
  - `history`: 交互历史（Agent 间消息、用户输入、系统提示）。
  - `status`: 当前状态（如 `pending`, `running`, `completed`, `failed`）。
  - `result`: 任务最终输出（供业务方或外部系统消费）。
  - `telemetry_trace_id`: 对应日志与监控系统中的 trace 标识。
- **Relationships**
  - 每个 Session 绑定一个 Workflow。
  - Session 与多个 Agent 调用记录相关联。
- **Validation rules**
  - `workflow_id` 必须指向有效的 Workflow。
  - 在 `completed` 状态下必须存在 `result`。
  - `history` 应记录支持对照测试所需的关键事件序列（至少包含各 Agent 的输入输出摘要）。

---

## Entity: ParityTestScenario

用于跨 Python 与 Go 实现的行为对照测试。

- **Fields**
  - `id`: 唯一标识，对应某个 cookbook 示例或业务用例。
  - `description`: 简要说明测试目标和覆盖场景。
  - `input_payload`: 固定的输入数据（如用户问题、上下文配置），可序列化为 JSON 或等价格式。
  - `expected_behavior`: 对行为的描述（例如“两个实现的业务结果和关键决策路径应一致”）。
  - `tolerance`: 对允许差异的定量定义（如字符串相似度阈值、列表顺序可忽略等）。
  - `severity`: 场景重要性级别（如 `must_match`, `should_match`）。
- **Relationships**
  - 一个 ParityTestScenario 关联一个或多个 Workflow/Session 模板。
- **Validation rules**
  - 所有 `must_match` 场景在 Go 实现上线前必须通过。
  - `tolerance` 的定义必须与对照测试工具实现兼容，避免不可测试的模糊表述。

---

## Entity: TelemetryEvent

用于在迁移期间观测行为与诊断差异。

- **Fields**
  - `id`: 事件唯一标识。
  - `timestamp`: 事件时间。
  - `session_id`: 关联的 Session。
  - `agent_id`: 如适用，关联的 Agent。
  - `provider_id`: 如适用，关联的 Provider。
  - `event_type`: 事件类型（如 `request_started`, `request_finished`, `provider_error`, `workflow_step_transition`）。
  - `payload`: 结构化内容，包含对照测试和故障排查所需的非敏感信息。
- **Relationships**
  - TelemetryEvent 关联 Session/Agent/Provider。
- **Validation rules**
  - 不得在 `payload` 中记录敏感数据（如明文凭证、个人隐私）。
  - 对于关键路径，应至少产生 `request_started` 与 `request_finished` 事件，以支持端到端延迟计算与问题定位。

