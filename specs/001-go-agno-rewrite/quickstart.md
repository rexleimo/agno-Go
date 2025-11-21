# Quickstart - Go 版 Agno 重构

绝对路径前缀：`/Users/rex/cool.cnb/agno-Go`

## 前置
- Go 1.25.1、make、bash
- 可选：GitHub Actions 等 CI，可复用 `make fmt lint test providers-test coverage bench gen-fixtures release constitution-check`
- Python 3.11 仅用于离线生成 fixtures，不参与运行/测试

## 环境变量
复制 `.env.example` 到 `.env` 并按需填写（缺失的供应商将被禁用并在健康检查/契约测试中跳过）：

```
OPENAI_API_KEY=
GEMINI_API_KEY=
GLM4_API_KEY=
OPENROUTER_API_KEY=
SILICONFLOW_API_KEY=
CEREBRAS_API_KEY=
MODELSCOPE_API_KEY=
GROQ_API_KEY=
OLLAMA_ENDPOINT=http://localhost:11434/v1
```

### 持久化存储路径
- 默认数据目录：`./data/{bolt|badger}/<namespace>`，namespace 未填或非法字符将回落 `default`。
- 自定义数据目录：设置 `AGNO_DATA_DIR=/abs/path`，路径下会创建 `bolt/<ns>.db` 或 `badger/<ns>/`。
- 清理策略：Badger 可选 `retention` TTL（在 config 中配置）；清理时先停服务再删除相应目录/文件。

## 初始化与构建
```bash
cd /Users/rex/cool.cnb/agno-Go
make fmt lint            # gofumpt + golangci-lint
make test                # 单元 + 契约基础测试
make providers-test      # 仅限已配置密钥的供应商集成测试
make coverage            # 汇总覆盖率 -> specs/001-go-agno-rewrite/artifacts/coverage/{coverage.out,coverage.txt}
make bench               # 基准 -> specs/001-go-agno-rewrite/artifacts/bench/{bench.txt,benchstat.txt}
make constitution-check  # 全量 fmt/lint/test/providers-test/coverage/bench + audit-no-python
```

## 启动 AgentOS（本地）
```bash
cd /Users/rex/cool.cnb/agno-Go
go run ./go/cmd/agno --config /Users/rex/cool.cnb/agno-Go/config/default.yaml
```

## 创建 Agent 与会话
```bash
# 创建 Agent
curl -X POST http://localhost:8080/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "go-agno",
    "model": {"provider": "openai", "modelId": "gpt-4.1-mini", "stream": true},
    "memory": {"storeType": "memory", "tokenWindow": 256}
  }'

# 创建 Session
curl -X POST http://localhost:8080/agents/<agentId>/sessions
```

## 发送消息（流式）
```bash
curl -N -X POST "http://localhost:8080/agents/<agentId>/sessions/<sessionId>/messages?stream=true" \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [{"role": "user", "content": "Hello, call the time tool"}]
  }'
```

## 工具禁用/降级示例
```bash
curl -X PATCH http://localhost:8080/agents/<agentId>/tools/time \
  -H "Content-Type: application/json" \
  -d '{"enabled": false}'
```

## 契约与基准
- 生成/验证治具：`./scripts/gen-fixtures.sh`（默认读取 `specs/001-go-agno-rewrite/contracts/fixtures-src`，写入 `contracts/fixtures`；`VERIFY_ONLY=true` 仅校验内容）。仓库内附带 OpenAI stub 示例（chat/embedding），替换为 Python 基线后再运行。
- 将 Python 生成的 fixtures 放入 `/Users/rex/cool.cnb/agno-Go/specs/001-go-agno-rewrite/contracts/fixtures/`（或放入 fixtures-src 后运行脚本复制）
- 运行 `make providers-test` 产出 `contracts/deviations.md` 与 parity 报告（覆盖率聚合于 artifacts/coverage）
- 运行 `make bench` 验证 100 并发、128 token、10 分钟基准并比较 Python 基线（报告在 artifacts/bench/，benchstat 汇总于 benchstat.txt）

## 健康检查
```bash
curl http://localhost:8080/health
```

缺密钥的供应商会显示为 `not-configured` 并在契约/集成测试中跳过，同时在日志中输出原因。
