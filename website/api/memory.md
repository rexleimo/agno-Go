# Memory API Reference

## NewInMemory

Create in-memory conversation storage.

**Signature:**
```go
func NewInMemory(maxSize int) *InMemory
```

**Parameters:**
- `maxSize`: Maximum number of messages to keep

**Methods:**
```go
func (m *InMemory) Add(msg *types.Message)
func (m *InMemory) GetMessages() []*types.Message
func (m *InMemory) Clear()
```

**Example:**
```go
mem := memory.NewInMemory(100) // Keep last 100 messages

ag, _ := agent.New(agent.Config{
    Memory: mem,
    // ...
})

// Later
ag.ClearMemory() // Clear all messages
```
