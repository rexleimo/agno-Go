# Session State Management | 会话状态管理

## 概述 | Overview

**Session State Management** 提供了 Workflow 执行过程中的状态管理能力,支持跨步骤的数据共享、并发安全的状态访问,以及并行分支的智能合并。

**Session State Management** provides state management capabilities during workflow execution, supporting cross-step data sharing, concurrency-safe state access, and intelligent merging of parallel branches.

---

## 核心概念 | Core Concepts

### 为什么需要 Session State? | Why Session State?

在复杂的 Workflow 中,步骤之间需要共享数据:

In complex workflows, steps need to share data:

```
Step1: 获取用户信息 | Get user info
  ↓
  保存到 SessionState: {"user_id": "123", "name": "Alice"}
  ↓
Step2: 根据用户信息查询订单 | Query orders based on user info
  ↓
  从 SessionState 读取: user_id = "123"
  保存到 SessionState: {"orders": [...]}
  ↓
Step3: 生成报告 | Generate report
  ↓
  从 SessionState 读取: user_id, name, orders
```

### 核心特性 | Core Features

1. **线程安全 | Thread-Safe**: 使用 `sync.RWMutex` 保护并发访问
2. **深拷贝 | Deep Copy**: 并行分支获得独立的状态副本
3. **智能合并 | Smart Merge**: 并行执行后自动合并状态变更
4. **类型灵活 | Flexible Types**: 支持任意 `interface{}` 类型数据

---

## 快速开始 | Quick Start

### 1. 基础用法 | Basic Usage

```go
package main

import (
    "context"
    "fmt"

    "github.com/rexleimo/agno-go/pkg/agno/workflow"
)

func main() {
    // 创建带会话状态的执行上下文 | Create execution context with session state
    execCtx := workflow.NewExecutionContextWithSession(
        "initial input",
        "session-123",  // sessionID
        "user-456",     // userID
    )

    // 设置会话状态 | Set session state
    execCtx.SetSessionState("user_name", "Alice")
    execCtx.SetSessionState("user_age", 30)
    execCtx.SetSessionState("preferences", map[string]string{
        "language": "zh-CN",
        "theme":    "dark",
    })

    // 读取会话状态 | Get session state
    if name, ok := execCtx.GetSessionState("user_name"); ok {
        fmt.Printf("User Name: %s\n", name)
    }

    if age, ok := execCtx.GetSessionState("user_age"); ok {
        fmt.Printf("User Age: %d\n", age)
    }
}
```

### 2. 在 Workflow 中使用 | Using in Workflow

```go
// 创建 Workflow | Create workflow
wf := workflow.NewWorkflow("user-workflow")

// Step 1: 获取用户信息 | Get user info
step1 := workflow.NewStep("get-user", func(ctx context.Context, execCtx *workflow.ExecutionContext) (*workflow.ExecutionContext, error) {
    // 模拟获取用户信息 | Simulate fetching user info
    userInfo := map[string]interface{}{
        "id":    "user-123",
        "name":  "Alice",
        "email": "alice@example.com",
    }

    // 保存到 SessionState | Save to SessionState
    execCtx.SetSessionState("user_info", userInfo)

    return execCtx, nil
})

// Step 2: 获取用户订单 | Get user orders
step2 := workflow.NewStep("get-orders", func(ctx context.Context, execCtx *workflow.ExecutionContext) (*workflow.ExecutionContext, error) {
    // 从 SessionState 读取用户信息 | Read user info from SessionState
    userInfoRaw, ok := execCtx.GetSessionState("user_info")
    if !ok {
        return execCtx, fmt.Errorf("user_info not found in session state")
    }

    userInfo := userInfoRaw.(map[string]interface{})
    userID := userInfo["id"].(string)

    // 模拟获取订单 | Simulate fetching orders
    orders := []string{"order-1", "order-2", "order-3"}
    execCtx.SetSessionState("orders", orders)

    fmt.Printf("Got %d orders for user %s\n", len(orders), userID)

    return execCtx, nil
})

// 链接步骤 | Chain steps
step1.Then(step2)
wf.AddStep(step1)

// 执行 Workflow | Execute workflow
execCtx := workflow.NewExecutionContextWithSession("", "session-123", "user-456")
result, err := wf.Execute(context.Background(), execCtx)
if err != nil {
    panic(err)
}

// 检查最终状态 | Check final state
orders, _ := result.GetSessionState("orders")
fmt.Printf("Final orders: %v\n", orders)
```

---

## 并行执行与状态合并 | Parallel Execution and State Merging

### 问题场景 | Problem Scenario

在并行执行时,多个分支可能同时修改 SessionState:

During parallel execution, multiple branches may modify SessionState simultaneously:

```
              ┌─→ Branch A: Set("key1", "value_A")
Parallel Step ├─→ Branch B: Set("key2", "value_B")
              └─→ Branch C: Set("key1", "value_C")  // ⚠️ 冲突! | Conflict!
```

### 解决方案 | Solution

Agno-Go 使用 **深拷贝 + 最终写入优先** 策略:

Agno-Go uses **deep copy + last-write-wins** strategy:

1. 每个并行分支获得独立的 SessionState 副本
2. 分支独立执行,互不干扰
3. 执行完成后,按顺序合并所有变更
4. 如果有冲突,后执行的分支覆盖先执行的分支

Each parallel branch gets an independent SessionState copy:

```go
// pkg/agno/workflow/parallel.go

func (p *Parallel) Execute(ctx context.Context, execCtx *ExecutionContext) (*ExecutionContext, error) {
    // 1. 为每个分支创建独立的 SessionState 副本 | Create independent SessionState copy for each branch
    sessionStateCopies := make([]*SessionState, len(p.Nodes))
    for i := range p.Nodes {
        if execCtx.SessionState != nil {
            sessionStateCopies[i] = execCtx.SessionState.Clone()  // 深拷贝 | Deep copy
        } else {
            sessionStateCopies[i] = NewSessionState()
        }
    }

    // 2. 并行执行各分支 | Execute branches in parallel
    // ... (goroutines execution)

    // 3. 合并所有分支的状态变更 | Merge all branch state changes
    execCtx.SessionState = MergeParallelSessionStates(
        originalSessionState,
        modifiedSessionStates,
    )

    return execCtx, nil
}
```

### 合并策略 | Merge Strategy

```go
// pkg/agno/workflow/session_state.go

func MergeParallelSessionStates(original *SessionState, modified []*SessionState) *SessionState {
    merged := NewSessionState()

    // 1. 复制原始状态 | Copy original state
    if original != nil {
        for k, v := range original.data {
            merged.data[k] = v
        }
    }

    // 2. 按顺序合并各分支的变更 | Merge changes from each branch in order
    for _, modState := range modified {
        if modState == nil {
            continue
        }
        for k, v := range modState.data {
            merged.data[k] = v  // 最终写入优先 | Last-write-wins
        }
    }

    return merged
}
```

### 示例 | Example

```go
// 并行执行 3 个分支 | Parallel execution of 3 branches
parallel := workflow.NewParallel()

// Branch A: 设置 counter = 1 | Branch A: Set counter = 1
branchA := workflow.NewStep("branch-a", func(ctx context.Context, execCtx *workflow.ExecutionContext) (*workflow.ExecutionContext, error) {
    execCtx.SetSessionState("counter", 1)
    execCtx.SetSessionState("branch_a_result", "done")
    return execCtx, nil
})

// Branch B: 设置 counter = 2 | Branch B: Set counter = 2
branchB := workflow.NewStep("branch-b", func(ctx context.Context, execCtx *workflow.ExecutionContext) (*workflow.ExecutionContext, error) {
    execCtx.SetSessionState("counter", 2)
    execCtx.SetSessionState("branch_b_result", "done")
    return execCtx, nil
})

// Branch C: 设置 counter = 3 | Branch C: Set counter = 3
branchC := workflow.NewStep("branch-c", func(ctx context.Context, execCtx *workflow.ExecutionContext) (*workflow.ExecutionContext, error) {
    execCtx.SetSessionState("counter", 3)
    execCtx.SetSessionState("branch_c_result", "done")
    return execCtx, nil
})

parallel.AddNode(branchA)
parallel.AddNode(branchB)
parallel.AddNode(branchC)

// 执行并行步骤 | Execute parallel step
execCtx := workflow.NewExecutionContextWithSession("", "session-123", "user-456")
result, _ := parallel.Execute(context.Background(), execCtx)

// 检查合并结果 | Check merged result
counter, _ := result.GetSessionState("counter")
fmt.Printf("Counter: %v\n", counter)  // 输出可能是 1, 2, 或 3 (取决于执行顺序) | Output may be 1, 2, or 3 (depends on execution order)

branchAResult, _ := result.GetSessionState("branch_a_result")
branchBResult, _ := result.GetSessionState("branch_b_result")
branchCResult, _ := result.GetSessionState("branch_c_result")
fmt.Printf("All branches completed: %v, %v, %v\n", branchAResult, branchBResult, branchCResult)
// 输出 | Output: All branches completed: done, done, done
```

---

## API 参考 | API Reference

### SessionState 类型 | SessionState Type

```go
type SessionState struct {
    mu   sync.RWMutex
    data map[string]interface{}
}
```

#### 方法 | Methods

##### NewSessionState()
```go
func NewSessionState() *SessionState
```
创建新的 SessionState 实例 | Create a new SessionState instance

##### Set(key string, value interface{})
```go
func (ss *SessionState) Set(key string, value interface{})
```
设置键值对 (线程安全) | Set key-value pair (thread-safe)

##### Get(key string) (interface{}, bool)
```go
func (ss *SessionState) Get(key string) (interface{}, bool)
```
获取键对应的值 (线程安全) | Get value for key (thread-safe)

**返回 | Returns**:
- `value`: 对应的值 | The value
- `exists`: 键是否存在 | Whether the key exists

##### Clone() *SessionState
```go
func (ss *SessionState) Clone() *SessionState
```
深拷贝 SessionState (使用 JSON 序列化) | Deep copy SessionState (using JSON serialization)

**注意 | Note**: 不可序列化的类型会回退到浅拷贝 | Non-serializable types fall back to shallow copy

##### GetAll() map[string]interface{}
```go
func (ss *SessionState) GetAll() map[string]interface{}
```
获取所有键值对的副本 (线程安全) | Get a copy of all key-value pairs (thread-safe)

##### MergeFrom(other *SessionState)
```go
func (ss *SessionState) MergeFrom(other *SessionState)
```
从另一个 SessionState 合并数据 | Merge data from another SessionState

##### Clear()
```go
func (ss *SessionState) Clear()
```
清空所有数据 | Clear all data

---

### ExecutionContext 扩展 | ExecutionContext Extensions

```go
type ExecutionContext struct {
    // 现有字段 | Existing fields
    Input    string
    Output   string
    Data     map[string]interface{}
    Metadata map[string]interface{}

    // 新增字段 | New fields
    SessionState *SessionState `json:"session_state,omitempty"`
    SessionID    string        `json:"session_id,omitempty"`
    UserID       string        `json:"user_id,omitempty"`
}
```

#### 新增方法 | New Methods

##### NewExecutionContextWithSession()
```go
func NewExecutionContextWithSession(input, sessionID, userID string) *ExecutionContext
```
创建带会话状态的执行上下文 | Create execution context with session state

**参数 | Parameters**:
- `input`: 初始输入 | Initial input
- `sessionID`: 会话 ID | Session ID
- `userID`: 用户 ID | User ID

##### SetSessionState(key string, value interface{})
```go
func (ec *ExecutionContext) SetSessionState(key string, value interface{})
```
设置会话状态 (便捷方法) | Set session state (convenience method)

##### GetSessionState(key string) (interface{}, bool)
```go
func (ec *ExecutionContext) GetSessionState(key string) (interface{}, bool)
```
获取会话状态 (便捷方法) | Get session state (convenience method)

---

## 高级用法 | Advanced Usage

### 1. 类型安全的访问 | Type-Safe Access

```go
// 使用类型断言确保类型安全 | Use type assertion for type safety
func getUserInfo(execCtx *workflow.ExecutionContext) (map[string]interface{}, error) {
    raw, ok := execCtx.GetSessionState("user_info")
    if !ok {
        return nil, fmt.Errorf("user_info not found")
    }

    userInfo, ok := raw.(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("user_info has invalid type")
    }

    return userInfo, nil
}
```

### 2. 结构化数据存储 | Structured Data Storage

```go
type UserProfile struct {
    ID    string
    Name  string
    Email string
    Age   int
}

// 存储结构体 | Store struct
profile := UserProfile{
    ID:    "user-123",
    Name:  "Alice",
    Email: "alice@example.com",
    Age:   30,
}
execCtx.SetSessionState("user_profile", profile)

// 读取结构体 | Read struct
raw, _ := execCtx.GetSessionState("user_profile")
profile := raw.(UserProfile)
fmt.Printf("User: %s (%s)\n", profile.Name, profile.Email)
```

### 3. 嵌套数据访问 | Nested Data Access

```go
// 存储嵌套数据 | Store nested data
execCtx.SetSessionState("config", map[string]interface{}{
    "database": map[string]interface{}{
        "host": "localhost",
        "port": 5432,
    },
    "cache": map[string]interface{}{
        "enabled": true,
        "ttl":     300,
    },
})

// 访问嵌套数据 | Access nested data
configRaw, _ := execCtx.GetSessionState("config")
config := configRaw.(map[string]interface{})
db := config["database"].(map[string]interface{})
fmt.Printf("Database: %s:%d\n", db["host"], db["port"])
```

### 4. 会话隔离 | Session Isolation

```go
// 不同 session 使用不同的 ExecutionContext | Different sessions use different ExecutionContext

// Session A
sessionA := workflow.NewExecutionContextWithSession("input-a", "session-a", "user-1")
sessionA.SetSessionState("session_name", "Session A")

// Session B
sessionB := workflow.NewExecutionContextWithSession("input-b", "session-b", "user-2")
sessionB.SetSessionState("session_name", "Session B")

// Session A 和 Session B 的状态完全隔离 | Session A and Session B states are completely isolated
nameA, _ := sessionA.GetSessionState("session_name")  // "Session A"
nameB, _ := sessionB.GetSessionState("session_name")  // "Session B"
```

---

## 性能考虑 | Performance Considerations

### 深拷贝性能 | Deep Copy Performance

`Clone()` 使用 JSON 序列化实现深拷贝:

`Clone()` uses JSON serialization for deep copying:

```go
// 高效的场景 | Efficient scenarios
execCtx.SetSessionState("counter", 42)           // 简单类型 | Simple types
execCtx.SetSessionState("name", "Alice")         // 字符串 | Strings
execCtx.SetSessionState("enabled", true)         // 布尔值 | Booleans

// 低效的场景 (避免频繁克隆) | Inefficient scenarios (avoid frequent cloning)
execCtx.SetSessionState("large_data", hugeSlice) // 大型数据结构 | Large data structures
```

### 最佳实践 | Best Practices

1. **避免存储大型数据** | Avoid storing large data
   ```go
   // ❌ 不推荐 | Not recommended
   execCtx.SetSessionState("all_users", []User{ /* 10000+ users */ })

   // ✅ 推荐 | Recommended
   execCtx.SetSessionState("user_ids", []string{"id1", "id2", "id3"})
   ```

2. **使用引用而非复制** | Use references instead of copies
   ```go
   // ❌ 不推荐 | Not recommended
   data, _ := execCtx.GetSessionState("config")
   config := data.(map[string]interface{})
   // 修改 config 不会影响 SessionState | Modifying config won't affect SessionState

   // ✅ 推荐 | Recommended
   // 如果需要修改,先读取,修改后重新设置 | If modification needed, read, modify, then set again
   configRaw, _ := execCtx.GetSessionState("config")
   config := configRaw.(map[string]interface{})
   config["new_key"] = "new_value"
   execCtx.SetSessionState("config", config)
   ```

3. **合理使用并行分支** | Use parallel branches wisely
   ```go
   // ✅ 并行分支独立处理不同数据 | Parallel branches process different data independently
   // Branch A: 处理用户数据 | Process user data
   // Branch B: 处理订单数据 | Process order data
   // Branch C: 处理日志数据 | Process log data

   // ⚠️ 避免并行分支修改同一个键 | Avoid parallel branches modifying the same key
   // (除非你理解最终写入优先策略) | (unless you understand last-write-wins strategy)
   ```

---

## 测试 | Testing

完整的测试覆盖了以下场景:

Complete test coverage includes the following scenarios:

- ✅ 基础 Get/Set 操作 | Basic Get/Set operations
- ✅ 深拷贝 (Clone) | Deep copy (Clone)
- ✅ 状态合并 (Merge) | State merging (Merge)
- ✅ 并发安全 (1000 goroutines) | Concurrency safety (1000 goroutines)
- ✅ Workflow 集成测试 | Workflow integration tests
- ✅ 并行分支隔离 | Parallel branch isolation
- ✅ 多租户隔离 | Multi-tenant isolation

**测试覆盖率 | Test Coverage**: 543 行测试代码 | 543 lines of test code

运行测试 | Run tests:
```bash
cd pkg/agno/workflow
go test -v -run TestSessionState
```

---

## 故障排查 | Troubleshooting

### 常见问题 | Common Issues

#### 1. SessionState 为 nil

**现象 | Symptom**:
```go
panic: runtime error: invalid memory address or nil pointer dereference
```

**原因 | Cause**: 未初始化 SessionState

**解决 | Solution**:
```go
// ❌ 错误 | Wrong
execCtx := &workflow.ExecutionContext{}
execCtx.SetSessionState("key", "value")  // panic!

// ✅ 正确 | Correct
execCtx := workflow.NewExecutionContextWithSession("", "session-id", "user-id")
execCtx.SetSessionState("key", "value")  // OK
```

#### 2. 类型断言失败

**现象 | Symptom**:
```go
panic: interface conversion: interface {} is string, not int
```

**原因 | Cause**: 类型不匹配

**解决 | Solution**:
```go
// ❌ 错误 | Wrong
execCtx.SetSessionState("age", "30")  // 存储字符串 | Store string
age := execCtx.GetSessionState("age").(int)  // 尝试读取为 int | Try to read as int - panic!

// ✅ 正确 | Correct
execCtx.SetSessionState("age", 30)  // 存储 int | Store int
raw, ok := execCtx.GetSessionState("age")
if !ok {
    // 键不存在 | Key doesn't exist
}
age, ok := raw.(int)
if !ok {
    // 类型不匹配 | Type mismatch
}
```

#### 3. 并行分支状态丢失

**现象 | Symptom**: 并行分支设置的状态在合并后丢失

**原因 | Cause**: 没有理解最终写入优先策略

**解决 | Solution**:
```go
// 避免并行分支修改同一个键 | Avoid parallel branches modifying the same key
// 使用不同的键名 | Use different key names

// ✅ 推荐 | Recommended
// Branch A
execCtx.SetSessionState("branch_a_counter", 1)

// Branch B
execCtx.SetSessionState("branch_b_counter", 2)

// Branch C
execCtx.SetSessionState("branch_c_counter", 3)
```

---

## 相关文档 | Related Documentation

- [A2A Interface](A2A_INTERFACE.md) - Agent间通信接口
- [Multi-Tenant Support](MULTI_TENANT.md) - 多租户支持
- [Workflow Guide](WORKFLOW_GUIDE.md) - Workflow 使用指南

---

## 版本历史 | Version History

- **v1.1.0** (2025-01-XX): Initial Session State implementation
  - Thread-safe SessionState with RWMutex
  - Deep copy support for parallel branches
  - Smart merge strategy (last-write-wins)
  - Integration with ExecutionContext

---

**更新时间 | Last Updated**: 2025-01-XX
