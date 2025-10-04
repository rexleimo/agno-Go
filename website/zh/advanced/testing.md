# 测试 / Testing

Agno-Go 开发的综合测试指南。

---

## 概述 / Overview

Agno-Go 通过全面的测试保持高质量,整个代码库的**测试覆盖率达 80.8%**。本指南涵盖测试标准、模式和最佳实践。

### 测试覆盖率状态 / Test Coverage Status

| 包 / Package | 覆盖率 / Coverage | 状态 / Status |
|---------|----------|--------|
| types | 100.0% | ✅ 优秀 / Excellent |
| memory | 93.1% | ✅ 优秀 / Excellent |
| team | 92.3% | ✅ 优秀 / Excellent |
| toolkit | 91.7% | ✅ 优秀 / Excellent |
| http | 88.9% | ✅ 良好 / Good |
| workflow | 80.4% | ✅ 良好 / Good |
| file | 76.2% | ✅ 良好 / Good |
| calculator | 75.6% | ✅ 良好 / Good |
| agent | 74.7% | ✅ 良好 / Good |
| anthropic | 50.9% | 🟡 需要改进 / Needs improvement |
| openai | 44.6% | 🟡 需要改进 / Needs improvement |
| ollama | 43.8% | 🟡 需要改进 / Needs improvement |

---

## 运行测试 / Running Tests

### 所有测试 / All Tests

```bash
# 运行所有测试(带覆盖率) / Run all tests with coverage
make test

# 等效于 / Equivalent to:
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
```

### 特定包 / Specific Package

```bash
# 测试 agent 包 / Test agent package
go test -v ./pkg/agno/agent/...

# 带覆盖率测试 / Test with coverage
go test -v -cover ./pkg/agno/agent/...
```

### 特定测试 / Specific Test

```bash
# 运行特定测试函数 / Run specific test function
go test -v -run TestAgentRun ./pkg/agno/agent/

# 运行匹配模式的测试 / Run tests matching pattern
go test -v -run TestAgent.* ./pkg/agno/agent/
```

### 覆盖率报告 / Coverage Report

```bash
# 生成 HTML 覆盖率报告 / Generate HTML coverage report
make coverage

# 在浏览器中打开 coverage.html
# 显示逐行覆盖率
```

---

## 测试标准 / Testing Standards

### 覆盖率要求 / Coverage Requirements

- **核心包 / Core packages** (agent, team, workflow): >70% 覆盖率
- **工具包 / Utility packages** (types, memory, toolkit): >80% 覆盖率
- **新功能 / New features**: 必须包含测试
- **错误修复 / Bug fixes**: 必须包含回归测试

### 测试结构 / Test Structure

每个包应该有:
- `*_test.go` 文件与源文件并列
- 所有公共函数的单元测试
- 复杂工作流的集成测试
- 性能关键代码的基准测试

---

## 编写单元测试 / Writing Unit Tests

### 基本单元测试 / Basic Unit Test

```go
package agent

import (
    "context"
    "testing"

    "github.com/rexleimo/agno-go/pkg/agno/models"
    "github.com/rexleimo/agno-go/pkg/agno/types"
)

func TestAgentRun(t *testing.T) {
    // 创建模拟模型 / Create mock model
    model := &MockModel{
        InvokeFunc: func(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
            return &types.ModelResponse{
                Content: "test response",
            }, nil
        },
    }

    // 创建 agent / Create agent
    agent, err := New(Config{
        Name:  "test-agent",
        Model: model,
    })
    if err != nil {
        t.Fatalf("Failed to create agent: %v", err)
    }

    // 运行 agent / Run agent
    output, err := agent.Run(context.Background(), "test input")
    if err != nil {
        t.Fatalf("Run failed: %v", err)
    }

    // 验证输出 / Verify output
    if output.Content != "test response" {
        t.Errorf("Expected 'test response', got '%s'", output.Content)
    }
}
```

### 表驱动测试 / Table-Driven Tests

```go
func TestCalculatorAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     float64
        expected float64
    }{
        {"positive numbers", 5.0, 3.0, 8.0},
        {"negative numbers", -5.0, -3.0, -8.0},
        {"mixed signs", 5.0, -3.0, 2.0},
        {"with zero", 5.0, 0.0, 5.0},
        {"decimals", 1.5, 2.3, 3.8},
    }

    calc := New()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := calc.add(map[string]interface{}{
                "a": tt.a,
                "b": tt.b,
            })
            if err != nil {
                t.Fatalf("add failed: %v", err)
            }

            if result != tt.expected {
                t.Errorf("Expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

### 错误处理测试 / Error Handling Tests

```go
func TestAgentInvalidConfig(t *testing.T) {
    tests := []struct {
        name   string
        config Config
        errMsg string
    }{
        {
            name:   "nil model",
            config: Config{Name: "test"},
            errMsg: "model is required",
        },
        {
            name:   "empty name with nil memory",
            config: Config{Model: &MockModel{}, Memory: nil},
            errMsg: "", // Should succeed with defaults
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := New(tt.config)
            if tt.errMsg == "" {
                if err != nil {
                    t.Errorf("Expected no error, got: %v", err)
                }
            } else {
                if err == nil {
                    t.Error("Expected error, got nil")
                } else if !strings.Contains(err.Error(), tt.errMsg) {
                    t.Errorf("Expected error containing '%s', got: %v", tt.errMsg, err)
                }
            }
        })
    }
}
```

---

## 模拟 / Mocking

### Mock Model

```go
type MockModel struct {
    InvokeFunc       func(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error)
    InvokeStreamFunc func(ctx context.Context, req *models.InvokeRequest) (<-chan types.ResponseChunk, error)
}

func (m *MockModel) Invoke(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
    if m.InvokeFunc != nil {
        return m.InvokeFunc(ctx, req)
    }
    return &types.ModelResponse{Content: "mock response"}, nil
}

func (m *MockModel) InvokeStream(ctx context.Context, req *models.InvokeRequest) (<-chan types.ResponseChunk, error) {
    if m.InvokeStreamFunc != nil {
        return m.InvokeStreamFunc(ctx, req)
    }
    return nil, nil
}

func (m *MockModel) GetProvider() string { return "mock" }
func (m *MockModel) GetID() string       { return "mock-model" }
```

### Mock Toolkit

```go
type MockToolkit struct {
    *toolkit.BaseToolkit
    callCount int
}

func NewMockToolkit() *MockToolkit {
    t := &MockToolkit{
        BaseToolkit: toolkit.NewBaseToolkit("mock"),
    }

    t.RegisterFunction(&toolkit.Function{
        Name:        "mock_function",
        Description: "Mock function for testing",
        Handler:     t.mockHandler,
    })

    return t
}

func (t *MockToolkit) mockHandler(args map[string]interface{}) (interface{}, error) {
    t.callCount++
    return "mock result", nil
}
```

---

## 基准测试 / Benchmark Tests

### 基本基准测试 / Basic Benchmark

```go
func BenchmarkAgentCreation(b *testing.B) {
    model := &MockModel{}

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        _, err := New(Config{
            Name:  "benchmark-agent",
            Model: model,
        })
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### 运行基准测试 / Running Benchmarks

```bash
# 运行所有基准测试 / Run all benchmarks
go test -bench=. ./pkg/agno/agent/

# 运行特定基准测试 / Run specific benchmark
go test -bench=BenchmarkAgentCreation ./pkg/agno/agent/

# 带内存分配统计 / With memory allocation stats
go test -bench=. -benchmem ./pkg/agno/agent/

# 多次运行以提高准确性 / Multiple runs for accuracy
go test -bench=. -benchtime=10s -count=5 ./pkg/agno/agent/
```

### 基准测试输出 / Benchmark Output

```
BenchmarkAgentCreation-8    5623174    180.1 ns/op    1184 B/op    14 allocs/op
```

解释 / Interpretation:
- 运行了 5,623,174 次迭代
- 每次操作 180.1 纳秒
- 每次操作分配 1,184 字节
- 每次操作 14 次分配

---

## 集成测试 / Integration Tests

### 使用真实 LLM 测试 / Testing with Real LLMs

```go
// +build integration

func TestAgentWithRealOpenAI(t *testing.T) {
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        t.Skip("OPENAI_API_KEY not set")
    }

    model, err := openai.New("gpt-4o-mini", openai.Config{
        APIKey: apiKey,
    })
    if err != nil {
        t.Fatal(err)
    }

    agent, _ := New(Config{
        Model:    model,
        Toolkits: []toolkit.Toolkit{calculator.New()},
    })

    output, err := agent.Run(context.Background(), "What is 25 * 17?")
    if err != nil {
        t.Fatal(err)
    }

    // 验证 agent 使用了计算器 / Verify agent used calculator
    if !strings.Contains(output.Content, "425") {
        t.Errorf("Expected answer to contain 425, got: %s", output.Content)
    }
}
```

运行集成测试 / Run integration tests:

```bash
# 仅运行集成测试 / Run only integration tests
go test -tags=integration ./...

# 跳过集成测试(默认) / Skip integration tests (default)
go test ./...
```

---

## 测试辅助工具 / Test Helpers

### 通用测试工具 / Common Test Utilities

```go
// test_helpers.go

package testutil

import (
    "testing"
    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models"
)

func CreateTestAgent(t *testing.T) *agent.Agent {
    t.Helper()

    model := &MockModel{}
    ag, err := agent.New(agent.Config{
        Name:  "test-agent",
        Model: model,
    })
    if err != nil {
        t.Fatalf("Failed to create test agent: %v", err)
    }

    return ag
}

func AssertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
}

func AssertError(t *testing.T, err error, expectedMsg string) {
    t.Helper()
    if err == nil {
        t.Fatal("Expected error, got nil")
    }
    if !strings.Contains(err.Error(), expectedMsg) {
        t.Fatalf("Expected error containing '%s', got: %v", expectedMsg, err)
    }
}
```

---

## 持续集成 / Continuous Integration

### GitHub Actions

测试在每次推送和拉取请求时自动运行:

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: go test -v -race -coverprofile=coverage.out ./...
      - uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

---

## 最佳实践 / Best Practices

### 1. 测试命名 / Test Naming

```go
// 好 / Good ✅
func TestAgentRun(t *testing.T)
func TestAgentRun_WithTools(t *testing.T)
func TestAgentRun_EmptyInput_ReturnsError(t *testing.T)

// 差 / Bad ❌
func Test1(t *testing.T)
func TestStuff(t *testing.T)
```

### 2. 使用 t.Helper()

标记辅助函数:

```go
func createAgent(t *testing.T) *Agent {
    t.Helper() // Stack traces skip this function
    // ...
}
```

### 3. 清理资源 / Clean Up Resources

```go
func TestWithTempFile(t *testing.T) {
    f, err := os.CreateTemp("", "test-*")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(f.Name()) // Clean up

    // Test code...
}
```

### 4. 使用子测试 / Use Subtests

```go
func TestAgent(t *testing.T) {
    t.Run("creation", func(t *testing.T) { /* ... */ })
    t.Run("run", func(t *testing.T) { /* ... */ })
    t.Run("clear memory", func(t *testing.T) { /* ... */ })
}
```

### 5. 并行测试 / Parallel Tests

```go
func TestParallel(t *testing.T) {
    tests := []struct{ /* ... */ }{}

    for _, tt := range tests {
        tt := tt // Capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // Run in parallel
            // Test code...
        })
    }
}
```

---

## 故障排查 / Troubleshooting

### 竞态检测器测试失败 / Tests Failing with Race Detector

```bash
# 使用竞态检测器运行 / Run with race detector
go test -race ./...

# 如果检测到竞态,修复代码
# 不要禁用竞态检测器
```

### 覆盖率未更新 / Coverage Not Updating

```bash
# 清除缓存并重新运行 / Clean cache and rerun
go clean -testcache
make test
```

### 测试缓慢 / Slow Tests

```bash
# 查找慢测试 / Find slow tests
go test -v ./... | grep PASS

# 添加超时 / Add timeout
go test -timeout 30s ./...
```

---

## 下一步 / Next Steps

- 查看[架构 / Architecture](/advanced/architecture)了解设计模式
- 检查[性能 / Performance](/advanced/performance)了解基准测试
- 阅读[部署 / Deployment](/advanced/deployment)了解生产设置
- 探索[贡献指南 / Contributing Guide](https://github.com/rexleimo/agno-Go/blob/main/CONTRIBUTING.md)

---

## 资源 / Resources

- [Go Testing](https://golang.org/pkg/testing/)
- [Table Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [Test Coverage](https://go.dev/blog/cover)
- [Benchmarking](https://pkg.go.dev/testing#hdr-Benchmarks)
