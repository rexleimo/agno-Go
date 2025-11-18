---
description: 基于现有设计文档生成可执行且按依赖排序的 tasks.md。
handoffs: 
  - label: 校验一致性
    agent: speckit.analyze
    prompt: 运行项目一致性分析
    send: true
  - label: 实施项目
    agent: speckit.implement
    prompt: 分阶段启动实现
    send: true
---

## 用户输入

```text
$ARGUMENTS
```

如有输入，你**必须**先参考用户输入后再继续。

## 执行概览

1. **环境准备**：在仓库根目录执行 `.specify/scripts/bash/check-prerequisites.sh --json`，解析 FEATURE_DIR 与 AVAILABLE_DOCS。所有路径需为绝对路径。若参数含单引号（如 "I'm Groot"），使用 `I'\''m Groot`，或视情况改用双引号。

2. **加载设计文档**：从 FEATURE_DIR 读取：
   - **必需**：plan.md（技术栈、库、结构）、spec.md（带优先级的用户故事）
   - **可选**：data-model.md（实体）、contracts/（API 端点）、research.md（决策）、quickstart.md（测试场景）
   - 提示：并非所有项目都有全部文档，按实际可用内容生成任务。

3. **执行任务生成流程**：
   - 解析 plan.md 的技术栈、依赖、项目结构
   - 解析 spec.md 的用户故事及优先级（P1、P2、P3 等）
   - 若存在 data-model.md：提取实体并映射到用户故事
   - 若存在 contracts/：将端点映射至相应用户故事
   - 若存在 research.md：提取决策以生成搭建类任务
   - 按用户故事组织任务（详见“任务生成规则”）
   - 绘制用户故事完成顺序的依赖图
   - 为每个故事提供可并行执行的示例
   - 校验任务覆盖度（每个故事都具备独立完成与测试所需的任务）

4. **生成 tasks.md**：遵循 `.specify/templates/tasks-template.md`，填入：
   - 从 plan.md 取得的正确功能名称
   - Phase 1：搭建任务（项目初始化）
   - Phase 2：基础任务（所有故事的阻塞前置）
   - Phase 3+：按 spec.md 中的优先级为每个故事单独设相位
   - 各阶段需包含：故事目标、独立测试准则、测试（如有要求）、实现任务
   - 最终阶段：Polish & cross-cutting concerns
   - 所有任务必须遵循严格清单格式（见下）
   - 为每个任务提供精确文件路径
   - Dependencies 小节要展示故事完成顺序
   - 每个故事都要提供并行执行示例
   - Implementation strategy 小节需体现“先 MVP，分步交付”

5. **输出报告**：给出生成的 tasks.md 路径与摘要：
   - 任务总数
   - 各用户故事的任务数量
   - 识别出的并行机会
   - 各故事的独立测试准则
   - 建议的 MVP 范围（通常仅故事 1）
   - 格式校验：确认所有任务都符合清单格式（复选框、ID、标签、路径）

任务生成上下文：$ARGUMENTS

tasks.md 必须“开箱即用”——每个任务都要具体到 LLM 可直接执行，无需额外上下文。

## 任务生成规则

**关键**：任务必须按用户故事组织，确保每个故事都能独立实现与测试。

**测试为可选**：仅当功能规格明确请求或用户要求 TDD 时才添加测试任务。

### 清单格式（必填）

每个任务必须严格遵循：

```text
- [ ] [TaskID] [P?] [Story?] 描述（含文件路径）
```

**格式要素**：

1. **复选框**：始终以 `- [ ]` 开头
2. **任务 ID**：按执行顺序递增（T001、T002……）
3. **[P] 标记**：仅在任务可并行时添加（不同文件、无未完成依赖）
4. **[Story] 标签**：只在用户故事阶段必填
   - 格式：[US1]、[US2] 等，对应 spec.md 中的故事
   - Setup / Foundational / Polish 阶段不加 story 标签
   - 用户故事阶段必须加
5. **描述**：明确动作与精确文件路径

**示例**：

- ✅ `- [ ] T001 Create project structure per implementation plan`
- ✅ `- [ ] T005 [P] Implement authentication middleware in src/middleware/auth.py`
- ✅ `- [ ] T012 [P] [US1] Create User model in src/models/user.py`
- ✅ `- [ ] T014 [US1] Implement UserService in src/services/user_service.py`
- ❌ `- [ ] Create User model`（缺 ID 与 Story）
- ❌ `T001 [US1] Create model`（缺复选框）
- ❌ `- [ ] [US1] Create User model`（缺 ID）
- ❌ `- [ ] T001 [US1] Create model`（缺文件路径）

### 任务组织

1. **来源：用户故事（spec.md）——主轴**
   - 每个故事（P1、P2、P3…）单独成相位
   - 将模型、服务、端点/UI 与该故事一一对应
   - 如存在测试需求，将对应测试置于该故事相位
   - 明确故事间依赖，大多数故事应保持独立

2. **来源：contracts/**
   - 每个契约/端点映射到它服务的故事
   - 若需测试：在实现前先添加 [P] 的契约测试任务

3. **来源：data-model.md**
   - 将实体映射到需要它的故事
   - 若同一实体服务多个故事，将其放到最早的相关故事或 Setup 阶段
   - 实体关系可化为相应故事中的服务层任务

4. **来源：搭建/基础设施**
   - 共享基础设施 → Phase 1
   - 阻塞性前置 → Phase 2
   - 某故事特有的搭建 → 放入该故事阶段

### 阶段结构

- **Phase 1**：Setup（项目初始化）
- **Phase 2**：Foundational（阻塞前置——必须先完成）
- **Phase 3+**：按优先级排列的用户故事（P1、P2、P3…）
  - 每个故事内部顺序：测试（如有）→ 模型 → 服务 → 端点 → 集成
  - 每个阶段都要构成一个可独立测试的增量
- **Final Phase**：Polish & Cross-Cutting Concerns
