# 实现计划：agno 核心 agents 能力迁移

**分支**：`001-migrate-agno-agents` | **日期**： 2025-11-19 | **规格**： /Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/spec.md
**输入**：来自 `/specs/[###-feature-name]/spec.md` 的功能规格

**说明**：该模板由 `/speckit.plan` 命令填充。执行流程见 `.specify/templates/commands/plan.md`。

## 摘要

本迭代的目标是在 Go 中迁移 agno 的核心 agents 能力，覆盖 cookbook 中已经出现的所有官方与第三方 agents 供应商，并在相同输入下实现与 Python 版本“尽可能严格等价”的协作行为。规划将围绕以下路径展开：以 `agno/libs/agno/agno/` 作为行为基线，通过针对 Agent、Provider、Workflow、Session 等核心实体的数据模型设计、跨语言对照测试（Python vs Go）和统一扩展协议，实现“一次扩展定义，尽可能两端复用”，并通过 specs 目录下的 research/data-model/contracts/quickstart 文档驱动后续实现与验证。

## 技术背景

<!--
  必须：将此处内容替换为该项目的技术细节。
  结构仅为建议，可根据迭代需要调整。
-->

**语言/版本**：Go 1.22（目标）+ Python 3.11（参考实现）  
**主要依赖**：标准库 + Go 官方工具链；Python 侧依赖 `agno/libs/agno/agno/` 及其 tests，作为行为与对照测试基线  
**存储**：本迭代聚焦库级行为迁移，不引入新的持久化依赖；沿用 Python 端已有的“外部存储适配器”作为参考，由后续存储规格单独覆盖  
**测试**：Go 侧使用 `go test ./...` 和表驱动测试；Python 侧沿用现有 pytest/内置测试，另加对照测试脚本将两端输出序列化为可比较格式  
**目标平台**：Linux/macOS 开发环境与常见服务器环境；Go 模块本身不依赖特定容器或编排方案  
**项目类型**：多模块库（library），以 Go 包形式暴露 Agent、Provider、Workflow 等能力  
**性能目标**：在代表性 agents 场景中，端到端延迟与 Python 版本保持同一数量级；Go 侧在 CPU/内存占用上不显著劣于 Python，并为后续优化预留基准测试钩子  
**约束条件**：保持与 Python 行为严格等价（在可控随机性前提下），并通过跨语言对照测试验证；单次迁移迭代内避免引入与存储、网络框架强绑定的新依赖  
**规模/范围**：首轮迁移覆盖 cookbook 中出现的全部官方与第三方 agents 供应商及其协作模式，重点聚焦核心路径与代表性示例；AgentOS 等高级能力在后续规格中单独规划
**当前覆盖率基线**：在本迭代已实现的 Go 包中，`go test ./... -cover` 的综合覆盖率约为 24.3%，其中 `go/session` 与 `go/workflow` 的 US1/US3 辅助代码已达到 100% 覆盖，`go/agent` 与 `go/providers` 主要集中在 US1/US3 相关函数。后续迭代将按照 T051 的任务要求逐步提升覆盖率以接近或达到 85% 目标。

## 宪章检查

*Gate：Phase 0 调研前必须通过；Phase 1 设计后需再次检查。*

- [x] **Python 参考实现与行为对齐**：本迭代以 `/Users/rex/cool.cnb/agno-Go/agno/libs/agno/agno/` 下的 `agent/`、`workflow/`、`team/`、`knowledge/`、`memory/`、`tools/`、`session/` 等模块以及 `cookbook/agents`、`cookbook/workflows`、`cookbook/teams` 示例作为行为基线。所有 Go 行为设计、数据模型和契约均以这些模块的公开行为为参照，对任何有意差异（例如 Go 特有的错误类型或资源管理方式）都将在后续实现任务中显式记录并配套测试。
- [x] **Go API 设计与包结构**：目标是围绕未来的 `/Users/rex/cool.cnb/agno-Go/go/agent`、`go/workflow`、`go/providers`、`go/knowledge`、`go/session` 等包构建清晰 API，采用显式错误返回与 `context.Context` 传递，避免直接照搬 Python 的动态调用模式。Plan 与 data-model/contracts 将以这些包边界为前提设计接口与数据结构。
- [x] **跨语言测试纪律与 85%+ 覆盖率**：规划在 Go 侧为 Agent、Provider、Workflow 等包新增单元与集成测试，并引入基于 Python 实现的对照测试脚本，在同一测试数据集上比较输出。Phase 1 的 contracts 将定义对照测试输入/输出格式，后续实现阶段通过 `go test ./...` 与覆盖率工具确保相关 Go 包覆盖率 ≥ 85%。
- [x] **性能与资源使用基线**：本计划仅在文档层面定义代表性场景（单 Agent、多 Agent 协作、长对话等）及其性能观测指标，Phase 1 data-model 中会标记需要基准测试的路径。实际基准测试与性能对比在后续实现任务中进行，并以 Python 版本为对照基线。
- [x] **安全、配置与 Telemetry**：本功能涉及的配置项主要为模型/供应商选择、超时/重试策略和 Telemetry 开关，均参考 Python 侧现有配置结构。Phase 1 将在 data-model 与 contracts 中标注需要暴露或记录的安全相关字段（如脱敏日志、错误分类），并确保不引入默认敏感信息上报。
- [x] **文档与示例对齐**：本计划要求在 `/Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/quickstart.md` 中提供面向 Go 使用者的迁移与上手指南，并在上游 `agno/cookbook/agents`、`cookbook/workflows`、`cookbook/teams` 对应位置规划 Go 示例的镜像版本。支持矩阵与差异点由后续文档任务在 agno 仓库与 Go 仓库中同步更新。

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
<!--
  必须：将占位树替换为本仓库的真实结构，补上真实路径（如 go/agent、go/workflow 等）。交付的计划中不得保留示例占位。
-->

```text
/Users/rex/cool.cnb/agno-Go/
├── agno/                                   # 上游 Python 项目镜像
│   ├── libs/agno/agno/                     # 核心 Python 包与子模块（行为基线）
│   └── cookbook/                           # 示例与教程（迁移对照样本）
├── specs/                                  # speckit 规格、计划与任务（本仓库新增）
│   └── 001-migrate-agno-agents/
│       ├── spec.md                         # 功能规格
│       ├── plan.md                         # 本实现计划
│       ├── research.md                     # Phase 0 调研输出
│       ├── data-model.md                   # Phase 1 数据模型设计
│       ├── quickstart.md                   # Phase 1 快速上手文档
│       └── contracts/                      # Phase 1 API/契约描述
└── .specify/
    ├── scripts/                            # speckit 辅助脚本（plan/tasks 等）
    └── templates/                          # 规格/计划/任务模板
```

**结构决策**：当前仓库以 `/Users/rex/cool.cnb/agno-Go/agno` 作为 Python 行为基线，以 `/Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents` 作为本功能的文档与规划根目录。后续 Go 源码将按宪章要求新增到 `/Users/rex/cool.cnb/agno-Go/go/agent`、`go/workflow`、`go/providers` 等包中，但本次 `/speckit.plan` 仅规划结构与契约，不创建具体实现文件。

## 复杂度追踪

> **仅当宪章检查存在违规且必须说明时填写**

| 违规项 | 必要原因 | 更简单方案被拒绝的理由 |
|--------|----------|--------------------------|
| （暂无） | 当前实现阶段尚未发现需要违反宪章原则的设计 | 如未来在性能、存储或部署方面需要特殊例外，将在此处记录并给出权衡说明 |

当前迭代尚未引入与宪章明显冲突的设计决策；后续如在 providers-inventory、迁移策略或性能基准中需要额外复杂度，将在更新计划时补充具体违规项与权衡理由。
