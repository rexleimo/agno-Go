<!--
版本变更：1.2.0 → 2.0.0
被修改的原则：
- Go + DDD 内核与热插拔服务 → 纯 Go Agno 核心与 Python 特性对齐
- Compose-First 可部署性、Makefile 驱动的自动化流水线 → 自动化与可重复交付（Go-only）
- GORM 数据治理与迁移（SQL 驱动矩阵） → 模型供应商迁移矩阵（ollama、gemini、openai、glm4、openrouter、siliconflow、cerebras、modelscope、groq）
- 多存储适配与数据服务矩阵 → 契约与基准验证（无 Python 运行时依赖）
- 全栈测试纪律与 85%+ 覆盖率 → 测试纪律与 85%+ 覆盖率（Go 聚焦）
- Remix + React Router V7 体验标准、Vitepress 文档与 GitHub 自动化发布 → 已移除（前端/文档站不在 Go 迁移范围）
新增章节：
- 技术栈与目录约束（聚焦 Go 重写 + Python 参考形态）
移除章节：
- Compose/多存储/前端/Vitepress 相关约束（并入或废止）
模板同步：
- ✅ .specify/templates/plan-template.md
- ✅ .specify/templates/spec-template.md
- ✅ .specify/templates/tasks-template.md
- ⚠️ .specify/templates/commands/（目录不存在，无文件可同步）
未完成的 TODO：无
-->
# Agno Go 宪章

## 核心原则

### 纯 Go Agno 核心与 Python 特性对齐
- Go 版本是唯一运行时，严禁通过 cgo/FFI/子进程桥接调用 `./agno` Python 代码；Python 代码仅用于参考与治具生成。
- 核心能力（Agent、Workflow/Step Engine、Memory/Knowledge、Tool/MCP、AgentOS API）必须以稳定的 Go 接口发布（建议 `go/internal/{agent,runtime,model,memory,tool}/` 与 `go/pkg/`），并在实现前先在 `specs/<feature>/` 记录与 Python 设计/行为的映射。
- Go 公共 API 需保持向后兼容；破坏性调整必须在规格中列出迁移指南与差异原因。

### 模型供应商迁移矩阵（ollama、gemini、openai、glm4、openrouter、siliconflow、cerebras、modelscope、groq）
- 仅上述九家供应商纳入 Go 适配器范围；每家必须实现统一的 `Chat`/`Embedding` 接口与错误规约，位于 `go/pkg/providers/<provider>/` 并通过接口在 `go/internal/model/` 路由。
- `.env.example` 必须列出各供应商必需/可选的密钥与 endpoint 变量，文档需标注行为差异（温度、token 计费、可流式能力）。
- 集成测试必须针对每个供应商运行（使用现有可用 key），验证聊天、流式/非流式与异常分支；未覆盖的接口视为不合规。

### 契约与基准验证（无 Python 运行时依赖）
- 所有行为必须以契约/治具驱动：先用 Python 参考实现生成脱敏 fixtures（`specs/<feature>/contracts/fixtures/*.json|yaml`），Go 侧以 golden/契约测试消费，禁止测试时再调用 Python。
- 与 Python 不一致时需在 `specs/<feature>/contracts/deviations.md` 记录原因与补偿控制（例如不支持某参数或响应格式差异）。
- 性能/资源门槛须在基准中固定（如 p95 latency、token 误差阈值），基准脚本与结果存放于 `specs/<feature>/artifacts/` 并纳入 CI。

### 自动化与可重复交付（Go-only）
- Makefile 是唯一入口，至少包含：`fmt`（gofmt/gofumpt）、`lint`（golangci-lint）、`test`（单元+契约）、`providers-test`（带真实 key 的集成）、`coverage`、`bench`、`gen-fixtures`、`release`；文档禁止裸命令。
- CI 与本地必须复用相同 make 目标，覆盖 fixtures 生成、签入/校验、providers 集成测试与发布工件（Go module/二进制）。
- 所有脚本必须纯 Go/标准工具链实现；禁止引入 Python 作为构建或运行依赖（允许离线一次性脚本生成 fixtures，但不可进入 runtime）。

### 测试纪律与 85%+ 覆盖率（Go 聚焦）
- `go/` 下每个包必须有 `_test.go`；契约、providers 集成、基准需纳入 `go/tests/{contract,providers,bench}/` 并在 PR 同步提交。
- `make test` + `make providers-test` + `make coverage` 必须在 CI 与本地通过，综合覆盖率 ≥85%；缺测试的 PR 不得合并，跳过测试需记录 issue 与补偿计划。
- 覆盖率报告、基准与 providers 测试日志须随 PR 附件或存于 `specs/<feature>/artifacts/`，低于阈值需立即补测或缩小改动范围。

## 技术栈与目录约束
- **语言与运行时**：Go 1.23+ 为唯一生产语言；Python 3.11+ 仅用于对照与治具生成，不可成为运行或构建依赖。
- **项目结构**：
  - `agno/`：Python 参考实现（只读，不得被 Go 运行时调用）
  - `go/`：Go 模块（`cmd/agno/`、`internal/{agent,runtime,model,memory,tool}/`、`pkg/providers/<provider>/`、`pkg/memory/`、`pkg/tools/`、`tests/{contract,providers,bench}/`）
  - `specs/<feature>/`：plan/research/data-model/contracts/quickstart/tasks 与治具/基准
  - `scripts/`：仅限 Go/标准工具辅助（治具生成、CI helper）
- **配置与密钥**：`.env.example` 列出全部供应商变量；密钥通过环境变量/secret manager 注入，禁止硬编码或提交真实 key。
- **文档与导航**：特性/供应商/契约必须在 `specs/<feature>/quickstart.md` 或等价文件记录调用示例、必要 env、差异；若需外部文档站，再行在规格中定义。
- **测试与工具链**：`go test ./...` 作为基础；`golangci-lint`、`gofumpt`、`benchstat` 为默认质量工具；不得引入与 Go 无关的构建链。

## 开发流程与质量门
1. **设计前置**：每个需求需在 `specs/<feature>/research.md` 记录 Python 行为、必要的 Go 抽象、涉及的供应商与治具计划；缺少对照的需求不得进入实现。
2. **计划 Gate**：Plan 中的 “宪章检查” 必须覆盖：纯 Go 路径/禁止桥接、供应商矩阵与 env、契约/基准与治具、Make/CI 入口、测试覆盖与密钥策略。
3. **实现 Gate**：
   - 先生成/更新 fixtures、契约测试，再实现 Go 代码；禁止直接调 Python。
   - 新增/修改供应商必须同步 `.env.example`、Makefile `providers-test`、契约/基准与文档。
   - 质量工具（fmt/lint/test/providers-test/coverage/bench）必须在 PR 中有证据（日志/工件）。
4. **发布 Gate**：`make release` 需产出 Go module/二进制、契约/基准结果、providers 测试报告与覆盖率；发布说明需列出支持的供应商版本、已知偏差与缺失功能。

## 治理
- 本宪章优先于其他指南；未覆盖内容遵循 Go 社区最佳实践。
- 修订流程：提出 RFC → 在 speckit 文档中描述影响（涵盖供应商矩阵、契约/基准、自动化与覆盖率）→ 更新宪章与相关模板 → 依据变更范围更新版本号（MAJOR：原则调整/移除；MINOR：新增原则/章节或显著扩写；PATCH：措辞或非语义微调）。
- 监督机制：PR 审查必须引用 Plan 的宪章检查项；CI 需提供 `make constitution-check` 聚合（至少 fmt/lint/test/providers-test/coverage/bench/fixture 校验）并在失败时阻塞。
- 合规审查：季度审查供应商矩阵覆盖率、治具新鲜度（距上次生成时间）、密钥管理、基准回归、覆盖率；发现违规需在两周内补齐或提交修订提案。

**Version**: 2.0.0 | **Ratified**: 2025-11-18 | **Last Amended**: 2025-11-21
