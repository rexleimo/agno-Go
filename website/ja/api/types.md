# Types APIリファレンス

## Messages

**メッセージタイプ:**
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

**メッセージの作成:**
```go
msg := types.NewSystemMessage("You are a helpful assistant")
msg := types.NewUserMessage("Hello")
msg := types.NewAssistantMessage("Hi there!")
msg := types.NewToolMessage("tool_id", "result")
```

## Errors

**エラータイプ:**
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

**エラーの作成:**
```go
err := types.NewInvalidInputError("input cannot be empty")
err := types.NewModelTimeoutError(30 * time.Second)
err := types.NewToolExecutionError("calculator", originalError)
```

**エラーのチェック:**
```go
if types.IsInvalidInputError(err) {
    // 無効な入力を処理
}

if types.IsRateLimitError(err) {
    // レート制限を処理
}
```
