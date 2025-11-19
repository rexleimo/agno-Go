# Quickstart：Go Runtime 核心迁移

**Feature**：/home/rex/codes/agno-Go/specs/001-migrate-agno-core/spec.md  
**Plan**：/home/rex/codes/agno-Go/specs/001-migrate-agno-core/plan.md  
**Date**：2025-11-19

本 Quickstart 面向需要在纯 Go 环境运行 Agno 核心模块（Agent、Workflow、Session、Telemetry）的工程师，步骤涵盖：对齐 Python 基线、运行 Go 版本、执行跨语言对照测试、收集性能/Telemetry 数据以及更新文档支持矩阵。

---

## 1. 准备环境并定位参考示例

1. 克隆包含 Python 基线与 Go 实现的仓库：
   ```bash
   git clone git@github.com:agno-agi/agno-go.git
   cd agno-go
   ```
2. 安装依赖：
   - Go ≥1.25.1（`go env GOVERSION` 验证）。
   - Python ≥3.11（供 parity 脚本引用 `agno/libs/agno`）。
   - 在根目录创建 `.env` 或导出模型/数据库凭据，沿用 Python 版的环境变量命名。
3. 选定一个 cookbook 场景作为迁移样例（推荐 `agno/cookbook/workflows/us1_basic_coordination`），记录涉及的 Agents、Providers、Workflow 结构、工具与 Session 配置。

---

## 2. 在 Go 中注册 AgentRuntime 与 WorkflowManifest

1. 根据 `specs/001-migrate-agno-core/data-model.md` 的 `AgentRuntime` 字段构建配置（可参考 `go/agent/us1_basic_coordination_agents.go`）：
   ```go
   agents := []agent.Agent{
       agent.Agent{
           ID:   "researcher",
           Name: "Research Analyst",
           AllowedProviders: []agent.ProviderID{
               agent.ProviderID(providers.US1OpenAIChat.ID),
           },
           MemoryPolicy: agent.MemoryPolicy{Persist: true, WindowSize: 10},
       },
       // ... more agents
   }
   ```
2. 组合 Workflow manifest：
   ```go
   wf := workflow.Workflow{
       ID:          "us1-basic-coordination",
       PatternType: workflow.PatternSequential,
       Steps: []workflow.Step{
           {ID: "gather", AgentID: "researcher"},
           {ID: "analyze", AgentID: "strategist"},
       },
       EntryPoints: []workflow.StepID{"gather"},
   }
   ```
3. 如果通过控制面 API 管理，调用 `POST /v1/runtime/go/agents` 与 `/v1/runtime/go/workflows/{id}:run`（见 `contracts/runtime-openapi.yaml`）。

---

## 3. 运行 Go Runtime 并输出 SessionRecord

1. 启动示例：
   ```bash
   go run ./go/examples/us1_basic_coordination \
     --providers=go/providers/providers.go \
     --workflow=go/workflow/us1_basic_coordination_workflow.go
   ```
2. 运行结束后，可通过以下命令检查 SessionRecord 是否与 Python schema 对齐：
   ```bash
   curl -s \
     "http://localhost:8080/v1/runtime/go/sessions/${SESSION_ID}" | jq
   ```
   输出需要包含 state/history/summary/metrics，并使用 `runtime":"go"` Telemetry 标签。

---

## 4. 执行跨语言对照测试（Parity）

1. 准备 fixture（YAML/JSON），结构参见 `specs/001-migrate-agno-core/data-model.md` 的 `ParityFixture`。
2. 运行脚本：
   ```bash
   ./scripts/ci/cross-language-parity.sh \
     --fixture specs/001-migrate-agno-core/fixtures/us1_basic_coordination.yaml \
     --python "python -m agno.tests.contracts.run" \
     --go "go test ./go/... -run TestParity"
   ```
3. 期望输出：
   - `status: pass`
   - `diffs: []`
   - `metrics_match: true`
4. 若失败，依据 `ParityTestResult.diffs` 中的路径（如 `outputs[2].content`）定位 Go 与 Python 的差异，并更新对应实现或 fixture 容差。

---

## 5. 收集 Telemetry 与性能基线

1. 在运行 parity/benchmark 时启用 Telemetry recorder：
   ```bash
   GO_TELEMETRY_EXPORTER=stdout go test ./go/agent -run TestWorkflowParity -count=1
   ```
2. 将事件导出到监控系统或文件：
   ```bash
   curl -G "http://localhost:8080/v1/runtime/go/telemetry/events" \
     --data-urlencode "sessionId=${SESSION_ID}" \
     --data-urlencode "runtime=go"
   ```
3. 记录性能对比：
   ```bash
   # Python baseline
   python agno/tests/benchmarks/run.py --workflow us1_basic_coordination --output /tmp/python.json
   # Go 运行
   go test ./go/agent -run TestUS1Bench -bench . -benchmem | tee /tmp/go.txt
   ```
4. 将结果提交给 `/v1/runtime/go/benchmarks`，以便控制面展示 Go vs Python 差异。

---

## 6. 更新文档与支持矩阵

1. 在 `agno/README.md` 与 cookbook 内添加 Go Runtime 入口：
   - 新增 “Go Runtime Quickstart” 小节，指向本指南与 `go/examples`。
   - 将 Python-only 功能标记为 “Not yet in Go（runtime=python-only）”。
2. 更新 `go/examples/us1_basic_coordination` README，列出：
   - 构建命令、配置文件示例；
   - Telemetry/benchmark 指标及如何阅读；
   - 常见问题（例如“尚未迁移的 Provider 会抛出 `not_migrated` 错误”）。
3. 将最新支持矩阵通过 `GET /v1/runtime/go/support-matrix` 生成 JSON，写入 docs 或网站。

---

## 7. 交付前检查

- [ ] `go test ./... -cover` ≥ 85%，且 parity/bench 测试未跳过。
- [ ] `/scripts/ci/cross-language-parity.sh` 在 CI 中可直接运行并产出 diff。
- [ ] Telemetry 中所有事件均带 `runtime=go` 标签，且 payload 无敏感信息。
- [ ] README / Cookbook / Quickstart 已描述 Go Runtime 支持矩阵与限制。
- [ ] 未迁移的模块在初始化阶段抛出 `not_migrated` 错误，并被记录在文档中。
