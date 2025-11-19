# Research: agno 核心 agents 能力迁移

**Feature**: /Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/spec.md  
**Plan**: /Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/plan.md  
**Date**: 2025-11-19  
**Context**: Phase 0 调研，用于支撑 Go 侧迁移 agno 核心 agents 能力（供应商与协作模式），并满足宪章与规格中的行为对齐要求。

---

## Decision 1: Go 版本与模块布局

- **Decision**: 采用 Go 1.22 作为目标版本，在 `/Users/rex/cool.cnb/agno-Go/go` 下按照功能域拆分包（如 `agent`、`workflow`、`providers`、`knowledge`、`session`），与 Python 侧 `agno/libs/agno/agno/` 中的模块一一对应。
- **Rationale**: Go 1.22 提供稳定的泛型与工具链支持，适合实现高并发、多协作模式的执行引擎；按功能域划分包可以与现有 Python 抽象（agent、workflow、team 等）建立清晰映射，方便按模块渐进迁移与测试。
- **Alternatives considered**:
  - 使用较低版本 Go（如 1.20）：虽然兼容面更广，但会限制对新语言特性的使用，对长期维护不利。
  - 将所有能力集中在单一包（如 `go/agno`）：短期实现简单，但会导致包接口膨胀、难以按模块对齐 Python 实现。

---

## Decision 2: 跨语言对照测试策略

- **Decision**: 通过对照测试脚本，将代表性输入发送给 Python 与 Go 实现，并将输出归一化为结构化格式（例如 JSON），保存在同一测试工件中，用 `go test` 驱动对照比较；Go 侧测试覆盖率目标为相关包 ≥ 85%。
- **Rationale**: 规格与宪章都要求“在可控随机性前提下行为严格等价”，对照测试是验证行为对齐的最直接方式；使用结构化输出可以规避日志格式差异，便于自动比较和回归。
- **Alternatives considered**:
  - 单独为 Go 实现编写全量业务测试而不与 Python 对比：难以保证与 Python 行为的一致性，且在 Python 演进时难以及时发现漂移。
  - 仅通过人工比对日志或交互式验证：效率低且不可复制，不符合自动化回归和覆盖率要求。

---

## Decision 3: 首批供应商迁移与分批策略

- **Decision**: 首批迁移范围覆盖 `/Users/rex/cool.cnb/agno-Go/agno/cookbook` 中出现的所有官方与第三方 agents 供应商；若个别供应商暂时无法迁移，则在文档与任务中显式记录“例外清单 + 后续批次计划”。
- **Rationale**: 规格中已将“所有 cookbook 中出现的供应商”设为首批目标，可以最大限度降低用户迁移阻力；通过例外清单控制范围，避免因个别供应商阻塞整个迭代交付。
- **Alternatives considered**:
  - 仅迁移官方供应商：短期工作量更小，但会导致大量 cookbook 示例无法直接迁移，削弱 agno-Go 的吸引力。
  - 一次性迁移所有潜在供应商（含未在 cookbook 出现的）：边界不清晰，工作量难以控制，不利于分阶段交付。

---

## Decision 4: 自定义扩展的复用模型

- **Decision**: 将自定义 agents 供应商与协作模式的扩展协议抽象为语言无关的契约（例如配置结构、接口行为描述与错误语义），在 Python 与 Go 两侧分别实现具体接口，力求在“同一契约”下尽量复用扩展定义，只在极端情况下引入语言专属实现差异。
- **Rationale**: 规格要求“尽可能两端复用”，而宪章强调两边行为对齐；通过契约抽象可以让扩展规范保持统一，同时保留为不同语言选择最佳实现策略的自由度。
- **Alternatives considered**:
  - 要求扩展代码在两种语言中 100% 复用：在现实中难以满足（语言特性、依赖和运行时差异大），会显著增加实现成本。
  - 完全放弃复用，仅要求“语义类似”：长期会造成生态割裂，使迁移成本随扩展数量增加而快速上升。

---

## Decision 5: 文档与 quickstart 覆盖范围

- **Decision**: 在 `/Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/quickstart.md` 中提供从“已有 Python 场景”到“Go 场景”的迁移向导，包括：如何选取一个 cookbook 示例、如何在 Go 包中重建等价场景、如何运行对照测试，以及如何阅读 Telemetry/日志来诊断差异。
- **Rationale**: 功能规格强调迁移窗口内需要可操作的迁移路径与回滚策略；quickstart 应以真实的 cookbook 示例为驱动，让团队可以按照文档一步步完成首个迁移案例。
- **Alternatives considered**:
  - 仅提供抽象概念性文档：难以指导真实迁移操作，用户需要自己在代码中摸索。
  - 在 agno-Go 仓库之外维护迁移指南：会让文档与代码容易脱节，增加同步与维护成本。

---

## Decision 6: Parity API 合同的使用方式

- **Decision**: `contracts/openapi-migration-parity.yaml` 作为对照测试编排的“目标合同”，在早期阶段主要用于约束内部脚本与测试输出结构，而非立即提供对外 HTTP 服务；后续若需要真实服务端，可直接以该合同为基础扩展。
- **Rationale**: 当前迁移重点在于行为对齐和测试纪律，先通过统一的结构化结果（ParityRun/ParityTestScenario）在本地脚本中实现即可满足验证需求；对外 API 可以在稳定后按需实现，以避免过早约束部署与基础设施形态。
- **Alternatives considered**:
  - 立即实现完整的 HTTP 服务并暴露对照测试 API：会提前引入部署和安全方面的复杂度，不利于先聚焦行为迁移本身。
  - 完全不定义合同，仅在脚本中随意选择 JSON 结构：短期灵活，但长期容易造成输出格式漂移，削弱对照测试结果的可重用性与可视化能力。

---

## Decision 7: US1 代表性迁移场景选择

- **Decision**: 将 `agno/cookbook/teams/basic_flows/01_basic_coordination.py` 选为 US1 的首个代表性迁移场景，对应的 ParityTestScenario ID 为 `teams-basic-coordination-us1`。该场景展示一个由团队 leader 协调的多 Agent 团队（至少包含 HackerNews Researcher 与 Article Reader），对给定主题进行联合研究与总结。
- **Rationale**: basic_flows 下的 “01_basic_coordination” 同时体现了多 Agent 协作与团队 leader 的决策逻辑，结构相对简单，便于在 Go 侧重建等价 Workflow/Session；同时其使用的供应商与工具具有代表性，可作为后续扩展到更多 teams/agents 示例的模板。
- **Alternatives considered**:
  - 选择更复杂的 distributed_rag 或 multimodal 场景：虽更贴近日常复杂需求，但会在首轮迁移中引入过多外部依赖与评估维度。
  - 选择单 Agent 示例：实现难度更低，但无法充分体现“协作模式”这一规格核心能力。

---

## Migration Status Summary (首批供应商与场景)

- **首批供应商（示例场景相关）**：
  - OpenAI Chat (`openai-chat-gpt-5-mini`)：已在 Go 侧实现 `US1OpenAIChat` 适配层，用于 US1 和 US2 场景中的文本生成。
  - HackerNewsTools (`hackernews-tools`)：已在 Go 侧实现 `US1HackerNewsTools`，作为 US1 场景中的新闻检索工具。
  - Newspaper4kTools (`newspaper4k-tools`)：已在 Go 侧实现 `US1Newspaper4kTools`，作为 US1 场景中的文章阅读工具。
  - CustomInternalSearch (`custom_internal_search`)：已在 Python 与 Go 两侧实现，用于 US3 自定义 Provider 协议与 parity 示例。

- **场景与供应商绑定**：
  - US1 (`teams-basic-coordination-us1`)：覆盖团队协作模式 + OpenAIChat + HackerNewsTools + Newspaper4kTools。
  - US3 (`custom-internal-search-us3`)：覆盖自定义内部搜索 Provider，在 Python 与 Go 之间进行精确 parity 比较。

- **例外与后续计划**：
  - 当前清单仅覆盖代表性的首批供应商；`providers-inventory.md` 将在后续迭代中扩展至 `agno/cookbook` 中所有出现的供应商，并按“首批必须迁移 / 后续批次 / 不再支持”分类。
  - 尚未迁移的供应商会在 `providers-inventory.md` 中标记为后续批次，并在迁移策略文档中给出优先级与计划节奏。

---

## Config Mapping 应用（US1）

- **Python 源配置**：
  - `OpenAIChat("gpt-5-mini")` + `HackerNewsTools()` + `Newspaper4kTools()` 注入到 US1 场景中的两个 Agent；
  - Team 中通过成员列表和 instructions 描述协作顺序。

- **Go 侧构造代码**：
  - 在 `go/providers/providers.go` 中按照 config-mapping 约定定义：
    - `US1OpenAIChat`（TypeLLM，Config: `{"model": "gpt-5-mini"}`）；
    - `US1HackerNewsTools`、`US1Newspaper4kTools`（TypeToolExecutor，Capability: `invoke_tool`）。
  - 在 `go/agent/us1_basic_coordination_agents.go` 中：
    - 使用 `AllowedProviders` 与 `AllowedTools` 将 Agent 绑定到上述 Provider ID；
    - 在 `Schema` 字段中明确输入/输出语义。
  - 在 `go/workflow/us1_basic_coordination_workflow.go` 中：
    - 将 Python Team 的成员执行顺序映射到 `PatternSequential` 的 Workflow，按步骤 ID 定义 routing rules。

- **偏离与说明**：
  - Python Team 的 instructions 字符串用于提示执行顺序和细节，Go 侧将其抽象为 `PatternType + Steps + RoutingRules`，不逐条复制文本指令；
  - Session 构造函数 `RunUS1Session` 当前使用占位 Result 数据（仅回显 query），未来接入真实执行引擎后会根据 Workflow 和 Providers 生成完整结果，但 Config 映射规则本身保持不变。*** End Patch

---

## 性能基线（T052 初始结果）

- **Go 侧基准测试**：
  - 基准位置：`go/agent/us1_basic_coordination_bench_test.go`
  - 运行命令：
    - `cd /Users/rex/cool.cnb/agno-Go`
    - `go test ./go/agent -run ^$ -bench BenchmarkRunUS1Example -benchtime=10x`
  - 在 Apple M4 开发环境上的一次样例结果：

    ```text
    BenchmarkRunUS1Example-10          10          ~8–30 ns/op
    ```

  - 说明：
    - 当前 `RunUS1Example` 实现仍为轻量占位（主要是结构体构造），因此基准值反映的是 Go 端入口函数与序列化壳层的开销，而非完整多智能体执行路径。
    - 随着后续接入 Workflow 执行与 Provider 调用，该 benchmark 将成为观察性能变化的基础。

- **Python 侧基准测试现状**：
  - 计划通过 `scripts/bench_us1.sh` 调用 `cookbook.scripts.us1_basic_coordination_parity.run_parity` 进行简单 timing。
  - 目前在本环境中运行该脚本时，由于 Python 模块导入路径未正确配置（`ModuleNotFoundError: No module named 'cookbook.teams.basic_flows._01_basic_coordination'`），Python 端 timing 尚未成功；需要在实际运行环境中先配置好 Python 的 `PYTHONPATH` 或安装方式。
  - 因此本迭代仅记录 Go 端基线数值，Python 端性能对比将在后续迭代或集成环境中补充。
