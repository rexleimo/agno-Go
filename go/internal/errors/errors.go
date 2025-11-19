package errors

import "fmt"

// Code represents a high-level, cross-cutting error classification that can be
// shared across providers, workflows and sessions.
type Code string

const (
	// CodeTimeout indicates a timeout interacting with an external system.
	CodeTimeout Code = "timeout"
	// CodeRateLimit indicates rate limiting by an upstream dependency.
	CodeRateLimit Code = "rate_limit"
	// CodeUnauthorized indicates missing or invalid authentication/authorization.
	CodeUnauthorized Code = "unauthorized"
	// CodeInternal indicates an unexpected internal failure.
	CodeInternal Code = "internal"
	// CodeNotMigrated indicates that the requested provider, collaboration mode
	// or AgentOS-specific capability has not yet been migrated to Go.
	CodeNotMigrated Code = "not_migrated"
)

// Error wraps a low-level error with a classification code and optional
// additional context for telemetry and logging.
type Error struct {
	Code    Code
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// New constructs a new classified Error without an underlying cause.
func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Wrap constructs a new classified Error with an underlying cause.
func Wrap(code Code, message string, cause error) *Error {
	return &Error{Code: code, Message: message, Cause: cause}
}

// NewNotMigrated constructs an Error indicating that a requested feature or
// provider has not yet been migrated to the Go implementation.
func NewNotMigrated(message string) *Error {
	return New(CodeNotMigrated, message)
}
