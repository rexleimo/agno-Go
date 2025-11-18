---
description: 在生成 tasks.md 之后，对 spec.md、plan.md、tasks.md 进行只读的跨文档一致性与质量分析。
---

## 用户输入

```text
$ARGUMENTS
```

如有输入，你**必须**先参考用户输入后再继续。

## 目标

在实施前识别三个核心工件（`spec.md`、`plan.md`、`tasks.md`）之间的矛盾、重复、模糊与未充分定义项。本命令仅允许在 `/speckit.tasks` 成功生成完整 `tasks.md` 后运行。

## 运行约束

**严格只读**：禁止修改任何文件，只能输出结构化分析报告。如需修复方案，须提示用户并待其明确同意后再运行其他编辑命令。

**宪章优先**：项目宪章（`.specify/memory/constitution.md`）在本分析范围内不可被质疑。凡与宪章 MUST 原则冲突的问题，一律判为 CRITICAL，必须通过调整 spec/plan/tasks 解决，而非淡化、曲解或忽略。若需修改宪章，应在 `/speckit.analyze` 之外单独进行。

## 执行步骤

### 1. 初始化上下文

在仓库根目录运行 `.specify/scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks`，解析 JSON 中的 FEATURE_DIR 与 AVAILABLE_DOCS，得到绝对路径：

- SPEC = FEATURE_DIR/spec.md
- PLAN = FEATURE_DIR/plan.md
- TASKS = FEATURE_DIR/tasks.md

若任一必备文件缺失，则直接报错并提示用户补齐前置命令。若参数含单引号（如 "I'm Groot"），需写成 `I'\''m Groot`，或改用双引号。

### 2. 渐进式加载工件

仅读取必要的最小上下文：

**来自 spec.md**：
- Overview/Context
- Functional Requirements
- Non-Functional Requirements
- User Stories
- Edge Cases（若存在）

**来自 plan.md**：
- 架构/技术栈选择
- 数据模型引用
- 阶段划分
- 技术约束

**来自 tasks.md**：
- 任务 ID
- 描述
- 所属阶段
- 并行标记 [P]
- 涉及的文件路径

**来自宪章**：
- 加载 `.specify/memory/constitution.md` 中的原则

### 3. 构建语义模型

（不在输出中粘贴原文）
- **需求清单**：为每个功能/非功能需求生成稳定 key（如 “User can upload file” → `user-can-upload-file`）
- **用户故事/动作清单**：列出所有可验收的用户行为
- **任务覆盖映射**：推断任务与需求/故事的关联（可用关键词或显式 ID）
- **宪章规则集**：提取原则名称及其中 MUST/SHOULD 语句

### 4. 检测流程（聚焦高信号）

发现上限 50 条，其余写入溢出摘要。

A. **重复**：找出近似重复的需求，并标注较弱版本待合并

B. **模糊**：定位含模糊形容词（fast、secure、intuitive 等）但缺乏量化的条目，或存在 TODO/TKTK/??? 等占位符

C. **定义不足**：动词明确但缺乏对象/指标的需求、缺少验收标准的用户故事、引用了 spec/plan 中未定义组件的任务

D. **宪章对齐**：凡与 MUST 原则冲突、或缺失宪章要求章节/质量门的情况

E. **覆盖缺口**：
- 没有任何任务关联的需求
- 没有关联需求/故事的任务
- 非功能需求在任务中未体现（如性能、安全）

F. **不一致**：
- 同一概念的术语漂移
- plan 中提到的数据实体在 spec 中缺失（或反之）
- 任务顺序与依赖矛盾（如在基础任务前执行集成任务）
- 互相冲突的要求（如一处指明用 Next.js，另一处指定用 Vue）

### 5. 评定严重级别

- **CRITICAL**：违反宪章 MUST、缺少核心工件、阻塞基本功能的零覆盖需求
- **HIGH**：重复/冲突需求、模糊的安全/性能属性、无法验证的验收标准
- **MEDIUM**：术语漂移、非功能任务缺失、边界情况定义不足
- **LOW**：风格/措辞问题、对执行顺序影响较小的冗余

### 6. 生成精简报告

以 Markdown 输出：

```
## Specification Analysis Report

| ID | Category | Severity | Location(s) | Summary | Recommendation |
|----|----------|----------|-------------|---------|----------------|
| A1 | Duplication | HIGH | spec.md:L120-134 | ... | ... |
```

- 每个发现一行，ID 以类别首字母+序号表示

**覆盖摘要表**：

```
| Requirement Key | Has Task? | Task IDs | Notes |
|-----------------|-----------|----------|-------|
```

额外块：
- **Constitution Alignment Issues**（如有）
- **Unmapped Tasks**（如有）

**指标**：
- Total Requirements
- Total Tasks
- Coverage %（>=1 任务的需求比例）
- Ambiguity Count
- Duplication Count
- Critical Issues Count

### 7. 下一步建议

在报告末尾输出 “Next Actions” 区块：
- 若存在 CRITICAL，建议在 `/speckit.implement` 前先修复，并指出应运行的命令（如 `/speckit.specify`、手动编辑 tasks.md 等）
- 若仅 LOW/MEDIUM，可提示可继续但需改善

### 8. 提供修复意向

最后询问用户：“需要我针对前 N 个问题给出具体修复建议吗？”（不要自动应用）。

## 运行原则

- **上下文节省**：仅聚焦高信号问题
- **渐进披露**：按需读取文件
- **输出轻量化**：列表 ≤50 行，超出则摘要
- **结果可重现**：相同输入应产出相同 ID 与统计

## 分析准则

- 永远不修改文件
- 不臆造不存在的章节
- 优先处理宪章冲突
- 以实例支撑结论，不给笼统规则
- 若完全无问题，仍需输出成功报告与覆盖数据

## 上下文

$ARGUMENTS
