# Agent APIリファレンス

## agent.New

新しいエージェントインスタンスを作成します。

**シグネチャ:**
```go
func New(config Config) (*Agent, error)
```

**パラメータ:**

```go
type Config struct {
    // 必須
    Model models.Model // 使用するLLMモデル

    // オプション
    Name         string            // エージェント名 (デフォルト: "Agent")
    Toolkits     []toolkit.Toolkit // 利用可能なツール
    Memory       memory.Memory     // 会話メモリ
    Instructions string            // システム指示
    MaxLoops     int               // 最大ツール呼び出しループ回数 (デフォルト: 10)
}
```

**戻り値:**
- `*Agent`: 作成されたエージェントインスタンス
- `error`: モデルがnilまたは設定が無効な場合のエラー

**例:**
```go
model, _ := openai.New("gpt-4", openai.Config{APIKey: apiKey})

ag, err := agent.New(agent.Config{
    Name:         "Assistant",
    Model:        model,
    Toolkits:     []toolkit.Toolkit{calculator.New()},
    Instructions: "You are a helpful assistant.",
    MaxLoops:     15,
})
```

## Agent.Run

入力でエージェントを実行します。

**シグネチャ:**
```go
func (a *Agent) Run(ctx context.Context, input string) (*RunOutput, error)
```

**パラメータ:**
- `ctx`: キャンセル/タイムアウト用のContext
- `input`: ユーザー入力文字列

**戻り値:**
```go
type RunOutput struct {
    Content  string                 // エージェントの応答
    Metadata map[string]interface{} // 追加のメタデータ
}
```

**エラー:**
- `InvalidInputError`: 入力が空
- `ModelTimeoutError`: LLMリクエストのタイムアウト
- `ToolExecutionError`: ツール実行の失敗
- `APIError`: LLM APIエラー

**例:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

output, err := ag.Run(ctx, "What is 2+2?")
if err != nil {
    log.Fatal(err)
}
fmt.Println(output.Content)
```

## Agent.ClearMemory

会話メモリをクリアします。

**シグネチャ:**
```go
func (a *Agent) ClearMemory()
```

**例:**
```go
ag.ClearMemory() // 新しい会話を開始
```
