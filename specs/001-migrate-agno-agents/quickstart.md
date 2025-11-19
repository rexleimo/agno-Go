# Quickstart: 迁移 agno 核心 agents 能力到 Go

**Feature**: /Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/spec.md  
**Plan**: /Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/plan.md  
**Date**: 2025-11-19

本指南面向已经熟悉 Python 版 agno 的工程师，展示如何在 Go 中迁移一个基于 agents 的场景，并通过对照测试验证行为严格等价。

---

## 1. 选择一个 Python 示例作为迁移起点

1. 在仓库中查找 Python 示例目录：

   - 根路径：`/Users/rex/cool.cnb/agno-Go/agno/cookbook`
   - 推荐从 `cookbook/agents` 或 `cookbook/workflows` 下选择一个涵盖多个 agents 和协作模式的示例。

2. 记录该示例使用的：

   - Agents 角色与职责；
   - 使用到的 Providers（模型、检索、工具等）；
   - 工作流/协作模式（串行、并行、协调者-执行者等）。

这些信息将映射到 Go 侧的数据模型与包设计中。

---

## 2. 在 Go 中规划等价结构

1. 在计划阶段，目标结构为：

   - `go/agent`：Agent 抽象与实现。
   - `go/workflow`：Workflow 与协作模式执行引擎。
   - `go/providers`：与 Python 供应商一一对应的 Provider 适配层。
   - `go/session`：会话与任务上下文。

2. 对照 `/Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/data-model.md`，为选择的示例确定：

   - 需要的 Agent 列表及其 `allowed_providers`、`input_schema`、`output_schema`；
   - 对应 Provider 的 `type`、`config` 与 `capabilities`；
   - Workflow 的 `pattern_type` 与步骤结构。

3. 在后续实现中，将上述实体映射到对应 Go 包的导出类型与构造函数。

---

## 3. 定义 ParityTestScenario

1. 为选定示例创建一个 ParityTestScenario 条目（在实现阶段可按 data-model 与 contracts 中的约定落地）：

   - `id`：例如 `"agents-migration-demo-001"`；
   - `description`：说明该场景覆盖的 agents 和协作模式；
   - `input_payload`：固定的输入数据（如用户问题、上下文配置）；
   - `severity`：对于关键业务流设为 `must_match`；
   - `tolerance`：如确有必要，可定义字符串相似度等可量化容差。

2. 确保该场景覆盖：

   - 至少一个多 agent 协作流程；
   - 至少一个对外部 Provider 的调用；
   - 完整的 Session 生命周期（启动、运行、完成/失败）。

---

## 4. 实现 Python 与 Go 的对照测试

> 具体脚本实现将在后续任务中完成，此处描述的是目标工作流。

1. 在 Python 侧：

   - 复用现有示例代码，封装为可接收 `ParityTestScenario.input_payload` 的函数；
   - 将输出归一化为结构化对象（例如 dict/JSON），避免与具体日志格式耦合。

2. 在 Go 侧：

   - 实现对应的 Agent、Provider 与 Workflow；
   - 提供一个公共入口函数，接受同样结构的输入，返回可序列化为 JSON 的结果。

3. 在对照测试脚本中：

   - 对同一 ParityTestScenario：
     - 先调用 Python 实现，记录 `python_output`；
     - 再调用 Go 实现，记录 `go_output`；
   - 按 `/Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/contracts/openapi-migration-parity.yaml` 中的契约生成 `ParityRun` 结果，并根据 `severity` 与 `tolerance` 判定通过/失败。

---

## 5. 运行测试与观察结果

1. 在 Go 侧运行单元与对照测试（示意）：

   ```bash
   cd /Users/rex/cool.cnb/agno-Go
   go test ./...  # 针对 agent/workflow/providers 等包
   ```

2. 运行 US1 的基础多智能体示例（适合作为新用户起步场景）：

   ```bash
   cd /Users/rex/cool.cnb/agno-Go
   go run ./go/examples/us1_basic_coordination
   ```

2. 在测试输出或 Telemetry 中检查：

   - `ParityRun.status` 是否为 `completed`；
   - 所有标记为 `must_match` 的场景是否通过；
   - 如有失败，查看 `diff_summary` 与事件日志找出差异点。

---

## 6. 文档与后续场景扩展

1. 将首个迁移成功的示例记录为 Go 侧的 cookbook 示例（例如 `go/examples/us1_basic_coordination`），并在 Python cookbook 对应位置添加“Go 实现可用”说明。
2. 以同样方式逐步覆盖更多 cookbook 示例，优先覆盖使用频率高、业务影响大的场景。
3. 对于暂时无法迁移的供应商，按照规格要求在文档中维护例外清单与计划批次。
4. 若需要扩展新的自定义 Provider，可遵循以下路径：
   - 阅读协议文档：`specs/001-migrate-agno-agents/contracts/custom-provider-protocol.md`
   - 参考 Python 示例：`agno/libs/agno/agno/tools/custom_internal_search.py` 与 `agno/cookbook/scripts/us3_custom_provider_parity.py`
   - 在 Go 中按相同契约实现对应 Provider（参见 `go/providers/custom_internal_search.go` 及其 Parity 测试）

## 7. 未迁移能力说明

在首批迁移完成前，仍可能存在以下情况：

- 部分 cookbook 示例中使用的供应商尚未在 Go 侧实现适配层；
- 部分协作模式或 AgentOS 能力暂未在 Go 中提供等价实现。

当调用尚未迁移的能力时建议：

- 在 Go 实现中返回带有 `not_migrated` 错误码的错误（参见 `go/internal/errors`）；
- 在文档与 UI 中清晰提示该能力尚未可用，并指向 `providers-inventory.md` 和 `migration-strategy.md` 中的计划批次与替代方案。
