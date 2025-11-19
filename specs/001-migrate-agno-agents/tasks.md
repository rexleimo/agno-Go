# 任务清单：agno 核心 agents 能力迁移

**输入**：`/specs/001-migrate-agno-agents/` 中的设计文档  
**前置**：plan.md（必需）、spec.md（用户故事必需）、research.md、data-model.md、contracts/

## Phase 1: Setup（共享基础设施）

**目的**：为 Python–Go 行为对齐搭建基础设施，包括 Go module、基础目录结构和最小对照测试支撑。

- [X] T001 初始化 Go module 与基础目录结构（go 目录）以支持 agents 迁移（在仓库根目录创建/更新 go.mod，并创建 go/ 目录作为未来包根）。
- [X] T002 配置 Go 开发工具链与基础脚本，支持在 /Users/rex/cool.cnb/agno-Go 下执行 go test ./... 和基本静态检查（例如新增或更新 scripts/ 下的工具脚本）。
- [X] T003 [P] 在 specs/001-migrate-agno-agents/contracts/openapi-migration-parity.yaml 基础上确认对照测试 API 合同可用，并在 specs/001-migrate-agno-agents/research.md 中记录任何需要的补充约定。
- [X] T004 [P] 根据 specs/001-migrate-agno-agents/quickstart.md 中的流程，验证从 agno/cookbook 选择 Python 示例并记录场景信息的操作步骤是否清晰可执行，如需说明在 quickstart.md 内补充示例占位。

---

## Phase 2: Foundational（阻塞性前置）

**目的**：为后续所有用户故事提供稳定的 Go 抽象、测试基线与对照测试支撑，完成前其他实现任务不可开始。

- [X] T005 在 go/agent 包中定义核心 Agent 抽象与基础结构（如 Agent 接口、配置结构），与 specs/001-migrate-agno-agents/data-model.md 中的 Agent 实体保持语义一致。
- [X] T006 在 go/providers 包中定义 Provider 抽象（能力类型、配置与错误语义），与 data-model.md 中的 Provider 实体对齐。
- [X] T007 在 go/workflow 包中定义 Workflow / Collaboration Pattern 的基础数据结构与执行接口，映射 data-model.md 中 Workflow 实体的字段与关系。
- [X] T008 在 go/session 包中定义 Session/Task 抽象（状态、历史、结果字段等），与 data-model.md 的 Session 实体保持一致。
- [X] T009 [P] 在 go/internal 或等价位置引入统一错误与 Telemetry 结构（对应数据模型中的 TelemetryEvent 与 error_semantics），为后续所有调用提供统一记录方式。
- [X] T010 [P] 编写基础 Go 测试工具（例如测试数据生成函数），放置在 go/internal/testutil 或等价位置，以便各包重用。
- [X] T011 设计并在 specs/001-migrate-agno-agents/data-model.md 中补充或确认 ParityTestScenario 与 TelemetryEvent 的最小字段集合可支持对照测试和回归分析（如已有则仅校验与计划一致，无需修改结构）。

---

## Phase 3: 用户故事 1 - 现有 agno 项目分阶段迁移到 agno-Go（优先级：P1）

**目标**：以一个或少量代表性现有 agno 场景为样本，在 Go 中搭建等价场景，并通过对照测试验证行为严格等价，为后续大规模迁移提供模板。

**独立测试准则**：在仅依赖 Phase 1–2 的基础设施和该阶段任务的前提下，能够对选定的 Python 示例在 Go 中建立等价 Workflow/Session，并通过 ParityTestScenario 与 ParityRun 比较两端结果，确认关键决策路径与业务输出一致。

### 测试与对照支撑（US1）

- [X] T012 [US1] 在 agno/cookbook 中选定至少一个代表性 P1 迁移场景，并在 specs/001-migrate-agno-agents/research.md 中记录对应 ParityTestScenario 的 id、输入与预期行为说明。
- [X] T013 [P] [US1] 在 agno 目录中为选定示例编写或整理 Python 端对照入口函数（例如新增或重构到 agno/cookbook/... 的可调用脚本），确保可接受结构化输入并返回结构化输出。
- [X] T014 [P] [US1] 在 go 目录中新建与 Python 示例对应的 Go 对照入口（例如 go/agent 或 go/workflow 下的示例函数），接受与 Python 相同结构的输入并返回结构化输出以供比较。
- [X] T015 [US1] 在 scripts/ 或等价位置创建对照测试驱动脚本，按照 contracts/openapi-migration-parity.yaml 中的 ParityRun/ParityTestScenario 结构，分别调用 Python 与 Go 实现并输出统一的比较结果。

### 实现与集成（US1）

- [X] T016 [US1] 在 go/providers 下实现与选定 Python 示例所用供应商对应的 Provider 适配层，确保 config、capabilities 与 error_semantics 与 Python 行为一致。
- [X] T017 [P] [US1] 在 go/agent 下实现与示例中各个 Agent 对应的结构与行为，包括角色、输入输出约束与 memory_policy。
- [X] T018 [P] [US1] 在 go/workflow 中定义与示例等价的 Workflow/Collaboration Pattern（steps、pattern_type、routing_rules），确保与 Python 配置结构相匹配。
- [X] T019 [US1] 在 go/session 中实现用于运行选定 Workflow 的 Session 管理逻辑，支持记录 history、result 与 TelemetryEvent。
- [X] T020 [US1] 为上述 Go 包添加单元测试（*_test.go），覆盖 Agent、Provider、Workflow、Session 的主路径与边界行为，并使用 testutil 工具简化构造。
- [X] T021 [P] [US1] 使用对照测试脚本运行至少一个 ParityRun，确保所有 `must_match` 级别的场景通过，并将结果链接记录在 specs/001-migrate-agno-agents/research.md 或 plan.md 的相关部分。

---

## Phase 4: 用户故事 2 - 新用户快速搭建多智能体协作体验（优先级：P2）

**目标**：为不熟悉 Python 实现的用户提供在 Go 中快速搭建多 agents 协作场景的路径，包括文档、示例与基本观测能力。

**独立测试准则**：一名熟悉业务但不熟悉 agno 的工程师，仅依赖 quickstart.md 与 Go 示例，即可在 1 个工作日内搭建包含 2–3 个 agents、至少一种协作模式和一种外部供应商的场景，并通过自测用例完成验证。

### 文档与示例（US2）

- [X] T022 [US2] 基于 Phase 3 完成的 P1 场景，在 go 目录下抽取或整理一个简化示例，作为新用户的起步案例（例如放置在 go/examples/agents 或等价位置）。
- [X] T023 [P] [US2] 更新 specs/001-migrate-agno-agents/quickstart.md，将选定示例的具体路径与运行命令写入，使用户可以按步骤完成从 clone 到运行的流程。
- [X] T024 [P] [US2] 在 agno/cookbook 对应示例目录中添加注释性说明文件或 README 片段，指向 Go 版本示例，并描述如何在 Go 实现中获得等价体验。

### 实现与体验（US2）

- [X] T025 [US2] 在 go/agent 与 go/workflow 中设计一个面向新用户的“基础多 agents 模板”，包含 2–3 个典型角色和至少一种协作模式（如串行+并行组合）。
- [X] T026 [P] [US2] 为上述模板提供最小可运行的 Provider 配置示例（如选择一个开箱即用的模型供应商），并在 README 或 quickstart 中写明配置步骤。
- [X] T027 [P] [US2] 为该模板编写简单的自测用例（可作为 Go 测试或脚本示例），让新用户可以通过运行命令快速验证环境是否配置正确且场景能完成。

---

## Phase 5: 用户故事 3 - 团队扩展自定义 agents 供应商（优先级：P3）

**目标**：定义并演示统一的自定义扩展协议，使团队能够在 agno 与 agno-Go 之间尽可能复用扩展定义，同时保持行为与错误语义的一致。

**独立测试准则**：团队能够基于扩展协议，为一个内部系统实现自定义 Provider；在 agno 与 agno-Go 中分别集成后，对相同用例的行为、错误反馈与审计记录在业务语义上保持一致。

### 扩展协议设计（US3）

- [X] T028 [US3] 在 specs/001-migrate-agno-agents/data-model.md 与 plan.md 的基础上，起草一份自定义 Provider 扩展协议草案（可作为 specs/001-migrate-agno-agents/contracts/custom-provider-protocol.md 或等价文档），明确配置结构、接口行为与错误语义。
- [X] T029 [P] [US3] 将扩展协议映射到 Python 侧的实现约定，在 agno/libs/agno/agno 下选取一个合适位置记录如何实现自定义 Provider（例如补充或创建文档/示例）。

### 示例实现与对照（US3）

- [X] T030 [US3] 在 agno 侧实现一个基于扩展协议的示例自定义 Provider（如访问内部搜索或业务 API），并提供最小示例代码与测试。
- [X] T031 [P] [US3] 在 go/providers 中实现对应的示例 Provider，遵循相同的扩展协议，并接入 go/agent、go/workflow 的调用链。
- [X] T032 [P] [US3] 为该示例 Provider 编写跨运行时对照测试，用 ParityTestScenario 与 ParityRun 检查 Python 与 Go 的行为与错误语义是否一致。
- [X] T033 [US3] 更新 quickstart.md 或新增迁移说明文档片段，指导团队如何从零开始实现第二个自定义 Provider，并说明两端复用与差异点。

---

## Phase 6: Polish & Cross-Cutting Concerns

**目的**：在核心功能与示例完成后，统一完善文档、测试覆盖率、性能与安全，确保满足宪章要求。

- [X] T034 汇总并更新 /Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/plan.md 与 research.md，记录首批迁移供应商实际完成情况与任何例外清单。
- [X] T035 [P] 在 go 目录中检查所有相关包的测试覆盖率，运行 go test ./... -cover，确保与宪章约定的 85% 目标相符，并在 README 或 CI 报告中记录结果。
- [X] T036 [P] 根据 TelemetryEvent 数据模型，在 Go 实现中补全关键路径的 Telemetry 事件记录，并验证不会泄露敏感信息（通过检查日志或测试用例）。
- [X] T037 对 quickstart.md 和任何新增文档进行一次端到端演练，确认从“选择示例”到“完成对照测试”的流程在当前实现下可顺利完成。
- [X] T038 在 specs/001-migrate-agno-agents/providers-inventory.md 中生成并维护供应商清单：遍历 agno/cookbook 下示例，列出出现过的所有官方与第三方 agents 供应商 ID、所在示例路径与使用频次。
- [X] T039 [P] 基于 providers-inventory.md 将每个供应商标记为“首批必须迁移 / 后续批次 / 不再支持”，并在文档中记录理由与计划批次。
- [X] T040 为所有标记为“首批必须迁移”的供应商在 go/providers 下创建对应适配层（如 go/providers/<provider_id>.go 或等价结构），并确保至少具备与 Python 行为等价的最小实现。
- [X] T041 [P] 为 providers-inventory.md 中所有“首批必须迁移”的供应商定义或关联 ParityTestScenario（可复用 US1 场景或新增），确保每个供应商至少在一个对照场景中被调用。
- [X] T042 在 specs/001-migrate-agno-agents/contracts/config-mapping.md 中设计 Python → Go 的配置/声明式映射规则（涵盖 Agent、Provider、Workflow/Session 的主要字段），并给出至少一个完整示例映射对（Python 配置片段 → Go 配置片段）。
- [X] T043 [US1] 将选定的 P1 Python 示例配置按照 config-mapping.md 中的规则迁移到 Go 配置文件或构造代码中（例如 go/agent 或 go/workflow 的构造函数），并在 specs/001-migrate-agno-agents/research.md 中记录实际迁移步骤与任何偏离规则的地方。
- [X] T044 在 go/internal/errors 或等价位置定义专门表示“未迁移供应商/协作模式/AgentOS 能力”的错误类型与错误码（例如 ErrProviderNotMigrated、ErrFeatureNotAvailableInGo），并在 TelemetryEvent 的 error_semantics 中加入对应分类。
- [X] T045 在用户文档中增加“未迁移能力说明”章节（例如更新 specs/001-migrate-agno-agents/quickstart.md 或在 agno/cookbook 对应目录新增说明文件），列出当前未在 Go 侧提供的供应商/协作模式/AgentOS 能力，以及推荐替代方案或后续计划。
- [X] T046 [US1] 在 go/providers 和 go/workflow 中对请求使用未迁移供应商或协作模式的情况统一返回“未迁移错误”（使用 T044 定义的错误类型），并通过 TelemetryEvent 记录事件，避免静默失败。
- [X] T047 在 specs/001-migrate-agno-agents/plan.md 或新增文档 specs/001-migrate-agno-agents/migration-strategy.md 中设计分阶段迁移策略：包括 Python 与 Go 双运行时期的流量路由方案、灰度发布步骤与回滚策略，并根据 providers-inventory.md 和 ParityTestScenario 清单定义“首轮迁移覆盖的场景集合”。
- [X] T048 基于 ParityTestScenario 清单统计当前覆盖的 agents 协作场景占比，确保至少达到规格中要求的 80% 目标，并在 migration-strategy.md 中记录覆盖情况。
- [X] T049 [P] 针对所有标记为 `must_match` 的 ParityTestScenario 运行 ParityRun，计算通过率并验证关键业务流程的通过率达到或接近 95%，将结果和任何未通过的场景记录在 research.md 或 migration-strategy.md 中。
- [X] T050 在 scripts/ 或文档中给出示例路由配置（例如通过环境变量或配置文件切换 Python/Go 实现），并描述如何在发现 Go 侧问题时快速回滚到 Python 实现。
- [X] T051 [P] 为与本功能相关的 Go 包（例如未来的 go/agent、go/workflow、go/providers 等）编写或补充缺失的 *_test.go，用 go test ./... -coverprofile 生成覆盖率报告，并确保相关包覆盖率单独统计时达到或超过 85%，将结果路径记录在 specs/001-migrate-agno-agents/plan.md 中。
- [X] T052 [P] 为一个典型的多 agents 协作场景编写 Go benchmark 测试（例如在 go/workflow 或 go/agent 下），对比 Python 与 Go 的端到端延迟与资源使用，并在 research.md 中记录结果与任何明显差异。
- [X] T053 在内部测试或灰度阶段记录因行为不一致导致任务无法完成的严重故障样本（例如通过 TelemetryEvent 标记），统计占比并在 migration-strategy.md 中记录，对超过阈值的情况给出改进计划。

---

## Dependencies（阶段依赖与并行机会概述）

- Phase 1（Setup）无前置依赖，可立即开始；T003、T004 标记为 [P] 可并行。
- Phase 2（Foundational）依赖 Phase 1 完成，是所有后续实现工作的阻塞前置。
- Phase 3（US1）依赖 Phase 2；其中标记为 [P] 的任务（如 T013、T014、T017、T018、T021）可在不同成员之间并行，只要共享抽象稳定。
- Phase 4（US2）依赖 Phase 3 已有至少一个稳定示例，可在完成 P1 场景后开始；内部标记为 [P] 的文档与示例任务可并行。
- Phase 5（US3）依赖 Phase 2 的抽象与至少一个 P1 场景的经验，可与 Phase 4 部分任务并行推进。
- Phase 6（Polish）依赖前述所有核心功能基本完成，用于统一收尾与验证。
