# Types API Reference

## Messages

**Message Types:**
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

**Create Messages:**
```go
msg := types.NewSystemMessage("You are a helpful assistant")
msg := types.NewUserMessage("Hello")
msg := types.NewAssistantMessage("Hi there!")
msg := types.NewToolMessage("tool_id", "result")
```

## Errors

**Error Types:**
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

**Create Errors:**
```go
err := types.NewInvalidInputError("input cannot be empty")
err := types.NewModelTimeoutError(30 * time.Second)
err := types.NewToolExecutionError("calculator", originalError)
```

**Check Errors:**
```go
if types.IsInvalidInputError(err) {
    // Handle invalid input
}

if types.IsRateLimitError(err) {
    // Handle rate limit
}
```
