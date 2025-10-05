# Changelog

All notable changes to Agno-Go will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.3] - 2025-10-06

### 🧪 Improved

#### Testing & Quality
- **Enhanced JSON Serialization Tests** - Achieved 100% test coverage for utils/serialize package
  - Added error handling tests for unserializable types (channels, functions)
  - Added panic behavior tests for MustToJSONString
  - Added edge case tests (nil pointers, empty collections)
  - Test coverage: 92.3% → 100% ✅

#### Performance Benchmarks
- **Optimized Performance Tests** - Aligned with Python Agno performance testing patterns
  - Simplified agent instantiation benchmark (removed unnecessary variable)
  - Cleaned up tool registration patterns
  - Renamed test for consistency: "Tool Instantiation Performance" → "Agent Instantiation"

#### Documentation
- **Comprehensive Package Documentation** - Added bilingual (English/中文) documentation
  - Package-level overview with usage examples
  - Detailed function documentation with examples
  - Performance metrics included in package docs
  - All public APIs now fully documented

### 📊 Performance

Current benchmark results on Apple M3:
- **ToJSON**: ~600ns/op, 760B/op, 15 allocs/op
- **ConvertValue**: ~180ns/op, 392B/op, 5 allocs/op
- **Agent Creation**: ~180ns/op (16x faster than Python)

### 🔧 Technical Highlights

- **100% Test Coverage** - utils/serialize package now has complete test coverage
- **Better Error Handling** - Comprehensive tests for edge cases and error conditions
- **Production Ready** - Serialization utilities validated for WebSocket and API usage
- **Python Compatibility** - Prevents the JSON serialization bug found in Python Agno (commit aea0fc129)

### 📝 Files Changed

- `pkg/agno/utils/serialize.go` - Enhanced documentation with examples and performance notes
- `pkg/agno/utils/serialize_test.go` - Added 3 new test cases for error handling
- `pkg/agno/agent/agent_bench_test.go` - Simplified benchmark following Python patterns

### ✅ Migration Status

Completed migration items from Python Agno:
- ✅ JSON serialization bug fix (aea0fc129) - Already prevented in Go implementation
- ✅ Performance test optimization (e639f4996) - Applied to Go benchmarks
- 🔄 Custom route prefix (06baed104) - Deferred to Week 7 (AgentOS expansion)
- 🔄 HN tools update (24c3ee688) - Documentation only, no action needed

## [1.0.2] - 2025-10-05

### ✨ Added

#### New LLM Provider
- **GLM (智谱AI)** - Full integration with Zhipu AI's GLM models
  - Support for GLM-4, GLM-4V (vision), GLM-3-Turbo
  - Custom JWT authentication (HMAC-SHA256)
  - Synchronous API calls (`Invoke`)
  - Streaming responses (`InvokeStream`)
  - Tool/Function calling support
  - Test coverage: 57.2%

#### Implementation Details
- **pkg/agno/models/glm/glm.go** - Main model implementation (410 lines)
- **pkg/agno/models/glm/auth.go** - JWT authentication logic (59 lines)
- **pkg/agno/models/glm/types.go** - GLM API type definitions (105 lines)
- **pkg/agno/models/glm/glm_test.go** - Comprehensive unit tests (320 lines)
- **pkg/agno/models/glm/README.md** - Complete usage documentation

#### Examples & Documentation
- **cmd/examples/glm_agent/** - GLM agent example with calculator tools
  - Chinese language support demonstration
  - Multi-step calculation examples
  - Tool calling integration
- Updated README.md with GLM provider information
- Updated CLAUDE.md with GLM configuration and usage
- Added bilingual comments (English/中文) throughout codebase

### 🔧 Technical Highlights

- **Custom JWT Authentication** - Implemented GLM-specific JWT token generation
  - 7-day token expiration
  - Secure HMAC-SHA256 signing
  - Automatic token regeneration per request

- **OpenAI-Compatible Format** - API structure similar to OpenAI for easy integration
  - Request/response format alignment
  - Tool calling compatibility
  - Streaming support via Server-Sent Events (SSE)

- **Type Safety** - Full Go type system integration
  - Strongly-typed request/response structures
  - Error handling with custom error types
  - Context support for cancellation

### 📊 Test Results

- ✅ All 7 GLM tests passing
- ✅ 57.2% code coverage
- ✅ Race detector: PASS
- ✅ Build verification: SUCCESS

### 🌍 Environment Variables

New environment variable for GLM:
```bash
export ZHIPUAI_API_KEY=your-key-id.your-key-secret
```

### 📦 Dependencies

Added:
- `github.com/golang-jwt/jwt/v5 v5.3.0` - For JWT authentication

### 🎯 Supported Models

Total LLM providers increased from 6 to 7:
- OpenAI (GPT-4, GPT-3.5, GPT-4 Turbo)
- Anthropic (Claude 3.5 Sonnet, Claude 3 Opus/Sonnet/Haiku)
- **GLM (智谱AI: GLM-4, GLM-4V, GLM-3-Turbo)** ⭐ NEW
- Ollama (Local models)
- DeepSeek (DeepSeek-V2, DeepSeek-Coder)
- Google Gemini (Gemini Pro, Flash)
- ModelScope (Qwen, Yi models)

### 📝 Documentation Updates

- README.md - Added GLM to supported models list with example code
- CLAUDE.md - Added GLM environment variables and configuration
- Created pkg/agno/models/glm/README.md with comprehensive usage guide
- All code comments are bilingual (English/中文)

## [1.0.0] - 2025-10-02

### 🎉 Initial Release

Agno-Go v1.0 is a high-performance Go implementation of the Agno multi-agent framework, designed for building production-ready AI agent systems.

### ✨ Features

#### Core Agent System
- **Agent** - Single autonomous agent with tool support
  - LLM model integration
  - Tool/function calling
  - Conversation memory
  - System instructions
  - Max loop protection
  - Coverage: 74.7%

- **Team** - Multi-agent collaboration with 4 coordination modes:
  - `Sequential` - Agents work one after another
  - `Parallel` - All agents work simultaneously
  - `LeaderFollower` - Leader delegates tasks to followers
  - `Consensus` - Agents discuss until reaching agreement
  - Coverage: 92.3%

- **Workflow** - Step-based orchestration with 5 primitives:
  - `Step` - Basic workflow step (agent or function)
  - `Condition` - Conditional branching
  - `Loop` - Iterative loops
  - `Parallel` - Parallel execution
  - `Router` - Dynamic routing
  - Coverage: 80.4%

#### LLM Providers
- **OpenAI** - GPT-4, GPT-3.5, GPT-4 Turbo
- **Anthropic** - Claude 3.5 Sonnet, Claude 3 Opus/Sonnet/Haiku
- **Ollama** - Local model support (llama3, mistral, etc.)

#### Tools
- **Calculator** - Basic math operations
- **HTTP** - GET/POST requests
- **File** - File operations with safety controls

#### Storage & Memory
- **Memory** - In-memory conversation storage with auto-truncation
  - Configurable message limits
  - Thread-safe operations
  - Coverage: 93.1%

- **Session** - Session management for multi-turn conversations
  - In-memory storage (default)
  - PostgreSQL support (via schema)
  - Redis caching support

#### Vector Database
- **ChromaDB** - Vector storage for RAG applications
  - Document embedding
  - Semantic search
  - Collection management

#### AgentOS - Production Server
- **RESTful API** - Full-featured HTTP server
  - Session CRUD operations
  - Agent registration and execution
  - Health check endpoint
  - Structured logging with slog
  - CORS support
  - Request timeout handling
  - Coverage: 65.0%

- **Agent Registry** - Thread-safe agent management
  - Dynamic agent registration
  - Concurrent access support
  - Agent lifecycle management

#### Developer Experience
- **Types** - Comprehensive type system
  - Message types (System, User, Assistant, Tool)
  - Error types with codes
  - Model request/response types
  - 100% test coverage ⭐

- **Documentation**
  - Complete API documentation (OpenAPI 3.0)
  - Deployment guide (Docker, K8s, native)
  - Architecture documentation
  - Performance benchmarks
  - Code examples

#### Deployment
- **Docker** - Production-ready Dockerfile
  - Multi-stage build (~15MB final image)
  - Non-root user
  - Health checks
  - Security best practices

- **Docker Compose** - Full stack deployment
  - AgentOS server
  - PostgreSQL database
  - Redis cache
  - ChromaDB (optional)
  - Ollama (optional)

- **Kubernetes** - K8s manifests included
  - Deployment, Service, ConfigMap, Secret
  - Health probes
  - Resource limits
  - Horizontal Pod Autoscaling ready

### 📊 Performance

- **Agent Creation:** ~180ns/op (16x faster than Python)
- **Memory Footprint:** ~1.2KB per agent
- **Test Coverage:** 80.8% average across core packages
- **Concurrent Operations:** Fully thread-safe with RWMutex

### 🧪 Testing

- **85+ test cases** across all core packages
- **100% pass rate** ✅
- All packages exceed 70% coverage target
- Comprehensive integration tests
- Concurrent access tests
- Performance benchmarks

### 📚 Examples

- `simple_agent` - Basic agent with calculator
- `claude_agent` - Anthropic Claude integration
- `ollama_agent` - Local model support
- `team_demo` - Multi-agent collaboration
- `workflow_demo` - Workflow orchestration
- `rag_demo` - RAG with ChromaDB

### 🔧 Technical Details

**Dependencies:**
- Go 1.21+
- Gin web framework
- PostgreSQL 15+ (optional)
- Redis 7+ (optional)
- ChromaDB (optional)

**Project Structure:**
```
agno-Go/
├── pkg/agno/          # Core framework
│   ├── agent/         # Agent implementation
│   ├── team/          # Team coordination
│   ├── workflow/      # Workflow engine
│   ├── models/        # LLM providers
│   ├── tools/         # Tool integrations
│   ├── memory/        # Conversation memory
│   └── types/         # Core types
├── pkg/agentos/       # Production server
│   ├── server.go      # HTTP server
│   ├── registry.go    # Agent registry
│   └── openapi.yaml   # API specification
├── cmd/examples/      # Example programs
└── docs/              # Documentation
```

### 🎯 Design Philosophy

Agno-Go follows the **KISS principle** (Keep It Simple, Stupid):
- Focus on quality over quantity
- Clear, maintainable code
- Comprehensive testing
- Production-ready from day one

### 🔒 Security

- Non-root Docker container
- Secret management best practices
- Input validation
- Error handling
- Rate limiting support
- HTTPS/TLS ready

### 📖 Documentation

- [README.md](README.md) - Getting started
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - Architecture overview
- [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) - Deployment guide
- [docs/PERFORMANCE.md](docs/PERFORMANCE.md) - Performance benchmarks
- [docs/TEST_REPORT.md](docs/TEST_REPORT.md) - Test coverage report
- [pkg/agentos/README.md](pkg/agentos/README.md) - AgentOS API guide
- [pkg/agentos/openapi.yaml](pkg/agentos/openapi.yaml) - OpenAPI specification

### 🙏 Acknowledgments

Agno-Go is inspired by and compatible with the design philosophy of:
- [Agno](https://github.com/agno-agi/agno) - Python multi-agent framework

### 📝 Migration from Python Agno

Agno-Go maintains API compatibility where possible, making migration straightforward:

**Python:**
```python
from agno.agent import Agent
from agno.models.openai import OpenAI

agent = Agent(
    name="Assistant",
    model=OpenAI(id="gpt-4"),
)
response = agent.run("Hello!")
```

**Go:**
```go
import (
    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
)

model, _ := openai.New("gpt-4", openai.Config{...})
ag, _ := agent.New(agent.Config{
    Name: "Assistant",
    Model: model,
})
output, _ := ag.Run(ctx, "Hello!")
```

### 🚀 Getting Started

**Installation:**
```bash
go get github.com/rexleimo/agno-go
```

**Quick Start:**
```bash
# Clone repository
git clone https://github.com/rexleimo/agno-go
cd agno-Go

# Run example
export OPENAI_API_KEY=sk-...
go run cmd/examples/simple_agent/main.go

# Or use Docker
docker-compose up -d
curl http://localhost:8080/health
```

### 🛣️ Roadmap

**v1.1** (Planned)
- Streaming response support
- More tool integrations
- Additional vector databases
- Enhanced monitoring/metrics

**v1.2** (Planned)
- gRPC API support
- WebSocket for real-time updates
- Plugin system
- Advanced workflow features

### 📄 License

MIT License - See [LICENSE](LICENSE) for details.

### 🔗 Links

- **GitHub:** https://github.com/rexleimo/agno-go
- **Documentation:** https://docs.agno.com
- **Issues:** https://github.com/rexleimo/agno-go/issues
- **Discussions:** https://github.com/rexleimo/agno-go/discussions

---

**Full Changelog:** https://github.com/rexleimo/agno-go/commits/v1.0.0
