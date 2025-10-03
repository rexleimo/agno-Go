# Agno-Go

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Test Coverage](https://img.shields.io/badge/coverage-80.8%25-brightgreen.svg)](docs/TEST_REPORT.md)
[![Release](https://img.shields.io/badge/release-v1.0.0-blue.svg)](CHANGELOG.md)

**Agno-Go** is a high-performance multi-agent system framework built with Go. Inheriting the Agno design philosophy, it leverages Go's concurrency model and performance advantages to build efficient, scalable AI agent systems.

## ✨ Highlights

- **🚀 High Performance**: 180ns agent instantiation, 1.2KB memory per agent ([16x faster than Python](docs/PERFORMANCE.md))
- **🤖 Production-Ready**: AgentOS HTTP server with RESTful API, session management, and agent registry
- **🧩 Flexible Architecture**: Agent, Team (4 modes), Workflow (5 primitives)
- **🔧 Extensible Tools**: Easy-to-extend toolkit system with built-in tools
- **🔌 Multi-Model Support**: OpenAI, Anthropic Claude, Ollama (local models)
- **💾 RAG Support**: ChromaDB integration and OpenAI embeddings
- **✅ Well-Tested**: 80.8% test coverage, 85+ test cases, 100% pass rate
- **📦 Easy Deployment**: Docker, Docker Compose, Kubernetes manifests included
- **📚 Complete Documentation**: API docs (OpenAPI 3.0), deployment guides, examples

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
Abstraction over different LLM providers. We support 6 major providers:
- ✅ **OpenAI** (GPT-4, GPT-3.5, etc.) - 44.6% test coverage
- ✅ **Anthropic Claude** (Claude 3 Opus, Sonnet, Haiku) - 50.9% test coverage
- ✅ **Ollama** (Llama 2, Mistral, CodeLlama, all local models) - 43.8% test coverage
- ✅ **DeepSeek** (DeepSeek-V2, DeepSeek-Coder)
- ✅ **Google Gemini** (Gemini Pro, Flash)
- ✅ **ModelScope** (Qwen, Yi models)

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

Following KISS principle, we provide essential tools with high quality:
- **Calculator**: Basic math operations (75.6% coverage)
- **HTTP**: Make HTTP GET/POST requests (88.9% coverage)
- **File Operations**: Read, write, list, delete files with security controls (76.2% coverage)
- **Search**: DuckDuckGo web search (92.1% coverage)

## 🧠 Knowledge & RAG

Build intelligent agents with knowledge bases and semantic search:

### Vector Database
- **ChromaDB**: Full integration with local and cloud instances
- Automatic embedding generation
- Metadata filtering and semantic search

### Embeddings
- **OpenAI**: text-embedding-3-small/large support
- Automatic batch processing
- 1536/3072 dimensional embeddings

### Example RAG Application

```go
// Create embedding function
embedFunc, _ := openai.New(openai.Config{
    APIKey: os.Getenv("OPENAI_API_KEY"),
    Model:  "text-embedding-3-small",
})

// Create vector database
db, _ := chromadb.New(chromadb.Config{
    CollectionName:    "knowledge_base",
    EmbeddingFunction: embedFunc,
})

// Add documents (embeddings generated automatically)
db.Add(ctx, []vectordb.Document{
    {ID: "doc1", Content: "AI is the future..."},
})

// Query with natural language
results, _ := db.Query(ctx, "What is AI?", 5, nil)
```

See [RAG Demo](cmd/examples/rag_demo/) for a complete example.

## 📁 Project Structure

```
agno-go/
├── pkg/agno/
│   ├── agent/          # Agent core (74.7% coverage)
│   ├── team/           # Multi-agent collaboration (92.3% coverage)
│   ├── workflow/       # Workflow engine (80.4% coverage)
│   ├── models/         # LLM providers (6 providers)
│   │   ├── openai/     # OpenAI (44.6% coverage)
│   │   ├── anthropic/  # Claude (50.9% coverage)
│   │   ├── ollama/     # Ollama (43.8% coverage)
│   │   ├── deepseek/   # DeepSeek
│   │   ├── gemini/     # Google Gemini
│   │   ├── modelscope/ # ModelScope
│   │   └── base.go     # Model interface
│   ├── tools/          # Tool system
│   │   ├── toolkit/    # Toolkit interface (91.7% coverage)
│   │   ├── calculator/ # Math tools (75.6% coverage)
│   │   ├── http/       # HTTP tools (88.9% coverage)
│   │   ├── file/       # File operations (76.2% coverage)
│   │   └── search/     # Web search (92.1% coverage)
│   ├── vectordb/       # Vector database
│   │   ├── base.go     # VectorDB interface
│   │   └── chromadb/   # ChromaDB implementation
│   ├── embeddings/     # Embedding functions
│   │   └── openai/     # OpenAI embeddings
│   ├── knowledge/      # Knowledge management
│   ├── memory/         # Memory management (93.1% coverage)
│   └── types/          # Core types (100% coverage ⭐)
├── cmd/examples/       # Example programs
│   ├── simple_agent/   # Basic agent example
│   ├── team_demo/      # Multi-agent collaboration
│   ├── workflow_demo/  # Workflow example
│   └── rag_demo/       # RAG pipeline example
├── docs/               # Documentation
│   ├── PERFORMANCE.md  # Performance benchmarks
│   └── PROGRESS.md     # Development progress
├── Makefile            # Build commands
└── go.mod              # Dependencies
```

## 🧪 Testing

We maintain **>70% test coverage** for all core packages:

```bash
# Run all tests
make test

# Generate coverage report (creates coverage.html)
make coverage

# Run linter
make lint
```

**Current Coverage**:
- Types: 100% ⭐
- Memory: 93.1%
- Team: 92.3%
- Toolkit: 91.7%
- HTTP Tools: 88.9%
- Workflow: 80.4%
- File Tools: 76.2%
- Calculator: 75.6%
- Agent: 74.7%

## 📚 Examples

See [`cmd/examples/`](cmd/examples/) for complete examples:
- `simple_agent`: Basic agent with OpenAI and calculator tools
- `claude_agent`: Anthropic Claude integration with tools
- `ollama_agent`: Local model support with Ollama
- `team_demo`: Multi-agent collaboration with 4 coordination modes
- `workflow_demo`: Workflow engine with 5 control flow primitives

## 🎯 Roadmap

> **KISS Principle**: Focus on quality over quantity. 3 core LLMs, 5 essential tools, 1 vector DB.

### ✅ M1: Core Framework (Week 1-2) - COMPLETED
- [x] Agent core with Run method (74.7% coverage)
- [x] OpenAI model integration (44.6% coverage)
- [x] Basic tools: Calculator, HTTP, File Operations
- [x] Memory management (93.1% coverage)
- [x] Types package (100% coverage ⭐)
- [x] Example programs

### ✅ M2: Extensions (Week 3-4) - 100% COMPLETE
- [x] Team (4 coordination modes, 92.3% coverage)
- [x] Workflow (5 primitives, 80.4% coverage)
- [x] Anthropic Claude integration (50.9% coverage)
- [x] Ollama local model support (43.8% coverage)
- [x] DuckDuckGo search tool (92.1% coverage)
- [x] Performance benchmarks ([details](docs/PERFORMANCE.md))
- [x] Model provider refactoring (common utilities, 84.8% coverage)
- [x] Documentation (README, CLAUDE.md, models/README.md)

**Performance Achieved**:
- ⚡ Agent instantiation: **180ns** (5x better than 1μs target)
- 💾 Memory per agent: **1.2KB** (60% better than 3KB target)
- 🚀 16x faster than Python version

### 🟢 M3: Knowledge & Storage (Week 5-6) - 60% COMPLETE
- [x] VectorDB interface design
- [x] Knowledge package - Document loaders (Text, Directory, Reader)
- [x] Knowledge package - Chunkers (Character, Sentence, Paragraph)
- [ ] Vector DB implementation (ChromaDB or alternative)
- [ ] RAG workflow example

### ⏰ M4: Production Ready (Week 7-8) - PLANNED
- [ ] Performance optimization
- [ ] Complete documentation and examples
- [ ] v1.0.0 release

**See [PROGRESS.md](docs/PROGRESS.md) for detailed milestone tracking.**

## 🤝 Contributing

Contributions are welcome! Please read:
- [CLAUDE.md](CLAUDE.md) - Development guide and architecture
- [Team Guide](docs/TEAM_GUIDE.md) - Development workflow
- [Performance Guide](docs/PERFORMANCE.md) - Benchmarking standards

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

Inspired by [Agno Python](https://github.com/agno-agi/agno) framework.

## 📞 Contact

- Issues: [GitHub Issues](https://github.com/yourusername/agno-go/issues)
- Discussions: [GitHub Discussions](https://github.com/yourusername/agno-go/discussions)
