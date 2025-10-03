# Agent API Reference

## agent.New

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

## Agent.Run

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

## Agent.ClearMemory

Clear conversation memory.

**Signature:**
```go
func (a *Agent) ClearMemory()
```

**Example:**
```go
ag.ClearMemory() // Start fresh conversation
```
