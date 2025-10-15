# AgentOS Server API 参考 / AgentOS Server API Reference

## NewServer

创建 HTTP 服务器。/ Create HTTP server.

**签名 / Signature:**
```go
func NewServer(config *Config) (*Server, error)

type Config struct {
    Address        string           // 服务器地址 (默认: :8080) / Server address (default: :8080)
    SessionStorage session.Storage  // 会话存储 (默认: memory) / Session storage (default: memory)
    Logger         *slog.Logger     // 日志记录器 (默认: slog.Default()) / Logger (default: slog.Default())
    Debug          bool             // 调试模式 (默认: false) / Debug mode (default: false)
    AllowOrigins   []string         // CORS 源 / CORS origins
    AllowMethods   []string         // CORS 方法 / CORS methods
    AllowHeaders   []string         // CORS 头 / CORS headers
    RequestTimeout time.Duration    // 请求超时 (默认: 30s) / Request timeout (default: 30s)
    MaxRequestSize int64            // 最大请求大小 (默认: 10MB) / Max request size (default: 10MB)

    // 知识库 API（可选）/ Knowledge API (optional)
    VectorDBConfig  *VectorDBConfig  // 向量数据库配置（如 chromadb）/ Vector DB configuration
    EmbeddingConfig *EmbeddingConfig // 嵌入模型配置（如 OpenAI）/ Embedding model configuration
}
 
type VectorDBConfig struct {
    Type           string // 例如 "chromadb" / e.g., "chromadb"
    BaseURL        string // 向量数据库地址 / Vector DB endpoint
    CollectionName string // 默认集合名 / Default collection
    Database       string // 可选数据库名 / Optional database
    Tenant         string // 可选租户名 / Optional tenant
}

type EmbeddingConfig struct {
    Provider string // 例如 "openai" / e.g., "openai"
    APIKey   string
    Model    string // 例如 "text-embedding-3-small"
    BaseURL  string // 例如 "https://api.openai.com/v1"
}
```

**示例 / Example:**
```go
server, err := agentos.NewServer(&agentos.Config{
    Address: ":8080",
    Debug:   true,
    RequestTimeout: 60 * time.Second,
})
```

## Server.RegisterAgent

注册一个智能体。/ Register an agent.

**签名 / Signature:**
```go
func (s *Server) RegisterAgent(agentID string, ag *agent.Agent) error
```

**示例 / Example:**
```go
err := server.RegisterAgent("assistant", myAgent)
```

## Server.Start / Shutdown

启动和停止服务器。/ Start and stop server.

**签名 / Signatures:**
```go
func (s *Server) Start() error
func (s *Server) Shutdown(ctx context.Context) error
```

**示例 / Example:**
```go
go func() {
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}()

// 优雅关闭 / Graceful shutdown
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
server.Shutdown(ctx)
```

## API 端点 / API Endpoints

查看完整的 API 文档请参阅 [OpenAPI 规范](../../pkg/agentos/openapi.yaml)。/ See [OpenAPI Specification](../../pkg/agentos/openapi.yaml) for complete API documentation.

**核心端点 / Core Endpoints:**
- `GET /health` - 健康检查 / Health check
- `POST /api/v1/sessions` - 创建会话 / Create session
- `GET /api/v1/sessions/{id}` - 获取会话 / Get session
- `PUT /api/v1/sessions/{id}` - 更新会话 / Update session
- `DELETE /api/v1/sessions/{id}` - 删除会话 / Delete session
- `GET /api/v1/sessions` - 列出会话 / List sessions
- `GET /api/v1/agents` - 列出智能体 / List agents
- `POST /api/v1/agents/{id}/run` - 运行智能体 / Run agent

**知识库端点（可选） / Knowledge Endpoints (optional):**
- `POST /api/v1/knowledge/search` — 在知识库中进行向量相似搜索 / Vector similarity search in knowledge base
- `GET  /api/v1/knowledge/config` — 返回可用分块器、向量库与嵌入模型信息 / Available chunkers, VectorDBs, embedding model info
- `POST /api/v1/knowledge/content` — 最小入库（text/plain 或 application/json）

示例请求 / Example:
```bash
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "如何创建 Agent?",
    "limit": 5,
    "filters": {"source": "documentation"}
  }'
```

最小服务器配置（启用知识库 API）/ Minimal server config (enable Knowledge API):
```go
server, err := agentos.NewServer(&agentos.Config{
  Address: ":8080",
  VectorDBConfig: &agentos.VectorDBConfig{
    Type:           "chromadb",
    BaseURL:        os.Getenv("CHROMADB_URL"),
    CollectionName: "agno_knowledge",
  },
  EmbeddingConfig: &agentos.EmbeddingConfig{
    Provider: "openai",
    APIKey:   os.Getenv("OPENAI_API_KEY"),
    Model:    "text-embedding-3-small",
  },
})
```

可运行示例 / Runnable example: `cmd/examples/knowledge_api/`

## 最佳实践 / Best Practices

### 1. 始终使用 Context / Always Use Contexts

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

output, err := ag.Run(ctx, input)
```

### 2. 正确处理错误 / Handle Errors Properly

```go
output, err := ag.Run(ctx, input)
if err != nil {
    switch {
    case types.IsInvalidInputError(err):
        // 处理无效输入 / Handle invalid input
    case types.IsRateLimitError(err):
        // 使用退避重试 / Retry with backoff
    default:
        // 处理其他错误 / Handle other errors
    }
}
```

### 3. 管理内存 / Manage Memory

```go
// 开始新话题时清除 / Clear when starting new topic
ag.ClearMemory()

// 或使用有限内存 / Or use limited memory
mem := memory.NewInMemory(50)
```

### 4. 设置适当的超时 / Set Appropriate Timeouts

```go
server, _ := agentos.NewServer(&agentos.Config{
    RequestTimeout: 60 * time.Second, // 用于复杂智能体 / For complex agents
})
```
