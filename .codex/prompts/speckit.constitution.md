---
description: 根据交互式或已提供的原则输入创建/更新项目宪章，并确保所有依赖模板保持同步。
handoffs: 
  - label: 构建规格
    agent: speckit.specify
    prompt: 基于更新后的宪章实现功能规格。我想构建...
---

## 用户输入

```text
$ARGUMENTS
```

如有输入，你**必须**先参考用户输入后再继续。

## 执行概览

你将更新 `.specify/memory/constitution.md` 这一项目宪章。该文件是一个模板，包含方括号占位符（如 `[PROJECT_NAME]`、`[PRINCIPLE_1_NAME]`）。你的任务是：(a) 收集或推导具体取值，(b) 精准填充模板，(c) 将修订同步到所有依赖的工件。

请按以下流程执行：

1. 加载现有宪章模板 `.specify/memory/constitution.md`。
   - 找出所有 `[ALL_CAPS_IDENTIFIER]` 形式的占位符。
   - **重要**：用户可能需要与模板不同数量的原则。若用户指定数量，请遵循其要求并更新文档。

2. 为占位符收集/推导取值：
   - 若用户输入已提供取值，直接使用。
   - 否则从现有仓库上下文（README、文档、旧宪章版本等）推断。
   - 治理日期：`RATIFICATION_DATE` 为首次通过日期（未知时请询问或标记 TODO），`LAST_AMENDED_DATE` 若本次有更改则填今日，否则沿用旧值。
   - `CONSTITUTION_VERSION` 需按语义化版本管理：
     - MAJOR：破坏性治理/原则移除或重定义
     - MINOR：新增原则/章节或显著扩充指南
     - PATCH：措辞澄清、错别字、非语义性微调
   - 若无法确定版本级别，在定稿前先说明理由。

3. 草拟更新后的宪章：
   - 替换每一个占位符，不留空白（除非项目选择暂不定义，需在文中说明）
   - 保留标题层级，若原文提示说明不再需要可删除
   - 每条原则需包含：简洁标题、阐述不可协商的规则段落/要点、必要时给出理由
   - 治理章节需明确修订流程、版本策略、合规审查期望

4. 一致性同步检查（将旧检查表转换为当前验证项）：
   - 阅读 `.specify/templates/plan-template.md`，确保其中的 “Constitution Check” 或相关规则与新原则一致
   - 阅读 `.specify/templates/spec-template.md`，若宪章新增/移除强制章节或约束，需同步更新
   - 阅读 `.specify/templates/tasks-template.md`，确认任务分类能体现新增/移除的原则驱动任务类型（如可观测性、版本化、测试纪律）
   - 检查 `.specify/templates/commands/*.md`（包含本文件）是否仍有过时引用（例如特定代理名）。若需一般化，立即更新
   - 检查运行时指导文档（如 `README.md`、`docs/quickstart.md`、特定代理指导文件），若引用了被改动的原则需同步更新

5. 生成同步影响报告（以 HTML 注释形式写在宪章文件顶部）：
   - 版本变更：旧 → 新
   - 被修改的原则（旧标题 → 新标题，如有更名）
   - 新增章节
   - 移除章节
   - 需更新的模板（✅ 已更新 / ⚠ 待处理）及路径
   - 若仍留占位符未填，列出后续 TODO

6. 最终检查：
   - 文中不得有未解释的方括号占位
   - 版本信息与影响报告一致
   - 日期使用 ISO 格式（YYYY-MM-DD）
   - 原则表述需具备可执行性，避免含糊（将 “should” 转为 MUST/SHOULD 并说明理由）

7. 将完成的宪章写回 `.specify/memory/constitution.md`。

8. 输出总结：
   - 新版本号及升级理由
   - 需要人工跟进的文件
   - 建议的提交信息（如 `docs: amend constitution to vX.Y.Z (principle additions + governance update)`）

## 格式与风格要求

- 保持模板中的 Markdown 标题级别
- 长段落尽量换行提升可读性（不必硬性 80 列）
- 各章节之间保持一个空行
- 末尾不要留多余空格

若用户只提供部分修改（如仅更新一条原则），也必须执行上述校验与版本判断。

若缺少关键信息（如确实不知道 ratification date），请写成 `TODO(<字段>): 说明`，并在同步报告中标记为待办。

不要新建模板，只能操作现有 `.specify/memory/constitution.md`。
