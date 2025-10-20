# Agno-Go

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Test Coverage](https://img.shields.io/badge/coverage-80.8%25-brightgreen.svg)](docs/DEVELOPMENT.md#testing-standards)
[![Release](https://img.shields.io/badge/release-v1.2.5-blue.svg)](CHANGELOG.md)

**Agno-Go** is a high-performance multi-agent framework written in Go. It keeps the KISS philosophy of the Agno project while embracing Go’s strengths: lightweight goroutines, a tiny memory footprint, single static binaries, and a batteries-included toolchain.

---

## Feature Highlights

- **🚀 Extreme performance** – agent instantiation in ~180 ns and (~1.2 KB) memory per agent, 16× faster than the Python version.
- **🤖 Production ready** – AgentOS REST server (OpenAPI 3.0), session storage, health checks, structured logging, CORS, request timeouts.
- **🧩 Flexible architecture** – build with Agents, Teams (4 coordination modes), or Workflows (5 primitives) and mix freely.
- **🔌 Multi-provider models** – OpenAI (incl. o-series reasoning), Anthropic Claude, Google Gemini, DeepSeek, GLM, ModelScope, Ollama, Cohere, Groq, Together, OpenRouter, LM Studio, Vercel, Portkey, InternLM, SambaNova.
- **🔧 Extensible tooling** – calculator, HTTP, file operations, search, plus an SDK for building bespoke toolkits or MCP connectors.
- **💾 Knowledge & RAG** – ChromaDB integration, batching utilities, and ingestion helpers.
- **🛡️ Guardrails & hooks** – prompt-injection guard, custom pre/post hooks, graceful degradation.
- **📊 Observability** – rich SSE event stream with reasoning snapshots, Logfire / OpenTelemetry sample included.

---

## Getting Started

```bash
go get github.com/rexleimo/agno-go
```

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rexleimo/agno-go/pkg/agno/agent"
	"github.com/rexleimo/agno-go/pkg/agno/models/openai"
	"github.com/rexleimo/agno-go/pkg/agno/tools/calculator"
	"github.com/rexleimo/agno-go/pkg/agno/tools/toolkit"
)

func main() {
	model, _ := openai.New("gpt-4o-mini", openai.Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	})

	ag, _ := agent.New(agent.Config{
		Name:     "Math Assistant",
		Model:    model,
		Toolkits: []toolkit.Toolkit{calculator.New()},
	})

	output, _ := ag.Run(context.Background(), "What is 25 * 4 + 15?")
	fmt.Println(output.Content)
}
```

Run the production server with Docker:

```bash
docker compose up -d
curl http://localhost:8080/health
```

---

## Documentation

| Resource | Link |
| --- | --- |
| Guides | https://rexleimo.github.io/agno-Go/guide/ |
| API Reference | https://rexleimo.github.io/agno-Go/api/ |
| Advanced Topics | https://rexleimo.github.io/agno-Go/advanced/ |
| Examples | https://rexleimo.github.io/agno-Go/examples/ |
| Release Notes | https://rexleimo.github.io/agno-Go/release-notes |
| Internal / WIP Docs | [`docs/`](docs/) |

`docs/README.md` explains the split between the public site (`website/`) and internal design notes (`docs/`).

---

## Observability & Reasoning

- **SSE Event Stream** – `POST /api/v1/agents/{id}/run/stream?types=run_start,reasoning,token,complete` emits structured events. `reasoning` events carry token counts, redacted transcripts, and provider metadata; `complete` events summarise the run.
- **Logfire Integration** – `cmd/examples/logfire_observability` shows how to export spans with OpenTelemetry (build with `-tags logfire`). Detailed walkthrough: [`docs/release/logfire_observability.md`](docs/release/logfire_observability.md).

---

## Example Catalogue

| Example | Highlights | Run |
| --- | --- | --- |
| **Simple Agent** (`cmd/examples/simple_agent/`) | GPT‑4o mini, calculator toolkit, single agent | `go run cmd/examples/simple_agent/main.go` |
| **Claude Agent** (`cmd/examples/claude_agent/`) | Anthropic Claude 3.5, HTTP + calculator tools | `go run cmd/examples/claude_agent/main.go` |
| **Ollama Agent** (`cmd/examples/ollama_agent/`) | Local Llama 3 via Ollama, file operations | `go run cmd/examples/ollama_agent/main.go` |
| **Team Demo** (`cmd/examples/team_demo/`) | 4 coordination modes, researcher + writer workflow | `go run cmd/examples/team_demo/main.go` |
| **Workflow Demo** (`cmd/examples/workflow_demo/`) | Step / condition / loop / parallel orchestration | `go run cmd/examples/workflow_demo/main.go` |
| **RAG Demo** (`cmd/examples/rag_demo/`) | ChromaDB, embeddings, document Q&A | `go run cmd/examples/rag_demo/main.go` |
| **Reasoning Demo** (`examples/reasoning/`) | OpenAI o1 / Gemini 2.5 thinking extraction | `go run examples/reasoning/main.go` |
| **Logfire Observability** (`cmd/examples/logfire_observability/`) | OpenTelemetry spans + reasoning metadata | `go run -tags logfire cmd/examples/logfire_observability/main.go` |

More details live in the [Examples documentation](website/examples/index.md).

---

## Development & Contribution

1. Read [`docs/DEVELOPMENT.md`](docs/DEVELOPMENT.md) for tooling, linting, and testing workflow.
2. Docs: follow the structure in [`docs/README.md`](docs/README.md) and update the VitePress site (`website/`) when promoting features.
3. Run targeted tests, then `go test ./...` (with `GOCACHE=$(pwd)/.gocache` when sandboxed).
4. Submit PRs with lint/test evidence. Adhere to Conventional Commits.

For ongoing changes and release scope, see [`CHANGELOG.md`](CHANGELOG.md) and the VitePress site’s release notes (`website/release-notes.md`).

---

## License

MIT © [Contributors](https://github.com/rexleimo/agno-Go/graphs/contributors). Inspired by the [Agno (Python)](https://github.com/agno-agi/agno) framework.
