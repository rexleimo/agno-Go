# Memory API 参考 / Memory API Reference

## NewInMemory

创建内存中的对话存储。/ Create in-memory conversation storage.

**签名 / Signature:**
```go
func NewInMemory(maxSize int) *InMemory
```

**参数 / Parameters:**
- `maxSize`: 保留的最大消息数 / Maximum number of messages to keep

**方法 / Methods:**
```go
func (m *InMemory) Add(msg *types.Message)
func (m *InMemory) GetMessages() []*types.Message
func (m *InMemory) Clear()
```

**示例 / Example:**
```go
mem := memory.NewInMemory(100) // 保留最后 100 条消息 / Keep last 100 messages

ag, _ := agent.New(agent.Config{
    Memory: mem,
    // ...
})

// 稍后 / Later
ag.ClearMemory() // 清除所有消息 / Clear all messages
```
