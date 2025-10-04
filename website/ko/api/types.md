# Types API 레퍼런스

## Messages

**메시지 타입:**
```go
const (
    RoleSystem    = "system"
    RoleUser      = "user"
    RoleAssistant = "assistant"
    RoleTool      = "tool"
)

type Message struct {
    Role      string
    Content   string
    ToolCalls []ToolCall
}
```

**메시지 생성:**
```go
msg := types.NewSystemMessage("You are a helpful assistant")
msg := types.NewUserMessage("Hello")
msg := types.NewAssistantMessage("Hi there!")
msg := types.NewToolMessage("tool_id", "result")
```

## Errors

**에러 타입:**
```go
type AgnoError struct {
    Code    ErrorCode
    Message string
    Cause   error
}

const (
    ErrCodeInvalidInput    ErrorCode = "INVALID_INPUT"
    ErrCodeInvalidConfig   ErrorCode = "INVALID_CONFIG"
    ErrCodeModelTimeout    ErrorCode = "MODEL_TIMEOUT"
    ErrCodeToolExecution   ErrorCode = "TOOL_EXECUTION"
    ErrCodeAPIError        ErrorCode = "API_ERROR"
    ErrCodeRateLimit       ErrorCode = "RATE_LIMIT"
)
```

**에러 생성:**
```go
err := types.NewInvalidInputError("input cannot be empty")
err := types.NewModelTimeoutError(30 * time.Second)
err := types.NewToolExecutionError("calculator", originalError)
```

**에러 확인:**
```go
if types.IsInvalidInputError(err) {
    // 잘못된 입력 처리
}

if types.IsRateLimitError(err) {
    // 속도 제한 처리
}
```
