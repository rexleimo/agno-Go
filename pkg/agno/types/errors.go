package types

import "fmt"

// ErrorCode represents different types of errors in the system
type ErrorCode string

const (
	ErrCodeModelTimeout   ErrorCode = "MODEL_TIMEOUT"
	ErrCodeToolExecution  ErrorCode = "TOOL_ERROR"
	ErrCodeInvalidInput   ErrorCode = "INVALID_INPUT"
	ErrCodeInvalidConfig  ErrorCode = "INVALID_CONFIG"
	ErrCodeAPIError       ErrorCode = "API_ERROR"
	ErrCodeRateLimitError ErrorCode = "RATE_LIMIT"
	ErrCodeUnknown        ErrorCode = "UNKNOWN"
)

// AgnoError represents a structured error in the Agno system
type AgnoError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

// Error implements the error interface
func (e *AgnoError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *AgnoError) Unwrap() error {
	return e.Cause
}

// NewError creates a new AgnoError
func NewError(code ErrorCode, message string, cause error) *AgnoError {
	return &AgnoError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewModelTimeoutError creates a model timeout error
func NewModelTimeoutError(message string, cause error) *AgnoError {
	return NewError(ErrCodeModelTimeout, message, cause)
}

// NewToolExecutionError creates a tool execution error
func NewToolExecutionError(message string, cause error) *AgnoError {
	return NewError(ErrCodeToolExecution, message, cause)
}

// NewInvalidInputError creates an invalid input error
func NewInvalidInputError(message string, cause error) *AgnoError {
	return NewError(ErrCodeInvalidInput, message, cause)
}

// NewInvalidConfigError creates an invalid config error
func NewInvalidConfigError(message string, cause error) *AgnoError {
	return NewError(ErrCodeInvalidConfig, message, cause)
}

// NewAPIError creates an API error
func NewAPIError(message string, cause error) *AgnoError {
	return NewError(ErrCodeAPIError, message, cause)
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message string, cause error) *AgnoError {
	return NewError(ErrCodeRateLimitError, message, cause)
}
