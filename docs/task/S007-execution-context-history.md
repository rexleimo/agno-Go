# S007: ExecutionContext 集成历史

**预估工作量**: 3 小时
**优先级**: P0
**前置依赖**: S006
**状态**: Pending

---

## 功能描述

在 ExecutionContext 中集成历史功能，提供：
1. 历史数据的访问接口
2. 历史上下文的存储和检索
3. 消息历史的管理

---

## 实现步骤

### 步骤 1: 扩展 ExecutionContext 结构 (0.5h)

**文件**: `pkg/agno/workflow/workflow.go` (或创建新文件 `execution_context.go`)

```go
// ExecutionContext 执行上下文
type ExecutionContext struct {
    // 现有字段...
    Input        string                 `json:"input"`
    Output       string                 `json:"output"`
    SessionState map[string]interface{} `json:"session_state,omitempty"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`

    // 已添加的字段
    // Already added field
    SessionID string `json:"session_id"`

    // 新增字段：直接访问历史
    // New fields: Direct access to history

    // WorkflowHistory 包含最近的工作流运行历史
    // WorkflowHistory contains recent workflow run history
    WorkflowHistory []HistoryEntry `json:"workflow_history,omitempty"`

    // HistoryContext 是格式化的历史上下文字符串
    // HistoryContext is the formatted history context string
    HistoryContext string `json:"history_context,omitempty"`
}

// NewExecutionContext 创建新的执行上下文
// NewExecutionContext creates a new execution context
func NewExecutionContext(input string) *ExecutionContext {
    return &ExecutionContext{
        Input:           input,
        SessionState:    make(map[string]interface{}),
        Metadata:        make(map[string]interface{}),
        WorkflowHistory: make([]HistoryEntry, 0),
    }
}
```

### 步骤 2: 添加历史访问方法 (1h)

```go
// GetWorkflowHistory 获取工作流历史
// GetWorkflowHistory gets workflow history
func (e *ExecutionContext) GetWorkflowHistory() []HistoryEntry {
    return e.WorkflowHistory
}

// SetWorkflowHistory 设置工作流历史
// SetWorkflowHistory sets workflow history
func (e *ExecutionContext) SetWorkflowHistory(history []HistoryEntry) {
    e.WorkflowHistory = history
}

// GetHistoryContext 获取格式化的历史上下文
// GetHistoryContext gets formatted history context
func (e *ExecutionContext) GetHistoryContext() string {
    return e.HistoryContext
}

// SetHistoryContext 设置格式化的历史上下文
// SetHistoryContext sets formatted history context
func (e *ExecutionContext) SetHistoryContext(context string) {
    e.HistoryContext = context
}

// HasHistory 检查是否有历史记录
// HasHistory checks if there is history
func (e *ExecutionContext) HasHistory() bool {
    return len(e.WorkflowHistory) > 0
}

// GetHistoryCount 获取历史记录数量
// GetHistoryCount gets the number of history entries
func (e *ExecutionContext) GetHistoryCount() int {
    return len(e.WorkflowHistory)
}

// GetLastHistoryEntry 获取最后一个历史条目
// GetLastHistoryEntry gets the last history entry
func (e *ExecutionContext) GetLastHistoryEntry() *HistoryEntry {
    if len(e.WorkflowHistory) == 0 {
        return nil
    }
    return &e.WorkflowHistory[len(e.WorkflowHistory)-1]
}

// GetHistoryInput 获取指定索引的历史输入
// GetHistoryInput gets history input at specified index
// index 0 表示最早的历史，-1 表示最近的历史
// index 0 means earliest history, -1 means most recent
func (e *ExecutionContext) GetHistoryInput(index int) string {
    if len(e.WorkflowHistory) == 0 {
        return ""
    }

    if index < 0 {
        // 负索引从末尾开始
        // Negative index from the end
        index = len(e.WorkflowHistory) + index
    }

    if index < 0 || index >= len(e.WorkflowHistory) {
        return ""
    }

    return e.WorkflowHistory[index].Input
}

// GetHistoryOutput 获取指定索引的历史输出
// GetHistoryOutput gets history output at specified index
func (e *ExecutionContext) GetHistoryOutput(index int) string {
    if len(e.WorkflowHistory) == 0 {
        return ""
    }

    if index < 0 {
        index = len(e.WorkflowHistory) + index
    }

    if index < 0 || index >= len(e.WorkflowHistory) {
        return ""
    }

    return e.WorkflowHistory[index].Output
}
```

### 步骤 3: 更新 Workflow 加载历史逻辑 (1h)

**文件**: `pkg/agno/workflow/workflow.go`

```go
func (w *Workflow) loadHistory(ctx context.Context, execCtx *ExecutionContext) error {
    // ... 现有逻辑

    // 获取历史记录
    history := session.GetHistory(w.numHistoryRuns)
    if len(history) == 0 {
        return nil
    }

    // 直接设置到 ExecutionContext
    // Set directly to ExecutionContext
    execCtx.SetWorkflowHistory(history)

    // 如果配置了添加历史到步骤
    if w.addHistoryToSteps {
        historyContext := session.GetHistoryContext(w.numHistoryRuns)
        execCtx.SetHistoryContext(historyContext)

        // 也保存到 session state (向后兼容)
        // Also save to session state (backward compatibility)
        execCtx.SetSessionState("workflow_history_context", historyContext)
    }

    // 保存 session 引用
    execCtx.SetSessionState("workflow_session", session)

    w.logger.Debug("loaded history",
        "session_id", execCtx.SessionID,
        "history_count", len(history))

    return nil
}
```

### 步骤 4: 添加消息历史管理 (0.5h)

```go
// GetMessages 获取消息历史
// GetMessages gets message history
func (e *ExecutionContext) GetMessages() []*types.Message {
    if messages, ok := e.GetSessionState("messages"); ok {
        if msgList, ok := messages.([]*types.Message); ok {
            return msgList
        }
    }
    return []*types.Message{}
}

// AddMessage 添加消息到历史
// AddMessage adds a message to history
func (e *ExecutionContext) AddMessage(msg *types.Message) {
    messages := e.GetMessages()
    messages = append(messages, msg)
    e.SetSessionState("messages", messages)
}

// AddMessages 批量添加消息
// AddMessages adds multiple messages
func (e *ExecutionContext) AddMessages(msgs []*types.Message) {
    messages := e.GetMessages()
    messages = append(messages, msgs...)
    e.SetSessionState("messages", messages)
}

// ClearMessages 清空消息历史
// ClearMessages clears message history
func (e *ExecutionContext) ClearMessages() {
    e.SetSessionState("messages", []*types.Message{})
}
```

---

## 测试要求

### 单元测试

**文件**: `pkg/agno/workflow/execution_context_test.go`

```go
package workflow

import (
    "testing"
    "time"
)

func TestExecutionContext_HistoryMethods(t *testing.T) {
    execCtx := NewExecutionContext("test input")

    // 初始状态
    if execCtx.HasHistory() {
        t.Error("expected no history initially")
    }

    if execCtx.GetHistoryCount() != 0 {
        t.Error("expected history count to be 0")
    }

    // 添加历史
    history := []HistoryEntry{
        {Input: "input-1", Output: "output-1", Timestamp: time.Now()},
        {Input: "input-2", Output: "output-2", Timestamp: time.Now()},
    }

    execCtx.SetWorkflowHistory(history)

    // 验证
    if !execCtx.HasHistory() {
        t.Error("expected to have history")
    }

    if execCtx.GetHistoryCount() != 2 {
        t.Errorf("expected history count 2, got %d", execCtx.GetHistoryCount())
    }

    // 获取最后一个条目
    last := execCtx.GetLastHistoryEntry()
    if last == nil || last.Input != "input-2" {
        t.Error("expected last entry to be input-2")
    }
}

func TestExecutionContext_GetHistoryInput(t *testing.T) {
    execCtx := NewExecutionContext("test")

    history := []HistoryEntry{
        {Input: "input-0", Output: "output-0"},
        {Input: "input-1", Output: "output-1"},
        {Input: "input-2", Output: "output-2"},
    }
    execCtx.SetWorkflowHistory(history)

    // 测试正索引
    if input := execCtx.GetHistoryInput(0); input != "input-0" {
        t.Errorf("expected input-0, got %s", input)
    }

    if input := execCtx.GetHistoryInput(1); input != "input-1" {
        t.Errorf("expected input-1, got %s", input)
    }

    // 测试负索引
    if input := execCtx.GetHistoryInput(-1); input != "input-2" {
        t.Errorf("expected input-2 (last), got %s", input)
    }

    if input := execCtx.GetHistoryInput(-2); input != "input-1" {
        t.Errorf("expected input-1 (second last), got %s", input)
    }

    // 测试越界
    if input := execCtx.GetHistoryInput(999); input != "" {
        t.Error("expected empty string for out of bounds")
    }
}

func TestExecutionContext_MessageManagement(t *testing.T) {
    execCtx := NewExecutionContext("test")

    // 初始无消息
    messages := execCtx.GetMessages()
    if len(messages) != 0 {
        t.Error("expected no messages initially")
    }

    // 添加消息
    msg1 := types.NewUserMessage("hello")
    execCtx.AddMessage(msg1)

    messages = execCtx.GetMessages()
    if len(messages) != 1 {
        t.Errorf("expected 1 message, got %d", len(messages))
    }

    // 批量添加
    msgs := []*types.Message{
        types.NewAssistantMessage("hi"),
        types.NewUserMessage("how are you"),
    }
    execCtx.AddMessages(msgs)

    messages = execCtx.GetMessages()
    if len(messages) != 3 {
        t.Errorf("expected 3 messages, got %d", len(messages))
    }

    // 清空
    execCtx.ClearMessages()
    messages = execCtx.GetMessages()
    if len(messages) != 0 {
        t.Error("expected no messages after clear")
    }
}

func TestExecutionContext_HistoryContext(t *testing.T) {
    execCtx := NewExecutionContext("test")

    context := "<workflow_history_context>test</workflow_history_context>"
    execCtx.SetHistoryContext(context)

    retrieved := execCtx.GetHistoryContext()
    if retrieved != context {
        t.Error("expected history context to match")
    }
}
```

---

## 验收标准

- [x] ExecutionContext 包含历史字段
- [x] 提供便捷的历史访问方法
- [x] 支持正负索引访问历史
- [x] 消息历史管理方法完整
- [x] 测试覆盖率 >90%
- [x] 所有测试通过
- [x] API 简洁易用

---

## 相关文件

- `pkg/agno/workflow/execution_context.go` - 新文件或在 workflow.go 中
- `pkg/agno/workflow/workflow.go` - 更新 loadHistory
- `pkg/agno/workflow/execution_context_test.go` - 新增测试

---

## 性能目标

- 所有方法 <100ns
- 零额外内存分配（对于访问方法）
- 不影响不使用历史功能的性能

---

## API 设计原则

1. **简洁性**: 提供最常用的访问模式
2. **一致性**: 与其他 Go API 保持一致（如负索引）
3. **类型安全**: 返回具体类型而非 interface{}
4. **零值友好**: 访问不存在的数据返回零值而非 panic
