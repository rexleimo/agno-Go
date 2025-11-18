---
description: 根据自然语言的功能描述来创建或更新功能规格。
handoffs: 
  - label: 构建技术计划
    agent: speckit.plan
    prompt: 为该规格创建实施计划。我正在构建...
  - label: 澄清规格需求
    agent: speckit.clarify
    prompt: 澄清规格中的需求
    send: true
---

## 用户输入

```text
$ARGUMENTS
```

如有输入，你**必须**先参考用户输入后再继续。

## 执行概览

触发命令中 `/speckit.specify` 后面的文字**即是**功能描述。即使下方仍然看到字面上的 `$ARGUMENTS`，也假设这些描述始终在当前对话中可用。除非用户命令为空，否则不要要求他们重复。

拿到该功能描述后，按以下步骤执行：

1. **为分支生成一个简洁的短名**（2-4 个单词）：
   - 分析描述提炼最重要的关键词
   - 组合出 2-4 个词的短名以概括该功能
   - 尽量使用“动作-名词”格式（如 "add-user-auth"、"fix-payment-bug"）
   - 保留技术术语与缩写（OAuth2、API、JWT 等）
   - 名称要简洁但足够描述性，便于一眼理解
   - 示例：
     - “I want to add user authentication” → “user-auth”
     - “Implement OAuth2 integration for the API” → “oauth2-api-integration”
     - “Create a dashboard for analytics” → “analytics-dashboard”
     - “Fix payment processing timeout bug” → “fix-payment-timeout”

2. **在创建新分支前，先检查是否已有同名分支**：
   
   a. 先拉取所有远程分支，确保信息最新：
      ```bash
      git fetch --all --prune
      ```
   
   b. 在所有来源中查找该短名对应的最大功能编号：
      - 远程分支：`git ls-remote --heads origin | grep -E 'refs/heads/[0-9]+-<short-name>$'`
      - 本地分支：`git branch | grep -E '^[* ]*[0-9]+-<short-name>$'`
      - 规格目录：查找 `specs/[0-9]+-<short-name>`
   
   c. 确定下一个可用编号：
      - 汇总上述三处查到的所有数字
      - 取最大值 N
      - 新分支编号为 N+1
   
   d. 运行脚本 `.specify/scripts/bash/create-new-feature.sh --json "$ARGUMENTS"`，并传入计算出的编号与短名：
      - 同时提供 `--number N+1` 与 `--short-name "your-short-name"`，并附上功能描述
      - Bash 示例：`.specify/scripts/bash/create-new-feature.sh --json "$ARGUMENTS" --json --number 5 --short-name "user-auth" "Add user authentication"`
      - PowerShell 示例：`.specify/scripts/bash/create-new-feature.sh --json "$ARGUMENTS" -Json -Number 5 -ShortName "user-auth" "Add user authentication"`
   
   **重要提示**：
   - 必须同时检查远程分支、本地分支与 specs 目录三类来源，才能找到最大编号
   - 仅匹配短名完全一致的分支/目录
   - 若该短名不存在任何分支/目录，则从 1 开始编号
   - 每个功能只能运行该脚本一次
   - 脚本会在终端输出 JSON，务必根据 JSON 获取所需信息
   - JSON 输出中包含 BRANCH_NAME 与 SPEC_FILE 路径
   - 参数中若包含单引号（如 "I'm Groot"），需用 `I'\''m Groot` 形式转义，或在可行时改用双引号 "I'm Groot"

3. 读取 `.specify/templates/spec-template.md`，了解所需章节。

4. 按以下执行流程：

    1. 从 Input 解析用户描述
       若为空：报错 "No feature description provided"
    2. 从描述中提取关键概念
       识别：参与者、行为、数据、约束
    3. 对不明确的部分：
       - 基于上下文与业界常规做出合理推断
       - 仅在以下情况使用 [NEEDS CLARIFICATION: 具体问题] 标记：
         - 该选择对功能范围或用户体验影响重大
         - 存在多种同样合理但含义不同的解释
         - 不存在合理默认值
       - **限制**：整个规格最多 3 个 [NEEDS CLARIFICATION] 标记
       - 优先级：范围 > 安全/隐私 > 用户体验 > 技术细节
    4. 填写 “User Scenarios & Testing” 部分
       若无法确定清晰的用户流程：报错 "Cannot determine user scenarios"
    5. 生成功能性需求
       每条需求都必须可测试
       未指明的细节用合理默认值，并在 Assumptions 中记录假设
    6. 定义成果衡量标准（Success Criteria）
       产出可衡量、与技术无关的结果
       同时包含量化指标（时间、性能、量级）与定性指标（用户满意度、任务完成度）
       每项标准都应在不了解实现细节的情况下即可验证
    7. 识别关键实体（若涉及数据）
    8. 返回：SUCCESS（规格已可进入规划阶段）

5. 将规格写入 SPEC_FILE，并沿用模板结构：用功能描述（参数）提炼出的具体信息替换占位符，同时保持章节顺序与标题不变。

6. **规格质量校验**：完成初稿后，按质量准则进行验证：

   a. **创建规格质量检查表**：使用 checklist 模板结构，在 `FEATURE_DIR/checklists/requirements.md` 生成以下校验内容：

      ```markdown
      # Specification Quality Checklist: [FEATURE NAME]
      
      **Purpose**: Validate specification completeness and quality before proceeding to planning
      **Created**: [DATE]
      **Feature**: [Link to spec.md]
      
      ## Content Quality
      
      - [ ] No implementation details (languages, frameworks, APIs)
      - [ ] Focused on user value and business needs
      - [ ] Written for non-technical stakeholders
      - [ ] All mandatory sections completed
      
      ## Requirement Completeness
      
      - [ ] No [NEEDS CLARIFICATION] markers remain
      - [ ] Requirements are testable and unambiguous
      - [ ] Success criteria are measurable
      - [ ] Success criteria are technology-agnostic (no implementation details)
      - [ ] All acceptance scenarios are defined
      - [ ] Edge cases are identified
      - [ ] Scope is clearly bounded
      - [ ] Dependencies and assumptions identified
      
      ## Feature Readiness
      
      - [ ] All functional requirements have clear acceptance criteria
      - [ ] User scenarios cover primary flows
      - [ ] Feature meets measurable outcomes defined in Success Criteria
      - [ ] No implementation details leak into specification
      
      ## Notes
      
      - Items marked incomplete require spec updates before `/speckit.clarify` or `/speckit.plan`
      ```

   b. **执行校验**：逐条审阅规格与检查表：
      - 对每一项判定是否通过
      - 记录发现的问题（引用规格中的对应段落）

   c. **处理校验结果**：

      - **若全部通过**：勾选检查表并跳至步骤 6

      - **若存在失败项（不含 [NEEDS CLARIFICATION]）**：
        1. 列出失败项与对应问题
        2. 更新规格以修复问题
        3. 重新执行校验，直至全部通过（最多 3 次）
        4. 若 3 次后仍未通过，在检查表的 Notes 中记录剩余问题，并提醒用户

      - **若仍存在 [NEEDS CLARIFICATION] 标记**：
        1. 提取规格中的所有 [NEEDS CLARIFICATION: ...] 标记
        2. **数量限制**：若超过 3 个，仅保留影响最大的 3 个（按范围/安全/体验排序），其余的需做出合理假设
        3. 对每个待澄清问题（最多 3 个）按以下格式呈现给用户：

           ```markdown
           ## Question [N]: [Topic]
           
           **Context**: [Quote relevant spec section]
           
           **What we need to know**: [Specific question from NEEDS CLARIFICATION marker]
           
           **Suggested Answers**:
           
           | Option | Answer | Implications |
           |--------|--------|--------------|
           | A      | [First suggested answer] | [What this means for the feature] |
           | B      | [Second suggested answer] | [What this means for the feature] |
           | C      | [Third suggested answer] | [What this means for the feature] |
           | Custom | Provide your own answer | [Explain how to provide custom input] |
           
           **Your choice**: _[Wait for user response]_
           ```

        4. **关键——表格格式**：必须保证 Markdown 表格格式正确：
           - 竖线两侧保持空格
           - 示例：`| Content |` 而非 `|Content|`
           - 表头分隔线至少 3 个连字符：`|--------|`
           - 在 Markdown 预览中确认表格可正常渲染
        5. 问题需按顺序编号（Q1、Q2、Q3，最多 3 个）
        6. 在等待用户回答前，将所有问题一次性列出
        7. 等待用户以 “Q1: A, Q2: Custom - [说明], Q3: B” 这类格式给出选择
        8. 根据用户回答，将规格中的每个 [NEEDS CLARIFICATION] 标记替换为对应答案
        9. 全部澄清后再次运行校验

   d. **更新检查表**：每次校验后，在检查表文件中记录当下的通过/失败情况

7. 报告完成情况，包括分支名、规格文件路径、检查表结果，以及是否已准备好进入下一阶段（`/speckit.clarify` 或 `/speckit.plan`）。

**注意**：脚本会自动创建并切换到新分支，并在写入前初始化规格文件。

## 通用指南

## 快速提示

- 聚焦用户需求与业务价值的 **WHAT** 与 **WHY**。
- 避免描述实现方式（不给出技术栈、API、代码结构）。
- 面向业务干系人撰写，而非开发者。
- 不要在规格中嵌入任何检查表，那将由单独命令生成。

### 章节要求

- **必填章节**：所有功能都必须填写
- **可选章节**：仅在与功能相关时包含
- 若某章节不适用，请直接删除该章节，而不是写 “N/A”

### 面向 AI 生成

当根据用户提示生成规格时：

1. **进行合理推断**：参考上下文、行业标准与常见模式补全缺口
2. **记录假设**：将在 Assumptions 中记录的合理默认值
3. **限制澄清数量**：最多 3 个 [NEEDS CLARIFICATION]，仅针对：
   - 对功能范围或用户体验影响重大的决定
   - 存在多个同样合理但含义不同的解释
   - 没有任何合理默认值
4. **澄清优先级**：范围 > 安全/隐私 > 用户体验 > 技术细节
5. **以测试思维审视**：任何模糊的需求都会导致“可测试且无歧义”检查项不通过
6. **常见需澄清的领域**（只有在缺乏合理默认值时才提问）：
   - 功能范围与边界（明确包含/排除的用例）
   - 用户类型与权限（如存在多种互斥解释）
   - 安全/合规要求（涉及法律或金融风险）

**合理默认值示例**（无需询问）：

- 数据保留：遵循对应领域的行业标准
- 性能目标：除非另有说明，采用常规 Web/移动应用预期
- 错误处理：友好的提示语与合理兜底
- 认证方式：Web 场景默认 session 或 OAuth2
- 集成模式：默认 RESTful API，除非另有说明

### 成功标准指南

成功标准必须满足：

1. **可量化**：包含明确指标（时间、百分比、数量、速率）
2. **与技术无关**：不提及框架、语言、数据库或工具
3. **以用户为中心**：用用户/业务角度描述结果，而非系统内部实现
4. **可验证**：在不了解实现细节时也能进行验证/测试

**良好示例**：

- “用户可以在 3 分钟内完成下单”
- “系统可同时支撑 10,000 名并发用户”
- “95% 的搜索会在 1 秒内返回结果”
- “任务完成率提升 40%”

**反面示例**（过于实现导向）：

- “API 响应时间小于 200ms”（过度技术化，换成“用户几乎即时看到结果”）
- “数据库可处理 1000 TPS”（实现细节，改用用户可感知的指标）
- “React 组件渲染效率高”（绑定具体框架）
- “Redis 缓存命中率达到 80% 以上”（过于技术细节）
