# Agno-Go Quick Start Guide

Get started with Agno-Go in less than 5 minutes!

## Prerequisites

- Go 1.21 or later
- OpenAI API key (or Anthropic/Ollama)
- Basic understanding of AI agents

## Installation

### Option 1: Using Go Get

```bash
go get github.com/rexleimo/agno-go
```

### Option 2: Clone Repository

```bash
git clone https://github.com/rexleimo/agno-go.git
cd agno-Go
go mod download
```

## Your First Agent

### 1. Simple Agent (No Tools)

Create a file `main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
)

func main() {
    // Get API key from environment
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("OPENAI_API_KEY environment variable is required")
    }

    // Create OpenAI model
    model, err := openai.New("gpt-4o-mini", openai.Config{
        APIKey: apiKey,
    })
    if err != nil {
        log.Fatalf("Failed to create model: %v", err)
    }

    // Create agent
    ag, err := agent.New(agent.Config{
        Name:         "Assistant",
        Model:        model,
        Instructions: "You are a helpful assistant.",
    })
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    // Run agent
    output, err := ag.Run(context.Background(), "What is the capital of France?")
    if err != nil {
        log.Fatalf("Agent run failed: %v", err)
    }

    fmt.Println("Agent:", output.Content)
}
```

**Run it:**

```bash
export OPENAI_API_KEY=sk-your-key-here
go run main.go
```

**Expected output:**

```
Agent: The capital of France is Paris.
```

### 2. Agent with Tools

Add calculator tools:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
    "github.com/rexleimo/agno-go/pkg/agno/tools/calculator"
    "github.com/rexleimo/agno-go/pkg/agno/tools/toolkit"
)

func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("OPENAI_API_KEY required")
    }

    // Create model
    model, _ := openai.New("gpt-4o-mini", openai.Config{
        APIKey: apiKey,
    })

    // Create agent WITH tools
    ag, _ := agent.New(agent.Config{
        Name:  "Calculator Agent",
        Model: model,
        Toolkits: []toolkit.Toolkit{
            calculator.New(),
        },
        Instructions: "You are a math assistant. Use the calculator tools for calculations.",
    })

    // Ask a math question
    output, _ := ag.Run(context.Background(), "What is 123 * 456 + 789?")

    fmt.Println("Question: What is 123 * 456 + 789?")
    fmt.Println("Agent:", output.Content)
}
```

**Run it:**

```bash
go run main.go
```

**Expected output:**

```
Question: What is 123 * 456 + 789?
Agent: The result is 56,877 (123 * 456 = 56,088, then 56,088 + 789 = 56,877)
```

### 3. Multi-Turn Conversation

Add memory for conversation:

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
)

func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("OPENAI_API_KEY required")
    }

    model, _ := openai.New("gpt-4o-mini", openai.Config{
        APIKey: apiKey,
    })

    ag, _ := agent.New(agent.Config{
        Name:         "Chat Assistant",
        Model:        model,
        Instructions: "You are a friendly chatbot. Remember context from previous messages.",
    })

    fmt.Println("Chat Assistant (type 'quit' to exit)")
    fmt.Println("=====================================")

    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Print("\nYou: ")
        if !scanner.Scan() {
            break
        }

        input := strings.TrimSpace(scanner.Text())
        if input == "quit" || input == "exit" {
            fmt.Println("Goodbye!")
            break
        }

        if input == "" {
            continue
        }

        // Run agent (memory is automatically maintained)
        output, err := ag.Run(context.Background(), input)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            continue
        }

        fmt.Printf("Agent: %s\n", output.Content)
    }
}
```

**Example conversation:**

```
You: My name is Alice
Agent: Nice to meet you, Alice! How can I help you today?

You: What's my name?
Agent: Your name is Alice!

You: quit
Goodbye!
```

## Using AgentOS (HTTP Server)

### 1. Start the Server

#### Using Docker Compose (Recommended)

```bash
# Copy environment template
cp .env.example .env

# Edit .env and add your API key
nano .env  # Add: OPENAI_API_KEY=sk-your-key

# Start server
docker-compose up -d

# Check health
curl http://localhost:8080/health
```

#### Using Go (Native)

```bash
# Build server
go build -o agentos cmd/server/main.go

# Run server
export OPENAI_API_KEY=sk-your-key
./agentos
```

### 2. Register an Agent

Create a simple server with registered agent:

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
    "github.com/rexleimo/agno-go/pkg/agentos"
)

func main() {
    // Create model
    model, err := openai.New("gpt-4o-mini", openai.Config{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create agent
    ag, err := agent.New(agent.Config{
        Name:         "Assistant",
        Model:        model,
        Instructions: "You are a helpful assistant.",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create server
    server, err := agentos.NewServer(&agentos.Config{
        Address: ":8080",
        Debug:   true,
    })
    if err != nil {
        log.Fatal(err)
    }

    // Register agent
    if err := server.RegisterAgent("assistant", ag); err != nil {
        log.Fatal(err)
    }

    // Start server
    go func() {
        if err := server.Start(); err != nil {
            log.Printf("Server error: %v", err)
        }
    }()

    log.Println("Server started on :8080")
    log.Println("Try: curl -X POST http://localhost:8080/api/v1/agents/assistant/run -H 'Content-Type: application/json' -d '{\"input\":\"Hello!\"}'")

    // Wait for interrupt
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    server.Shutdown(ctx)
}
```

### 3. Use the API

#### Health Check

```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "agentos",
  "time": 1704067200
}
```

#### List Agents

```bash
curl http://localhost:8080/api/v1/agents
```

**Response:**
```json
{
  "agents": [
    {
      "id": "assistant",
      "name": "Assistant"
    }
  ],
  "count": 1
}
```

#### Run Agent

```bash
curl -X POST http://localhost:8080/api/v1/agents/assistant/run \
  -H "Content-Type: application/json" \
  -d '{
    "input": "What is 2+2?"
  }'
```

**Response:**
```json
{
  "content": "2 + 2 equals 4.",
  "metadata": {
    "agent_id": "assistant"
  }
}
```

#### Create Session

```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "assistant",
    "user_id": "user-123",
    "name": "My Chat Session"
  }'
```

**Response:**
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "agent_id": "assistant",
  "user_id": "user-123",
  "name": "My Chat Session",
  "created_at": "2025-10-02T00:00:00Z",
  "updated_at": "2025-10-02T00:00:00Z"
}
```

#### Run Agent with Session

```bash
curl -X POST http://localhost:8080/api/v1/agents/assistant/run \
  -H "Content-Type: application/json" \
  -d '{
    "input": "My name is Alice",
    "session_id": "550e8400-e29b-41d4-a716-446655440000"
  }'

# Later, in the same session...
curl -X POST http://localhost:8080/api/v1/agents/assistant/run \
  -H "Content-Type: application/json" \
  -d '{
    "input": "What is my name?",
    "session_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

**Second Response:**
```json
{
  "content": "Your name is Alice!",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "metadata": {
    "agent_id": "assistant"
  }
}
```

## Advanced Examples

### Multi-Agent Team

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
    "github.com/rexleimo/agno-go/pkg/agno/team"
)

func main() {
    model, _ := openai.New("gpt-4o-mini", openai.Config{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })

    // Create researcher agent
    researcher, _ := agent.New(agent.Config{
        Name:         "Researcher",
        Model:        model,
        Instructions: "You are a research specialist. Gather information.",
    })

    // Create writer agent
    writer, _ := agent.New(agent.Config{
        Name:         "Writer",
        Model:        model,
        Instructions: "You are a writing specialist. Create compelling content.",
    })

    // Create team
    tm, _ := team.New(team.Config{
        Name:  "Content Team",
        Agents: []*agent.Agent{researcher, writer},
        Mode:  team.ModeSequential, // Researcher → Writer
    })

    // Run team
    output, err := tm.Run(context.Background(), "Write a short article about Go programming")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(output.Content)
}
```

### Workflow with Conditions

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
    "github.com/rexleimo/agno-go/pkg/agno/workflow"
)

func main() {
    model, _ := openai.New("gpt-4o-mini", openai.Config{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })

    classifier, _ := agent.New(agent.Config{
        Name:         "Classifier",
        Model:        model,
        Instructions: "Classify the sentiment as positive or negative. Respond with just 'positive' or 'negative'.",
    })

    positiveHandler, _ := agent.New(agent.Config{
        Name:         "Positive Handler",
        Model:        model,
        Instructions: "Respond enthusiastically to positive feedback.",
    })

    negativeHandler, _ := agent.New(agent.Config{
        Name:         "Negative Handler",
        Model:        model,
        Instructions: "Respond empathetically to negative feedback.",
    })

    // Create workflow
    wf, _ := workflow.New(workflow.Config{
        Name: "Sentiment Workflow",
        Steps: []workflow.Primitive{
            workflow.NewStep("classify", classifier),
            workflow.NewCondition("route", func(ctx *workflow.ExecutionContext) bool {
                result := ctx.GetResult("classify")
                return result != nil && result.Content == "positive"
            },
                workflow.NewStep("positive", positiveHandler),
                workflow.NewStep("negative", negativeHandler),
            ),
        },
    })

    // Test with positive feedback
    output, _ := wf.Run(context.Background(), "I love this product!")
    fmt.Println("Positive:", output.Content)

    // Test with negative feedback
    output, _ = wf.Run(context.Background(), "This is terrible.")
    fmt.Println("Negative:", output.Content)
}
```

## Next Steps

### Learn More

1. **[Architecture Guide](ARCHITECTURE.md)** - Understand the design
2. **[API Documentation](../pkg/agentos/README.md)** - Complete API reference
3. **[Deployment Guide](DEPLOYMENT.md)** - Production deployment
4. **[Examples](../cmd/examples/)** - More code examples

### Try Advanced Features

- **RAG**: Try `cmd/examples/rag_demo`
- **Local Models**: Try `cmd/examples/ollama_agent`
- **Claude**: Try `cmd/examples/claude_agent`

### Get Help

- 📖 [Full Documentation](https://docs.agno.com)
- 💬 [GitHub Discussions](https://github.com/rexleimo/agno-go/discussions)
- 🐛 [Report Issues](https://github.com/rexleimo/agno-go/issues)

## Troubleshooting

### Common Issues

**1. "OPENAI_API_KEY not set"**

```bash
export OPENAI_API_KEY=sk-your-key-here
```

**2. "Module not found"**

```bash
go mod download
go mod tidy
```

**3. "Port 8080 already in use"**

```bash
# Change port in .env or Config
AGENTOS_ADDRESS=:9090
```

**4. "Context deadline exceeded"**

```bash
# Increase timeout
export REQUEST_TIMEOUT=60
```

### Getting Debug Logs

```bash
export AGENTOS_DEBUG=true
export LOG_LEVEL=debug
```

## Tips & Best Practices

### 1. Use Environment Variables

```bash
# .env file
OPENAI_API_KEY=sk-your-key
ANTHROPIC_API_KEY=sk-ant-your-key
LOG_LEVEL=info
```

### 2. Error Handling

```go
output, err := ag.Run(ctx, input)
if err != nil {
    // Check error type
    if agno.IsInvalidInputError(err) {
        // Handle invalid input
    } else if agno.IsRateLimitError(err) {
        // Handle rate limit
    } else {
        // Handle other errors
    }
}
```

### 3. Use Contexts

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

output, err := ag.Run(ctx, input)
```

### 4. Manage Memory

```go
// Clear memory when needed
ag.ClearMemory()

// Or use custom memory with limits
mem := memory.NewInMemory(100) // Keep last 100 messages
ag, _ := agent.New(agent.Config{
    Memory: mem,
    // ...
})
```

## Quick Reference

### Common Imports

```go
import (
    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
    "github.com/rexleimo/agno-go/pkg/agno/tools/calculator"
    "github.com/rexleimo/agno-go/pkg/agno/team"
    "github.com/rexleimo/agno-go/pkg/agno/workflow"
    "github.com/rexleimo/agno-go/pkg/agentos"
)
```

### Agent Creation Template

```go
model, err := openai.New("gpt-4o-mini", openai.Config{
    APIKey: os.Getenv("OPENAI_API_KEY"),
})

ag, err := agent.New(agent.Config{
    Name:         "Agent Name",
    Model:        model,
    Toolkits:     []toolkit.Toolkit{/* tools */},
    Instructions: "System instructions",
    MaxLoops:     10,
})

output, err := ag.Run(context.Background(), "input")
```

### Server Creation Template

```go
server, err := agentos.NewServer(&agentos.Config{
    Address: ":8080",
    Debug:   true,
})

server.RegisterAgent("agent-id", agent)
server.Start()
```

---

**Happy Coding! 🚀**

For more examples, check out the [examples directory](../cmd/examples/).
