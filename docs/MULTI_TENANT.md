# Multi-Tenant Support | 多租户支持

## 概述 | Overview

**Multi-Tenant Support** 为 Agno-Go 提供了多租户数据隔离能力,允许同一个 Agent 实例为不同用户提供服务,确保每个用户的对话历史、会话状态完全隔离。

**Multi-Tenant Support** provides multi-tenant data isolation capabilities for Agno-Go, allowing the same Agent instance to serve different users while ensuring complete isolation of conversation history and session state for each user.

---

## 核心概念 | Core Concepts

### 什么是多租户? | What is Multi-Tenancy?

多租户架构允许单个应用实例服务多个用户(租户),每个租户的数据相互隔离:

Multi-tenant architecture allows a single application instance to serve multiple users (tenants), with each tenant's data completely isolated:

```
                 ┌─────────────────┐
                 │  Agent Instance │
                 └────────┬────────┘
                          │
         ┌────────────────┼────────────────┐
         ▼                ▼                ▼
   ┌──────────┐     ┌──────────┐     ┌──────────┐
   │  User A  │     │  User B  │     │  User C  │
   │ Messages │     │ Messages │     │ Messages │
   └──────────┘     └──────────┘     └──────────┘
```

### 为什么需要多租户? | Why Multi-Tenancy?

**Without Multi-Tenant | 不使用多租户**:
```go
// ❌ 每个用户需要单独的 Agent 实例 | Each user needs a separate Agent instance
userAgents := make(map[string]*agent.Agent)

agent1, _ := agent.New(config)  // User 1
agent2, _ := agent.New(config)  // User 2
agent3, _ := agent.New(config)  // User 3
// ... 1000+ users = 1000+ Agent instances
```

**问题 | Problems**:
- 高内存占用: 1000 用户 = 1000 个 Agent 实例
- 难以管理: 需要手动管理 Agent 生命周期
- 资源浪费: 每个 Agent 都有独立的配置副本

**With Multi-Tenant | 使用多租户**:
```go
// ✅ 单个 Agent 实例服务所有用户 | Single Agent instance serves all users
sharedAgent, _ := agent.New(config)

// 不同用户使用不同的 userID | Different users use different userID
output1, _ := sharedAgent.Run(ctx, "user-1 input", "user-1")
output2, _ := sharedAgent.Run(ctx, "user-2 input", "user-2")
output3, _ := sharedAgent.Run(ctx, "user-3 input", "user-3")
```

**优势 | Advantages**:
- ✅ 低内存占用: 单个 Agent 实例
- ✅ 易于管理: 统一配置和更新
- ✅ 高效资源利用: 共享模型和工具

---

## 快速开始 | Quick Start

### 1. 创建多租户 Agent | Create Multi-Tenant Agent

```go
package main

import (
    "context"
    "fmt"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/memory"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
)

func main() {
    // 创建模型 | Create model
    model, _ := openai.New("gpt-4", openai.Config{
        APIKey: "your-api-key",
    })

    // 创建多租户 Memory | Create multi-tenant Memory
    mem := memory.NewInMemory(100)  // 自动支持多租户 | Automatically supports multi-tenancy

    // 创建 Agent | Create Agent
    myAgent, _ := agent.New(&agent.Config{
        Name:         "customer-service",
        Model:        model,
        Memory:       mem,
        Instructions: "You are a helpful customer service agent.",
    })

    // 不同用户的对话 | Conversations for different users
    ctx := context.Background()

    // User A 的对话 | User A's conversation
    myAgent.UserID = "user-a"
    output1, _ := myAgent.Run(ctx, "My name is Alice")
    fmt.Printf("User A: %s\n", output1.Content)

    output2, _ := myAgent.Run(ctx, "What's my name?")  // "Your name is Alice"
    fmt.Printf("User A: %s\n", output2.Content)

    // User B 的对话 | User B's conversation
    myAgent.UserID = "user-b"
    output3, _ := myAgent.Run(ctx, "My name is Bob")
    fmt.Printf("User B: %s\n", output3.Content)

    output4, _ := myAgent.Run(ctx, "What's my name?")  // "Your name is Bob"
    fmt.Printf("User B: %s\n", output4.Content)

    // User A 再次对话 | User A talks again
    myAgent.UserID = "user-a"
    output5, _ := myAgent.Run(ctx, "What's my name?")  // "Your name is Alice"
    fmt.Printf("User A: %s\n", output5.Content)
}
```

### 2. Web API 示例 | Web API Example

```go
package main

import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/rexleimo/agno-go/pkg/agno/agent"
)

var sharedAgent *agent.Agent

func main() {
    // 初始化 Agent | Initialize Agent
    sharedAgent, _ = agent.New(&agent.Config{
        Name:   "api-agent",
        Model:  model,
        Memory: memory.NewInMemory(100),
    })

    // 设置路由 | Setup routes
    router := gin.Default()
    router.POST("/chat", handleChat)
    router.Run(":8080")
}

type ChatRequest struct {
    UserID  string `json:"user_id"`
    Message string `json:"message"`
}

type ChatResponse struct {
    UserID  string `json:"user_id"`
    Reply   string `json:"reply"`
}

func handleChat(c *gin.Context) {
    var req ChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 设置当前用户 ID | Set current user ID
    sharedAgent.UserID = req.UserID

    // 执行对话 | Run conversation
    output, err := sharedAgent.Run(context.Background(), req.Message)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, ChatResponse{
        UserID: req.UserID,
        Reply:  output.Content,
    })
}
```

**测试 | Test**:
```bash
# User A 的对话 | User A's conversation
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user-a", "message": "My name is Alice"}'

curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user-a", "message": "What is my name?"}'
# Response: {"user_id":"user-a","reply":"Your name is Alice"}

# User B 的对话 | User B's conversation
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user-b", "message": "My name is Bob"}'

curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user-b", "message": "What is my name?"}'
# Response: {"user_id":"user-b","reply":"Your name is Bob"}
```

---

## 内存管理 | Memory Management

### Memory 接口扩展 | Memory Interface Extensions

```go
// pkg/agno/memory/memory.go

type Memory interface {
    // 添加消息 (支持可选的 userID) | Add message (supports optional userID)
    Add(message *types.Message, userID ...string)

    // 获取消息历史 (支持可选的 userID) | Get message history (supports optional userID)
    GetMessages(userID ...string) []*types.Message

    // 清空特定用户的消息 | Clear messages for specific user
    Clear(userID ...string)

    // 清空所有用户的消息 | Clear messages for all users
    ClearAll()

    // 获取特定用户的消息数量 | Get message count for specific user
    Size(userID ...string) int
}
```

### InMemory 实现 | InMemory Implementation

```go
type InMemory struct {
    userMessages map[string][]*types.Message  // 用户 ID → 消息列表 | User ID → Message list
    maxSize      int
    mu           sync.RWMutex
}

// 默认用户 ID | Default user ID
const defaultUserID = "default"

// 获取用户 ID (支持向后兼容) | Get user ID (backward compatible)
func getUserID(userID ...string) string {
    if len(userID) > 0 && userID[0] != "" {
        return userID[0]
    }
    return defaultUserID
}
```

### 使用示例 | Usage Examples

#### 基础用法 | Basic Usage

```go
mem := memory.NewInMemory(100)

// User A 的消息 | User A's messages
mem.Add(types.NewUserMessage("Hello from Alice"), "user-a")
mem.Add(types.NewAssistantMessage("Hi Alice!"), "user-a")

// User B 的消息 | User B's messages
mem.Add(types.NewUserMessage("Hello from Bob"), "user-b")
mem.Add(types.NewAssistantMessage("Hi Bob!"), "user-b")

// 获取各用户的消息 | Get messages for each user
messagesA := mem.GetMessages("user-a")  // 2 messages
messagesB := mem.GetMessages("user-b")  // 2 messages

fmt.Printf("User A has %d messages\n", len(messagesA))  // 2
fmt.Printf("User B has %d messages\n", len(messagesB))  // 2
```

#### 向后兼容 | Backward Compatibility

```go
mem := memory.NewInMemory(100)

// 不指定 userID (使用默认 "default") | No userID specified (uses default "default")
mem.Add(types.NewUserMessage("Hello"))
messages := mem.GetMessages()

// 等价于 | Equivalent to:
mem.Add(types.NewUserMessage("Hello"), "default")
messages := mem.GetMessages("default")
```

#### 清空操作 | Clear Operations

```go
mem := memory.NewInMemory(100)

// 添加不同用户的消息 | Add messages for different users
mem.Add(types.NewUserMessage("User A msg"), "user-a")
mem.Add(types.NewUserMessage("User B msg"), "user-b")

// 清空特定用户 | Clear specific user
mem.Clear("user-a")
fmt.Printf("User A: %d messages\n", mem.Size("user-a"))  // 0
fmt.Printf("User B: %d messages\n", mem.Size("user-b"))  // 1

// 清空所有用户 | Clear all users
mem.ClearAll()
fmt.Printf("User A: %d messages\n", mem.Size("user-a"))  // 0
fmt.Printf("User B: %d messages\n", mem.Size("user-b"))  // 0
```

---

## Agent 集成 | Agent Integration

### Agent 配置 | Agent Configuration

```go
type Agent struct {
    ID           string
    Name         string
    Model        models.Model
    Toolkits     []toolkit.Toolkit
    Memory       memory.Memory
    Instructions string
    MaxLoops     int

    UserID string  // ⭐ 新增: 多租户用户 ID | NEW: Multi-tenant user ID
}

type Config struct {
    Name         string
    Model        models.Model
    Toolkits     []toolkit.Toolkit
    Memory       memory.Memory
    Instructions string
    MaxLoops     int

    UserID string  // ⭐ 新增: 多租户用户 ID | NEW: Multi-tenant user ID
}
```

### Run 方法调用 | Run Method Calls

```go
// pkg/agno/agent/agent.go

func (a *Agent) Run(ctx context.Context, input string) (*RunOutput, error) {
    // ...

    // 所有 Memory 调用都传递 UserID | All Memory calls pass UserID
    userMsg := types.NewUserMessage(input)
    a.Memory.Add(userMsg, a.UserID)  // ⭐ 传递 UserID | Pass UserID

    // ...

    messages := a.Memory.GetMessages(a.UserID)  // ⭐ 传递 UserID | Pass UserID

    // ...

    a.Memory.Add(types.NewAssistantMessage(content), a.UserID)  // ⭐ 传递 UserID | Pass UserID
}
```

### 使用方式 | Usage Patterns

#### 方式 1: 共享 Agent + 切换 UserID | Shared Agent + Switch UserID

```go
agent, _ := agent.New(&agent.Config{
    Name:   "shared-agent",
    Model:  model,
    Memory: memory.NewInMemory(100),
})

// 处理 User A 的请求 | Handle User A's request
agent.UserID = "user-a"
output, _ := agent.Run(ctx, "User A message")

// 处理 User B 的请求 | Handle User B's request
agent.UserID = "user-b"
output, _ := agent.Run(ctx, "User B message")
```

⚠️ **注意 | Note**: 此方式需要在并发环境下小心处理 UserID 切换 | This approach requires careful UserID switching in concurrent environments

#### 方式 2: 每个用户独立 Agent (推荐用于高并发) | Separate Agent per User (Recommended for High Concurrency)

```go
// 创建 Agent 工厂 | Create Agent factory
func createUserAgent(userID string) (*agent.Agent, error) {
    return agent.New(&agent.Config{
        Name:   "user-agent",
        Model:  sharedModel,  // 可以共享 Model | Can share Model
        Memory: memory.NewInMemory(100),
        UserID: userID,  // 设置固定的 UserID | Set fixed UserID
    })
}

// 使用 Agent 池 | Use Agent pool
userAgents := make(map[string]*agent.Agent)

// User A
if _, exists := userAgents["user-a"]; !exists {
    userAgents["user-a"], _ = createUserAgent("user-a")
}
output, _ := userAgents["user-a"].Run(ctx, "User A message")

// User B
if _, exists := userAgents["user-b"]; !exists {
    userAgents["user-b"], _ = createUserAgent("user-b")
}
output, _ := userAgents["user-b"].Run(ctx, "User B message")
```

---

## Workflow 集成 | Workflow Integration

### Session State + UserID | Session State + UserID

```go
// 创建带 UserID 的 ExecutionContext | Create ExecutionContext with UserID
execCtx := workflow.NewExecutionContextWithSession(
    "input",
    "session-123",
    "user-a",  // ⭐ UserID
)

// ExecutionContext 包含 UserID | ExecutionContext contains UserID
type ExecutionContext struct {
    Input        string
    Output       string
    Data         map[string]interface{}
    Metadata     map[string]interface{}
    SessionState *SessionState
    SessionID    string
    UserID       string  // ⭐ 多租户用户 ID | Multi-tenant user ID
}
```

### 完整示例 | Complete Example

```go
// 创建 Workflow | Create Workflow
wf := workflow.NewWorkflow("user-workflow")

// Step: 使用 Agent 处理用户输入 | Step: Process user input with Agent
step := workflow.NewStepWithAgent("process", myAgent)
wf.AddStep(step)

// User A 执行 | User A execution
execCtxA := workflow.NewExecutionContextWithSession("User A input", "session-a", "user-a")
resultA, _ := wf.Execute(context.Background(), execCtxA)

// User B 执行 | User B execution
execCtxB := workflow.NewExecutionContextWithSession("User B input", "session-b", "user-b")
resultB, _ := wf.Execute(context.Background(), execCtxB)

// User A 和 User B 的状态完全隔离 | User A and User B states are completely isolated
```

---

## 数据隔离保证 | Data Isolation Guarantees

### 1. Memory 隔离 | Memory Isolation

```go
// 测试: 多租户隔离 | Test: Multi-tenant isolation
mem := memory.NewInMemory(100)

// User A 添加 10 条消息 | User A adds 10 messages
for i := 0; i < 10; i++ {
    mem.Add(types.NewUserMessage(fmt.Sprintf("User A message %d", i)), "user-a")
}

// User B 添加 5 条消息 | User B adds 5 messages
for i := 0; i < 5; i++ {
    mem.Add(types.NewUserMessage(fmt.Sprintf("User B message %d", i)), "user-b")
}

// 验证隔离 | Verify isolation
assert.Equal(t, 10, mem.Size("user-a"))  // ✅
assert.Equal(t, 5, mem.Size("user-b"))   // ✅
assert.Equal(t, 0, mem.Size("user-c"))   // ✅ 不存在的用户 | Non-existent user

messagesA := mem.GetMessages("user-a")
messagesB := mem.GetMessages("user-b")

// User A 看不到 User B 的消息 | User A cannot see User B's messages
for _, msg := range messagesA {
    assert.NotContains(t, msg.Content, "User B")  // ✅
}
```

### 2. SessionState 隔离 | SessionState Isolation

```go
// 不同 session 完全独立 | Different sessions are completely independent
sessionA := workflow.NewExecutionContextWithSession("", "session-a", "user-a")
sessionB := workflow.NewExecutionContextWithSession("", "session-b", "user-b")

sessionA.SetSessionState("shared_key", "value_from_user_a")
sessionB.SetSessionState("shared_key", "value_from_user_b")

valueA, _ := sessionA.GetSessionState("shared_key")
valueB, _ := sessionB.GetSessionState("shared_key")

assert.Equal(t, "value_from_user_a", valueA)  // ✅
assert.Equal(t, "value_from_user_b", valueB)  // ✅
```

### 3. 并发安全 | Concurrency Safety

```go
// 测试: 1000 并发请求 | Test: 1000 concurrent requests
mem := memory.NewInMemory(100)
var wg sync.WaitGroup

// 10 个用户,每个用户 100 个并发请求 | 10 users, 100 concurrent requests each
for userID := 0; userID < 10; userID++ {
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(uid, msgID int) {
            defer wg.Done()
            userIDStr := fmt.Sprintf("user-%d", uid)
            msg := types.NewUserMessage(fmt.Sprintf("Message %d", msgID))
            mem.Add(msg, userIDStr)
        }(userID, i)
    }
}

wg.Wait()

// 验证每个用户都有正确数量的消息 | Verify each user has correct number of messages
for userID := 0; userID < 10; userID++ {
    userIDStr := fmt.Sprintf("user-%d", userID)
    assert.Equal(t, 100, mem.Size(userIDStr))  // ✅
}
```

---

## 性能优化 | Performance Optimization

### 1. Memory 容量管理 | Memory Capacity Management

```go
// 设置合理的 maxSize 避免内存溢出 | Set reasonable maxSize to avoid memory overflow
mem := memory.NewInMemory(100)  // 每个用户最多 100 条消息 | Max 100 messages per user

// 如果用户数量很多,考虑使用外部存储 | If many users, consider external storage
// 例如: PostgreSQL, Redis, 等 | e.g., PostgreSQL, Redis, etc.
```

### 2. Agent 池管理 | Agent Pool Management

```go
// 使用 LRU 缓存管理 Agent 实例 | Use LRU cache to manage Agent instances
type AgentPool struct {
    agents   map[string]*agent.Agent
    mu       sync.RWMutex
    maxSize  int
}

func (p *AgentPool) GetOrCreate(userID string) (*agent.Agent, error) {
    p.mu.RLock()
    ag, exists := p.agents[userID]
    p.mu.RUnlock()

    if exists {
        return ag, nil
    }

    p.mu.Lock()
    defer p.mu.Unlock()

    // Double-check
    if ag, exists := p.agents[userID]; exists {
        return ag, nil
    }

    // 创建新 Agent | Create new Agent
    ag, err := createUserAgent(userID)
    if err != nil {
        return nil, err
    }

    // 检查容量,必要时淘汰 | Check capacity, evict if necessary
    if len(p.agents) >= p.maxSize {
        // LRU eviction logic
        // ...
    }

    p.agents[userID] = ag
    return ag, nil
}
```

### 3. 内存监控 | Memory Monitoring

```go
// 定期清理不活跃用户的数据 | Periodically clean up inactive user data
func cleanupInactiveUsers(mem memory.Memory, threshold time.Duration) {
    // 伪代码 | Pseudocode
    for _, userID := range getInactiveUsers(threshold) {
        mem.Clear(userID)
    }
}

// 在后台运行清理任务 | Run cleanup task in background
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()

    for range ticker.C {
        cleanupInactiveUsers(sharedMemory, 24*time.Hour)
    }
}()
```

---

## 最佳实践 | Best Practices

### 1. UserID 命名规范 | UserID Naming Convention

```go
// ✅ 推荐: 使用统一的命名规范 | Recommended: Use consistent naming convention
"user-{uuid}"           // user-123e4567-e89b-12d3-a456-426614174000
"org-{org_id}-user-{id}" // org-acme-user-001
"tenant-{id}"           // tenant-12345

// ❌ 避免: 使用不稳定的标识 | Avoid: Use unstable identifiers
"{ip_address}"          // IP 可能变化 | IP may change
"{session_id}"          // Session 会过期 | Session expires
```

### 2. 错误处理 | Error Handling

```go
// 验证 UserID 合法性 | Validate UserID
func validateUserID(userID string) error {
    if userID == "" {
        return fmt.Errorf("userID cannot be empty")
    }
    if len(userID) > 255 {
        return fmt.Errorf("userID too long (max 255 chars)")
    }
    // 可以添加更多验证规则 | Can add more validation rules
    return nil
}

// 在 API 层验证 | Validate at API layer
func handleChat(c *gin.Context) {
    var req ChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    if err := validateUserID(req.UserID); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // ...
}
```

### 3. 日志和监控 | Logging and Monitoring

```go
// 记录每个请求的 UserID | Log UserID for each request
logger.Info("Processing request",
    "user_id", userID,
    "input_length", len(input),
    "timestamp", time.Now(),
)

// 监控指标 | Monitoring metrics
metrics.RecordUserRequest(userID)
metrics.RecordMessageCount(userID, mem.Size(userID))
```

### 4. 安全考虑 | Security Considerations

```go
// 使用加密的 UserID | Use encrypted UserID
func encryptUserID(plainUserID string) string {
    // 使用加密算法 | Use encryption algorithm
    return encryptedID
}

// 限制访问 | Access control
func checkUserPermission(userID string, action string) bool {
    // 实现权限检查逻辑 | Implement permission check logic
    return hasPermission
}
```

---

## 测试 | Testing

完整的测试覆盖了以下场景:

Complete test coverage includes the following scenarios:

- ✅ 多用户数据隔离 | Multi-user data isolation
- ✅ 并发安全 (1000 goroutines) | Concurrency safety (1000 goroutines)
- ✅ Agent 集成测试 | Agent integration tests
- ✅ Memory 容量管理 | Memory capacity management
- ✅ 清空操作正确性 | Clear operation correctness

**测试覆盖率 | Test Coverage**: 93.1% (Memory module)

运行测试 | Run tests:
```bash
cd pkg/agno/memory
go test -v -run TestInMemory

cd pkg/agno/agent
go test -v -run TestAgent_MultiTenant
```

---

## 故障排查 | Troubleshooting

### 常见问题 | Common Issues

#### 1. 用户数据混乱

**现象 | Symptom**: User A 看到了 User B 的消息

**原因 | Cause**: UserID 未正确传递

**解决 | Solution**:
```go
// ❌ 错误 | Wrong
agent.Run(ctx, input)  // UserID 未设置 | UserID not set

// ✅ 正确 | Correct
agent.UserID = "user-a"
agent.Run(ctx, input)
```

#### 2. 内存占用过高

**现象 | Symptom**: 内存持续增长

**原因 | Cause**: 未清理不活跃用户的数据

**解决 | Solution**:
```go
// 定期清理 | Periodic cleanup
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    for range ticker.C {
        cleanupInactiveUsers(mem, 24*time.Hour)
    }
}()
```

#### 3. 并发竞争

**现象 | Symptom**: 数据偶尔丢失或重复

**原因 | Cause**: 共享 Agent 的 UserID 字段在并发环境下被多个 goroutine 修改

**解决 | Solution**:
```go
// ❌ 错误: 并发修改共享 Agent | Wrong: Concurrent modification of shared Agent
var sharedAgent *agent.Agent
go func() { sharedAgent.UserID = "user-a"; sharedAgent.Run(ctx, input) }()
go func() { sharedAgent.UserID = "user-b"; sharedAgent.Run(ctx, input) }()

// ✅ 正确: 每个用户独立 Agent | Correct: Separate Agent per user
agentA := createUserAgent("user-a")
agentB := createUserAgent("user-b")
go func() { agentA.Run(ctx, input) }()
go func() { agentB.Run(ctx, input) }()
```

---

## 与其他功能的集成 | Integration with Other Features

### A2A Interface + Multi-Tenant

```go
// A2A 请求包含 contextID,可以作为 userID | A2A request contains contextID, can be used as userID
type Message struct {
    MessageID string `json:"messageId"`
    Role      string `json:"role"`
    AgentID   string `json:"agentId"`
    ContextID string `json:"contextId"`  // ⭐ 可以作为 userID | Can be used as userID
    Parts     []Part `json:"parts"`
}

// 映射时设置 UserID | Set UserID during mapping
func MapA2ARequestToRunInput(req *JSONRPC2Request) (*RunInput, error) {
    // ...
    agent.UserID = req.Params.Message.ContextID  // ⭐ 使用 contextID 作为 userID
    // ...
}
```

### Session State + Multi-Tenant

```go
// ExecutionContext 同时支持 SessionID 和 UserID | ExecutionContext supports both SessionID and UserID
execCtx := workflow.NewExecutionContextWithSession(
    "input",
    "session-123",  // SessionID: 单次会话的标识 | SessionID: Identifier for single session
    "user-a",       // UserID: 用户的标识 | UserID: User identifier
)

// SessionID: 用于会话状态管理 | SessionID: For session state management
// UserID: 用于多租户数据隔离 | UserID: For multi-tenant data isolation
```

---

## 相关文档 | Related Documentation

- [A2A Interface](A2A_INTERFACE.md) - Agent间通信接口
- [Session State Management](SESSION_STATE.md) - 会话状态管理
- [Memory Guide](MEMORY_GUIDE.md) - Memory 使用指南

---

## 版本历史 | Version History

- **v1.1.0** (2025-01-XX): Initial Multi-Tenant implementation
  - UserID support in Memory interface
  - Backward-compatible variadic parameters
  - Complete data isolation between users
  - Integration with Agent and Workflow

---

**更新时间 | Last Updated**: 2025-01-XX
