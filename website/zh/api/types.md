# Types API 参考 / Types API Reference

## 消息 / Messages

**消息类型 / Message Types:**
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

**创建消息 / Create Messages:**
```go
msg := types.NewSystemMessage("You are a helpful assistant")
msg := types.NewUserMessage("Hello")
msg := types.NewAssistantMessage("Hi there!")
msg := types.NewToolMessage("tool_id", "result")
```

## 错误 / Errors

**错误类型 / Error Types:**
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

**创建错误 / Create Errors:**
```go
err := types.NewInvalidInputError("input cannot be empty")
err := types.NewModelTimeoutError(30 * time.Second)
err := types.NewToolExecutionError("calculator", originalError)
```

**检查错误 / Check Errors:**
```go
if types.IsInvalidInputError(err) {
    // 处理无效输入 / Handle invalid input
}

if types.IsRateLimitError(err) {
    // 处理速率限制 / Handle rate limit
}
```
