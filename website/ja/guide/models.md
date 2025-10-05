# Models - LLMプロバイダー

Agno-Goは統一されたインターフェースで複数のLLMプロバイダーをサポートしています。

---

## サポートされているモデル

### OpenAI
- GPT-4o、GPT-4o-mini、GPT-4 Turbo、GPT-3.5 Turbo
- 完全なストリーミングサポート
- 関数呼び出し

### Anthropic Claude
- Claude 3.5 Sonnet、Claude 3 Opus、Claude 3 Sonnet、Claude 3 Haiku
- ストリーミングサポート
- ツール使用

### GLM (智谱AI) ⭐ v1.0.2で追加
- GLM-4、GLM-4V（ビジョン）、GLM-3-Turbo
- 中国語に最適化
- カスタムJWT認証
- 関数呼び出しサポート

### Ollama
- ローカルでモデルを実行（Llama、Mistral等）
- プライバシー重視
- APIコストなし

---

## OpenAI

### セットアップ

```go
import "github.com/rexleimo/agno-go/pkg/agno/models/openai"

model, err := openai.New("gpt-4o-mini", openai.Config{
    APIKey:      os.Getenv("OPENAI_API_KEY"),
    Temperature: 0.7,
    MaxTokens:   1000,
})
```

### 設定

```go
type Config struct {
    APIKey      string  // 必須: OpenAI APIキー
    BaseURL     string  // オプション: カスタムエンドポイント（デフォルト: https://api.openai.com/v1）
    Temperature float64 // オプション: 0.0-2.0（デフォルト: 0.7）
    MaxTokens   int     // オプション: 最大応答トークン数
}
```

### サポートされているモデル

| モデル | コンテキスト | 最適な用途 |
|-------|---------|----------|
| `gpt-4o` | 128K | 最も高性能、マルチモーダル |
| `gpt-4o-mini` | 128K | 高速、コスト効率的 |
| `gpt-4-turbo` | 128K | 高度な推論 |
| `gpt-3.5-turbo` | 16K | シンプルなタスク、高速 |

### 例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
)

func main() {
    model, err := openai.New("gpt-4o-mini", openai.Config{
        APIKey:      os.Getenv("OPENAI_API_KEY"),
        Temperature: 0.7,
    })
    if err != nil {
        log.Fatal(err)
    }

    agent, _ := agent.New(agent.Config{
        Name:  "Assistant",
        Model: model,
    })

    output, _ := agent.Run(context.Background(), "Hello!")
    fmt.Println(output.Content)
}
```

---

## Anthropic Claude

### セットアップ

```go
import "github.com/rexleimo/agno-go/pkg/agno/models/anthropic"

model, err := anthropic.New("claude-3-5-sonnet-20241022", anthropic.Config{
    APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
    MaxTokens: 2048,
})
```

### 設定

```go
type Config struct {
    APIKey      string  // 必須: Anthropic APIキー
    Temperature float64 // オプション: 0.0-1.0
    MaxTokens   int     // オプション: 最大応答トークン数（デフォルト: 4096）
}
```

### サポートされているモデル

| モデル | コンテキスト | 最適な用途 |
|-------|---------|----------|
| `claude-3-5-sonnet-20241022` | 200K | 最も高性能、コーディング |
| `claude-3-opus-20240229` | 200K | 複雑なタスク |
| `claude-3-sonnet-20240229` | 200K | バランスの取れたパフォーマンス |
| `claude-3-haiku-20240307` | 200K | 高速応答 |

### 例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/anthropic"
)

func main() {
    model, err := anthropic.New("claude-3-5-sonnet-20241022", anthropic.Config{
        APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
        MaxTokens: 2048,
    })
    if err != nil {
        log.Fatal(err)
    }

    agent, _ := agent.New(agent.Config{
        Name:         "Claude",
        Model:        model,
        Instructions: "You are a helpful assistant.",
    })

    output, _ := agent.Run(context.Background(), "Explain quantum computing")
    fmt.Println(output.Content)
}
```

---

## GLM (智谱AI)

### セットアップ

```go
import "github.com/rexleimo/agno-go/pkg/agno/models/glm"

model, err := glm.New("glm-4", glm.Config{
    APIKey:      os.Getenv("ZHIPUAI_API_KEY"),  // 形式: {key_id}.{key_secret}
    Temperature: 0.7,
    MaxTokens:   1024,
})
```

### 設定

```go
type Config struct {
    APIKey      string  // 必須: APIキー、形式は {key_id}.{key_secret}
    BaseURL     string  // オプション: カスタムエンドポイント（デフォルト: https://open.bigmodel.cn/api/paas/v4）
    Temperature float64 // オプション: 0.0-1.0
    MaxTokens   int     // オプション: 最大応答トークン数
    TopP        float64 // オプション: Top-pサンプリングパラメータ
    DoSample    bool    // オプション: サンプリングを使用するか
}
```

### サポートされているモデル

| モデル | コンテキスト | 最適な用途 |
|-------|---------|----------|
| `glm-4` | 128K | 一般的な会話、中国語 |
| `glm-4v` | 128K | ビジョンタスク、マルチモーダル |
| `glm-3-turbo` | 128K | 高速応答、コスト最適化 |

### APIキー形式

GLMは特別なAPIキー形式を使用し、2つの部分で構成されています：

```
{key_id}.{key_secret}
```

APIキーの取得先: https://open.bigmodel.cn/

### 例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/glm"
    "github.com/rexleimo/agno-go/pkg/agno/tools/calculator"
    "github.com/rexleimo/agno-go/pkg/agno/tools/toolkit"
)

func main() {
    model, err := glm.New("glm-4", glm.Config{
        APIKey:      os.Getenv("ZHIPUAI_API_KEY"),
        Temperature: 0.7,
        MaxTokens:   1024,
    })
    if err != nil {
        log.Fatal(err)
    }

    agent, _ := agent.New(agent.Config{
        Name:         "GLM 助手",
        Model:        model,
        Toolkits:     []toolkit.Toolkit{calculator.New()},
        Instructions: "你是一个有用的 AI 助手。",
    })

    // 中国語サポート
    output, _ := agent.Run(context.Background(), "你好！请计算 123 * 456")
    fmt.Println(output.Content)
}
```

### 認証

GLMはJWT（JSON Web Token）認証を使用します：

1. APIキーが`key_id`と`key_secret`に解析されます
2. HMAC-SHA256署名を使用してJWTトークンを生成します
3. トークンの有効期限は7日間です
4. リクエストごとに自動的に再生成されます

これらはすべてSDKによって自動的に処理されます。

---

## Ollama（ローカルモデル）

### セットアップ

1. Ollamaをインストール: https://ollama.ai
2. モデルをプル: `ollama pull llama2`
3. Agno-Goで使用:

```go
import "github.com/rexleimo/agno-go/pkg/agno/models/ollama"

model, err := ollama.New("llama2", ollama.Config{
    BaseURL: "http://localhost:11434",  // Ollamaサーバー
})
```

### 設定

```go
type Config struct {
    BaseURL     string  // オプション: OllamaサーバーURL（デフォルト: http://localhost:11434）
    Temperature float64 // オプション: 0.0-1.0
}
```

### サポートされているモデル

Ollamaで利用可能な任意のモデル:
- `llama2`、`llama3`、`llama3.1`
- `mistral`、`mixtral`
- `codellama`、`deepseek-coder`
- `qwen2`、`gemma2`

### 例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/ollama"
)

func main() {
    // Ollamaが実行中でモデルがプルされていることを確認
    model, err := ollama.New("llama2", ollama.Config{
        BaseURL: "http://localhost:11434",
    })
    if err != nil {
        log.Fatal(err)
    }

    agent, _ := agent.New(agent.Config{
        Name:  "Local Assistant",
        Model: model,
    })

    output, _ := agent.Run(context.Background(), "What is Go?")
    fmt.Println(output.Content)
}
```

---

## モデル比較

### パフォーマンス

| プロバイダー | 速度 | コスト | プライバシー | コンテキスト |
|----------|-------|------|---------|---------|
| OpenAI GPT-4o-mini | ⚡⚡⚡ | 💰 | ☁️ クラウド | 128K |
| OpenAI GPT-4o | ⚡⚡ | 💰💰💰 | ☁️ クラウド | 128K |
| Anthropic Claude | ⚡⚡ | 💰💰 | ☁️ クラウド | 200K |
| GLM-4 | ⚡⚡⚡ | 💰 | ☁️ クラウド | 128K |
| Ollama | ⚡ | 🆓 無料 | 🏠 ローカル | 可変 |

### それぞれをいつ使用するか

**OpenAI GPT-4o-mini**
- 開発とテスト
- 大量アプリケーション
- コストに敏感なユースケース

**OpenAI GPT-4o**
- 複雑な推論タスク
- マルチモーダルアプリケーション
- プロダクションシステム

**Anthropic Claude**
- 長いコンテキストのニーズ（200Kトークン）
- コーディング支援
- 複雑な分析

**GLM-4**
- 中国語アプリケーション
- 中国国内での展開要件
- 高速応答と高品質
- 中国ユーザー向けコスト最適化

**Ollama**
- プライバシー要件
- インターネット接続なし
- APIコストゼロ
- 開発/テスト

---

## モデルの切り替え

モデル間の切り替えは簡単です:

```go
// OpenAI
openaiModel, _ := openai.New("gpt-4o-mini", openai.Config{
    APIKey: os.Getenv("OPENAI_API_KEY"),
})

// Claude
claudeModel, _ := anthropic.New("claude-3-5-sonnet-20241022", anthropic.Config{
    APIKey: os.Getenv("ANTHROPIC_API_KEY"),
})

// GLM
glmModel, _ := glm.New("glm-4", glm.Config{
    APIKey: os.Getenv("ZHIPUAI_API_KEY"),
})

// Ollama
ollamaModel, _ := ollama.New("llama2", ollama.Config{})

// 同じAgentコードを使用
agent, _ := agent.New(agent.Config{
    Model: openaiModel,  // または claudeModel、glmModel、ollamaModel
})
```

---

## 高度な設定

### Temperature

ランダム性を制御（0.0 = 決定論的、1.0+ = 創造的）:

```go
model, _ := openai.New("gpt-4o-mini", openai.Config{
    Temperature: 0.0,  // 一貫した応答
})

model, _ := openai.New("gpt-4o-mini", openai.Config{
    Temperature: 1.5,  // 創造的な応答
})
```

### Max Tokens

応答の長さを制限:

```go
model, _ := openai.New("gpt-4o-mini", openai.Config{
    MaxTokens: 500,  // 短い応答
})
```

### カスタムエンドポイント

互換性のあるAPIを使用:

```go
model, _ := openai.New("gpt-4o-mini", openai.Config{
    BaseURL: "https://your-proxy.com/v1",  // カスタムエンドポイント
    APIKey:  "your-key",
})
```

---

## ベストプラクティス

### 1. 環境変数

APIキーを安全に保存:

```go
// 良い例 ✅
APIKey: os.Getenv("OPENAI_API_KEY")

// 悪い例 ❌
APIKey: "sk-proj-..." // ハードコード
```

### 2. エラー処理

常にエラーをチェック:

```go
model, err := openai.New("gpt-4o-mini", openai.Config{
    APIKey: os.Getenv("OPENAI_API_KEY"),
})
if err != nil {
    log.Fatalf("Failed to create model: %v", err)
}
```

### 3. モデル選択

ニーズに基づいて選択:

```go
// 開発: 高速で安価
devModel, _ := openai.New("gpt-4o-mini", config)

// プロダクション: より高性能
prodModel, _ := openai.New("gpt-4o", config)
```

### 4. コンテキスト管理

コンテキスト制限に注意:

```go
// 長い会話の場合、定期的にメモリをクリア
if messageCount > 50 {
    agent.ClearMemory()
}
```

---

## 環境設定

`.env`ファイルを作成:

```bash
# OpenAI
OPENAI_API_KEY=sk-proj-...

# Anthropic
ANTHROPIC_API_KEY=sk-ant-...

# GLM (智谱AI) - 形式: {key_id}.{key_secret}
ZHIPUAI_API_KEY=your-key-id.your-key-secret

# Ollama（オプション、デフォルトはlocalhost）
OLLAMA_BASE_URL=http://localhost:11434
```

コードで読み込む:

```go
import "github.com/joho/godotenv"

func init() {
    godotenv.Load()
}
```

---

## 次のステップ

- モデル機能を拡張するには[Tools](/guide/tools)を追加
- 会話履歴については[Memory](/guide/memory)を参照
- 混合モデルで[Teams](/guide/team)を構築
- 実際の使用法については[Examples](/examples/)を参照

---

## 関連例

- [Simple Agent](/examples/simple-agent) - OpenAIの例
- [Claude Agent](/examples/claude-agent) - Anthropicの例
- [GLM Agent](/examples/glm-agent) - GLM (智谱AI)の例
- [Ollama Agent](/examples/ollama-agent) - ローカルモデルの例
