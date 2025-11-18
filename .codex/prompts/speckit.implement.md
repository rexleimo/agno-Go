---
description: 执行实现计划，处理并完成 tasks.md 中定义的所有任务。
---

## 用户输入

```text
$ARGUMENTS
```

如有输入，你**必须**先参考用户输入后再继续。

## 执行概览

1. 在仓库根目录运行 `.specify/scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks`，解析 FEATURE_DIR 与 AVAILABLE_DOCS。所有路径必须是绝对路径。若参数包含单引号（如 "I'm Groot"），需写成 `I'\''m Groot`，或可行时改用双引号。

2. **检查 checklists 状态**（若 FEATURE_DIR/checklists/ 存在）：
   - 扫描 checklists/ 目录内的所有检查表文件
   - 对每个检查表统计：
     - 总项数：所有符合 `- [ ]`、`- [X]` 或 `- [x]` 的行
     - 已完成项：匹配 `- [X]` 或 `- [x]` 的行
     - 未完成项：匹配 `- [ ]` 的行
   - 生成如下状态表：

     ```text
     | Checklist | Total | Completed | Incomplete | Status |
     |-----------|-------|-----------|------------|--------|
     | ux.md     | 12    | 12        | 0          | ✓ PASS |
     | test.md   | 8     | 5         | 3          | ✗ FAIL |
     | security.md | 6   | 6         | 0          | ✓ PASS |
     ```

   - 计算整体状态：
     - **PASS**：所有检查表未完成项为 0
     - **FAIL**：任一检查表存在未完成项

   - **若有检查表未完成**：
     - 输出带有未完成数量的表格
     - **停止**并询问：“部分检查表尚未完成。仍要继续实现吗？(yes/no)”
     - 等待用户回复
     - 若用户回答 “no / wait / stop” 等同拒绝，则终止执行
     - 若用户回答 “yes / proceed / continue”，再进入步骤 3

   - **若全部完成**：
     - 展示所有检查表通过的表格
     - 自动进入步骤 3

3. 加载并分析实现上下文：
   - **必读**：tasks.md，用于获取完整任务列表与执行计划
   - **必读**：plan.md，了解技术栈、架构、文件结构
   - **若存在**：data-model.md，了解实体与关系
   - **若存在**：contracts/，获取 API 规格与测试要求
   - **若存在**：research.md，了解技术决策与约束
   - **若存在**：quickstart.md，获取集成场景

4. **项目设置校验**：
   - **必需**：根据实际项目情况创建/校验忽略文件

   **探测与创建逻辑**：
   - 运行以下命令判断当前是否为 git 仓库（若是则需要创建/校验 .gitignore）：

     ```sh
     git rev-parse --git-dir 2>/dev/null
     ```

   - 若存在 Dockerfile* 或 plan.md 中提到 Docker → 创建/校验 .dockerignore
   - 若存在 .eslintrc* → 创建/校验 .eslintignore
   - 若存在 eslint.config.* → 检查配置中的 `ignores` 是否覆盖所需模式
   - 若存在 .prettierrc* → 创建/校验 .prettierignore
   - 若存在 .npmrc 或 package.json → 若需要发布，则创建/校验 .npmignore
   - 若存在 Terraform 文件（*.tf）→ 创建/校验 .terraformignore
   - 若需要 .helmignore（存在 helm chart）→ 创建/校验 .helmignore

   **若忽略文件已存在**：确认包含关键模式，若缺少则附加
   **若忽略文件缺失**：按检测到的技术栈写入完整模式

   **常见技术的模式（参考 plan.md 的技术栈）**：
   - **Node.js/JavaScript/TypeScript**：`node_modules/`、`dist/`、`build/`、`*.log`、`.env*`
   - **Python**：`__pycache__/`、`*.pyc`、`.venv/`、`venv/`、`dist/`、`*.egg-info/`
   - **Java**：`target/`、`*.class`、`*.jar`、`.gradle/`、`build/`
   - **C#/.NET**：`bin/`、`obj/`、`*.user`、`*.suo`、`packages/`
   - **Go**：`*.exe`、`*.test`、`vendor/`、`*.out`
   - **Ruby**：`.bundle/`、`log/`、`tmp/`、`*.gem`、`vendor/bundle/`
   - **PHP**：`vendor/`、`*.log`、`*.cache`、`*.env`
   - **Rust**：`target/`、`debug/`、`release/`、`*.rs.bk`、`*.rlib`、`*.prof*`、`.idea/`、`*.log`、`.env*`
   - **Kotlin**：`build/`、`out/`、`.gradle/`、`.idea/`、`*.class`、`*.jar`、`*.iml`、`*.log`、`.env*`
   - **C++**：`build/`、`bin/`、`obj/`、`out/`、`*.o`、`*.so`、`*.a`、`*.exe`、`*.dll`、`.idea/`、`*.log`、`.env*`
   - **C**：`build/`、`bin/`、`obj/`、`out/`、`*.o`、`*.a`、`*.so`、`*.exe`、`Makefile`、`config.log`、`.idea/`、`*.log`、`.env*`
   - **Swift**：`.build/`、`DerivedData/`、`*.swiftpm/`、`Packages/`
   - **R**：`.Rproj.user/`、`.Rhistory`、`.RData`、`.Ruserdata`、`*.Rproj`、`packrat/`、`renv/`
   - **通用**：`.DS_Store`、`Thumbs.db`、`*.tmp`、`*.swp`、`.vscode/`、`.idea/`

   **工具专属模式**：
   - **Docker**：`node_modules/`、`.git/`、`Dockerfile*`、`.dockerignore`、`*.log*`、`.env*`、`coverage/`
   - **ESLint**：`node_modules/`、`dist/`、`build/`、`coverage/`、`*.min.js`
   - **Prettier**：`node_modules/`、`dist/`、`build/`、`coverage/`、`package-lock.json`、`yarn.lock`、`pnpm-lock.yaml`
   - **Terraform**：`.terraform/`、`*.tfstate*`、`*.tfvars`、`.terraform.lock.hcl`
   - **Kubernetes/k8s**：`*.secret.yaml`、`secrets/`、`.kube/`、`kubeconfig*`、`*.key`、`*.crt`

5. 解析 tasks.md 的结构，提取：
   - **任务阶段**：Setup、Tests、Core、Integration、Polish
   - **任务依赖**：串行或并行执行规则
   - **任务详情**：ID、描述、文件路径、并行标记 [P]
   - **执行流程**：任务顺序与依赖要求

6. 按任务计划执行实现：
   - **逐阶段执行**：完成当前阶段后才能进入下一阶段
   - **遵守依赖**：串行任务按顺序执行，带 [P] 的任务可并行
   - **遵循 TDD**：需要测试的地方先写测试
   - **按文件协调**：修改同一文件的任务需串行处理
   - **阶段验收**：每完成一个阶段都需校验

7. 实施规则：
   - **先做搭建**：初始化结构、依赖与配置
   - **先测后码**：在需要时先写 contract/entity/integration 测试
   - **核心开发**：实现模型、服务、CLI、端点
   - **集成工作**：数据库连接、中间件、日志、外部服务
   - **打磨与验证**：单测、性能优化、文档

8. 进度追踪与错误处理：
   - 每完成一个任务就汇报进展
   - 如串行任务失败则停止
   - 对并行任务 [P]：成功的继续，失败的单独汇报
   - 错误信息需包含上下文以便排查
   - 若无法继续，实现阻塞时需给出下一步建议
   - **重要**：完成的任务要在 tasks.md 中改为 `[X]`

9. 完成校验：
   - 确认所有必需任务已完成
   - 检查实现是否符合原始规格
   - 确认测试通过且覆盖率满足要求
   - 确认实现遵循技术计划
   - 汇报最终状态与工作摘要

注意：该命令假定 tasks.md 中已有完整的任务拆解。若任务缺失或不完整，请先运行 `/speckit.tasks` 重新生成清单。
