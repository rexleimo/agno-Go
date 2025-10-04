# テスト

Agno-Go開発のための包括的なテストガイド。

---

## 概要

Agno-Goは、コードベース全体で**80.8%のテストカバレッジ**を持つ包括的なテストを通じて高品質を維持しています。このガイドでは、テスト基準、パターン、ベストプラクティスをカバーします。

### テストカバレッジの状況

| パッケージ | カバレッジ | ステータス |
|---------|----------|--------|
| types | 100.0% | ✅ 優秀 |
| memory | 93.1% | ✅ 優秀 |
| team | 92.3% | ✅ 優秀 |
| toolkit | 91.7% | ✅ 優秀 |
| http | 88.9% | ✅ 良好 |
| workflow | 80.4% | ✅ 良好 |
| file | 76.2% | ✅ 良好 |
| calculator | 75.6% | ✅ 良好 |
| agent | 74.7% | ✅ 良好 |
| anthropic | 50.9% | 🟡 改善が必要 |
| openai | 44.6% | 🟡 改善が必要 |
| ollama | 43.8% | 🟡 改善が必要 |

---

## テストの実行

### すべてのテスト

```bash
# カバレッジ付きですべてのテストを実行
make test

# 同等のコマンド:
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
```

### 特定のパッケージ

```bash
# agentパッケージをテスト
go test -v ./pkg/agno/agent/...

# カバレッジ付き
go test -v -cover ./pkg/agno/agent/...
```

### 特定のテスト

```bash
# 特定のテスト関数を実行
go test -v -run TestAgentRun ./pkg/agno/agent/

# パターンに一致するテストを実行
go test -v -run TestAgent.* ./pkg/agno/agent/
```

### カバレッジレポート

```bash
# HTMLカバレッジレポートを生成
make coverage

# ブラウザでcoverage.htmlを開く
# 行ごとのカバレッジを表示
```

---

## テスト基準

### カバレッジ要件

- **コアパッケージ**（agent、team、workflow）: >70% カバレッジ
- **ユーティリティパッケージ**（types、memory、toolkit）: >80% カバレッジ
- **新機能**: テストを含める必要がある
- **バグ修正**: リグレッションテストを含める必要がある

### テスト構造

すべてのパッケージには以下が必要:
- ソースファイルと並んで`*_test.go`ファイル
- すべての公開関数のユニットテスト
- 複雑なワークフローの統合テスト
- パフォーマンスクリティカルなコードのベンチマークテスト

---

## ユニットテストの作成

### 基本ユニットテスト

```go
package agent

import (
    "context"
    "testing"

    "github.com/rexleimo/agno-go/pkg/agno/models"
    "github.com/rexleimo/agno-go/pkg/agno/types"
)

func TestAgentRun(t *testing.T) {
    // モックモデルを作成
    model := &MockModel{
        InvokeFunc: func(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
            return &types.ModelResponse{
                Content: "test response",
            }, nil
        },
    }

    // エージェントを作成
    agent, err := New(Config{
        Name:  "test-agent",
        Model: model,
    })
    if err != nil {
        t.Fatalf("Failed to create agent: %v", err)
    }

    // エージェントを実行
    output, err := agent.Run(context.Background(), "test input")
    if err != nil {
        t.Fatalf("Run failed: %v", err)
    }

    // 出力を検証
    if output.Content != "test response" {
        t.Errorf("Expected 'test response', got '%s'", output.Content)
    }
}
```

### テーブル駆動テスト

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

### エラーハンドリングテスト

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
            errMsg: "", // デフォルトで成功するはず
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

## モッキング

### モックモデル

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

### モックツールキット

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

## ベンチマークテスト

### 基本ベンチマーク

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

### ベンチマークの実行

```bash
# すべてのベンチマークを実行
go test -bench=. ./pkg/agno/agent/

# 特定のベンチマークを実行
go test -bench=BenchmarkAgentCreation ./pkg/agno/agent/

# メモリアロケーション統計付き
go test -bench=. -benchmem ./pkg/agno/agent/

# 精度のために複数回実行
go test -bench=. -benchtime=10s -count=5 ./pkg/agno/agent/
```

### ベンチマーク出力

```
BenchmarkAgentCreation-8    5623174    180.1 ns/op    1184 B/op    14 allocs/op
```

解釈:
- 5,623,174回の反復を実行
- 操作あたり180.1ナノ秒
- 操作あたり1,184バイト割り当て
- 操作あたり14回の割り当て

---

## 統合テスト

### 実際のLLMでのテスト

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

    // エージェントが計算機を使用したことを確認
    if !strings.Contains(output.Content, "425") {
        t.Errorf("Expected answer to contain 425, got: %s", output.Content)
    }
}
```

統合テストを実行:

```bash
# 統合テストのみ実行
go test -tags=integration ./...

# 統合テストをスキップ（デフォルト）
go test ./...
```

---

## テストヘルパー

### 共通テストユーティリティ

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

## 継続的インテグレーション

### GitHub Actions

すべてのプッシュとプルリクエストでテストが自動実行されます:

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

## ベストプラクティス

### 1. テスト命名

```go
// 良い ✅
func TestAgentRun(t *testing.T)
func TestAgentRun_WithTools(t *testing.T)
func TestAgentRun_EmptyInput_ReturnsError(t *testing.T)

// 悪い ❌
func Test1(t *testing.T)
func TestStuff(t *testing.T)
```

### 2. t.Helper()を使用

ヘルパー関数をマーク:

```go
func createAgent(t *testing.T) *Agent {
    t.Helper() // スタックトレースでこの関数をスキップ
    // ...
}
```

### 3. リソースをクリーンアップ

```go
func TestWithTempFile(t *testing.T) {
    f, err := os.CreateTemp("", "test-*")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(f.Name()) // クリーンアップ

    // テストコード...
}
```

### 4. サブテストを使用

```go
func TestAgent(t *testing.T) {
    t.Run("creation", func(t *testing.T) { /* ... */ })
    t.Run("run", func(t *testing.T) { /* ... */ })
    t.Run("clear memory", func(t *testing.T) { /* ... */ })
}
```

### 5. 並列テスト

```go
func TestParallel(t *testing.T) {
    tests := []struct{ /* ... */ }{}

    for _, tt := range tests {
        tt := tt // 範囲変数をキャプチャ
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // 並列実行
            // テストコード...
        })
    }
}
```

---

## トラブルシューティング

### 競合検出器でテストが失敗

```bash
# 競合検出器で実行
go test -race ./...

# 競合が検出された場合、コードを修正
# 競合検出器を無効にしない
```

### カバレッジが更新されない

```bash
# キャッシュをクリアして再実行
go clean -testcache
make test
```

### 遅いテスト

```bash
# 遅いテストを見つける
go test -v ./... | grep PASS

# タイムアウトを追加
go test -timeout 30s ./...
```

---

## 次のステップ

- 設計パターンについては[アーキテクチャ](/advanced/architecture)を確認
- ベンチマークについては[パフォーマンス](/advanced/performance)を確認
- プロダクションセットアップについては[デプロイメント](/advanced/deployment)を確認
- [コントリビューティングガイド](https://github.com/rexleimo/agno-Go/blob/main/CONTRIBUTING.md)を探索

---

## リソース

- [Goテスティング](https://golang.org/pkg/testing/)
- [テーブル駆動テスト](https://go.dev/wiki/TableDrivenTests)
- [テストカバレッジ](https://go.dev/blog/cover)
- [ベンチマーキング](https://pkg.go.dev/testing#hdr-Benchmarks)
