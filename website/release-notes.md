---
title: Release Notes
description: Version history and release notes for Agno-Go
outline: deep
---

# Release Notes

## Version 1.2.6 (2025-10-31)

### ✨ Highlights
- Session parity: session reuse endpoint, `GET/POST /sessions/{id}/summary` with sync/async modes, history filters (`num_messages`, `stream_events`), and run metadata for cache hits and cancellation reasons.
- Response caching for agents and teams with in-memory LRU store plus configurable summary manager defaults.
- Media attachments pipeline for agents, teams, and workflows (`WithMediaPayload`) with validation helpers.
- Storage adapters for MongoDB and SQLite alongside Postgres, sharing the same JSON contracts.
- Toolkit expansions: Tavily Reader/Search, Claude Agent Skills, Gmail mark-as-read, Jira worklogs, ElevenLabs speech synthesis, and enhanced file tooling.
- Culture knowledge manager for curating organisational knowledge asynchronously.

### 🔧 Improvements
- Workflow engine persists cancellation reasons, supports resume-from checkpoints, and handles media-only payloads.
- AgentOS session APIs expose summary endpoints, reuse semantics, and history pagination with SSE toggles.
- MCP client forwards media attachments and caches capability manifests for faster tool execution.

### 🧪 Tests
- Added coverage for cache layer, summary manager, storage drivers, workflow resumptions, and new toolkits.

### ✅ Compatibility
- Additive changes; existing APIs remain backward compatible.


## Version 1.2.5 (2025-10-20)

### ✨ Highlights
- 新增 8 个模型提供商：Cohere、Together、OpenRouter、LM Studio、Vercel、Portkey、InternLM、SambaNova（支持同步与流式、函数调用）
- 新增评估系统（场景评测、指标汇总、多模型对比），媒体处理（图片元数据；音/视频占位），调试工具（请求/响应精简转储），云部署占位（NoopDeployer）
- 集成注册表（注册/列表/健康检查）与通用 utils（JSONPretty、Retry）

### 🔧 修复与改进
- Airflow 工具返回结构对齐 Airflow REST API v2：`total_entries`、`dag_run_id`、`logical_date`
- 站点首页图片缺失：将 hero 图片从 `/logo.svg` 更改为 `/logo.png`

### 🧪 测试
- 为新模型与模块补充了聚焦单测；维持现有基准测试

### ✅ 兼容性
- 全部为增量功能，无破坏性变更


## Version 1.2.1 (2025-10-15)

### 🧭 Documentation Reorganization

- Adopted clear separation between implemented docs and design/WIP:
  - `website/` → Implemented, user-facing documentation (VitePress)
  - `docs/` → Design drafts, migration plans, tasks, and developer/internal docs
- Added `docs/README.md` to document the policy and entry points
- Added `CONTRIBUTING.md` for contributor onboarding

### 🔗 Link Fixes

- Updated README, CLAUDE, CHANGELOG, and release notes to point to canonical pages under `website/advanced/*` and `website/guide/*`
- Removed outdated references to duplicated files under `docs/`

### 🌐 Website Updates

- API: Added Knowledge API details to AgentOS page (`/api/agentos`)
- Ensured Workflow History and Performance pages are the canonical references

### ✅ No Behavior Changes

- This release updates documentation only; runtime behavior is unchanged

### ✨ New in 1.2.1 (Implemented)

- SSE event filtering for streaming endpoints (A2A)
  - `POST /api/v1/agents/:id/run/stream?types=token,complete`
  - Emits only requested event types; standard SSE format; context cancel supported
- Content extraction middleware for AgentOS
  - JSON/Form → context injection of `content/metadata/user_id/session_id`
  - Request size guard via `MaxRequestSize`; skip paths supported
- Google Sheets toolkit (service account)
  - `read_range`, `write_range`, `append_rows`; JSON or file credentials
- Minimal knowledge ingestion endpoint
  - `POST /api/v1/knowledge/content` supports `text/plain` and `application/json`

Enterprise validation steps: see [`docs/ENTERPRISE_MIGRATION_PLAN.md`](https://github.com/rexleimo/agno-Go/blob/main/docs/ENTERPRISE_MIGRATION_PLAN.md).

## Version 1.1.0 (2025-10-08)

### 🎉 Highlights

This release brings powerful new features for production-ready multi-agent systems:

- **A2A Interface** - Standardized agent-to-agent communication protocol
- **Session State Management** - Persistent state across workflow steps
- **Multi-Tenant Support** - Serve multiple users with a single agent instance
- **Model Timeout Configuration** - Fine-grained timeout control for LLM calls

---

### ✨ New Features

#### A2A (Agent-to-Agent) Interface

A standardized communication protocol for agent-to-agent interactions based on JSON-RPC 2.0.

**Key Features:**
- REST API endpoints (`/a2a/message/send`, `/a2a/message/stream`)
- Multi-media support (text, images, files, JSON data)
- Server-Sent Events (SSE) for streaming
- Compatible with Python Agno A2A implementation

**Quick Example:**
```go
import "github.com/rexleimo/agno-go/pkg/agentos/a2a"

// Create A2A interface
a2a := a2a.New(a2a.Config{
    Agents: []a2a.Entity{myAgent},
    Prefix: "/a2a",
})

// Register routes (Gin)
router := gin.Default()
a2a.RegisterRoutes(router)
```

📚 **Learn More:** [A2A Interface Documentation](/api/a2a)

---

#### Workflow Session State Management

Thread-safe session management for maintaining state across workflow steps.

**Key Features:**
- Cross-step persistent state storage
- Thread-safe with `sync.RWMutex`
- Deep copy for parallel branch isolation
- Smart merge strategy to prevent data loss
- Fixes Python Agno v2.1.2 race condition

**Quick Example:**
```go
// Create context with session info
execCtx := workflow.NewExecutionContextWithSession(
    "input",
    "session-123",  // Session ID
    "user-a",       // User ID
)

// Access session state
execCtx.SetSessionState("key", "value")
value, _ := execCtx.GetSessionState("key")
```

📚 **Learn More:** [Session State Documentation](/guide/session-state)

---

#### Multi-Tenant Support

Serve multiple users with a single Agent instance while ensuring complete data isolation.

**Key Features:**
- User-isolated conversation history
- Optional `userID` parameter in Memory interface
- Backward compatible with existing code
- Thread-safe concurrent operations
- `ClearAll()` method for cleanup

**Quick Example:**
```go
// Create multi-tenant agent
agent, _ := agent.New(&agent.Config{
    Name:   "customer-service",
    Model:  model,
    Memory: memory.NewInMemory(100),
})

// User A's conversation
agent.UserID = "user-a"
output, _ := agent.Run(ctx, "My name is Alice")

// User B's conversation
agent.UserID = "user-b"
output, _ := agent.Run(ctx, "My name is Bob")
```

📚 **Learn More:** [Multi-Tenant Documentation](/advanced/multi-tenant)

---

#### Model Timeout Configuration

Configure request timeout for LLM calls with fine-grained control.

**Key Features:**
- Default: 60 seconds
- Range: 1 second to 10 minutes
- Supported models: OpenAI, Anthropic Claude
- Context-aware timeout handling

**Quick Example:**
```go
// OpenAI with custom timeout
model, _ := openai.New("gpt-4", openai.Config{
    APIKey:  apiKey,
    Timeout: 30 * time.Second,
})

// Claude with custom timeout
claude, _ := anthropic.New("claude-3-opus", anthropic.Config{
    APIKey:  apiKey,
    Timeout: 45 * time.Second,
})
```

📚 **Learn More:** [Model Configuration](/guide/models#timeout-configuration)

---

### 🐛 Bug Fixes

- **Workflow Race Condition** - Fixed parallel step execution data race
  - Python Agno v2.1.2 had shared `session_state` dict causing overwrites
  - Go implementation uses independent SessionState clones per branch
  - Smart merge strategy prevents data loss in concurrent execution

---

### 📚 Documentation

All new features include comprehensive bilingual documentation (English/中文):

- [A2A Interface Guide](/api/a2a) - Complete protocol specification
- [Session State Guide](/guide/session-state) - Workflow state management
- [Multi-Tenant Guide](/advanced/multi-tenant) - Data isolation patterns
- [Model Configuration](/guide/models#timeout-configuration) - Timeout settings

---

### 🧪 Testing

**New Test Suites:**
- `session_state_test.go` - 543 lines of session state tests
- `memory_test.go` - Multi-tenant memory tests (4 new test cases)
- `agent_test.go` - Multi-tenant agent test
- `openai_test.go` - Timeout configuration test
- `anthropic_test.go` - Timeout configuration test

**Test Results:**
- ✅ All tests passing with `-race` detector
- ✅ Workflow coverage: 79.4%
- ✅ Memory coverage: 93.1%
- ✅ Agent coverage: 74.7%

---

### 📊 Performance

**No Performance Regression** - All benchmarks remain consistent:
- Agent instantiation: ~180ns/op (16x faster than Python)
- Memory footprint: ~1.2KB per agent
- Thread-safe concurrent operations

---

### ⚠️ Breaking Changes

**None.** This release is fully backward compatible with v1.0.x.

---

### 🔄 Migration Guide

**No migration needed** - All new features are additive and backward compatible.

**Optional Enhancements:**

1. **Enable Multi-Tenant Support:**
   ```go
   // Add UserID to agent configuration
   agent := agent.New(agent.Config{
       UserID: "user-123",  // NEW
       Memory: memory.NewInMemory(100),
   })
   ```

2. **Use Session State in Workflows:**
   ```go
   // Create context with session
   ctx := workflow.NewExecutionContextWithSession(
       "input",
       "session-id",
       "user-id",
   )
   ```

3. **Configure Model Timeout:**
   ```go
   // Add Timeout to model config
   model, _ := openai.New("gpt-4", openai.Config{
       APIKey:  apiKey,
       Timeout: 30 * time.Second,  // NEW
   })
   ```

---

### 📦 Installation

```bash
go get github.com/rexleimo/agno-go@v1.1.0
```

---

### 🔗 Links

- **GitHub Release:** [v1.1.0](https://github.com/rexleimo/agno-go/releases/tag/v1.1.0)
- **Full Changelog:** [CHANGELOG.md](https://github.com/rexleimo/agno-go/blob/main/CHANGELOG.md)
- **Documentation:** [https://agno-go.dev](https://agno-go.dev)

---

## Version 1.0.3 (2025-10-06)

### 🧪 Improved

#### Testing & Quality
- **Enhanced JSON Serialization Tests** - Achieved 100% test coverage for utils/serialize package
- **Performance Benchmarks** - Aligned with Python Agno performance testing patterns
- **Comprehensive Documentation** - Added bilingual package documentation

#### Performance
- **ToJSON**: ~600ns/op, 760B/op, 15 allocs/op
- **ConvertValue**: ~180ns/op, 392B/op, 5 allocs/op
- **Agent Creation**: ~180ns/op (16x faster than Python)

---

## Version 1.0.2 (2025-10-05)

### ✨ Added

#### GLM (智谱AI) Provider
- Full integration with Zhipu AI's GLM models
- Support for GLM-4, GLM-4V (vision), GLM-3-Turbo
- Custom JWT authentication (HMAC-SHA256)
- Synchronous and streaming API calls
- Tool/function calling support
- Test coverage: 57.2%

**Quick Example:**
```go
model, _ := glm.New("glm-4", glm.Config{
    APIKey:      os.Getenv("ZHIPUAI_API_KEY"),
    Temperature: 0.7,
})
```

---

## Version 1.0.0 (2025-10-02)

### 🎉 Initial Release

Agno-Go v1.0 is a high-performance Go implementation of the Agno multi-agent framework.

#### Core Features
- **Agent** - Single autonomous agent with tool support (74.7% coverage)
- **Team** - Multi-agent collaboration with 4 modes (92.3% coverage)
- **Workflow** - Step-based orchestration with 5 primitives (80.4% coverage)

#### LLM Providers
- OpenAI (GPT-4, GPT-3.5, GPT-4 Turbo)
- Anthropic (Claude 3.5 Sonnet, Claude 3 Opus/Sonnet/Haiku)
- Ollama (Local models)

#### Tools & Storage
- Calculator, HTTP, File tools
- In-memory conversation storage (93.1% coverage)
- Session management
- ChromaDB vector database

#### Performance
- Agent creation: ~180ns/op (16x faster than Python)
- Memory footprint: ~1.2KB per agent
- Test coverage: 80.8% average

---

## Previous Versions

See [CHANGELOG.md](https://github.com/rexleimo/agno-go/blob/main/CHANGELOG.md) for complete version history.

---

**Last Updated:** 2025-10-08
