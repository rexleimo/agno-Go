# Agno-Go API Reference

Complete API reference for Agno-Go v1.0.

## Table of Contents

- [Agent](#agent)
- [Team](#team)
- [Workflow](#workflow)
- [Models](#models)
- [Tools](#tools)
- [Memory](#memory)
- [Types](#types)
- [AgentOS Server](#agentos-server)

---

## Agent

### agent.New

Create a new agent instance.

**Signature:**
```go
func New(config Config) (*Agent, error)
```

**Parameters:**

```go
type Config struct {
    // Required
    Model models.Model // LLM model to use

    // Optional
    Name         string            // Agent name (default: "Agent")
    Toolkits     []toolkit.Toolkit // Available tools
    Memory       memory.Memory     // Conversation memory
    Instructions string            // System instructions
    MaxLoops     int               // Max tool call loops (default: 10)
}
```

**Returns:**
- `*Agent`: Created agent instance
- `error`: Error if model is nil or config is invalid

**Example:**
```go
model, _ := openai.New("gpt-4", openai.Config{APIKey: apiKey})

ag, err := agent.New(agent.Config{
    Name:         "Assistant",
    Model:        model,
    Toolkits:     []toolkit.Toolkit{calculator.New()},
    Instructions: "You are a helpful assistant.",
    MaxLoops:     15,
})
```

### Agent.Run

Execute the agent with input.

**Signature:**
```go
func (a *Agent) Run(ctx context.Context, input string) (*RunOutput, error)
```

**Parameters:**
- `ctx`: Context for cancellation/timeout
- `input`: User input string

**Returns:**
```go
type RunOutput struct {
    Content  string                 // Agent's response
    Metadata map[string]interface{} // Additional metadata
}
```

**Errors:**
- `InvalidInputError`: Input is empty
- `ModelTimeoutError`: LLM request timeout
- `ToolExecutionError`: Tool execution failed
- `APIError`: LLM API error

**Example:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

output, err := ag.Run(ctx, "What is 2+2?")
if err != nil {
    log.Fatal(err)
}
fmt.Println(output.Content)
```

### Agent.ClearMemory

Clear conversation memory.

**Signature:**
```go
func (a *Agent) ClearMemory()
```

**Example:**
```go
ag.ClearMemory() // Start fresh conversation
```

---

## Team

### team.New

Create a new team of agents.

**Signature:**
```go
func New(config Config) (*Team, error)
```

**Parameters:**
```go
type Config struct {
    // Required
    Agents []*agent.Agent // Team members

    // Optional
    Name   string         // Team name (default: "Team")
    Mode   CoordinationMode // Coordination mode (default: Sequential)
    Leader *agent.Agent   // Leader (for LeaderFollower mode)
}

type CoordinationMode string

const (
    ModeSequential     CoordinationMode = "sequential"
    ModeParallel       CoordinationMode = "parallel"
    ModeLeaderFollower CoordinationMode = "leader-follower"
    ModeConsensus      CoordinationMode = "consensus"
)
```

**Returns:**
- `*Team`: Created team instance
- `error`: Error if agents list is empty or invalid config

**Example:**
```go
tm, err := team.New(team.Config{
    Name:   "Research Team",
    Agents: []*agent.Agent{researcher, writer, editor},
    Mode:   team.ModeSequential,
})
```

### Team.Run

Execute the team with input.

**Signature:**
```go
func (t *Team) Run(ctx context.Context, input string) (*RunOutput, error)
```

**Behavior by Mode:**

- **Sequential**: Agents execute one after another, output feeds to next
- **Parallel**: All agents execute simultaneously, results combined
- **LeaderFollower**: Leader delegates tasks to followers
- **Consensus**: Agents discuss until reaching agreement

**Example:**
```go
output, err := tm.Run(context.Background(), "Write an article about AI")
```

### Team.AddAgent / RemoveAgent

Manage team members dynamically.

**Signatures:**
```go
func (t *Team) AddAgent(ag *agent.Agent)
func (t *Team) RemoveAgent(name string) error
func (t *Team) GetAgents() []*agent.Agent
```

**Example:**
```go
tm.AddAgent(newAgent)
tm.RemoveAgent("OldAgent")
agents := tm.GetAgents()
```

---

## Workflow

### workflow.New

Create a new workflow.

**Signature:**
```go
func New(config Config) (*Workflow, error)
```

**Parameters:**
```go
type Config struct {
    Name  string      // Workflow name
    Steps []Primitive // Workflow steps
}
```

**Example:**
```go
wf, err := workflow.New(workflow.Config{
    Name: "Data Processing",
    Steps: []workflow.Primitive{
        workflow.NewStep("fetch", fetchAgent),
        workflow.NewStep("process", processAgent),
        workflow.NewStep("output", outputAgent),
    },
})
```

### Workflow Primitives

#### 1. Step

Execute an agent or function.

**Signature:**
```go
func NewStep(name string, target interface{}) *Step
```

**Target types:**
- `*agent.Agent`: Run agent
- `func(ctx *ExecutionContext) (*RunOutput, error)`: Custom function

**Example:**
```go
step := workflow.NewStep("analyze", analyzerAgent)

// Or custom function
step := workflow.NewStep("transform", func(ctx *workflow.ExecutionContext) (*workflow.RunOutput, error) {
    input := ctx.Input
    // Transform input
    return &workflow.RunOutput{Content: transformed}, nil
})
```

#### 2. Condition

Conditional branching.

**Signature:**
```go
func NewCondition(name string, condition func(*ExecutionContext) bool,
                   thenStep, elseStep Primitive) *Condition
```

**Example:**
```go
cond := workflow.NewCondition("check_sentiment",
    func(ctx *workflow.ExecutionContext) bool {
        result := ctx.GetResult("classify")
        return strings.Contains(result.Content, "positive")
    },
    workflow.NewStep("positive_handler", positiveAgent),
    workflow.NewStep("negative_handler", negativeAgent),
)
```

#### 3. Loop

Iterative loops.

**Signature:**
```go
func NewLoop(name string, condition func(*ExecutionContext) bool,
             body Primitive, maxIterations int) *Loop
```

**Example:**
```go
loop := workflow.NewLoop("retry",
    func(ctx *workflow.ExecutionContext) bool {
        result := ctx.GetResult("attempt")
        return result == nil || strings.Contains(result.Content, "error")
    },
    workflow.NewStep("attempt", retryAgent),
    5, // Max 5 iterations
)
```

#### 4. Parallel

Parallel execution.

**Signature:**
```go
func NewParallel(name string, steps []Primitive) *Parallel
```

**Example:**
```go
parallel := workflow.NewParallel("gather",
    []workflow.Primitive{
        workflow.NewStep("source1", agent1),
        workflow.NewStep("source2", agent2),
        workflow.NewStep("source3", agent3),
    },
)
```

#### 5. Router

Dynamic routing.

**Signature:**
```go
func NewRouter(name string, routes map[string]Primitive,
               selector func(*ExecutionContext) string) *Router
```

**Example:**
```go
router := workflow.NewRouter("route_by_type",
    map[string]workflow.Primitive{
        "email":  workflow.NewStep("email", emailAgent),
        "chat":   workflow.NewStep("chat", chatAgent),
        "phone":  workflow.NewStep("phone", phoneAgent),
    },
    func(ctx *workflow.ExecutionContext) string {
        // Determine route based on input
        if strings.Contains(ctx.Input, "@") {
            return "email"
        }
        return "chat"
    },
)
```

### ExecutionContext

Access workflow context.

**Methods:**
```go
func (ctx *ExecutionContext) GetResult(stepName string) *RunOutput
func (ctx *ExecutionContext) SetData(key string, value interface{})
func (ctx *ExecutionContext) GetData(key string) interface{}
```

**Example:**
```go
step := workflow.NewStep("use_context", func(ctx *workflow.ExecutionContext) (*workflow.RunOutput, error) {
    previous := ctx.GetResult("previous_step")
    userData := ctx.GetData("user_data")

    // Use previous results
    return &workflow.RunOutput{Content: result}, nil
})
```

---

## Models

### OpenAI

**Create:**
```go
func New(modelID string, config Config) (*OpenAI, error)

type Config struct {
    APIKey      string  // Required
    BaseURL     string  // Optional (default: https://api.openai.com/v1)
    Temperature float64 // Optional (default: 0.7)
    MaxTokens   int     // Optional
}
```

**Supported Models:**
- `gpt-4`
- `gpt-4-turbo`
- `gpt-4o`
- `gpt-4o-mini`
- `gpt-3.5-turbo`

**Example:**
```go
model, err := openai.New("gpt-4o-mini", openai.Config{
    APIKey:      os.Getenv("OPENAI_API_KEY"),
    Temperature: 0.7,
    MaxTokens:   1000,
})
```

### Anthropic

**Create:**
```go
func New(modelID string, config Config) (*Anthropic, error)

type Config struct {
    APIKey      string  // Required
    Temperature float64 // Optional
    MaxTokens   int     // Optional (default: 4096)
}
```

**Supported Models:**
- `claude-3-5-sonnet-20241022`
- `claude-3-opus-20240229`
- `claude-3-sonnet-20240229`
- `claude-3-haiku-20240307`

**Example:**
```go
model, err := anthropic.New("claude-3-5-sonnet-20241022", anthropic.Config{
    APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
    MaxTokens: 2048,
})
```

### Ollama

**Create:**
```go
func New(modelID string, config Config) (*Ollama, error)

type Config struct {
    BaseURL     string  // Optional (default: http://localhost:11434)
    Temperature float64 // Optional
}
```

**Supported Models:**
- Any model available in local Ollama installation
- Common: `llama3`, `mistral`, `codellama`, `phi`

**Example:**
```go
model, err := ollama.New("llama3", ollama.Config{
    BaseURL:     "http://localhost:11434",
    Temperature: 0.8,
})
```

---

## Tools

### Calculator

Basic math operations.

**Create:**
```go
func New() *Calculator
```

**Functions:**
- `add(a, b)`: Addition
- `subtract(a, b)`: Subtraction
- `multiply(a, b)`: Multiplication
- `divide(a, b)`: Division

**Example:**
```go
calc := calculator.New()

ag, _ := agent.New(agent.Config{
    Toolkits: []toolkit.Toolkit{calc},
    // ...
})
```

### HTTP

HTTP client for GET/POST requests.

**Create:**
```go
func New() *HTTPToolkit
```

**Functions:**
- `http_get(url)`: HTTP GET request
- `http_post(url, body)`: HTTP POST request

**Example:**
```go
http := http.New()

ag, _ := agent.New(agent.Config{
    Toolkits: []toolkit.Toolkit{http},
    Instructions: "You can make HTTP requests to fetch data.",
})
```

### File

File operations with safety controls.

**Create:**
```go
func New(config Config) *FileToolkit

type Config struct {
    AllowedPaths []string // Whitelist of allowed directories
    MaxFileSize  int64    // Max file size in bytes (default: 10MB)
}
```

**Functions:**
- `read_file(path)`: Read file content
- `write_file(path, content)`: Write file
- `list_files(directory)`: List directory
- `delete_file(path)`: Delete file

**Example:**
```go
file := file.New(file.Config{
    AllowedPaths: []string{"/data", "/tmp"},
    MaxFileSize:  5 * 1024 * 1024, // 5MB
})

ag, _ := agent.New(agent.Config{
    Toolkits: []toolkit.Toolkit{file},
})
```

### Custom Tools

Create custom tools:

```go
type MyToolkit struct {
    *toolkit.BaseToolkit
}

func NewMyToolkit() *MyToolkit {
    t := &MyToolkit{
        BaseToolkit: toolkit.NewBaseToolkit("my_tools"),
    }

    t.RegisterFunction(&toolkit.Function{
        Name:        "my_function",
        Description: "Description of what this function does",
        Parameters: map[string]toolkit.Parameter{
            "input": {
                Type:        "string",
                Description: "Input parameter description",
                Required:    true,
            },
            "optional": {
                Type:        "number",
                Description: "Optional parameter",
                Required:    false,
            },
        },
        Handler: t.myHandler,
    })

    return t
}

func (t *MyToolkit) myHandler(args map[string]interface{}) (interface{}, error) {
    input := args["input"].(string)
    // Process input
    return result, nil
}
```

---

## Memory

### NewInMemory

Create in-memory conversation storage.

**Signature:**
```go
func NewInMemory(maxSize int) *InMemory
```

**Parameters:**
- `maxSize`: Maximum number of messages to keep

**Methods:**
```go
func (m *InMemory) Add(msg *types.Message)
func (m *InMemory) GetMessages() []*types.Message
func (m *InMemory) Clear()
```

**Example:**
```go
mem := memory.NewInMemory(100) // Keep last 100 messages

ag, _ := agent.New(agent.Config{
    Memory: mem,
    // ...
})

// Later
ag.ClearMemory() // Clear all messages
```

---

## Types

### Messages

**Message Types:**
```go
const (
    RoleSystem    = "system"
    RoleUser      = "user"
    RoleAssistant = "assistant"
    RoleTool      = "tool"
)

type Message struct {
    Role      string
    Content   string
    ToolCalls []ToolCall
}
```

**Create Messages:**
```go
msg := types.NewSystemMessage("You are a helpful assistant")
msg := types.NewUserMessage("Hello")
msg := types.NewAssistantMessage("Hi there!")
msg := types.NewToolMessage("tool_id", "result")
```

### Errors

**Error Types:**
```go
type AgnoError struct {
    Code    ErrorCode
    Message string
    Cause   error
}

const (
    ErrCodeInvalidInput    ErrorCode = "INVALID_INPUT"
    ErrCodeInvalidConfig   ErrorCode = "INVALID_CONFIG"
    ErrCodeModelTimeout    ErrorCode = "MODEL_TIMEOUT"
    ErrCodeToolExecution   ErrorCode = "TOOL_EXECUTION"
    ErrCodeAPIError        ErrorCode = "API_ERROR"
    ErrCodeRateLimit       ErrorCode = "RATE_LIMIT"
)
```

**Create Errors:**
```go
err := types.NewInvalidInputError("input cannot be empty")
err := types.NewModelTimeoutError(30 * time.Second)
err := types.NewToolExecutionError("calculator", originalError)
```

**Check Errors:**
```go
if types.IsInvalidInputError(err) {
    // Handle invalid input
}

if types.IsRateLimitError(err) {
    // Handle rate limit
}
```

---

## AgentOS Server

### NewServer

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

### Server.RegisterAgent

Register an agent.

**Signature:**
```go
func (s *Server) RegisterAgent(agentID string, ag *agent.Agent) error
```

**Example:**
```go
err := server.RegisterAgent("assistant", myAgent)
```

### Server.Start / Shutdown

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

### API Endpoints

See [OpenAPI Specification](../pkg/agentos/openapi.yaml) for complete API documentation.

**Core Endpoints:**
- `GET /health` - Health check
- `POST /api/v1/sessions` - Create session
- `GET /api/v1/sessions/{id}` - Get session
- `PUT /api/v1/sessions/{id}` - Update session
- `DELETE /api/v1/sessions/{id}` - Delete session
- `GET /api/v1/sessions` - List sessions
- `GET /api/v1/agents` - List agents
- `POST /api/v1/agents/{id}/run` - Run agent

---

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

---

For more examples, see [Quick Start Guide](QUICK_START.md) and [examples directory](../cmd/examples/).
