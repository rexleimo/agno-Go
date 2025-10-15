---
title: リリースノート
description: Agno-Goのバージョン履歴とリリースノート
outline: deep
---

# リリースノート

## バージョン 1.2.1 (2025-10-15)

### 🧭 ドキュメント再編

- 明確に分離：
  - `website/` → 実装済みの対外ドキュメント（VitePress サイト）
  - `docs/` → 設計ドラフト、移行計画、タスク、開発者/内部ドキュメント
- `docs/README.md` を追加（方針と入口を記載）
- 貢献者向けに `CONTRIBUTING.md` を追加

### 🔗 リンク修正

- README、CLAUDE、CHANGELOG、リリースノートのリンクを `website/advanced/*` と `website/guide/*` に統一
- `docs/` 配下の重複実装ドキュメントへの旧リンクを削除

### 🌐 サイト更新

- API：AgentOS ページにナレッジ API を追記（/api/agentos）
- Workflow History と Performance ページを正規参照に

### ✅ 動作変更

- なし（ドキュメントと構成のみ更新）

## バージョン 1.1.0 (2025-10-08)

### 🎉 ハイライト

本リリースは、本番環境対応のマルチエージェントシステムのための強力な新機能をもたらします：

- **A2Aインターフェース** - 標準化されたエージェント間通信プロトコル
- **セッション状態管理** - ワークフローステップ間の永続的な状態
- **マルチテナントサポート** - 単一エージェントインスタンスで複数のユーザーにサービスを提供
- **モデルタイムアウト設定** - LLM呼び出しのための細かいタイムアウト制御

---

### ✨ 新機能

#### A2A (Agent-to-Agent) インターフェース

JSON-RPC 2.0に基づくエージェント間相互作用の標準化された通信プロトコル。

**主な機能:**
- RESTful APIエンドポイント（`/a2a/message/send`, `/a2a/message/stream`）
- マルチメディアサポート（テキスト、画像、ファイル、JSONデータ）
- ストリーミング用Server-Sent Events (SSE)
- Python Agno A2A実装と互換性あり

**クイック例:**
```go
import "github.com/rexleimo/agno-go/pkg/agentos/a2a"

// A2Aインターフェースを作成
a2a := a2a.New(a2a.Config{
    Agents: []a2a.Entity{myAgent},
    Prefix: "/a2a",
})

// ルートを登録（Gin）
router := gin.Default()
a2a.RegisterRoutes(router)
```

📚 **詳細を学ぶ:** [A2Aインターフェースドキュメント](/ja/api/a2a)

---

#### ワークフローセッション状態管理

ワークフローステップ間で状態を維持するためのスレッドセーフなセッション管理。

**主な機能:**
- ステップ間の永続的な状態ストレージ
- `sync.RWMutex`によるスレッドセーフ
- 並列ブランチ分離のためのディープコピー
- データ損失を防ぐスマートマージ戦略
- Python Agno v2.1.2の競合状態を修正

**クイック例:**
```go
// セッション情報付きコンテキストを作成
execCtx := workflow.NewExecutionContextWithSession(
    "input",
    "session-123",  // セッションID
    "user-a",       // ユーザーID
)

// セッション状態にアクセス
execCtx.SetSessionState("key", "value")
value, _ := execCtx.GetSessionState("key")
```

📚 **詳細を学ぶ:** [セッション状態ドキュメント](/ja/guide/session-state)

---

#### マルチテナントサポート

単一のAgentインスタンスで複数のユーザーにサービスを提供し、完全なデータ分離を保証。

**主な機能:**
- ユーザー分離された会話履歴
- Memoryインターフェースのオプション`userID`パラメータ
- 既存コードとの下位互換性
- スレッドセーフな並行操作
- クリーンアップ用`ClearAll()`メソッド

**クイック例:**
```go
// マルチテナントエージェントを作成
agent, _ := agent.New(&agent.Config{
    Name:   "customer-service",
    Model:  model,
    Memory: memory.NewInMemory(100),
})

// ユーザーAの会話
agent.UserID = "user-a"
output, _ := agent.Run(ctx, "My name is Alice")

// ユーザーBの会話
agent.UserID = "user-b"
output, _ := agent.Run(ctx, "My name is Bob")
```

📚 **詳細を学ぶ:** [マルチテナントドキュメント](/ja/advanced/multi-tenant)

---

#### モデルタイムアウト設定

細かい制御でLLM呼び出しのリクエストタイムアウトを設定。

**主な機能:**
- デフォルト: 60秒
- 範囲: 1秒から10分
- サポートモデル: OpenAI、Anthropic Claude
- コンテキストを考慮したタイムアウト処理

**クイック例:**
```go
// カスタムタイムアウト付きOpenAI
model, _ := openai.New("gpt-4", openai.Config{
    APIKey:  apiKey,
    Timeout: 30 * time.Second,
})

// カスタムタイムアウト付きClaude
claude, _ := anthropic.New("claude-3-opus", anthropic.Config{
    APIKey:  apiKey,
    Timeout: 45 * time.Second,
})
```

📚 **詳細を学ぶ:** [モデル設定](/ja/guide/models#timeout-configuration)

---

### 🐛 バグ修正

- **ワークフロー競合状態** - 並列ステップ実行のデータ競合を修正
  - Python Agno v2.1.2には共有`session_state` dictによる上書きの問題がありました
  - Go実装はブランチごとに独立したSessionStateクローンを使用
  - スマートマージ戦略により並行実行でのデータ損失を防止

---

### 📚 ドキュメント

すべての新機能には包括的なバイリンガルドキュメント（英語/中文）が含まれています：

- [A2Aインターフェースガイド](/ja/api/a2a) - 完全なプロトコル仕様
- [セッション状態ガイド](/ja/guide/session-state) - ワークフロー状態管理
- [マルチテナントガイド](/ja/advanced/multi-tenant) - データ分離パターン
- [モデル設定](/ja/guide/models#timeout-configuration) - タイムアウト設定

---

### 🧪 テスト

**新しいテストスイート:**
- `session_state_test.go` - セッション状態テスト543行
- `memory_test.go` - マルチテナントメモリテスト（新しいテストケース4つ）
- `agent_test.go` - マルチテナントエージェントテスト
- `openai_test.go` - タイムアウト設定テスト
- `anthropic_test.go` - タイムアウト設定テスト

**テスト結果:**
- ✅ すべてのテストが`-race`検出器で合格
- ✅ ワークフローカバレッジ: 79.4%
- ✅ メモリカバレッジ: 93.1%
- ✅ エージェントカバレッジ: 74.7%

---

### 📊 パフォーマンス

**パフォーマンス低下なし** - すべてのベンチマークが一貫しています：
- Agent実例化: ~180ns/op（Pythonより16倍高速）
- メモリフットプリント: ~1.2KB/エージェント
- スレッドセーフな並行操作

---

### ⚠️ 破壊的変更

**なし。** 本リリースはv1.0.xと完全に下位互換性があります。

---

### 🔄 移行ガイド

**移行は不要** - すべての新機能は追加的で下位互換性があります。

**オプションの拡張:**

1. **マルチテナントサポートを有効化:**
   ```go
   // エージェント設定にUserIDを追加
   agent := agent.New(agent.Config{
       UserID: "user-123",  // NEW
       Memory: memory.NewInMemory(100),
   })
   ```

2. **ワークフローでセッション状態を使用:**
   ```go
   // セッション付きコンテキストを作成
   ctx := workflow.NewExecutionContextWithSession(
       "input",
       "session-id",
       "user-id",
   )
   ```

3. **モデルタイムアウトを設定:**
   ```go
   // モデル設定にTimeoutを追加
   model, _ := openai.New("gpt-4", openai.Config{
       APIKey:  apiKey,
       Timeout: 30 * time.Second,  // NEW
   })
   ```

---

### 📦 インストール

```bash
go get github.com/rexleimo/agno-go@v1.1.0
```

---

### 🔗 リンク

- **GitHubリリース:** [v1.1.0](https://github.com/rexleimo/agno-go/releases/tag/v1.1.0)
- **完全な変更ログ:** [CHANGELOG.md](https://github.com/rexleimo/agno-go/blob/main/CHANGELOG.md)
- **ドキュメント:** [https://agno-go.dev](https://agno-go.dev)

---

## バージョン 1.0.3 (2025-10-06)

### 🧪 改善

- **JSONシリアライゼーションテストの強化** - utils/serializeパッケージで100%のテストカバレッジを達成
- **パフォーマンスベンチマーク** - Python Agnoパフォーマンステストパターンと整合
- **包括的なドキュメント** - バイリンガルパッケージドキュメントを追加

---

## バージョン 1.0.2 (2025-10-05)

### ✨ 追加

#### GLM (智谱AI) プロバイダー

- Zhipu AIのGLMモデルとの完全統合
- GLM-4、GLM-4V（ビジョン）、GLM-3-Turboのサポート
- カスタムJWT認証（HMAC-SHA256）
- 同期およびストリーミングAPI呼び出し
- ツール/関数呼び出しサポート

---

## バージョン 1.0.0 (2025-10-02)

### 🎉 初回リリース

Agno-Go v1.0は、Agnoマルチエージェントフレームワークの高性能Go実装です。

#### コア機能
- **Agent** - ツールサポート付き単一自律エージェント
- **Team** - 4つのモードによるマルチエージェント協力
- **Workflow** - 5つのプリミティブによるステップベースのオーケストレーション

#### LLMプロバイダー
- OpenAI（GPT-4、GPT-3.5、GPT-4 Turbo）
- Anthropic（Claude 3.5 Sonnet、Claude 3 Opus/Sonnet/Haiku）
- Ollama（ローカルモデル）

---

**最終更新:** 2025-10-08
