# agno-Go 开发指引

自动汇总于所有功能计划。最后更新： 2025-11-19

## 在用技术

- Go 1.25.1：核心实现语言，`go/agent`, `go/workflow`, `go/session`, `go/providers`, `go/internal` 等包通过 Go modules 组织，要求 ≥85% `go test ./... -cover` 覆盖率。
- Python ≥3.11：仅用于行为基线与 parity，对照 `agno/libs/agno/agno/` 的 Agent、Workflow、Session、Telemetry 实现。
- 可插拔存储：通过 `go/session.Store` 接口支持内存、Sqlite、Postgres、Redis 驱动，遵循 Python `SessionRecord` schema。
- Provider/Toolkit 能力：`go/providers` 解析 YAML/JSON manifest，为 LLM、工具、Knowledge、Memory、Guardrail 提供注册入口，并在未迁移时返回 `internal/errors.CodeNotMigrated`。
- 观测与性能：`go/internal/telemetry` 输出带 `runtime=go` 标签的 RunEvent；`scripts/benchmarks` 和 `go/agent/us1_basic_coordination_bench_test.go` 记录 CPU/RSS/tokens/sec。

## 项目结构

```text
/home/rex/codes/agno-Go/
├── agno/
│   ├── libs/agno/agno/                # Python 行为基线
│   └── cookbook/                      # Cookbook 场景
├── go/
│   ├── agent/                         # Go Agent API + parity/bench tests
│   ├── workflow/                      # Workflow orchestration
│   ├── session/                       # Session 数据结构 + Store 接口
│   ├── providers/                     # Provider/Toolkit/Knowledge registry
│   ├── internal/{errors,telemetry,testutil}
│   └── examples/us1_basic_coordination
├── scripts/
│   └── ci/                            # cross-language parity 与 benchmark 脚本
├── specs/001-migrate-agno-core/       # 规格、plan、research、tasks、fixtures
└── .specify/                          # speckit 模板与检查脚本
```

## 可用命令

- `./.specify/scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks`：校验当前分支、feature 路径与 Go/Python 版本。
- `go test ./... -cover`：运行 Go 端单测、parity 测试与基准测试，覆盖率需 ≥85%。
- `python -m agno.tests.contracts.run --scenario <name>`：运行 Python 参考实现契约测试，生成 parity 对比数据。
- `./scripts/ci/cross-language-parity.sh --fixture <fixture> --python "<cmd>" --go "<cmd>"`：以 YAML/JSON fixture 驱动 Go/Python，对比 RunOutput/Metrics。
- `GO_TELEMETRY_EXPORTER=stdout go test ./go/agent -run TestUS1Bench -bench . -benchmem`：采集性能指标（p95 延迟、tokens/sec、RSS）。
- `./scripts/go-ci.sh`：CI/Fleet gate，串联 parity、benchmarks、lint/format 检查。

## 代码风格

- Go 包遵循 `context.Context` 作为首个参数、显式 `error` 返回、不可变配置 struct，避免 Python 风格的隐式 kwargs；错误需映射到 `go/internal/errors` 的枚举。
- Provider/Session/Workflow 结构体需与 `specs/001-migrate-agno-core/data-model.md` 定义字段 1:1 对齐，并保持 JSON 序列化兼容 Python schema。
- Telemetry/日志务必添加 `runtime=go` 标签和去敏字段，工具/会话配置必须通过接口注入，禁止硬编码 secrets。
- 测试采用 TDD：优先编写 parity fixture/contract 测试，再实现功能；`go/internal/testutil/parity` 封装 fixture 解析与 diff。
- Benchmark/Parity 输出使用 JSON，保持 determinism（随机种子来自 fixture），并记录在 `scripts/benchmarks/data/`。

## 最新变更

- **US1：Go Runtime 复用既有 Agent** — 在 plan 中锁定 `AgentRuntime` 结构、workflow orchestration、session record mapper、toolchain registry 以及跨语言 fixture/test，确保无 Python 依赖下运行 cookbook 场景。
- **US2：可靠性与性能观测** — 规划 `go/internal/telemetry` 扩展、`scripts/benchmarks` 数据收集与 `go/agent/us1_basic_coordination_bench_test.go`，目标 p95 延迟 ≤ Python 的 70%，100 并发 CPU < 75%。
- **US3：生态与文档支持** — 要求更新 `agno/README.md`、cookbook、`go/examples` Quickstart，发布支持矩阵、迁移 FAQ 与 issue 模板，帮助社区理解 Go/Python 差异。

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
