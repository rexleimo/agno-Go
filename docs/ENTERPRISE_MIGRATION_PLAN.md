# 企业迁移计划（agno → agno-Go）

本文档梳理 agno 近期 60 次提交中可迁移到 agno-Go 的内容、不可迁移项、实施 TODOs 与技术方案，聚焦企业场景的隐私、稳定与可观测性。

## 可迁移项（按优先级）
- P0｜知识库搜索与配置 API
  - 新增 `POST /api/v1/knowledge/search`、`GET /api/v1/knowledge/config`，复用 `pkg/agno/knowledge`、`pkg/agno/vectordb/chromadb`。
- P1｜事件流过滤（SSE）
  - A2A/AgentOS 暴露可筛选的执行事件（token/tool_call/step/error/complete）。
- P1｜内容抽取中间件
  - 在 Gin 中统一抽取 `content/metadata` 注入上下文，便于审计与落库。
- P1｜Google Sheets 工具（服务账号）
  - 新增工具包，支持服务账号 JSON，提供读写接口。
- P2｜最小化知识入库接口
  - 支持 `text/plain` 上传→分块→嵌入→向量库写入（先覆盖文本/CSV）。

## 不可/不建议迁移
- Python v2 DB 迁移与 `updated_at` 细节；Go 侧暂无 DB 迁移框架。
- 多向量库修复（Pinecone/Milvus/Weaviate/Qdrant/PGVector 等）；当前仅 ChromaDB。
- PDF/DOCX Reader 修复与 ScrapeGraph 相关变更；Go 侧暂无同类实现。

## TODOs 清单（落地项）
- 路由与 OpenAPI
  - 在 `pkg/agentos/server.go` 注册 `/api/v1/knowledge` 组。
  - 更新 `pkg/agentos/openapi.yaml` 增加 Knowledge 与事件流过滤参数。
- 类型与处理器
  - 新增 `pkg/agentos/knowledge_types.go`（DTO）、`pkg/agentos/knowledge_handlers.go`（Handler）。
- 向量检索
  - 处理器中注入 `vectordb.VectorDB`（默认 `chromadb`），实现分页与过滤。
- 事件流过滤（SSE）
  - 在 `pkg/agentos/a2a/handlers.go` 增参 `?types=token,complete`；必要时新增 `pkg/agentos/events_handlers.go`。
- 中间件
  - 新增 `pkg/agentos/middleware/extract_content.go` 并在 `NewServer` 启用。
- Google Sheets 工具
  - 新增 `pkg/agno/tools/googlesheets/*`，示例与测试覆盖读/写/追加。
- 示例与测试
  - 新增 `cmd/examples/knowledge_api`；补充 handler/middleware/tool 单测与分页过滤集成测。

## 技术实现方案（摘要）
- 知识库 API
  - 定义 `VectorSearchRequest{Query string, Meta{Page,Limit}, Filters map[string]any}` 与 `VectorSearchResult{ID,Content,Metadata,Score}`。
  - `POST /knowledge/search` 调用 `vectordb.Query(ctx, query, limit, filter)`，做分页返回 `meta.total_count/total_pages`。
  - `GET /knowledge/config` 返回可用 chunker（Character/Sentence/Paragraph）与向量库（chromadb）。
- 事件流过滤
  - 事件枚举：`run_start/tool_call/token/step_start/step_end/error/complete`；SSE 端点 `?types=` 过滤、按行 flush、支持 ctx 取消。
- 内容抽取中间件
  - 解析 JSON/Form，抽取 `content/metadata/user_id/session_id`，`c.Set("extracted_content", v)`；结合 `MaxRequestSize` 进行输入约束。
- Google Sheets 工具
  - 基于 `google.golang.org/api/sheets/v4`；支持 ENV/JSON 文本加载服务账号；注册 `read_range/write_range/append_rows` 函数。
- 最小化知识入库（可选）
  - `POST /knowledge/content` 支持文本上传→`CharacterChunker` 分块→`chromadb.Add` 写入；声明暂不支持 PDF/DOCX。

## 交付与验证
- 更新 OpenAPI、添加示例与文档；
- 单元/集成测试覆盖：搜索分页、事件过滤、工具 I/O、处理中间件；
- 本地验证：`make test`、`make build`，Chromadb 参考 `docker-compose.yml`。

## 小目标路线图（针对“不可/不建议迁移”的分步实现）
- M1｜持久化会话与迁移基础（SQLite）
  - 范围：实现 `session.Storage` 的 SQLite 实现与迁移框架，保留 `updated_at` 语义。
  - 任务：`pkg/agno/session/sqlite/*`、`scripts/init-db.sql`、在 `pkg/agentos/server.go` 提供存储选择；`cmd/tools/session_migrate`（可选）。
  - 验收：CRUD + List 通过；并发读写测试；重启后数据一致；`updated_at` 不回退。
- M2｜多向量库适配器（PGVector 或 Qdrant 二选一）
  - 范围：扩展 `vectordb.VectorDB`（含 Upsert/Delete/Filter/IDs），新增 `pgvector/*` 或 `qdrant/*` 实现。
  - 任务：适配 `knowledge/search` 的分页与过滤；提供 Docker Compose 与最小使用示例。
  - 验收：增删改查与语义检索通过；删除与元数据写入有回归测试。
- M3｜Reader 与内容管道（最小支持 PDF/DOCX）
  - 范围：定义 Reader 接口并实现 `Text/Markdown`；新增 `PDFReader`（unidoc/unipdf）与 `DocxReader`（gooxml）。
  - 任务：`pkg/agno/knowledge/reader/*`、扩展 `POST /knowledge/content`；与 chunker、向量库串联。
  - 验收：小/大文件解析稳定、分块边界合理、资源占用受控；失败路径具备可观测日志。
- M4｜抓取与动态渲染（Scrape 替代方案）
  - 范围：先静态抓取（`goquery`/readability），后可选动态渲染（`chromedp`）。
  - 任务：`pkg/agno/tools/webfetch/*`；`/knowledge/content` 支持 URL 输入并调用抓取。
  - 验收：静态页面正文抽取准确；启用渲染能抓到主内容；具备限流与超时控制。
