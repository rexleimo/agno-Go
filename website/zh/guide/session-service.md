# Go 会话服务

Go 会话服务完整复刻了 Python AgentOS 的 `/sessions` REST API，并以独立的 Go
二进制形式运行。将它与 AgentOS 一起部署，即可获得基于 PostgreSQL 的高性能
会话 CRUD、生产级 HTTP 中间件，以及覆盖本地与 Kubernetes 场景的部署资产。

## 功能亮点

- **端点保持一致**：`/sessions` 列表/创建、`/sessions/{id}` 详情、重命名、删除，
  `/sessions/{id}/runs` 历史记录，以及 `/healthz` 健康探针。
- **Postgres 存储**：强类型 DTO 层配合事务安全读写，与现有 AgentOS JSON 协议保持一致。
- **多数据库路由**：通过 `AGNO_SESSION_DSN_MAP` 提供多个 DSN，并使用 `db_id`
  查询参数按需选择目标存储。
- **运维友好**：基于 Chi 的路由器默认启用请求 ID、结构化日志、真实 IP 解析、
  panic 恢复与 60s 超时保护。
- **部署资产齐全**：独立 Dockerfile、本地用的 Docker Compose 栈、Helm Chart 以及 curl
  的冒烟脚本一应俱全。

## 本地快速起步

```bash
export AGNO_PG_DSN="postgres://user:pass@localhost:5432/agentos?sslmode=disable"
export AGNO_SERVICE_PORT=8080
go run ./cmd/agentos-session
```

启动日志会显示 `Go session service listening on :8080`，访问
`http://localhost:8080/healthz` 即可获得 JSON 健康探针响应。

### 合同测试对齐

- `make contract-test`：运行端到端合同测试，对比 Go 与 Python 的响应数据。
- `./scripts/test-session-api.sh http://localhost:8080`：使用 curl + jq 完成列表、创建、重命名、删除的基本流程。

## 配置参考

| 环境变量               | 说明                                                                                 | 默认值      |
|------------------------|--------------------------------------------------------------------------------------|-------------|
| `AGNO_PG_DSN`          | 主 Postgres DSN；在未提供 `AGNO_SESSION_DSN_MAP` 时必填。                           | –           |
| `DATABASE_URL`         | 备用 DSN（Heroku 风格）。在 `AGNO_PG_DSN` 未设置时回退使用。                         | –           |
| `AGNO_SESSION_DSN_MAP` | JSON 映射 `{"dbID":"dsn"}`，开启多数据库路由并允许通过 `db_id` 查询参数选择存储。 | –           |
| `AGNO_DEFAULT_DB_ID`   | 使用 `AGNO_SESSION_DSN_MAP` 时可选的默认存储标识。                                    | 映射首个键  |
| `AGNO_SERVICE_PORT`    | HTTP 监听端口。                                                                       | `8080`      |

提供 `AGNO_SESSION_DSN_MAP` 后，可通过
`/sessions?type=agent&db_id=analytics` 将请求定向到指定库。

## API 概览

| Endpoint                | 方法  | 描述                                                                                   |
|-------------------------|-------|----------------------------------------------------------------------------------------|
| `/healthz`              | GET   | 返回 `{"status":"ok"}` 的健康检查。                                                  |
| `/sessions`             | GET   | 支持分页与多种筛选（`type`、`component_id`、`user_id`、`session_name`、`sort_by`、`db_id`）。 |
| `/sessions`             | POST  | 创建会话，可携带状态、元数据以及预置的 runs 或 summary。                             |
| `/sessions/{id}`        | GET   | 按类型与 ID 查询会话详情，可附带 `db_id`。                                             |
| `/sessions/{id}`        | DELETE| 删除会话及其历史记录。                                                                |
| `/sessions/{id}/rename` | POST  | 更新 `session_name`。                                                                 |
| `/sessions/{id}/runs`   | GET   | 获取历史 run 数据，结构与 Python AgentOS 保持一致。                                    |

所有写操作均需提交符合 Python 基准数据的 JSON 结构，并在合同测试中自动校验。

## Docker Compose 集成

使用仓库自带的 `docker-compose.session.yml` 启动 Postgres 与 Go 会话服务：

```bash
docker compose -f docker-compose.session.yml up --build
```

容器就绪后可直接访问 `http://localhost:8080`，再运行脚本完成验证。

## Helm 部署

`deploy/helm/agno-session/` 提供了 Kubernetes 部署模板，可按需覆盖镜像与 DSN：

```bash
helm upgrade --install agno-session ./deploy/helm/agno-session \
  --set image.repository=ghcr.io/<org>/agno-session \
  --set image.tag=v1.2.9 \
  --set config.dsn="postgres://user:pass@postgres:5432/agentos?sslmode=disable"
```

更多可调参数（探针、实例数、Service 类型等）见 `values.yaml`。

## 生产检查清单

- 在预发布环境执行 `make contract-test` 确认协议一致。
- 引流真实流量并与 Python 服务的 JSON 响应对比验证。
- 关注 Postgres 连接数与延迟，必要时接入连接池或限流策略。
- 基于 `/healthz` 与 HTTP 日志配置监控与告警。
- 在生产稳定前保留 Python 版本的回退策略。
