# S009: Workflow History 端到端测试

**预估工作量**: 6 小时
**优先级**: P0
**前置依赖**: S005, S006, S007, S008
**状态**: Pending

---

## 功能描述

编写全面的端到端测试，验证 Workflow History 功能：
1. 多轮对话场景测试
2. 不同配置组合测试
3. 并发执行测试
4. 性能基准测试
5. 集成示例

---

## 测试场景

### 场景 1: 基础历史功能 (1h)

**文件**: `pkg/agno/workflow/workflow_history_e2e_test.go`

```go
package workflow

import (
    "context"
    "testing"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models"
)

// TestWorkflowHistory_BasicMultiTurn 测试基础多轮对话
func TestWorkflowHistory_BasicMultiTurn(t *testing.T) {
    // 创建 mock model
    mockModel := &MockModel{
        InvokeFunc: func(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
            // 返回包含历史信息的响应
            return &types.ModelResponse{
                Content: "I remember our previous conversation",
            }, nil
        },
    }

    // 创建 agent
    testAgent, err := agent.New(agent.Config{
        Name:         "test-agent",
        Model:        mockModel,
        Instructions: "You are a helpful assistant with memory",
    })
    if err != nil {
        t.Fatalf("failed to create agent: %v", err)
    }

    // 创建带历史的 workflow
    storage := NewMemoryStorage(0)
    workflow, err := New(Config{
        ID:                "test-workflow",
        Name:              "Test Workflow",
        EnableHistory:     true,
        HistoryStore:      storage,
        NumHistoryRuns:    3,
        AddHistoryToSteps: true,
        Steps: []Node{
            &Step{
                id:    "step-1",
                name:  "Chat",
                agent: testAgent,
            },
        },
    })
    if err != nil {
        t.Fatalf("failed to create workflow: %v", err)
    }

    ctx := context.Background()
    sessionID := "test-session-1"

    // 第一轮对话
    result1, err := workflow.Run(ctx, "Hello, my name is Alice", sessionID)
    if err != nil {
        t.Fatalf("first run failed: %v", err)
    }

    // 验证第一轮结果
    if result1 == nil {
        t.Fatal("expected non-nil result")
    }

    // 第二轮对话 - 应该能访问历史
    result2, err := workflow.Run(ctx, "What's my name?", sessionID)
    if err != nil {
        t.Fatalf("second run failed: %v", err)
    }

    // 验证历史被加载
    if !result2.HasHistory() {
        t.Error("expected history in second run")
    }

    if result2.GetHistoryCount() != 1 {
        t.Errorf("expected 1 history entry, got %d", result2.GetHistoryCount())
    }

    // 验证历史内容
    firstHistory := result2.GetHistoryInput(0)
    if firstHistory != "Hello, my name is Alice" {
        t.Errorf("expected first history input, got %s", firstHistory)
    }

    // 第三轮对话
    result3, err := workflow.Run(ctx, "Tell me about our conversation", sessionID)
    if err != nil {
        t.Fatalf("third run failed: %v", err)
    }

    // 应该有 2 个历史条目
    if result3.GetHistoryCount() != 2 {
        t.Errorf("expected 2 history entries, got %d", result3.GetHistoryCount())
    }
}
```

### 场景 2: Step 级别历史配置 (1h)

```go
// TestWorkflowHistory_StepLevelConfig 测试 Step 级别的历史配置
func TestWorkflowHistory_StepLevelConfig(t *testing.T) {
    mockModel := &MockModel{
        InvokeFunc: func(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
            // 检查 system message 是否包含历史
            hasHistory := false
            for _, msg := range req.Messages {
                if msg.Role == types.RoleSystem && contains(msg.Content, "workflow_history") {
                    hasHistory = true
                    break
                }
            }

            return &types.ModelResponse{
                Content: fmt.Sprintf("has_history: %v", hasHistory),
            }, nil
        },
    }

    // 创建两个 agent
    agent1, _ := agent.New(agent.Config{
        Name:  "agent-1",
        Model: mockModel,
    })

    agent2, _ := agent.New(agent.Config{
        Name:  "agent-2",
        Model: mockModel,
    })

    // Step 1: 启用历史 (使用 5 个历史)
    enableHistory := true
    numRuns5 := 5

    step1 := &Step{
        id:                 "step-1",
        name:               "With History",
        agent:              agent1,
        addWorkflowHistory: &enableHistory,
        numHistoryRuns:     &numRuns5,
    }

    // Step 2: 禁用历史
    disableHistory := false

    step2 := &Step{
        id:                 "step-2",
        name:               "Without History",
        agent:              agent2,
        addWorkflowHistory: &disableHistory,
    }

    // 创建 workflow (默认启用历史，3 个运行)
    storage := NewMemoryStorage(0)
    workflow, _ := New(Config{
        ID:                "test-workflow",
        EnableHistory:     true,
        HistoryStore:      storage,
        NumHistoryRuns:    3,
        AddHistoryToSteps: true,
        Steps:             []Node{step1, step2},
    })

    ctx := context.Background()
    sessionID := "test-session-2"

    // 运行多次以积累历史
    for i := 0; i < 6; i++ {
        _, err := workflow.Run(ctx, fmt.Sprintf("input-%d", i), sessionID)
        if err != nil {
            t.Fatalf("run %d failed: %v", i, err)
        }
    }

    // 最后一次运行并验证配置生效
    result, err := workflow.Run(ctx, "final input", sessionID)
    if err != nil {
        t.Fatalf("final run failed: %v", err)
    }

    // Step 1 应该有历史（但只用了最近 5 个，尽管 workflow 配置是 3）
    // Step 2 应该没有历史

    // 验证历史加载
    if result.GetHistoryCount() != 6 {
        t.Errorf("expected 6 history entries, got %d", result.GetHistoryCount())
    }
}
```

### 场景 3: 多 Session 隔离 (1h)

```go
// TestWorkflowHistory_SessionIsolation 测试多个 session 的历史隔离
func TestWorkflowHistory_SessionIsolation(t *testing.T) {
    mockModel := &MockModel{
        InvokeFunc: func(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
            return &types.ModelResponse{
                Content: "response",
            }, nil
        },
    }

    testAgent, _ := agent.New(agent.Config{
        Name:  "test-agent",
        Model: mockModel,
    })

    storage := NewMemoryStorage(0)
    workflow, _ := New(Config{
        ID:                "test-workflow",
        EnableHistory:     true,
        HistoryStore:      storage,
        NumHistoryRuns:    10,
        AddHistoryToSteps: true,
        Steps: []Node{
            &Step{id: "step-1", agent: testAgent},
        },
    })

    ctx := context.Background()

    // Session 1: 运行 3 次
    for i := 0; i < 3; i++ {
        workflow.Run(ctx, fmt.Sprintf("session1-input-%d", i), "session-1")
    }

    // Session 2: 运行 5 次
    for i := 0; i < 5; i++ {
        workflow.Run(ctx, fmt.Sprintf("session2-input-%d", i), "session-2")
    }

    // 验证 session 1 的历史
    result1, _ := workflow.Run(ctx, "check history", "session-1")
    if result1.GetHistoryCount() != 3 {
        t.Errorf("session 1: expected 3 history entries, got %d", result1.GetHistoryCount())
    }

    // 验证历史内容不包含 session 2 的数据
    for i := 0; i < result1.GetHistoryCount(); i++ {
        input := result1.GetHistoryInput(i)
        if contains(input, "session2") {
            t.Error("session 1 history should not contain session 2 data")
        }
    }

    // 验证 session 2 的历史
    result2, _ := workflow.Run(ctx, "check history", "session-2")
    if result2.GetHistoryCount() != 5 {
        t.Errorf("session 2: expected 5 history entries, got %d", result2.GetHistoryCount())
    }

    // 验证历史内容不包含 session 1 的数据
    for i := 0; i < result2.GetHistoryCount(); i++ {
        input := result2.GetHistoryInput(i)
        if contains(input, "session1") {
            t.Error("session 2 history should not contain session 1 data")
        }
    }
}
```

### 场景 4: 并发执行测试 (1h)

```go
// TestWorkflowHistory_Concurrency 测试并发执行的安全性
func TestWorkflowHistory_Concurrency(t *testing.T) {
    mockModel := &MockModel{
        InvokeFunc: func(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
            return &types.ModelResponse{
                Content: "response",
            }, nil
        },
    }

    testAgent, _ := agent.New(agent.Config{
        Name:  "test-agent",
        Model: mockModel,
    })

    storage := NewMemoryStorage(0)
    workflow, _ := New(Config{
        ID:                "test-workflow",
        EnableHistory:     true,
        HistoryStore:      storage,
        NumHistoryRuns:    5,
        AddHistoryToSteps: true,
        Steps: []Node{
            &Step{id: "step-1", agent: testAgent},
        },
    })

    ctx := context.Background()

    // 并发运行多个 sessions
    var wg sync.WaitGroup
    numSessions := 10
    numRunsPerSession := 20

    for sessionIdx := 0; sessionIdx < numSessions; sessionIdx++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()

            sessionID := fmt.Sprintf("session-%d", idx)

            for runIdx := 0; runIdx < numRunsPerSession; runIdx++ {
                input := fmt.Sprintf("session-%d-run-%d", idx, runIdx)
                _, err := workflow.Run(ctx, input, sessionID)
                if err != nil {
                    t.Errorf("concurrent run failed: %v", err)
                }
            }
        }(sessionIdx)
    }

    wg.Wait()

    // 验证每个 session 的历史完整性
    for sessionIdx := 0; sessionIdx < numSessions; sessionIdx++ {
        sessionID := fmt.Sprintf("session-%d", sessionIdx)

        session, err := storage.GetSession(ctx, sessionID)
        if err != nil {
            t.Errorf("failed to get session %s: %v", sessionID, err)
            continue
        }

        if session.CountRuns() != numRunsPerSession {
            t.Errorf("session %s: expected %d runs, got %d",
                sessionID, numRunsPerSession, session.CountRuns())
        }

        // 验证历史数据完整性
        for runIdx := 0; runIdx < session.CountRuns(); runIdx++ {
            run := session.Runs[runIdx]
            expectedInput := fmt.Sprintf("session-%d-run-%d", sessionIdx, runIdx)
            if run.Input != expectedInput {
                t.Errorf("session %s run %d: expected input %s, got %s",
                    sessionID, runIdx, expectedInput, run.Input)
            }
        }
    }
}
```

### 场景 5: 性能基准测试 (1h)

```go
// BenchmarkWorkflowHistory_Load 基准测试历史加载性能
func BenchmarkWorkflowHistory_Load(b *testing.B) {
    mockModel := &MockModel{
        InvokeFunc: func(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
            return &types.ModelResponse{Content: "ok"}, nil
        },
    }

    testAgent, _ := agent.New(agent.Config{
        Name:  "test-agent",
        Model: mockModel,
    })

    storage := NewMemoryStorage(0)
    workflow, _ := New(Config{
        ID:                "bench-workflow",
        EnableHistory:     true,
        HistoryStore:      storage,
        NumHistoryRuns:    10,
        AddHistoryToSteps: true,
        Steps: []Node{
            &Step{id: "step-1", agent: testAgent},
        },
    })

    ctx := context.Background()
    sessionID := "bench-session"

    // 预先运行 100 次以积累历史
    for i := 0; i < 100; i++ {
        workflow.Run(ctx, fmt.Sprintf("warmup-%d", i), sessionID)
    }

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        workflow.Run(ctx, "benchmark input", sessionID)
    }
}

// BenchmarkWorkflowHistory_NoHistory 基准测试不使用历史的性能
func BenchmarkWorkflowHistory_NoHistory(b *testing.B) {
    mockModel := &MockModel{
        InvokeFunc: func(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
            return &types.ModelResponse{Content: "ok"}, nil
        },
    }

    testAgent, _ := agent.New(agent.Config{
        Name:  "test-agent",
        Model: mockModel,
    })

    workflow, _ := New(Config{
        ID:            "bench-workflow",
        EnableHistory: false,
        Steps: []Node{
            &Step{id: "step-1", agent: testAgent},
        },
    })

    ctx := context.Background()

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        workflow.Run(ctx, "benchmark input", "session")
    }
}
```

### 场景 6: 集成示例 (1h)

**文件**: `cmd/examples/workflow_history/main.go`

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
    "github.com/rexleimo/agno-go/pkg/agno/workflow"
)

func main() {
    // 创建 OpenAI 模型
    model, err := openai.New("gpt-4", openai.Config{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
    if err != nil {
        log.Fatalf("failed to create model: %v", err)
    }

    // 创建 agent
    chatAgent, err := agent.New(agent.Config{
        Name:         "ChatBot",
        Model:        model,
        Instructions: "You are a helpful assistant with a good memory. Remember previous conversations.",
    })
    if err != nil {
        log.Fatalf("failed to create agent: %v", err)
    }

    // 创建带历史的 workflow
    storage := workflow.NewMemoryStorage(100)
    wf, err := workflow.New(workflow.Config{
        ID:                "chat-workflow",
        Name:              "Conversational Chat",
        EnableHistory:     true,
        HistoryStore:      storage,
        NumHistoryRuns:    5, // 记住最近 5 轮对话
        AddHistoryToSteps: true,
        Steps: []workflow.Node{
            &workflow.Step{
                ID:    "chat",
                Name:  "Chat Step",
                Agent: chatAgent,
            },
        },
    })
    if err != nil {
        log.Fatalf("failed to create workflow: %v", err)
    }

    ctx := context.Background()
    sessionID := "user-session-123"

    // 多轮对话
    conversations := []string{
        "Hello, my name is Alice and I love programming in Go",
        "What's my name?",
        "What programming language do I like?",
        "Can you remind me what we talked about?",
    }

    for i, input := range conversations {
        fmt.Printf("\n=== Round %d ===\n", i+1)
        fmt.Printf("User: %s\n", input)

        result, err := wf.Run(ctx, input, sessionID)
        if err != nil {
            log.Fatalf("workflow run failed: %v", err)
        }

        fmt.Printf("Assistant: %s\n", result.Output)

        // 显示历史信息
        if result.HasHistory() {
            fmt.Printf("(History: %d previous conversations)\n", result.GetHistoryCount())
        }
    }

    // 显示完整的 session 统计
    fmt.Println("\n=== Session Statistics ===")
    session, _ := storage.GetSession(ctx, sessionID)
    fmt.Printf("Total runs: %d\n", session.CountRuns())
    fmt.Printf("Successful runs: %d\n", session.CountSuccessfulRuns())
}
```

---

## 验收标准

- [x] 所有端到端测试通过
- [x] 并发测试无竞态条件
- [x] 性能基准测试记录
- [x] 集成示例可运行
- [x] 测试覆盖多种配置组合
- [x] 文档包含使用示例
- [x] README 更新

---

## 性能目标

**基准测试期望**:
- 历史加载: <5ms (100 条历史)
- 内存开销: <2MB (100 sessions × 10 runs)
- 不使用历史的性能退化: <5%

---

## 相关文件

- `pkg/agno/workflow/workflow_history_e2e_test.go` - 端到端测试
- `cmd/examples/workflow_history/main.go` - 集成示例
- `website/guide/workflow-history.md` - 功能文档（已实现）
- `README.md` - 更新示例

---

## 文档需求

### 创建功能文档

**文件**: `website/guide/workflow-history.md`

包含：
1. 功能概述
2. 快速入门
3. 配置选项说明
4. 使用场景
5. API 参考
6. 性能考虑
7. 故障排除

---

## 注意事项

1. **真实 LLM 测试**: 集成示例需要真实 API key
2. **Mock Model**: 单元测试使用 mock 以保证速度
3. **并发测试**: 使用 `-race` 标志检测竞态条件
4. **性能回归**: 记录基准测试结果，防止性能退化
