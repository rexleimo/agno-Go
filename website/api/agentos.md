# AgentOS Server API Reference

## NewServer

Create HTTP server.

**Signature:**
```go
func NewServer(config *Config) (*Server, error)

type Config struct {
    Address        string           // Server address (default: :8080)
    SessionStorage session.Storage  // Session storage (default: memory)
    Logger         *slog.Logger     // Logger (default: slog.Default())
    Debug          bool             // Debug mode (default: false)
    AllowOrigins   []string         // CORS origins
    AllowMethods   []string         // CORS methods
    AllowHeaders   []string         // CORS headers
    RequestTimeout time.Duration    // Request timeout (default: 30s)
    MaxRequestSize int64            // Max request size (default: 10MB)
}
```

**Example:**
```go
server, err := agentos.NewServer(&agentos.Config{
    Address: ":8080",
    Debug:   true,
    RequestTimeout: 60 * time.Second,
})
```

## Server.RegisterAgent

Register an agent.

**Signature:**
```go
func (s *Server) RegisterAgent(agentID string, ag *agent.Agent) error
```

**Example:**
```go
err := server.RegisterAgent("assistant", myAgent)
```

## Server.Start / Shutdown

Start and stop server.

**Signatures:**
```go
func (s *Server) Start() error
func (s *Server) Shutdown(ctx context.Context) error
```

**Example:**
```go
go func() {
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}()

// Graceful shutdown
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
server.Shutdown(ctx)
```

## API Endpoints

See [OpenAPI Specification](../../pkg/agentos/openapi.yaml) for complete API documentation.

**Core Endpoints:**
- `GET /health` - Health check
- `POST /api/v1/sessions` - Create session
- `GET /api/v1/sessions/{id}` - Get session
- `PUT /api/v1/sessions/{id}` - Update session
- `DELETE /api/v1/sessions/{id}` - Delete session
- `GET /api/v1/sessions` - List sessions
- `GET /api/v1/agents` - List agents
- `POST /api/v1/agents/{id}/run` - Run agent

## Best Practices

### 1. Always Use Contexts

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

output, err := ag.Run(ctx, input)
```

### 2. Handle Errors Properly

```go
output, err := ag.Run(ctx, input)
if err != nil {
    switch {
    case types.IsInvalidInputError(err):
        // Handle invalid input
    case types.IsRateLimitError(err):
        // Retry with backoff
    default:
        // Handle other errors
    }
}
```

### 3. Manage Memory

```go
// Clear when starting new topic
ag.ClearMemory()

// Or use limited memory
mem := memory.NewInMemory(50)
```

### 4. Set Appropriate Timeouts

```go
server, _ := agentos.NewServer(&agentos.Config{
    RequestTimeout: 60 * time.Second, // For complex agents
})
```
