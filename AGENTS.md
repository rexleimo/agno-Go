# agno-Go 开发指引

自动汇总于所有功能计划。最后更新： 2025-11-21

## 在用技术
- Go 1.25.1 作为唯一运行时；Python 3.11 仅用于离线治具生成
- 标准库 `net/http` + `github.com/go-chi/chi/v5` 路由/中间件
- 自研 REST 客户端覆盖九家模型供应商（Ollama、Gemini、OpenAI、GLM4、OpenRouter、SiliconFlow、Cerebras、ModelScope、Groq）
- 质量工具：golangci-lint、gofumpt、benchstat，统一入口为 Makefile
- 记忆存储接口 `MemoryStore`，默认内存实现，提供 Bolt/Badger 可选持久化

## 项目结构

```text
/Users/rex/cool.cnb/agno-Go/
├── agno/                         # Python 参考实现（只读）
├── specs/001-go-agno-rewrite/    # 当前迭代规格、计划、契约、研究
├── go/                           # 计划新增的 Go 模块根（cmd/internal/pkg/tests）
├── scripts/                      # Go/标准工具脚本
├── Makefile                      # fmt/lint/test/providers-test/coverage/bench/gen-fixtures/release/constitution-check
├── .github/workflows/ci.yml      # 运行 make fmt/lint/test/providers-test/coverage/bench/constitution-check 并上传工件
└── .env.example                  # 待补全的供应商占位
```

## 可用命令
- `make fmt` / `make lint` / `make test` / `make providers-test` / `make coverage` / `make bench` / `make gen-fixtures` / `make release` / `make constitution-check`
- `go run ./go/cmd/agno --config /Users/rex/cool.cnb/agno-Go/config/default.yaml`（启动 AgentOS，本地）
- `./scripts/gen-fixtures.sh`（从 `specs/001-go-agno-rewrite/contracts/fixtures-src` 复制/验证治具到 fixtures；`VERIFY_ONLY=true` 时仅校验）

## 代码风格
- gofumpt 格式化，golangci-lint 默认配置；所有包需有 `_test.go`
- API 与错误语义需与 Python 版保持兼容，禁止任何运行时 Python/cgo/子进程依赖

## 最新变更
- 001-go-agno-rewrite：规划纯 Go AgentOS，定义数据模型、OpenAPI 契约、记忆存储策略与性能目标（20% p95、25% 峰值内存），新增 CI（Go 1.25.1 + make fmt/lint/test/providers-test/coverage/bench/constitution-check）与 gofumpt/golangci-lint 配置

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
