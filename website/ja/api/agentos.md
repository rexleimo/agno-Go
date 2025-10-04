# AgentOS Server APIリファレンス

## NewServer

HTTPサーバーを作成します。

**シグネチャ:**
```go
func NewServer(config *Config) (*Server, error)

type Config struct {
    Address        string           // サーバーアドレス (デフォルト: :8080)
    SessionStorage session.Storage  // セッションストレージ (デフォルト: memory)
    Logger         *slog.Logger     // ロガー (デフォルト: slog.Default())
    Debug          bool             // デバッグモード (デフォルト: false)
    AllowOrigins   []string         // CORSオリジン
    AllowMethods   []string         // CORSメソッド
    AllowHeaders   []string         // CORSヘッダー
    RequestTimeout time.Duration    // リクエストタイムアウト (デフォルト: 30秒)
    MaxRequestSize int64            // 最大リクエストサイズ (デフォルト: 10MB)
}
```

**例:**
```go
server, err := agentos.NewServer(&agentos.Config{
    Address: ":8080",
    Debug:   true,
    RequestTimeout: 60 * time.Second,
})
```

## Server.RegisterAgent

エージェントを登録します。

**シグネチャ:**
```go
func (s *Server) RegisterAgent(agentID string, ag *agent.Agent) error
```

**例:**
```go
err := server.RegisterAgent("assistant", myAgent)
```

## Server.Start / Shutdown

サーバーを起動および停止します。

**シグネチャ:**
```go
func (s *Server) Start() error
func (s *Server) Shutdown(ctx context.Context) error
```

**例:**
```go
go func() {
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}()

// グレースフルシャットダウン
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
server.Shutdown(ctx)
```

## APIエンドポイント

完全なAPIドキュメントは[OpenAPI仕様](../../pkg/agentos/openapi.yaml)を参照してください。

**コアエンドポイント:**
- `GET /health` - ヘルスチェック
- `POST /api/v1/sessions` - セッション作成
- `GET /api/v1/sessions/{id}` - セッション取得
- `PUT /api/v1/sessions/{id}` - セッション更新
- `DELETE /api/v1/sessions/{id}` - セッション削除
- `GET /api/v1/sessions` - セッション一覧
- `GET /api/v1/agents` - エージェント一覧
- `POST /api/v1/agents/{id}/run` - エージェント実行

## ベストプラクティス

### 1. 常にContextを使用

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

output, err := ag.Run(ctx, input)
```

### 2. エラーを適切に処理

```go
output, err := ag.Run(ctx, input)
if err != nil {
    switch {
    case types.IsInvalidInputError(err):
        // 無効な入力を処理
    case types.IsRateLimitError(err):
        // バックオフして再試行
    default:
        // その他のエラーを処理
    }
}
```

### 3. メモリを管理

```go
// 新しいトピックを開始する際にクリア
ag.ClearMemory()

// または制限付きメモリを使用
mem := memory.NewInMemory(50)
```

### 4. 適切なタイムアウトを設定

```go
server, _ := agentos.NewServer(&agentos.Config{
    RequestTimeout: 60 * time.Second, // 複雑なエージェント用
})
```
