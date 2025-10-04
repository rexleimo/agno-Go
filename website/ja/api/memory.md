# Memory APIリファレンス

## NewInMemory

インメモリ会話ストレージを作成します。

**シグネチャ:**
```go
func NewInMemory(maxSize int) *InMemory
```

**パラメータ:**
- `maxSize`: 保持する最大メッセージ数

**メソッド:**
```go
func (m *InMemory) Add(msg *types.Message)
func (m *InMemory) GetMessages() []*types.Message
func (m *InMemory) Clear()
```

**例:**
```go
mem := memory.NewInMemory(100) // 最新100メッセージを保持

ag, _ := agent.New(agent.Config{
    Memory: mem,
    // ...
})

// 後で
ag.ClearMemory() // すべてのメッセージをクリア
```
