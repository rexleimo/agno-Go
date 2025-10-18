---
layout: home

hero:
  name: "Agno-Go"
  text: "High-Performance Multi-Agent Framework"
  tagline: "16x faster than Python | 180ns instantiation | 1.2KB memory per agent"
  image:
    src: /logo.svg
    alt: Agno-Go
  actions:
    - theme: brand
      text: Get Started
      link: /guide/quick-start
    - theme: alt
      text: View on GitHub
      link: https://github.com/rexleimo/agno-Go

features:
  - icon: 🚀
    title: Extreme Performance
    details: Agent instantiation in ~180ns, 16x faster than Python version. Memory footprint of just 1.2KB per agent with native Go concurrency support.

  - icon: 🤖
    title: Production-Ready
    details: AgentOS HTTP server with RESTful API, session management, agent registry, health monitoring, and comprehensive error handling out of the box.

  - icon: 🧩
    title: Flexible Architecture
    details: Choose from Agent (autonomous), Team (4 coordination modes), or Workflow (5 control primitives) to build your multi-agent system.

  - icon: 🔌
    title: Multi-Model Support
    details: Built-in support for OpenAI (GPT-4), Anthropic Claude, Ollama (local models), DeepSeek, Google Gemini, and ModelScope.

  - icon: 🔧
    title: Extensible Tools
    details: Easy-to-extend toolkit system with built-in Calculator, HTTP Client, File Operations, and DuckDuckGo Search. MCP integration for connecting to any MCP-compatible server.

  - icon: 💾
    title: RAG & Knowledge
    details: ChromaDB vector database integration with OpenAI embeddings. Build intelligent agents with semantic search and knowledge bases.

  - icon: ✅
    title: Well-Tested
    details: 80.8% test coverage with 85+ test cases and 100% pass rate. Production-quality code you can trust.

  - icon: 📦
    title: Easy Deployment
    details: Docker, Docker Compose, and Kubernetes manifests included. Deploy to any cloud platform in minutes with complete deployment guides.

  - icon: 📚
    title: Complete Documentation
    details: OpenAPI 3.0 specification, deployment guides, architecture docs, performance benchmarks, and working examples for every feature.
---

## Quick Example

Create an AI agent with tools in just a few lines:

```go
package main

import (
    "context"
    "fmt"
    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
    "github.com/rexleimo/agno-go/pkg/agno/tools/calculator"
)

func main() {
    // Create model
    model, _ := openai.New("gpt-4o-mini", openai.Config{
        APIKey: "your-api-key",
    })

    // Create agent with tools
    ag, _ := agent.New(agent.Config{
        Name:     "Math Assistant",
        Model:    model,
        Toolkits: []toolkit.Toolkit{calculator.New()},
    })

    // Run agent
    output, _ := ag.Run(context.Background(), "What is 25 * 4 + 15?")
    fmt.Println(output.Content) // Output: 115
}
```

## Performance Comparison

| Metric | Python Agno | Agno-Go | Improvement |
|--------|-------------|---------|-------------|
| Agent Creation | ~3μs | ~180ns | **16x faster** |
| Memory/Agent | ~6.5KB | ~1.2KB | **5.4x less** |
| Concurrency | GIL limited | Native goroutines | **Unlimited** |

## Why Agno-Go?

### Built for Production

Agno-Go isn't just a framework—it's a complete production system. The included **AgentOS** server provides:

- RESTful API with OpenAPI 3.0 specification
- Session management for multi-turn conversations
- Thread-safe agent registry
- Health monitoring and structured logging
- CORS support and request timeout handling

### KISS Principle

Following the **Keep It Simple, Stupid** philosophy:

- **3 core LLM providers** (not 45+) - OpenAI, Anthropic, Ollama
- **Essential tools** (not 115+) - Calculator, HTTP, File, Search
- **Quality over quantity** - Focus on production-ready features

### Developer Experience

- **Type-Safe**: Go's strong typing catches errors at compile time
- **Fast Builds**: Go's compilation speed enables rapid iteration
- **Easy Deployment**: Single binary with no runtime dependencies
- **Great Tooling**: Built-in testing, profiling, and race detection

## Get Started in 5 Minutes

```bash
# Clone repository
git clone https://github.com/rexleimo/agno-Go.git
cd agno-Go

# Set API key
export OPENAI_API_KEY=sk-your-key-here

# Run example
go run cmd/examples/simple_agent/main.go

# Or start AgentOS server
docker-compose up -d
curl http://localhost:8080/health
```

## What's Included

- **Core Framework**: Agent, Team (4 modes), Workflow (5 primitives)
- **Models**: OpenAI, Anthropic Claude, Ollama, DeepSeek, Gemini, ModelScope
- **Tools**: Calculator (75.6%), HTTP (88.9%), File (76.2%), Search (92.1%)
- **MCP Integration**: Model Context Protocol support for connecting to any MCP server
- **RAG**: ChromaDB integration + OpenAI embeddings
- **AgentOS**: Production HTTP server (65.0% coverage)
- **Examples**: 8 working examples covering all features (Simple Agent, Claude Agent, Ollama Agent, Team Demo, Workflow Demo, RAG Demo, MCP Demo, Logfire Observability)
- **Docs**: Complete guides, API reference, deployment instructions

## Community

- **GitHub**: [rexleimo/agno-Go](https://github.com/rexleimo/agno-Go)
- **Issues**: [Report bugs and request features](https://github.com/rexleimo/agno-Go/issues)
- **Discussions**: [Ask questions and share ideas](https://github.com/rexleimo/agno-Go/discussions)

## License

Agno-Go is released under the [MIT License](https://github.com/rexleimo/agno-Go/blob/main/LICENSE).

Inspired by [Agno (Python)](https://github.com/agno-agi/agno) framework.
