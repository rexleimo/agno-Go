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
