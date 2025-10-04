---
layout: home

hero:
  name: "Agno-Go"
  text: "高性能マルチエージェントフレームワーク"
  tagline: "Python より 16 倍高速 | 180ns インスタンス化 | エージェントあたり 1.2KB メモリ"
  image:
    src: /logo.svg
    alt: Agno-Go
  actions:
    - theme: brand
      text: はじめる
      link: /ja/guide/quick-start
    - theme: alt
      text: GitHub で見る
      link: https://github.com/rexleimo/agno-Go

features:
  - icon: 🚀
    title: 極限のパフォーマンス
    details: エージェントのインスタンス化は約 180ns、Python 版より 16 倍高速。エージェントあたりのメモリフットプリントはわずか 1.2KB で、Go のネイティブ並行性をサポート。

  - icon: 🤖
    title: 本番環境対応
    details: AgentOS HTTP サーバーは、RESTful API、セッション管理、エージェントレジストリ、ヘルスモニタリング、包括的なエラー処理を標準装備。

  - icon: 🧩
    title: 柔軟なアーキテクチャ
    details: Agent(自律型)、Team(4 つの協調モード)、Workflow(5 つの制御プリミティブ)から選択して、マルチエージェントシステムを構築。

  - icon: 🔌
    title: マルチモデル対応
    details: OpenAI(GPT-4)、Anthropic Claude、Ollama(ローカルモデル)、DeepSeek、Google Gemini、ModelScope を標準サポート。

  - icon: 🔧
    title: 拡張可能なツール
    details: 拡張が簡単なツールキットシステムで、計算機、HTTP クライアント、ファイル操作、DuckDuckGo 検索を標準装備。数分でカスタムツールを作成可能。

  - icon: 💾
    title: RAG とナレッジベース
    details: OpenAI 埋め込みによる ChromaDB ベクトルデータベース統合。セマンティック検索とナレッジベースを備えたインテリジェントエージェントを構築。

  - icon: ✅
    title: 十分なテスト
    details: 80.8% のテストカバレッジ、85 以上のテストケース、100% の合格率。信頼できる本番品質のコード。

  - icon: 📦
    title: 簡単なデプロイ
    details: Docker、Docker Compose、Kubernetes マニフェストを同梱。完全なデプロイガイド付きで、数分で任意のクラウドプラットフォームにデプロイ可能。

  - icon: 📚
    title: 完全なドキュメント
    details: OpenAPI 3.0 仕様、デプロイガイド、アーキテクチャドキュメント、パフォーマンスベンチマーク、すべての機能の実用例。
---

## クイック例

わずか数行のコードで、ツール付き AI エージェントを作成:

```go
package main

import (
    "context"
    "fmt"
    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
    "github.com/rexleimo/agno-go/pkg/agno/tools/calculator"
)

func main() {
    // モデルを作成
    model, _ := openai.New("gpt-4o-mini", openai.Config{
        APIKey: "your-api-key",
    })

    // ツール付きエージェントを作成
    ag, _ := agent.New(agent.Config{
        Name:     "数学アシスタント",
        Model:    model,
        Toolkits: []toolkit.Toolkit{calculator.New()},
    })

    // エージェントを実行
    output, _ := ag.Run(context.Background(), "25 * 4 + 15 はいくつ?")
    fmt.Println(output.Content) // 出力: 115
}
```

## パフォーマンス比較

| 指標 | Python Agno | Agno-Go | 改善 |
|--------|-------------|---------|-------------|
| エージェント作成 | ~3μs | ~180ns | **16 倍高速** |
| メモリ/エージェント | ~6.5KB | ~1.2KB | **5.4 倍削減** |
| 並行性 | GIL 制限 | ネイティブ goroutine | **無制限** |

## なぜ Agno-Go?

### 本番環境向けに構築

Agno-Go は単なるフレームワークではなく、完全な本番システムです。付属の **AgentOS** サーバーは以下を提供:

- OpenAPI 3.0 仕様の RESTful API
- マルチターン会話のセッション管理
- スレッドセーフなエージェントレジストリ
- ヘルスモニタリングと構造化ロギング
- CORS サポートとリクエストタイムアウト処理

### KISS 原則

**Keep It Simple, Stupid** の哲学に従う:

- **3 つのコア LLM プロバイダー**(45+ ではない) - OpenAI、Anthropic、Ollama
- **必須ツール**(115+ ではない) - 計算機、HTTP、ファイル、検索
- **量より質** - 本番環境対応機能に焦点

### 開発者体験

- **型安全**: Go の強い型付けでコンパイル時にエラーを検出
- **高速ビルド**: Go のコンパイル速度で迅速な反復開発
- **簡単なデプロイ**: ランタイム依存なしの単一バイナリ
- **優れたツール**: 組み込みのテスト、プロファイリング、競合検出

## 5 分でスタート

```bash
# リポジトリをクローン
git clone https://github.com/rexleimo/agno-Go.git
cd agno-Go

# API キーを設定
export OPENAI_API_KEY=sk-your-key-here

# サンプルを実行
go run cmd/examples/simple_agent/main.go

# または AgentOS サーバーを起動
docker-compose up -d
curl http://localhost:8080/health
```

## 含まれるもの

- **コアフレームワーク**: Agent、Team(4 モード)、Workflow(5 プリミティブ)
- **モデル**: OpenAI、Anthropic Claude、Ollama、DeepSeek、Gemini、ModelScope
- **ツール**: Calculator(75.6%)、HTTP(88.9%)、File(76.2%)、Search(92.1%)
- **RAG**: ChromaDB 統合 + OpenAI 埋め込み
- **AgentOS**: 本番環境向け HTTP サーバー(65.0% カバレッジ)
- **サンプル**: すべての機能をカバーする 6 つの実用例
- **ドキュメント**: 完全ガイド、API リファレンス、デプロイ手順

## コミュニティ

- **GitHub**: [rexleimo/agno-Go](https://github.com/rexleimo/agno-Go)
- **Issues**: [バグ報告と機能リクエスト](https://github.com/rexleimo/agno-Go/issues)
- **Discussions**: [質問とアイデア共有](https://github.com/rexleimo/agno-Go/discussions)

## ライセンス

Agno-Go は [MIT ライセンス](https://github.com/rexleimo/agno-Go/blob/main/LICENSE) でリリースされています。

[Agno (Python)](https://github.com/agno-agi/agno) フレームワークからインスピレーションを得ています。
