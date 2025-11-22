# 实现计划：Go 版 Agno 重构

**分支**：`001-go-agno-rewrite` | **日期**： 2025-11-21 | **规格**： /Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/spec.md
**输入**：来自 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/spec.md` 的功能规格

**说明**：该模板由 `/speckit.plan` 命令填充并在本次迭代中补全。

## 摘要

重构 Python 版 Agno 为纯 Go 版本，保持 Agent/AgentOS 接口、聊天/工具调用/记忆/工作流行为一致，无任何运行时 Python 依赖。需要覆盖九家模型供应商（ollama、gemini、openai、glm4、openrouter、siliconflow、cerebras、modelscope、groq）的聊天与嵌入能力，契约测试与治具从 Python 版生成并在 Go 侧消费，匹配率 ≥95%。在 100 并发、128-token 输入、10 分钟压测下 p95 延迟较 Python 降低 ≥20%、峰值常驻内存下降 ≥25%，并通过 Make 入口统一 fmt/lint/test/providers-test/coverage/bench/gen-fixtures/release 与报告产出。

## 技术背景

<!--
  必须：将此处内容替换为该项目的技术细节。
  结构仅为建议，可根据迭代需要调整。
-->

**语言/版本**： Go 1.25.1（唯一运行时），Python 3.11 仅用于离线治具生成  
**主要依赖**： 标准库 `net/http` + `github.com/go-chi/chi/v5`（路由/中间件）、自研 REST 客户端封装九家模型供应商（openai/gemini/glm4/openrouter/siliconflow/cerebras/modelscope/groq/ollama）、golangci-lint、gofumpt、benchstat  
**存储**： 抽象 `MemoryStore` 接口，默认线程安全内存实现 + 可选嵌入式 Bolt/Badger 持久化实现，用于会话历史/工具结果/向量缓存  
**数据目录**：持久化默认落在 `./data/{bolt|badger}/<namespace>`，可通过 `AGNO_DATA_DIR` 重定位；Namespace 非法字符将被过滤并回退为 `default`。Badger 支持 `retention` TTL 控制，清理需先停服后删除目录。  
**测试**： `go test`, golden/契约测试，providers 集成测试，bench 基准，golangci-lint，benchstat，覆盖率聚合  
**目标平台**： Linux server & macOS 开发机（容器/裸机均需运行），CI 参考 GitHub Actions  
**项目类型**： 后端服务 + CLI（单仓库 Go 模块，与 Python 参考共存）  
**性能目标**： 100 并发聊天会话（128 token 输入、流式）持续 10 分钟，p95 延迟较 Python 改善 ≥20%，峰值常驻内存下降 ≥25%  
**约束条件**： 纯 Go 不调用 Python/cgo/子进程；接口/错误语义与 Python 版一致；`.env.example` 列出九家供应商变量；高负载下流式不截断并有背压提示  
**规模/范围**： 九家模型供应商、至少 95% 契约匹配率、综合覆盖率 ≥85%、压测与覆盖率报告存放于 `specs/001-go-agno-rewrite/artifacts/`

## 宪章检查

*Gate：Phase 0 调研前必须通过；Phase 1 设计后需再次检查。*

- [x] **纯 Go / 禁止桥接**：迁移 Agent/Workflow/Memory/Tool/MCP/AgentOS API 至 `/Users/rex/cool.cnb/agno-Go/go/internal/{agent,runtime,model,memory,tool}/` 与 `/Users/rex/cool.cnb/agno-Go/go/pkg/`（骨架已建），禁止 cgo/FFI/子进程调用 `/Users/rex/cool.cnb/agno-Go/agno`。Python 仅作为行为参考与离线治具生成，不进入运行/构建链。
- [x] **模型供应商矩阵（ollama、gemini、openai、glm4、openrouter、siliconflow、cerebras、modelscope、groq）**：每家实现统一 `Chat`/`Embedding`/流式接口，适配器位于 `/Users/rex/cool.cnb/agno-Go/go/pkg/providers/<provider>/` 并由 `/Users/rex/cool.cnb/agno-Go/go/internal/model/` 路由；需确认各自 SDK/stream 能力与 env 变量命名（列入 research 任务），`.env.example` 覆盖必需/可选项。
- [x] **契约/治具与基准**：Python 参考在 `/Users/rex/cool.cnb/agno-Go/agno` 生成脱敏 fixtures 至 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/contracts/fixtures/`，Go 契约测试与 golden 位于 `/Users/rex/cool.cnb/agno-Go/go/tests/contract/`（占位目录），偏差记录 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/contracts/deviations.md`，压测/覆盖率报告存放 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/artifacts/`。运行与测试阶段不再调用 Python。
- [x] **自动化与 Make**：单一入口 `/Users/rex/cool.cnb/agno-Go/Makefile`（已添加）提供 `fmt lint test providers-test coverage bench gen-fixtures release constitution-check`，CI 复用同名目标；脚本仅使用 Go/标准工具，禁止 Python 依赖。
- [x] **测试纪律 + 85% 覆盖率**：所有包提供 `_test.go`，`go/tests/{contract,providers,bench}/` 覆盖契约、供应商、基准并计入 `make coverage` 聚合，目标综合覆盖率 ≥85%；缺口通过新增单元/端到端/基准测试补齐。
- [x] **密钥与安全**：`.env.example` 仅含占位与必需/可选说明，密钥通过环境变量/secret manager 注入；压测与契约报告脱敏（不含真 key/token），缺密钥的供应商测试需显式跳过并输出原因。

## 项目结构

### 文档（当前功能）

```text
/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/
├── plan.md              # 本文件（/speckit.plan 输出）
├── spec.md              # 功能规格
├── research.md          # Phase 0 输出（/speckit.plan，已生成）
├── data-model.md        # Phase 1 输出（/speckit.plan，已生成）
├── quickstart.md        # Phase 1 输出（/speckit.plan，已生成）
├── contracts/           # Phase 1 输出（/speckit.plan，openapi + deviations + fixtures 目录已建）
├── artifacts/           # 覆盖率/压测报告（目录已建）
└── tasks.md             # Phase 2 输出（/speckit.tasks，已生成）
```

### 源码（仓库根目录）
<!--
  必须：将占位树替换为真实结构，删除未用选项，补上真实路径（如 apps/admin、packages/foo）。交付的计划中不得保留 Option 标签。
-->

```text
/Users/rex/cool.cnb/agno-Go/
├── agno/                         # Python 参考实现（只读，当前存在）
│   ├── README.md 等             # Python 版文档与代码
│   └── libs/、scripts/          # 参考实现与工具
├── specs/
│   └── 001-go-agno-rewrite/     # 本次迭代规格与计划
│       ├── spec.md
│       ├── plan.md
│       └── checklists/requirements.md
├── go/                          # Go 模块根（已初始化 go.mod）
│   ├── cmd/agno/                # Go CLI/服务入口（占位 main.go）
│   ├── internal/{agent,runtime,model,memory,tool}/  # 核心引擎占位
│   ├── pkg/{providers,memory,tools}/                 # 适配器与可插拔组件占位
│   ├── tests/{contract,providers,bench}/             # 契约/供应商/基准测试占位
│   └── scripts/                                     # 纯 Go 工具脚本（如 gen_fixtures）
├── config/default.yaml          # 本地启动/基准配置占位
├── scripts/                     # Go/标准工具脚本（治具生成等，待对齐 Go-only）
├── Makefile                     # fmt/lint/test/providers-test/coverage/bench/gen-fixtures/release
└── .env.example                 # 供应商 env 占位
```

**结构决策**：沿用单仓库结构，保持 Python 参考在 `/Users/rex/cool.cnb/agno-Go/agno` 只读，`/Users/rex/cool.cnb/agno-Go/go` 已作为唯一运行目录骨架，测试/治具/报告集中在 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/`，Makefile 与 scripts 统一使用 Go/标准工具。

## 复杂度追踪

> **仅当宪章检查存在违规且必须说明时填写**

| 违规项 | 必要原因 | 更简单方案被拒绝的理由 |
|--------|----------|--------------------------|
| - | - | - |
