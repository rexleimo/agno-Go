# Memory API 레퍼런스

## NewInMemory

인메모리 대화 저장소를 생성합니다.

**함수 시그니처:**
```go
func NewInMemory(maxSize int) *InMemory
```

**매개변수:**
- `maxSize`: 유지할 최대 메시지 수

**메서드:**
```go
func (m *InMemory) Add(msg *types.Message)
func (m *InMemory) GetMessages() []*types.Message
func (m *InMemory) Clear()
```

**예제:**
```go
mem := memory.NewInMemory(100) // 최근 100개 메시지 유지

ag, _ := agent.New(agent.Config{
    Memory: mem,
    // ...
})

// 나중에
ag.ClearMemory() // 모든 메시지 초기화
```
