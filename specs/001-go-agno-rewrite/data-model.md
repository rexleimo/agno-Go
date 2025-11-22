# Data Model - Go 版 Agno 重构

## 目标
提供与 Python 版一致的 Agent/AgentOS 行为，覆盖九家模型供应商的聊天与嵌入接口，支持会话记忆、工具调用与流式输出，并输出契约/基准/覆盖率报告。

## 核心实体

### Agent
- `id` (uuid) / `name` (string, unique)
- `model`：`provider`（enum: ollama, gemini, openai, glm4, openrouter, siliconflow, cerebras, modelscope, groq）、`model_id`、`stream`(bool)、`temperature`、`max_tokens`、`timeout_ms`
- `tools`：`registered`(list)、`enabled`(list)、`mcp_endpoints`(list)、`tool_timeout_ms`
- `memory`：`store_type`(enum: memory, bolt, badger)、`namespace`、`retention`(ttl)、`token_window`
- `knowledge`：可选向量库/索引引用（占位），需要与 Python 行为对齐
- `metadata`：tags、created_at、updated_at

关系：`Agent` 1—N `Session`；`Agent` 1—N `ToolBinding`；`Agent` 1—N `KnowledgeIndex`（可选）。

校验：`name` 非空唯一；`model.provider` 必须在支持矩阵内；`tools.enabled` ⊆ `tools.registered`；`memory.store_type` 合合法定枚举且在 `.env` 支持范围内。

### Session
- `session_id` (uuid)
- `agent_id` (uuid, fk)
- `state` (enum: `idle`→`streaming`→`completed` / `errored` / `cancelled`)
- `history`：消息/工具调用记录（有序 append-only）
- `user`：可选 user_id/tenant_id
- `context`：会话级变量、工具输出缓存
- `last_activity_at` / `expired_at`

关系：`Session` 属于 `Agent`，1—N `Message`/`ToolCall`。

状态流转：`idle`（创建）→ `streaming`（接收/发送）→ `completed`（正常结束）；遇错误进入 `errored`；可人工 `cancelled`。

### Message
- `id` (uuid)、`session_id`(fk)、`agent_id`(fk)
- `role` (user/assistant/system/tool)；`content`（text/structured）
- `tool_calls`：列表包含 `tool_call_id`、`name`、`args`
- `attachments`：可选二进制/URL
- `usage`：prompt_tokens、completion_tokens、latency_ms（用于契约/基准）

校验：`role` 受限枚举；当 `role=assistant` 且含 `tool_calls` 时必须存在对应 `ToolCallResult`。

### ToolCall / ToolCallResult
- `tool_call_id`、`name`、`args`、`issued_at`
- `result`：`status`(success/error/timeout/disabled)、`output`、`duration_ms`
- `tool_origin`：MCP/内置/自定义

关系：`Message` 1—N `ToolCall`，每个 `ToolCall` 0..1 `ToolCallResult`。

### ProviderAdapter
- `provider`（enum 九家）
- `capabilities`：chat/embedding/streaming (bool)
- `endpoints`：base_url、region、api_version
- `auth`：env var keys、token/endpoint 是否必需
- `limits`：rate_limits、timeout、重试策略
- `status`：available/not-configured/disabled

校验：缺少必需 env 时标记 `not-configured` 并在契约/集成测试中跳过且记录原因。

### Fixture / ContractCase
- `id`、`provider`、`type`(chat|embedding)、`input`、`expected`、`tolerance`（tokens: ±2、embedding cos ≥0.98 默认）
- `source`：对应 Python 版本 commit/hash
- `created_at`、`notes`

关系：与 `ProviderAdapter` 关联，用于契约测试；偏差记录于 deviations。

### BenchmarkRun
- `run_id`、`scenario`（100 并发 128 tokens 10 分钟基准）
- `metrics`：p50/p95 latency、peak_rss_mb、error_rate、throughput
- `environment`：CPU/内存/Go 版本/GC 设置、启用的 providers
- `artifacts_path`：`/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/artifacts/bench/<run_id>/`

校验：与 Python 基线对比，需达到 ≥20% p95 降低、≥25% 峰值内存降低。

### CoverageReport
- `report_id`、`coverage_pct`（目标 ≥85%）、`packages_missing` 列表
- 产出路径：`/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/artifacts/coverage/`
- 来源：`make coverage` 聚合单元/契约/供应商/基准测试

校验：低于阈值即 fail；缺测试需补充或缩小改动。

## 关系图（文本）
- Agent → Session (1:N)
- Session → Message (1:N)
- Message → ToolCall (1:N)
- ToolCall → ToolCallResult (1:0..1)
- Agent → ProviderAdapter (1:N，通过模型配置绑定)
- ProviderAdapter ↔ Fixture/ContractCase (1:N，用于契约校验)
- Agent/Session → BenchmarkRun/CoverageReport（统计与报告关联，不直接持久化在运行态）

## 校验与一致性规则
- 缺少 provider 必需 env：禁用相应 ProviderAdapter，API 返回 503+详细错误，契约/集成测试显式跳过并记录。
- 流式响应：当 `stream=true` 时返回 SSE/分块响应；任何中途错误应发出终止事件并在 Session 状态中记录。
- 工具禁用/失败：`ToolCallResult.status` 反映禁用/超时/错误，不得 fallback 至 Python；会话继续且提供降级提示。
- 记忆存储切换：`MemoryStore` 初始化时校验 store_type；持久化实现必须提供 WAL/刷盘保证，防止在压测下丢失消息。

## 数据保真与审计
- 契约 fixtures 需携带来源 commit 与时间戳，用于追踪 Python 参考版本。
- 基准与覆盖率报告落盘于 artifacts，并在 CI 上传；报告需脱敏（不含密钥/用户输入原文）。
