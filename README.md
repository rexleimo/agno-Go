# agno-Go

Agno-Go: A High-Performance Multi-Agent System Framework Based on Golang. Inheriting the Agno design philosophy, it leverages Golang's concurrency model and performance advantages to build efficient, scalable AI agent systems.

## 🚀 Features

- **Simple & Powerful**: Clean API design following KISS principle
- **High Performance**: ~1μs agent instantiation, <3KB memory per agent
- **Flexible Tools**: Easy-to-extend toolkit system
- **Multi-Model Support**: OpenAI and more LLM providers
- **Production Ready**: Built-in error handling, logging, and testing

## 📦 Installation

```bash
go get github.com/yourusername/agno-go
```

## 🎯 Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/yourusername/agno-go/pkg/agno/agent"
    "github.com/yourusername/agno-go/pkg/agno/models/openai"
    "github.com/yourusername/agno-go/pkg/agno/tools/calculator"
)

func main() {
    // Create model
    model, _ := openai.New("gpt-4o-mini", openai.Config{
        APIKey: "your-api-key",
    })

    // Create agent with tools
    ag, _ := agent.New(agent.Config{
        Name:     "Assistant",
        Model:    model,
        Toolkits: []toolkit.Toolkit{calculator.New()},
    })

    // Run agent
    output, _ := ag.Run(context.Background(), "What is 25 * 4 + 15?")
    fmt.Println(output.Content) // Output: 115
}
```

## 📖 Core Concepts

### Agent
An autonomous AI agent that can use tools and maintain conversation context.

```go
agent, err := agent.New(agent.Config{
    Name:         "My Agent",
    Model:        model,
    Toolkits:     []toolkit.Toolkit{httpTools, calcTools},
    Instructions: "You are a helpful assistant",
    MaxLoops:     10,
})
```

### Models
Abstraction over different LLM providers. Currently supports:
- ✅ OpenAI (GPT-4, GPT-3.5, etc.)
- 🚧 Anthropic (Coming soon)
- 🚧 Google (Coming soon)

```go
model, err := openai.New("gpt-4o-mini", openai.Config{
    APIKey:      os.Getenv("OPENAI_API_KEY"),
    Temperature: 0.7,
    MaxTokens:   1000,
})
```

### Tools
Extend agent capabilities with custom functions.

```go
type MyToolkit struct {
    *toolkit.BaseToolkit
}

func New() *MyToolkit {
    t := &MyToolkit{
        BaseToolkit: toolkit.NewBaseToolkit("my_tools"),
    }

    t.RegisterFunction(&toolkit.Function{
        Name:        "my_function",
        Description: "Does something useful",
        Parameters: map[string]toolkit.Parameter{
            "input": {Type: "string", Required: true},
        },
        Handler: t.myHandler,
    })

    return t
}
```

### Memory
Manages conversation history with automatic truncation.

```go
memory := memory.NewInMemory(100) // Keep last 100 messages
```

## 🛠️ Built-in Tools

- **Calculator**: Basic math operations (add, subtract, multiply, divide)
- **HTTP**: Make HTTP GET/POST requests
- More coming soon...

## 📁 Project Structure

```
agno-go/
├── pkg/agno/
│   ├── agent/          # Agent core
│   ├── models/         # LLM providers
│   │   ├── openai/     # OpenAI implementation
│   │   └── base.go     # Model interface
│   ├── tools/          # Tool system
│   │   ├── toolkit/    # Toolkit interface
│   │   ├── calculator/ # Calculator tools
│   │   └── http/       # HTTP tools
│   ├── memory/         # Memory management
│   └── types/          # Core types
├── cmd/examples/       # Example programs
├── docs/               # Documentation
├── Makefile            # Build commands
└── go.mod              # Dependencies
```

## 🧪 Testing

Run tests:
```bash
make test
```

Run tests with coverage:
```bash
make coverage
```

Run linter:
```bash
make lint
```

## 📚 Examples

See [`cmd/examples/`](cmd/examples/) for complete examples:
- `simple_agent`: Basic agent with calculator tools
- `team_demo`: Multi-agent collaboration with 4 coordination modes
- `workflow_demo`: Workflow engine with 5 control flow primitives

## 🎯 Roadmap

### Week 1-2: Core Framework ✅
- [x] Project setup
- [x] Core types (Message, Response, Errors)
- [x] Model interface + OpenAI implementation
- [x] Toolkit system + basic tools
- [x] Agent with Run method
- [x] Unit tests
- [x] Example programs

### Week 3-4: Extensions (🟡 40% Complete)
- [x] Team (multi-agent collaboration) - 4 modes, 92.3% test coverage
- [x] Workflow engine - 5 primitives, 80.4% test coverage
- [ ] More LLM providers (Anthropic, Google, Groq, Ollama)
- [ ] More tools (10+ tools)

### Week 5-6: Storage & Knowledge
- [ ] Vector database integrations
- [ ] Knowledge base
- [ ] Session management

### Week 7: Web API
- [ ] RESTful API (Gin framework)
- [ ] WebSocket streaming
- [ ] Authentication

### Week 8: Production Ready
- [ ] Performance optimization
- [ ] Complete documentation
- [ ] v1.0.0 release

## 🤝 Contributing

Contributions are welcome! Please read our [Team Guide](docs/TEAM_GUIDE.md) for development workflow.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

Inspired by [Agno Python](https://github.com/agno-agi/agno) framework.

## 📞 Contact

- Issues: [GitHub Issues](https://github.com/yourusername/agno-go/issues)
- Discussions: [GitHub Discussions](https://github.com/yourusername/agno-go/discussions)
