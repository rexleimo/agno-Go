---
description: 使用 plan 模板执行实现规划工作流，生成设计工件。
handoffs: 
  - label: 创建任务
    agent: speckit.tasks
    prompt: 将计划拆解为任务
    send: true
  - label: 创建检查表
    agent: speckit.checklist
    prompt: 为以下领域生成检查表...
---

## 用户输入

```text
$ARGUMENTS
```

如有输入，你**必须**先参考用户输入后再继续。

## 执行概览

1. **环境准备**：在仓库根目录运行 `.specify/scripts/bash/setup-plan.sh --json`，解析 FEATURE_SPEC、IMPL_PLAN、SPECS_DIR、BRANCH。若参数含单引号（如 "I'm Groot"），需写成 `I'\''m Groot`，或改用双引号。

2. **加载上下文**：读取 FEATURE_SPEC 与 `.specify/memory/constitution.md`，并加载（已复制好的）IMPL_PLAN 模板。

3. **执行规划流程**：根据 IMPL_PLAN 模板的结构：
   - 填写 Technical Context（未知项标记 “NEEDS CLARIFICATION”）
   - 将宪章内容同步到 Constitution Check
   - 执行 gate 校验（若无法解释的违反项直接报错）
   - Phase 0：生成 research.md（解决所有 NEEDS CLARIFICATION）
   - Phase 1：生成 data-model.md、contracts/、quickstart.md
   - Phase 1：运行代理脚本更新 agent context
   - 设计完成后再次评估 Constitution Check

4. **停止并汇报**：命令在 Phase 2 规划后结束，需报告分支、IMPL_PLAN 路径与已生成的工件。

## 分阶段说明

### Phase 0：大纲与调研

1. **从 Technical Context 中提取未知项**：
   - 每个 NEEDS CLARIFICATION → research 任务
   - 每个依赖 → best practices 任务
   - 每个集成 → patterns 任务

2. **生成并派发调研代理**：

   ```text
   对每个未知项：
     Task: "Research {unknown} for {feature context}"
   对每个技术选型：
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **整理研究结果**到 `research.md`，格式：
   - Decision: [选择]
   - Rationale: [原因]
   - Alternatives considered: [评估过的其他方案]

**输出**：覆盖所有 NEEDS CLARIFICATION 的 research.md

### Phase 1：设计与契约

**前置条件**：`research.md` 完成

1. **从规格提取实体** → `data-model.md`：
   - 实体名称、字段、关系
   - 根据需求得出的校验规则
   - 若适用，加入状态流转

2. **生成 API 契约**：
   - 基于每个用户行为 → 端点
   - 采用标准 REST/GraphQL 模式
   - 输出 OpenAPI/GraphQL 描述至 `/contracts/`

3. **更新代理上下文**：
   - 运行 `.specify/scripts/bash/update-agent-context.sh codex`
   - 脚本会检测当前使用的 AI 代理
   - 仅追加本次计划中的新增技术
   - 保留 markers 之间的人工增补

**输出**：data-model.md、/contracts/*、quickstart.md、代理特定上下文文件

## 关键规则

- 所有路径使用绝对路径
- 对 gate 检查失败或未解决澄清项必须立即报错
