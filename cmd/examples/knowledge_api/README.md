# Knowledge API 示例

本示例演示如何使用 agno-Go 的知识库 API 进行向量搜索。

## 前置条件

### 1. 安装并启动 ChromaDB

使用 Docker 启动 ChromaDB:

```bash
docker run -p 8000:8000 chromadb/chroma:latest
```

或者使用 Docker Compose (如果项目根目录有 `docker-compose.yml`):

```bash
docker-compose up -d chromadb
```

### 2. 设置 OpenAI API Key

```bash
export OPENAI_API_KEY=sk-your-api-key
```

可选：自定义 ChromaDB URL

```bash
export CHROMADB_URL=http://localhost:8000
```

## 运行示例

### 启动服务器

```bash
cd /Users/molei/codes/aiagent/agno-Go
go run cmd/examples/knowledge_api/main.go
```

服务器将在 `http://localhost:8080` 启动。

## API 使用示例

### 1. 获取配置信息

查询可用的分块器、向量数据库和嵌入模型信息：

```bash
curl http://localhost:8080/api/v1/knowledge/config
```

**响应示例：**

```json
{
  "available_chunkers": [
    {
      "name": "character",
      "description": "按字符数量分块，适合通用文本",
      "default_chunk_size": 1000,
      "default_overlap": 100
    },
    {
      "name": "sentence",
      "description": "按句子分块，适合对话和文档",
      "default_chunk_size": 1000,
      "default_overlap": 0
    },
    {
      "name": "paragraph",
      "description": "按段落分块，适合长文档",
      "default_chunk_size": 2000,
      "default_overlap": 0
    }
  ],
  "available_vector_dbs": ["chromadb"],
  "default_chunker": "character",
  "default_vector_db": "chromadb",
  "embedding_model": {
    "provider": "openai",
    "model": "text-embedding-3-small",
    "dimensions": 1536
  }
}
```

### 2. 搜索知识库

基本搜索：

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "如何创建 Agent?",
    "limit": 5
  }'
```

**响应示例：**

```json
{
  "results": [
    {
      "id": "doc_123",
      "content": "To create an Agent in agno-Go, use the agent.New() function...",
      "metadata": {
        "source": "documentation",
        "page": 5
      },
      "score": 0.95,
      "distance": 0.05
    }
  ],
  "meta": {
    "total_count": 10,
    "page": 1,
    "page_size": 5,
    "total_pages": 2
  }
}
```

### 3. 带分页的搜索

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "Agent 配置",
    "limit": 10,
    "offset": 10
  }'
```

这将返回第 2 页的结果（跳过前 10 个结果）。

### 4. 带过滤器的搜索

使用元数据过滤搜索结果：

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "Agent tools",
    "limit": 10,
    "filters": {
      "source": "documentation",
      "category": "tools"
    }
  }'
```

## 添加测试数据

在使用搜索功能之前，您需要向 ChromaDB 添加一些文档。可以使用以下 Python 脚本：

```python
import chromadb

# 连接到 ChromaDB
client = chromadb.HttpClient(host="localhost", port=8000)

# 获取或创建集合
collection = client.get_or_create_collection(
    name="agno_knowledge",
    metadata={"description": "Agno-Go knowledge base"}
)

# 添加文档
documents = [
    "To create an Agent in agno-Go, use the agent.New() function with a configuration.",
    "Agents can use tools like Calculator, HTTP, and File operations.",
    "Memory management in agno-Go uses the memory.Memory interface.",
]

metadatas = [
    {"source": "documentation", "page": 1},
    {"source": "documentation", "page": 2},
    {"source": "documentation", "page": 3},
]

ids = ["doc1", "doc2", "doc3"]

collection.add(
    documents=documents,
    metadatas=metadatas,
    ids=ids
)

print("Documents added successfully!")
```

## 配置选项

示例程序支持以下环境变量：

| 变量 | 描述 | 默认值 | 必需 |
|------|------|--------|------|
| `OPENAI_API_KEY` | OpenAI API 密钥 | - | 是 |
| `CHROMADB_URL` | ChromaDB 服务器 URL | `http://localhost:8000` | 否 |

## 架构说明

### 知识库服务组件

```
┌─────────────────────────────────────────┐
│         Knowledge API Handler           │
│  (handleKnowledgeSearch/Config)         │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│       KnowledgeService                  │
│  - VectorDB                             │
│  - EmbeddingFunction                    │
│  - Configuration                        │
└────────────────┬────────────────────────┘
                 │
         ┌───────┴────────┐
         │                │
         ▼                ▼
┌────────────────┐  ┌──────────────┐
│   ChromaDB     │  │   OpenAI     │
│  (Vector DB)   │  │ (Embeddings) │
└────────────────┘  └──────────────┘
```

### 搜索流程

1. **接收搜索请求** - POST /api/v1/knowledge/search
2. **验证参数** - 检查 query、limit、offset 等参数
3. **执行向量搜索** - 调用 VectorDB.Query()
4. **应用分页** - 根据 offset 和 limit 切片结果
5. **返回响应** - 包含结果和分页元数据

## 故障排除

### ChromaDB 连接失败

```
ERROR: failed to initialize knowledge service: failed to create chromadb: ...
```

**解决方案：**
1. 确保 ChromaDB 正在运行：`docker ps | grep chroma`
2. 检查端口是否正确：`curl http://localhost:8000/api/v1/heartbeat`
3. 验证 CHROMADB_URL 环境变量

### OpenAI API 密钥无效

```
FATAL: OPENAI_API_KEY environment variable is required
```

**解决方案：**
1. 设置有效的 OpenAI API 密钥：`export OPENAI_API_KEY=sk-...`
2. 验证密钥：`echo $OPENAI_API_KEY`

### 搜索返回空结果

**原因：**
- 知识库中没有文档
- 查询与现有文档不匹配

**解决方案：**
1. 使用上面的 Python 脚本添加测试数据
2. 尝试不同的查询词
3. 检查过滤器是否过于严格

## 相关文档

- [AgentOS Server 文档](../../../docs/AGENTOS.md)
- [Vector DB 使用指南](../../../docs/VECTORDB.md)
- [Knowledge API OpenAPI 规范](../../../pkg/agentos/openapi.yaml)
- [企业迁移计划](../../../docs/ENTERPRISE_MIGRATION_PLAN.md)

## 下一步

- 实现 P2 任务：最小化知识入库接口 (POST /api/v1/knowledge/content)
- 添加事件流过滤 (SSE)
- 实现内容抽取中间件
- 添加 Google Sheets 工具集成
