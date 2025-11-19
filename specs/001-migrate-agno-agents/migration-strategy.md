# Migration Strategy: Python → Go (agno 核心 agents 能力)

**Feature**: /Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/spec.md  
**Plan**: /Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/plan.md  
**Date**: 2025-11-19

本文档描述在迁移 agno 核心 agents 能力到 Go 时的分阶段迁移策略、流量路由与回滚方案，以及基于 ParityTestScenario 的覆盖率与通过率管理方式。

---

## 1. 分阶段迁移策略

### 1.1 阶段划分

- **阶段 0：对照测试与样例落地（US1, US2, US3）**
  - 完成代表性场景（US1/US2/US3）的 Python 与 Go 实现。
  - 建立 ParityTestScenario 与 ParityRun 配置（如 `teams-basic-coordination-us1`、`custom-internal-search-us3`）。
  - 确立基础多智能体模板与自定义 Provider 协议。

- **阶段 1：扩展首批供应商与场景**
  - 基于 `providers-inventory.md` 进一步扩展首批 suppliers 集合。
  - 为每个“首批必须迁移”的供应商至少绑定一个 ParityTestScenario。
  - 按本策略文件中的路由规则在内部环境中进行灰度演练。

- **阶段 2：双运行时灰度与回滚**
  - 在生产或接近生产的环境中同时部署 Python 与 Go 实现。
  - 通过配置（例如环境变量或路由服务）按比例导入流量到 Go 实现。
  - 监控行为差异、性能与严重故障占比，并在必要时回滚至 Python 实现。

### 1.2 路由与回滚原则

- 默认情况：所有新创建场景仍路由到 Python 实现。
- 对于已完成 parity 验证的场景（如 US1、US3 及后续扩展的首批场景）：
  - 在内部环境中以 0%→10%→50%→100% 的梯度导入读流量到 Go 实现。
  - 对写操作或有副作用的流程，在达到 50% 前需确保幂等性或具备回滚机制。
- 若监控显示：
  - 行为不一致导致任务失败的严重故障占比超过 2%；
  - 或关键性能指标明显劣于 Python；
  - 则应立即将该场景路由回 Python，实现快速回滚。

路由配置示例见 `scripts/route_python_go_example.sh`。

---

## 2. 场景覆盖率与通过率管理

### 2.1 覆盖率目标

- 代表性场景覆盖率目标：至少覆盖当前 agents 与协作模式场景的 80%。
- 对于每个 `must_match` 场景，Go 实现上线前必须通过 Parity 测试。

### 2.2 当前场景清单

| Scenario ID                     | Description                                      | Severity    | Providers Involved                                  |
|---------------------------------|--------------------------------------------------|-------------|-----------------------------------------------------|
| teams-basic-coordination-us1    | 团队协调 HackerNews + article 阅读场景           | must_match | openai-chat-gpt-5-mini, hackernews-tools, newspaper4k-tools |
| custom-internal-search-us3      | 自定义内部搜索 Provider 行为对齐                  | must_match | custom_internal_search                              |

> 后续迭代将根据 `providers-inventory.md` 和 cookbook 内容扩展场景清单。

### 2.3 覆盖率与通过率统计（示例流程）

- 通过一个统一的 parity runner（可由脚本或 CI Job 实现）：
  - 读取所有 ParityTestScenario 定义；
  - 对每个场景执行 Python 与 Go 实现；
  - 比较行为并记录通过/失败。
- 统计指标：
  - 覆盖率：`已实现 parity 的场景数 / 目标场景总数`（目标 ≥80%）；
  - 通过率：`通过场景数 / 所有 must_match 场景数`（目标接近或达到 95%）。
- 当前迭代提供了一个极简的 parity 统计脚本：
  - 路径：`scripts/parity_stats.sh`
  - 当前配置的场景与命令：
    - `teams-basic-coordination-us1` → `go test ./go/agent -run TestUS1ParityConfigScript`
    - `custom-internal-search-us3` → `go test ./go/providers -run TestUS3CustomProviderParity`
  - 样例运行结果（2025-11-19）：
    - 总场景数：2（全部为 must_match）
    - 已实现场景数：2
    - 通过场景数：2
    - 覆盖率：100%
    - 通过率：100%


---

## 3. 严重故障记录与占比统计

在内部测试或灰度阶段，需要记录“因行为不一致导致任务无法完成的严重故障”，并计算其占比是否满足 SC-003 中提出的“≤2%”目标。

### 3.1 严重故障判定

- 当出现以下情况之一，即认为是严重故障：
  - Go 实现给出与 Python 明显不一致的业务结果，导致用户任务无法完成；
  - Go 实现缺失关键错误信息或提示，使得用户无法采取补救措施；
  - Go 实现触发的错误在 Python 中不存在，且影响主要业务流程。

### 3.2 记录方式（基于 TelemetryEvent）

- 在 Go 实现中，对严重故障事件以 TelemetryEvent 结构记录：

  - `event_type`: `provider_error` 或其它合适类型；
  - `payload`: 至少包含 `scenario_id`、`provider_id`、`error_code`（例如 `internal`、`not_migrated`）。

- 可在后续实现中引入一个简单的聚合脚本或仪表盘统计：

  - 总请求数（分场景）；
  - 严重故障数；
  - 故障占比（用于对照 2% 阈值）。

### 3.3 结果记录

- 针对每个灰度阶段（例如小流量试运行），建议通过脚本计算并记录结果：
  - 统计命令示例：

    ```bash
    cd /Users/rex/cool.cnb/agno-Go
    ./scripts/incidents_stats.sh logs/telemetry.jsonl
    ```

  - 样例输出（当前尚无 Telemetry 数据时）：

    ```text
    Telemetry file not found: logs/telemetry.jsonl
    Total requests:        0
    Severe incidents:      0
    Severe incident rate:  0% (no data)
    ```

  - 在实际灰度阶段，应将带有 `error_code` 字段的 Telemetry 事件写入 `logs/telemetry.jsonl`，再使用上述脚本计算严重故障占比，并将结果记录到本文件或 `research.md` 中，以便后续迭代参考和优化。
