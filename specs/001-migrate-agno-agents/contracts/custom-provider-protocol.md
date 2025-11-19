# Custom Provider Protocol: Agents 供应商扩展协议

**Feature**: /Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/spec.md  
**Plan**: /Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/plan.md  
**Date**: 2025-11-19

本协议约定了在 Python 与 Go 两侧实现自定义 Agents 供应商（Provider）时应遵循的公共契约，以便在两种运行时之间尽可能复用扩展定义，并通过 ParityTestScenario 验证行为与错误语义一致。

---

## 1. Provider 标识与元数据

- `id`：在系统内唯一的字符串标识，例如 `internal-search`、`billing-api`。
- `type`：能力类型，取值来自数据模型中的 Provider.Type：
  - `llm`：通用大模型服务
  - `retriever`：检索/搜索服务
  - `tool-executor`：工具执行器（包装一个或多个工具）
  - `business-api`：业务系统 API 封装
- `display_name`：面向用户与 Telemetry 的名称，用于日志、监控与文档。
- `telemetry_tags`（可选）：一组键值对，用于将 Provider 相关信息附加到 TelemetryEvent 中（例如 `{"provider":"internal-search","team":"payments"}`）。

**约束**：

- `id` 在整个系统中必须唯一。
- `display_name` 不得为空，应能清晰反映 Provider 的用途。

---

## 2. 配置结构（config）

配置采用键值结构，在 Python 与 Go 中的表示分别为：

- Python：`Dict[str, Any]`（通常由 pydantic/TypedDict 或 dataclass 封装）
- Go：`providers.Config`，底层类型为 `map[string]any`

**建议字段**：

- `endpoint`：业务 API 地址（如适用）
- `timeout_ms`：调用超时时间
- `retries`：重试次数或策略
- `auth_mode`：认证方式（如 `api_key`、`oauth`）
- `extras`：其他与业务相关但不影响协议的字段

**协议约束**：

- 配置字段应在 Python 与 Go 两侧保持语义一致（例如 `timeout_ms` 表示同一含义）。
- 如某字段仅在单侧实现中使用，应在文档中标注为“运行时特定”并说明影响范围。

---

## 3. 能力集合（capabilities）

`capabilities` 用于描述 Provider 支持的操作集合，基于数据模型中建议的枚举：

- `generate`：生成文本或结构化内容
- `embed`：生成向量表示
- `search`：执行检索或搜索操作
- `invoke_tool`：调用底层工具或操作

对于自定义 Provider，至少应明确为每个公开方法标注其能力类型。例如：

- 内部搜索 API：`["search"]`
- 内部审批 API：`["invoke_tool"]`

---

## 4. 错误语义（error_semantics）

所有自定义 Provider 的错误应映射到统一的错误代码集合，以便在 Python 与 Go 之间进行对照测试与集中处理。

推荐错误代码：

- `timeout`：调用超时
- `rate_limit`：被上游限流
- `unauthorized`：认证/授权失败
- `internal`：未分类的内部错误
- `not_migrated`：请求使用尚未迁移到 Go 的能力（由 Go 侧统一返回）

**约定**：

- Python 侧应通过异常类型或错误对象记录原始错误，再映射到上述错误代码之一。
- Go 侧应使用 `go/internal/errors` 包中的 `Code` 与 `Error` 类型表示错误，并在 TelemetryEvent 中记录 `Code`。

---

## 5. 行为接口

自定义 Provider 至少应定义一个“核心操作函数”，例如：

- 内部搜索示例：
  - Python：`search_documents(query: str) -> str`（返回 JSON 字符串）
  - Go：`SearchDocuments(query string) (SearchResult, error)`

**行为约束**：

- 对于相同的输入（例如同一个 `query`），Python 与 Go 实现应在业务语义层面返回等价结果：
  - 结果结构字段名称一致（如 `results`、`id`、`title`）
  - 关键字段值一致，或在 ParityTestScenario 的容差规则内相等
- 方法的副作用（如 Telemetry 记录）应在类型与字段上保持一致，便于对照测试与排查。

---

## 6. 版本与演进

- `version` 字段可选，建议用于在业务方扩展协议时区分适配器版本（如 `v1`、`v2`）。
- 协议演进原则：
  - 新增字段应保持向后兼容（默认值合理）。
  - 如需改变字段含义，必须在文档中记录版本差异，并在 ParityTestScenario 中补充新的场景。

---

## 7. ParityTestScenario 集成

每个自定义 Provider 至少应关联一个 ParityTestScenario，内容包括：

- `id`：例如 `custom-internal-search-us3`
- `description`：例如“内部搜索 Provider 在关键字查询上的行为对齐”
- `input_payload`：例如 `{"query": "internal-api"}`（固定输入）
- `severity`：通常为 `must_match`
- `tolerance`：如允许结果排序差异，可在此注明

测试步骤：

1. 使用 Python 实现执行核心操作，得到 `python_output`。
2. 使用 Go 实现执行同一操作，得到 `go_output`。
3. 按照 ParityRun 约定序列化并比较关键字段，若超过容差则视为失败。

