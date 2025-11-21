# 任务清单：Go 版 Agno 重构

**输入**：`/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/` 中的设计文档  
**前置**：plan.md（必需）、spec.md（用户故事必需）、research.md、data-model.md、contracts/、quickstart.md  
**测试与覆盖率**：`make test`、`make providers-test`、`make coverage` 必须可运行并产出 ≥85% 综合覆盖率；契约/供应商测试需引用 `contracts/fixtures/`，基准输出至 `artifacts/bench/`。禁止任何运行时 Python 依赖。

## Phase 1: Setup（共享基础设施）

目的：完善 Go-only 工程基线、自动化入口、配置占位，解除后续阻塞。

- [X] T001 在 `/Users/rex/cool.cnb/agno-Go/go/go.mod` 补充依赖与版本约束（chi、yaml、uuid、gofumpt/golangci-lint/benchstat），运行 `go mod tidy` 生成 `go/go.sum`
- [X] T002 [P] 扩充 `/Users/rex/cool.cnb/agno-Go/Makefile`：为 lint/test/coverage/bench 添加 `help`、`coverage` 使用 `-coverpkg=./...`，将 benchstat 汇总输出到 `specs/001-go-agno-rewrite/artifacts/bench/benchstat.txt`
- [X] T003 [P] 新建 `/Users/rex/cool.cnb/agno-Go/.golangci.yml`：启用 gofumpt、revive, govet, staticcheck 等规则，确保 `make lint` 可用
- [X] T004 [P] 补充 `/Users/rex/cool.cnb/agno-Go/.env.example` 变量说明（必需/可选、默认 endpoint），修正格式并标注缺失时的行为
- [X] T005 [P] 新建 `/Users/rex/cool.cnb/agno-Go/.github/workflows/ci.yml` 复用 `make fmt lint test providers-test coverage bench constitution-check`，将覆盖率/基准产物上传至 `specs/001-go-agno-rewrite/artifacts/`
- [X] T006 [P] 在 `/Users/rex/cool.cnb/agno-Go/scripts/` 添加 Go/标准工具脚本生成脱敏 fixtures（基于 Python 参考的离线输出），落盘至 `specs/001-go-agno-rewrite/contracts/fixtures/`
- [X] T007 更新 `/Users/rex/cool.cnb/agno-Go/AGENTS.md` 与 `specs/001-go-agno-rewrite/quickstart.md`，反映实际命令与路径
- [X] T040 [P] 在 `/Users/rex/cool.cnb/agno-Go/go/scripts/gen_fixtures.go` 实现纯 Go fixture 复制/验证，并在 `/Users/rex/cool.cnb/agno-Go/Makefile` 的 `gen-fixtures` 调用中落盘到 `specs/001-go-agno-rewrite/contracts/fixtures/`

## Phase 2: Foundational（阻塞性前置）

目的：核心接口、配置与测试基线，完成后用户故事可并行。

- [X] T008 在 `/Users/rex/cool.cnb/agno-Go/go/internal/agent/types.go` 定义 Agent/Session/Message/ToolCall/状态机类型，映射 `data-model.md`
- [X] T009 [P] 在 `/Users/rex/cool.cnb/agno-Go/go/internal/memory/store.go` 扩展接口（history、tool result、token window），实现线程安全内存版于 `/Users/rex/cool.cnb/agno-Go/go/pkg/memory/memory_store.go`
- [X] T010 [P] 在 `/Users/rex/cool.cnb/agno-Go/go/internal/model/router.go` 定义 Chat/Embedding 接口、错误规约与 provider 路由器
- [X] T011 [P] 在 `/Users/rex/cool.cnb/agno-Go/go/internal/runtime/config/config.go` 实现 config/env 加载（`.env`、`config/default.yaml`），支持 provider 缺失 env 时标记 not-configured
- [X] T012 [P] 在 `/Users/rex/cool.cnb/agno-Go/go/tests/contract/fixtures_loader_test.go` 搭建契约测试基架（加载 `specs/001-go-agno-rewrite/contracts/fixtures/`，校验 token/embedding 容差）
- [X] T013 [P] 在 `/Users/rex/cool.cnb/agno-Go/go/tests/providers/env_gating_test.go` 添加供应商测试骨架：无密钥时跳过并记录原因，输出到 `specs/001-go-agno-rewrite/artifacts/coverage/`
- [X] T014 在 `/Users/rex/cool.cnb/agno-Go/go/internal/runtime/server.go` 搭建 chi 路由与 SSE/分块基础（健康检查、空 handler/501 占位），对齐 `contracts/openapi.yaml`
- [X] T015 更新 `/Users/rex/cool.cnb/agno-Go/Makefile` 的 `constitution-check` 聚合，确保 fmt/lint/test/providers-test/coverage/bench 全执行并写日志至 `specs/001-go-agno-rewrite/artifacts/`
- [X] T041 [P] 在 `/Users/rex/cool.cnb/agno-Go/Makefile` 与 CI 工作流中添加 `audit-no-python` 目标，禁止 cgo/子进程调用 `./agno`，并在 constitution-check 中执行

**检查点**：核心骨架齐备，可着手用户故事。

## Phase 3: 用户故事 1 - 启动 Go 版代理服务（P1）🎯 MVP

目标：启动 Go AgentOS，提供与 Python 版一致的接口/行为（聊天、工具调用、记忆、流式），支持九家供应商的基本能力并可独立测试。

独立测试：填 `.env`，`go run ./go/cmd/agno --config /Users/rex/cool.cnb/agno-Go/config/default.yaml`，用 cURL 调用 `/agents`→`/agents/{id}/sessions/{id}/messages?stream=true`，验证流式、工具调用与会话持久，契约/集成测试通过。

### 测试（先写后实现）
- [X] T016 [P] [US1] 在 `/Users/rex/cool.cnb/agno-Go/go/tests/contract/agentos_contract_test.go` 覆盖 create agent/session/message 流式/非流式契约，引用 fixtures 与 `contracts/openapi.yaml`
- [X] T017 [P] [US1] 在 `/Users/rex/cool.cnb/agno-Go/go/tests/providers/openai_smoke_test.go` 编写最小 provider 集成测试（已配置 key 时运行，验证流式/错误分支）
- [X] T042 [P] [US1] 在 `/Users/rex/cool.cnb/agno-Go/go/tests/runtime/health_env_test.go` 添加缺失密钥/禁用 provider 的健康检查与错误提示回归，确保返回可见错误且调用被阻止

### 实现
- [X] T018 [P] [US1] 在 `/Users/rex/cool.cnb/agno-Go/go/internal/agent/service.go` 实现 Agent/Session 管理、状态流转、消息与工具调用挂钩记忆存储
- [X] T019 [P] [US1] 在 `/Users/rex/cool.cnb/agno-Go/go/internal/memory/badger_store.go` 和 `/Users/rex/cool.cnb/agno-Go/go/internal/memory/bolt_store.go` 实现可选持久化存储（与 MemoryStore 接口对齐）
- [X] T020 [P] [US1] 在 `/Users/rex/cool.cnb/agno-Go/go/pkg/providers/openai/client.go` 建立 REST/SSE 客户端与错误映射，复用到路由器；为其他八家创建占位 `client.go` + `errors.go`
- [X] T021 [P] [US1] 在 `/Users/rex/cool.cnb/agno-Go/go/internal/runtime/server.go` 完成 `/agents`、`/agents/{id}/sessions`、`/agents/{id}/sessions/{sid}/messages`（SSE/分块）、工具启停、health 路由，挂接中间件（日志/限流/鉴权预留）
- [X] T022 [US1] 更新 `/Users/rex/cool.cnb/agno-Go/go/cmd/agno/main.go` 启动流程：加载 config/env，初始化 providers/memory/agent runtime，启动 HTTP
- [X] T023 [US1] 更新 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/contracts/fixtures/` 与 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/contracts/deviations.md`，并同步 `quickstart.md` 示例

**检查点**：US1 可独立运行，契约/基本集成测试通过。

## Phase 4: 用户故事 2 - 高并发性能验证（P2）

目标：在 100 并发、128-token、10 分钟压测下，p95 延迟较 Python 改善 ≥20%，峰值常驻内存下降 ≥25%，无错误率上升。

独立测试：运行 `make bench`，记录 p95/峰值内存在 `specs/001-go-agno-rewrite/artifacts/bench/`，与 Python 基线对比。

### 测试/基准
- [X] T024 [P] [US2] 在 `/Users/rex/cool.cnb/agno-Go/go/tests/bench/perf_bench_test.go` 实现 100 并发、128-token 输入、持续 10m 的流式基准，参数从 `/Users/rex/cool.cnb/agno-Go/config/default.yaml` 读取并输出原始数据
- [X] T025 [P] [US2] 在 `/Users/rex/cool.cnb/agno-Go/go/tests/bench/benchstat_test.go` 使用 benchstat 对比 Go 结果与基线，写入 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/artifacts/bench/benchstat.txt`
- [ ] T043 [P] [US2] 在 `/Users/rex/cool.cnb/agno-Go/go/tests/runtime/stream_backpressure_test.go` 验证高负载下流式不中断且返回背压/限流提示

### 实现/优化
- [ ] T026 [P] [US2] 在 `/Users/rex/cool.cnb/agno-Go/go/internal/runtime/middleware/` 添加限流/背压与请求追踪，确保流式不断流
- [ ] T027 [P] [US2] 在 `/Users/rex/cool.cnb/agno-Go/go/internal/model/router.go` 增加请求池/复用（连接重用、超时、重试），减少 GC 压力
- [ ] T028 [US2] 配置 GC/内存优化（如 GOMEMLIMIT）与 provider 并发控制，更新 `/Users/rex/cool.cnb/agno-Go/config/default.yaml` 与 `Makefile bench` 命令
- [ ] T029 [US2] 在 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/artifacts/bench/` 汇总压测报告与与 Python 基线对比说明
- [ ] T044 [US2] 在 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/artifacts/bench/python_baseline.txt` 记录预先生成的 Python 版基准（同治具/场景，脱敏文件，不在运行时调用 Python），并在 benchstat 报告中体现对比

**检查点**：US2 达标或记录改进行动。

## Phase 5: 用户故事 3 - 行为一致性验证（P3）

目标：九家供应商在契约治具下匹配率 ≥95%，偏差有记录并提供解释/替代。

独立测试：运行 `make providers-test`（有 key 的供应商），契约匹配率 ≥95%，偏差记录在 `contracts/deviations.md`；缺 key 需跳过并输出原因。

### 测试
- [ ] T030 [P] [US3] 在 `/Users/rex/cool.cnb/agno-Go/go/tests/contract/providers_parity_test.go` 覆盖九家供应商 chat/embedding 契约，使用 fixtures 容差（tokens ±2 / cos≥0.98）
- [ ] T031 [P] [US3] 在 `/Users/rex/cool.cnb/agno-Go/go/tests/providers/providers_integration_test.go` 针对已配置 key 运行正/异常分支，输出报告到 `specs/001-go-agno-rewrite/artifacts/coverage/providers.log`

### 实现
- [ ] T032 [P] [US3] 在 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/contracts/fixtures/` 导入/生成 Python 参考治具，并在 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/contracts/deviations.md` 记录差异
- [ ] T033 [P] [US3] 在 `/Users/rex/cool.cnb/agno-Go/go/pkg/providers/{gemini,glm4,openrouter,siliconflow,cerebras,modelscope,groq,ollama}/client.go` 实现/完善 REST 客户端与错误映射，对齐路由器接口
- [ ] T034 [US3] 更新 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/quickstart.md` 与 `/Users/rex/cool.cnb/agno-Go/AGENTS.md`，加入契约/供应商测试运行示例与常见偏差

**检查点**：US3 契约匹配率达标或差异已记录。

## Phase 6: 抛光与跨领域事项

- [ ] T035 [P] 在 `/Users/rex/cool.cnb/agno-Go/go/internal/runtime/` 与 `go/pkg/` 清理占位、补充错误处理/日志
- [ ] T036 [P] 在 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/quickstart.md` 与 `contracts/openapi.yaml` 同步最新端点/示例
- [ ] T037 运行全量 fmt/lint/test/providers-test/coverage/bench，修复遗留警告，并整理提交说明
- [ ] T046 在 `/Users/rex/cool.cnb/agno-Go/Makefile` 实现 `release` 目标（构建二进制输出到 `dist/` 并准备可发布工件），确保符合宪章的发布 Gate

## 覆盖率 Gate（所有故事完成后执行）

- [ ] T038 运行 `make test`, `make providers-test`, `make coverage`, 如需性能验证运行 `make bench`，确保覆盖率 ≥85%，工件写入 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/artifacts/`
- [ ] T039 在 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/artifacts/coverage/coverage.txt`（或 CI 工件）记录覆盖率、供应商测试与基准结果链接
- [ ] T045 确认 `make coverage` 使用 `-coverpkg=./...` 并汇总到 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/artifacts/coverage/coverage.txt`，在 CI 产物中上传同一报告以满足 FR-007/宪章

## 依赖与执行顺序

- 阶段依赖：Setup → Foundational → US1（P1/MVP） → US2（P2） → US3（P3） → 抛光 → 覆盖率 Gate
- 用户故事依赖：US1 完成后可同时推进 US2/US3；US2/US3 互不阻塞
- 并行机会：所有标记 [P] 的任务；Foundational 完成后 US2/US3 可与 US1 后续工作并行

## 并行示例

- 同时编写契约测试与供应商集成测试：`go/tests/contract/agentos_contract_test.go` 与 `go/tests/providers/openai_smoke_test.go`
- 并行实现内存存储与 provider 客户端：`go/internal/memory/badger_store.go` 与 `go/pkg/providers/openai/client.go`
- 并行性能与行为工作：`go/tests/bench/perf_bench_test.go` 与 `go/tests/contract/providers_parity_test.go`

## 实施策略

- MVP：完成 Setup + Foundational + US1，先验证流式聊天/工具/记忆与 OpenAPI 契约
- 渐进：在 MVP 基础上推进 US2（性能基准）与 US3（契约匹配），每个阶段都可独立测试与演示
