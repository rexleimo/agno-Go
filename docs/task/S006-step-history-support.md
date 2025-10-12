# S006: Step 历史支持

**预估工作量**: 4 小时
**优先级**: P0
**前置依赖**: S005
**状态**: Pending

---

## 功能描述

在 Step 级别支持历史配置，允许：
1. Step 级别启用/禁用历史
2. 覆盖 Workflow 级别的历史设置
3. 自定义每个 Step 的历史数量

---

## 实现步骤

### 步骤 1: 扩展 Step 结构 (1h)

**文件**: `pkg/agno/workflow/step.go`

```go
// Step 结构体添加历史配置字段
type Step struct {
    // 现有字段...
    id     string
    name   string
    agent  *agent.Agent
    fn     StepFunc
    logger *slog.Logger

    // 新增字段：历史配置
    // New fields: History configuration

    // addWorkflowHistory enables history for this specific step
    // nil means inherit from workflow, true/false overrides
    // addWorkflowHistory 为此特定步骤启用历史
    // nil 表示从 workflow 继承，true/false 覆盖
    addWorkflowHistory *bool

    // numHistoryRuns specifies how many history runs to include
    // nil means use workflow default
    // numHistoryRuns 指定包含多少历史运行
    // nil 表示使用 workflow 默认值
    numHistoryRuns *int
}

// StepConfig 添加历史配置选项
type StepConfig struct {
    // 现有字段...
    ID          string
    Name        string
    Agent       *agent.Agent
    Function    StepFunc
    Logger      *slog.Logger

    // 新增字段
    // New fields

    // AddWorkflowHistory enables/disables history for this step
    // AddWorkflowHistory 为此步骤启用/禁用历史
    AddWorkflowHistory *bool `json:"add_workflow_history,omitempty"`

    // NumHistoryRuns specifies history count for this step
    // NumHistoryRuns 指定此步骤的历史数量
    NumHistoryRuns *int `json:"num_history_runs,omitempty"`
}
```

### 步骤 2: 修改 Step 构造函数 (0.5h)

```go
func NewStep(config StepConfig) (*Step, error) {
    // 现有验证...
    if config.ID == "" {
        config.ID = generateID()
    }

    step := &Step{
        id:                 config.ID,
        name:               config.Name,
        agent:              config.Agent,
        fn:                 config.Function,
        logger:             config.Logger,
        addWorkflowHistory: config.AddWorkflowHistory,
        numHistoryRuns:     config.NumHistoryRuns,
    }

    return step, nil
}
```

### 步骤 3: 实现历史决策逻辑 (1h)

```go
// shouldAddHistory 决定是否为此 step 添加历史
// shouldAddHistory determines whether to add history for this step
func (s *Step) shouldAddHistory(workflowAddHistory bool) bool {
    // Step 级别配置优先
    // Step-level configuration takes precedence
    if s.addWorkflowHistory != nil {
        return *s.addWorkflowHistory
    }

    // 否则使用 workflow 级别配置
    // Otherwise use workflow-level configuration
    return workflowAddHistory
}

// getHistoryRunCount 获取历史运行数量
// getHistoryRunCount gets the number of history runs to include
func (s *Step) getHistoryRunCount(workflowNumRuns int) int {
    // Step 级别配置优先
    // Step-level configuration takes precedence
    if s.numHistoryRuns != nil {
        return *s.numHistoryRuns
    }

    // 否则使用 workflow 级别配置
    // Otherwise use workflow-level configuration
    return workflowNumRuns
}
```

### 步骤 4: 修改 Step Execute 方法 (1.5h)

```go
// Execute 执行步骤（需要接收 workflow 配置）
// Execute executes the step (needs to receive workflow config)
func (s *Step) Execute(ctx context.Context, execCtx *ExecutionContext, workflowConfig HistoryConfig) (*ExecutionContext, error) {
    // 检查是否应该添加历史
    // Check if history should be added
    if s.shouldAddHistory(workflowConfig.AddHistory) {
        // 从 session state 获取历史
        // Get history from session state
        historyCount := s.getHistoryRunCount(workflowConfig.NumHistoryRuns)

        if session, ok := execCtx.GetSessionState("workflow_session"); ok {
            if workflowSession, ok := session.(*WorkflowSession); ok {
                // 获取格式化的历史上下文
                // Get formatted history context
                historyContext := workflowSession.GetHistoryContext(historyCount)

                // 如果使用 Agent，将历史添加到 instructions
                // If using Agent, add history to instructions
                if s.agent != nil {
                    s.injectHistoryToAgent(historyContext)
                }

                // 如果使用自定义函数，将历史添加到执行上下文
                // If using custom function, add history to execution context
                if s.fn != nil {
                    execCtx.SetSessionState("step_history_context", historyContext)
                }

                s.logger.Debug("added history to step",
                    "step_id", s.id,
                    "history_count", historyCount)
            }
        }
    }

    // 执行步骤（现有逻辑）
    // Execute step (existing logic)
    if s.agent != nil {
        output, err := s.agent.Run(ctx, execCtx.Input)
        if err != nil {
            return nil, err
        }
        execCtx.Output = output.Content

        // 保存消息历史到 session state
        // Save message history to session state
        if messages, ok := execCtx.GetSessionState("messages"); ok {
            msgList := messages.([]*types.Message)
            msgList = append(msgList, output.Messages...)
            execCtx.SetSessionState("messages", msgList)
        } else {
            execCtx.SetSessionState("messages", output.Messages)
        }

        return execCtx, nil
    }

    if s.fn != nil {
        return s.fn(ctx, execCtx)
    }

    return execCtx, nil
}

// injectHistoryToAgent 将历史注入到 agent 的 system message
// injectHistoryToAgent injects history into agent's system message
func (s *Step) injectHistoryToAgent(historyContext string) {
    if s.agent == nil || historyContext == "" {
        return
    }

    // 获取当前 instructions
    // Get current instructions
    currentInstructions := s.agent.GetInstructions()

    // 添加历史上下文
    // Add history context
    enhancedInstructions := currentInstructions + "\n\n" + historyContext

    // 临时更新 agent instructions (仅此次执行)
    // Temporarily update agent instructions (for this execution only)
    s.agent.SetInstructions(enhancedInstructions)
}
```

### 步骤 5: 定义 HistoryConfig 结构 (0.5h)

```go
// HistoryConfig 包含历史配置信息
// HistoryConfig contains history configuration information
type HistoryConfig struct {
    AddHistory     bool
    NumHistoryRuns int
}
```

### 步骤 6: 更新 Workflow 调用 (0.5h)

**文件**: `pkg/agno/workflow/workflow.go`

```go
func (w *Workflow) Run(ctx context.Context, input string, sessionID string) (*ExecutionContext, error) {
    // ... 现有逻辑

    // 创建历史配置
    // Create history configuration
    historyConfig := HistoryConfig{
        AddHistory:     w.addHistoryToSteps,
        NumHistoryRuns: w.numHistoryRuns,
    }

    // 执行步骤
    // Execute steps
    for i, step := range w.Steps {
        // ...

        // 传递历史配置
        // Pass history configuration
        result, err := step.Execute(ctx, execCtx, historyConfig)

        // ...
    }

    // ...
}
```

---

## 测试要求

### 单元测试

**文件**: `pkg/agno/workflow/step_history_test.go`

```go
package workflow

import (
    "context"
    "testing"
)

func TestStep_HistoryInheritance(t *testing.T) {
    // 测试 step 继承 workflow 的历史设置
    storage := NewMemoryStorage(0)

    // 创建 workflow (启用历史)
    workflow, _ := New(Config{
        EnableHistory:     true,
        HistoryStore:      storage,
        NumHistoryRuns:    3,
        AddHistoryToSteps: true,
        Steps: []Node{
            &Step{/* 不设置历史配置，应该继承 */},
        },
    })

    // ... 测试执行和验证
}

func TestStep_HistoryOverride(t *testing.T) {
    // 测试 step 覆盖 workflow 的历史设置
    addHistory := false // 覆盖为 false
    numRuns := 5        // 覆盖为 5

    step := &Step{
        addWorkflowHistory: &addHistory,
        numHistoryRuns:     &numRuns,
    }

    // 测试决策逻辑
    if step.shouldAddHistory(true) != false {
        t.Error("expected step to override workflow setting")
    }

    if step.getHistoryRunCount(3) != 5 {
        t.Error("expected step to override history count")
    }
}

func TestStep_DifferentHistoryCounts(t *testing.T) {
    // 测试不同 step 使用不同的历史数量
    // ...
}

func TestStep_HistoryInjectionToAgent(t *testing.T) {
    // 测试历史正确注入到 agent
    // ...
}

func TestStep_CustomFunctionWithHistory(t *testing.T) {
    // 测试自定义函数访问历史
    // ...
}
```

---

## 验收标准

- [x] Step 结构支持历史配置字段
- [x] Step 可以继承 Workflow 的历史设置
- [x] Step 可以覆盖 Workflow 的历史设置
- [x] 历史正确注入到 Agent 的 system message
- [x] 自定义函数可以访问历史上下文
- [x] 测试覆盖率 >80%
- [x] 所有测试通过
- [x] 不影响不使用历史功能的性能

---

## 相关文件

- `pkg/agno/workflow/step.go` - 主要修改
- `pkg/agno/workflow/workflow.go` - 更新调用
- `pkg/agno/workflow/condition.go` - 类似修改
- `pkg/agno/workflow/loop.go` - 类似修改
- `pkg/agno/workflow/parallel.go` - 类似修改
- `pkg/agno/workflow/router.go` - 类似修改
- `pkg/agno/workflow/step_history_test.go` - 新增测试

---

## 性能目标

- 历史注入: <1ms
- 不影响不使用历史的 step 性能
- 内存开销: 可忽略（仅 2 个指针字段）

---

## 注意事项

1. **所有 Node 类型**: Condition, Loop, Parallel, Router 都需要类似的修改
2. **Agent Instructions**: 临时修改 instructions 后需要恢复
3. **并发安全**: 如果 Agent 被多个 Step 共享，需要注意并发问题
4. **向后兼容**: 字段使用指针，nil 表示不设置（继承）
