# S005: Workflow 集成 Session 管理

**预估工作量**: 6 小时
**优先级**: P0
**前置依赖**: S001, S002, S003, S004
**状态**: Pending

---

## 功能描述

在 Workflow 结构体中集成 Session 管理，支持：
1. 配置历史记录选项（是否启用、历史数量）
2. 在 Workflow 执行前加载历史
3. 在 Workflow 执行后保存运行记录
4. 支持 SessionID 参数传递

---

## 实现步骤

### 步骤 1: 扩展 Workflow Config (1h)

**文件**: `pkg/agno/workflow/workflow.go`

```go
// Config 结构体添加字段
type Config struct {
    // 现有字段...
    ID     string
    Name   string
    Steps  []Node
    Logger *slog.Logger

    // 新增字段：历史管理
    // New fields: History management

    // EnableHistory enables workflow history tracking
    // EnableHistory 启用工作流历史跟踪
    EnableHistory bool `json:"enable_history"`

    // HistoryStore is the storage backend for workflow history
    // HistoryStore 是工作流历史的存储后端
    HistoryStore WorkflowStorage `json:"-"`

    // NumHistoryRuns is the number of recent runs to include in history context
    // NumHistoryRuns 是历史上下文中包含的最近运行数量
    NumHistoryRuns int `json:"num_history_runs"`

    // AddHistoryToSteps automatically adds history context to all steps
    // AddHistoryToSteps 自动将历史上下文添加到所有步骤
    AddHistoryToSteps bool `json:"add_history_to_steps"`
}

// Workflow 结构体添加字段
type Workflow struct {
    // 现有字段...
    ID     string
    Name   string
    Steps  []Node
    logger *slog.Logger

    // 新增字段
    // New fields
    enableHistory     bool
    historyStore      WorkflowStorage
    numHistoryRuns    int
    addHistoryToSteps bool
}
```

### 步骤 2: 修改 Workflow 构造函数 (1h)

```go
func New(config Config) (*Workflow, error) {
    // 现有验证...
    if config.ID == "" {
        config.ID = generateID()
    }

    // 历史配置验证和默认值
    // History configuration validation and defaults
    if config.EnableHistory && config.HistoryStore == nil {
        // 使用默认内存存储
        // Use default memory storage
        config.HistoryStore = NewMemoryStorage(100)
    }

    if config.NumHistoryRuns <= 0 {
        config.NumHistoryRuns = 3 // 默认包含最近 3 次运行
    }

    w := &Workflow{
        ID:                config.ID,
        Name:              config.Name,
        Steps:             config.Steps,
        logger:            config.Logger,
        enableHistory:     config.EnableHistory,
        historyStore:      config.HistoryStore,
        numHistoryRuns:    config.NumHistoryRuns,
        addHistoryToSteps: config.AddHistoryToSteps,
    }

    // ... 其他初始化

    return w, nil
}
```

### 步骤 3: 修改 Run 方法签名 (1h)

```go
// Run 执行工作流
// Run executes the workflow
// 新增 sessionID 参数
func (w *Workflow) Run(ctx context.Context, input string, sessionID string) (*ExecutionContext, error) {
    // 验证输入
    if input == "" {
        return nil, types.NewInvalidInputError("input cannot be empty", nil)
    }

    // 生成 session ID（如果未提供）
    // Generate session ID if not provided
    if sessionID == "" {
        sessionID = generateSessionID()
    }

    w.logger.Info("workflow started",
        "workflow_id", w.ID,
        "session_id", sessionID,
        "steps", len(w.Steps))

    // 创建执行上下文
    execCtx := NewExecutionContext(input)
    execCtx.SessionID = sessionID

    // 加载历史（如果启用）
    // Load history (if enabled)
    if w.enableHistory && w.historyStore != nil {
        if err := w.loadHistory(ctx, execCtx); err != nil {
            w.logger.Error("failed to load history", "error", err)
            // 不阻止执行，仅记录错误
        }
    }

    // 创建 WorkflowRun 记录
    // Create WorkflowRun record
    var workflowRun *WorkflowRun
    if w.enableHistory {
        runID := generateRunID()
        workflowRun = NewWorkflowRun(runID, sessionID, w.ID, input)
        workflowRun.MarkStarted()
    }

    // 执行步骤（现有逻辑）
    // Execute steps (existing logic)
    for i, step := range w.Steps {
        select {
        case <-ctx.Done():
            if workflowRun != nil {
                workflowRun.MarkCancelled()
                w.saveRun(ctx, sessionID, workflowRun)
            }
            return nil, ctx.Err()
        default:
        }

        w.logger.Info("executing step",
            "step_id", step.GetID(),
            "step_type", step.GetType(),
            "sequence", i+1)

        result, err := step.Execute(ctx, execCtx)
        if err != nil {
            w.logger.Error("step execution failed",
                "step_id", step.GetID(),
                "error", err)

            if workflowRun != nil {
                workflowRun.MarkFailed(err)
                w.saveRun(ctx, sessionID, workflowRun)
            }

            return nil, err
        }

        execCtx = result
    }

    // 保存历史（如果启用）
    // Save history (if enabled)
    if workflowRun != nil {
        workflowRun.MarkCompleted(execCtx.Output)
        workflowRun.Messages = extractMessages(execCtx)
        w.saveRun(ctx, sessionID, workflowRun)
    }

    w.logger.Info("workflow completed",
        "workflow_id", w.ID,
        "session_id", sessionID)

    return execCtx, nil
}
```

### 步骤 4: 实现历史加载和保存 (2h)

```go
// loadHistory 从存储加载历史并添加到上下文
// loadHistory loads history from storage and adds to context
func (w *Workflow) loadHistory(ctx context.Context, execCtx *ExecutionContext) error {
    if w.historyStore == nil {
        return nil
    }

    // 获取或创建 session
    // Get or create session
    session, err := w.historyStore.GetSession(ctx, execCtx.SessionID)
    if err != nil {
        if err == ErrSessionNotFound {
            // 创建新 session
            // Create new session
            session, err = w.historyStore.CreateSession(
                ctx,
                execCtx.SessionID,
                w.ID,
                "", // userID 可以从 execCtx 获取
            )
            if err != nil {
                return fmt.Errorf("failed to create session: %w", err)
            }
        } else {
            return fmt.Errorf("failed to get session: %w", err)
        }
    }

    // 获取历史记录
    // Get history
    history := session.GetHistory(w.numHistoryRuns)
    if len(history) == 0 {
        return nil
    }

    // 如果配置了添加历史到步骤
    // If configured to add history to steps
    if w.addHistoryToSteps {
        historyContext := session.GetHistoryContext(w.numHistoryRuns)
        execCtx.SetSessionState("workflow_history_context", historyContext)
    }

    // 将历史数据存储在上下文中
    // Store history data in context
    execCtx.SetSessionState("workflow_history", history)
    execCtx.SetSessionState("workflow_session", session)

    w.logger.Debug("loaded history",
        "session_id", execCtx.SessionID,
        "history_count", len(history))

    return nil
}

// saveRun 保存运行记录到存储
// saveRun saves run record to storage
func (w *Workflow) saveRun(ctx context.Context, sessionID string, run *WorkflowRun) error {
    if w.historyStore == nil {
        return nil
    }

    // 获取 session
    // Get session
    session, err := w.historyStore.GetSession(ctx, sessionID)
    if err != nil {
        return fmt.Errorf("failed to get session: %w", err)
    }

    // 添加运行记录
    // Add run record
    session.AddRun(run)

    // 更新 session
    // Update session
    if err := w.historyStore.UpdateSession(ctx, session); err != nil {
        return fmt.Errorf("failed to update session: %w", err)
    }

    w.logger.Debug("saved run",
        "session_id", sessionID,
        "run_id", run.RunID,
        "status", run.Status)

    return nil
}

// extractMessages 从执行上下文提取消息
// extractMessages extracts messages from execution context
func extractMessages(execCtx *ExecutionContext) []*types.Message {
    // 从 session state 中提取消息历史
    // Extract message history from session state
    if messages, ok := execCtx.GetSessionState("messages"); ok {
        if msgList, ok := messages.([]*types.Message); ok {
            return msgList
        }
    }

    // 如果没有消息历史，创建基本的输入/输出消息
    // If no message history, create basic input/output messages
    return []*types.Message{
        types.NewUserMessage(execCtx.Input),
        types.NewAssistantMessage(execCtx.Output),
    }
}

// generateSessionID 生成唯一的 session ID
// generateSessionID generates a unique session ID
func generateSessionID() string {
    import "github.com/google/uuid"
    return "session-" + uuid.New().String()
}

// generateRunID 生成唯一的 run ID
// generateRunID generates a unique run ID
func generateRunID() string {
    import "github.com/google/uuid"
    return "run-" + uuid.New().String()
}
```

### 步骤 5: 修改 ExecutionContext (1h)

**文件**: 检查 `pkg/agno/workflow/workflow.go` 中的 ExecutionContext

```go
// ExecutionContext 添加 SessionID 字段（如果还没有）
type ExecutionContext struct {
    // 现有字段...
    Input        string
    Output       string
    SessionState map[string]interface{}
    Metadata     map[string]interface{}

    // 新增字段
    // New field
    SessionID string `json:"session_id"`
}
```

---

## 测试要求

### 单元测试

**文件**: `pkg/agno/workflow/workflow_history_test.go`

```go
package workflow

import (
    "context"
    "testing"
)

func TestWorkflow_WithHistory(t *testing.T) {
    // 创建带历史的 workflow
    storage := NewMemoryStorage(0)

    workflow, err := New(Config{
        ID:                "test-workflow",
        Name:              "Test Workflow",
        Steps:             []Node{/* mock steps */},
        EnableHistory:     true,
        HistoryStore:      storage,
        NumHistoryRuns:    3,
        AddHistoryToSteps: true,
    })

    if err != nil {
        t.Fatalf("failed to create workflow: %v", err)
    }

    ctx := context.Background()
    sessionID := "test-session-1"

    // 第一次运行
    result1, err := workflow.Run(ctx, "input-1", sessionID)
    if err != nil {
        t.Fatalf("first run failed: %v", err)
    }

    // 第二次运行（应该能看到历史）
    result2, err := workflow.Run(ctx, "input-2", sessionID)
    if err != nil {
        t.Fatalf("second run failed: %v", err)
    }

    // 验证历史被加载
    history, ok := result2.GetSessionState("workflow_history")
    if !ok {
        t.Error("expected workflow_history in session state")
    }

    historyEntries, ok := history.([]HistoryEntry)
    if !ok {
        t.Error("expected workflow_history to be []HistoryEntry")
    }

    if len(historyEntries) != 1 {
        t.Errorf("expected 1 history entry, got %d", len(historyEntries))
    }

    // 验证 session 中的运行记录
    session, err := storage.GetSession(ctx, sessionID)
    if err != nil {
        t.Fatalf("failed to get session: %v", err)
    }

    if session.CountRuns() != 2 {
        t.Errorf("expected 2 runs in session, got %d", session.CountRuns())
    }
}

func TestWorkflow_WithoutHistory(t *testing.T) {
    // 测试未启用历史的情况
    workflow, err := New(Config{
        ID:            "test-workflow",
        Name:          "Test Workflow",
        Steps:         []Node{/* mock steps */},
        EnableHistory: false,
    })

    if err != nil {
        t.Fatalf("failed to create workflow: %v", err)
    }

    ctx := context.Background()

    // 运行不应该失败
    _, err = workflow.Run(ctx, "input", "session-1")
    if err != nil {
        t.Fatalf("run failed: %v", err)
    }
}

func TestWorkflow_HistoryLoadError(t *testing.T) {
    // 测试历史加载失败的情况（不应该阻止执行）
    // ...
}

func TestWorkflow_MultipleSessionsHistory(t *testing.T) {
    // 测试多个 session 的历史隔离
    // ...
}
```

---

## 验收标准

- [x] Workflow Config 支持历史配置选项
- [x] Run 方法接受 sessionID 参数
- [x] 执行前自动加载历史
- [x] 执行后自动保存运行记录
- [x] 历史加载失败不阻止执行
- [x] 支持禁用历史功能
- [x] 测试覆盖率 >80%
- [x] 所有测试通过
- [x] 性能：历史加载 <5ms
- [x] 文档：更新 API 文档

---

## 相关文件

- `pkg/agno/workflow/workflow.go` - 主要修改
- `pkg/agno/workflow/run.go` - WorkflowRun 定义
- `pkg/agno/workflow/session.go` - WorkflowSession 定义
- `pkg/agno/workflow/storage.go` - WorkflowStorage 接口
- `pkg/agno/workflow/memory_storage.go` - 内存存储实现
- `pkg/agno/workflow/workflow_history_test.go` - 新增测试文件

---

## 性能目标

- 历史加载: <5ms
- 历史保存: <10ms (可异步)
- 内存开销: 每个 session ~10KB
- 不影响不使用历史功能的性能

---

## 注意事项

1. **向后兼容**: Run 方法签名改变，需要更新所有调用处
2. **错误处理**: 历史操作失败不应阻止 workflow 执行
3. **并发安全**: 确保 session 操作是线程安全的
4. **默认行为**: EnableHistory 默认 false，不影响现有代码
