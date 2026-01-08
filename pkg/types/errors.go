// Package types provides type definitions for the Chucky SDK.
package types

import "fmt"

// ErrorCode represents the type of error that occurred.
type ErrorCode string

const (
	ErrCodeConnection       ErrorCode = "CONNECTION_ERROR"
	ErrCodeAuthentication   ErrorCode = "AUTHENTICATION_ERROR"
	ErrCodeBudgetExceeded   ErrorCode = "BUDGET_EXCEEDED"
	ErrCodeConcurrencyLimit ErrorCode = "CONCURRENCY_LIMIT"
	ErrCodeRateLimit        ErrorCode = "RATE_LIMIT"
	ErrCodeSession          ErrorCode = "SESSION_ERROR"
	ErrCodeToolExecution    ErrorCode = "TOOL_EXECUTION_ERROR"
	ErrCodeTimeout          ErrorCode = "TIMEOUT_ERROR"
	ErrCodeValidation       ErrorCode = "VALIDATION_ERROR"
	ErrCodeProtocol         ErrorCode = "PROTOCOL_ERROR"
	ErrCodeUnknown          ErrorCode = "UNKNOWN_ERROR"
)

// ChuckyError is the base error type for all SDK errors.
type ChuckyError struct {
	Code    ErrorCode
	Message string
	Details map[string]any
	Err     error // Wrapped error
}

func (e *ChuckyError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *ChuckyError) Unwrap() error {
	return e.Err
}

// NewChuckyError creates a new ChuckyError with the given code and message.
func NewChuckyError(code ErrorCode, message string) *ChuckyError {
	return &ChuckyError{
		Code:    code,
		Message: message,
	}
}

// WithDetails adds details to the error.
func (e *ChuckyError) WithDetails(details map[string]any) *ChuckyError {
	e.Details = details
	return e
}

// Wrap wraps another error.
func (e *ChuckyError) Wrap(err error) *ChuckyError {
	e.Err = err
	return e
}

// ConnectionError creates a connection error.
func ConnectionError(message string) *ChuckyError {
	return NewChuckyError(ErrCodeConnection, message)
}

// AuthenticationError creates an authentication error.
func AuthenticationError(message string) *ChuckyError {
	return NewChuckyError(ErrCodeAuthentication, message)
}

// BudgetExceededError creates a budget exceeded error.
func BudgetExceededError(message string) *ChuckyError {
	return NewChuckyError(ErrCodeBudgetExceeded, message)
}

// ConcurrencyLimitError creates a concurrency limit error.
func ConcurrencyLimitError(message string) *ChuckyError {
	return NewChuckyError(ErrCodeConcurrencyLimit, message)
}

// RateLimitError creates a rate limit error.
func RateLimitError(message string) *ChuckyError {
	return NewChuckyError(ErrCodeRateLimit, message)
}

// SessionError creates a session error.
func SessionError(message string) *ChuckyError {
	return NewChuckyError(ErrCodeSession, message)
}

// ToolExecutionError creates a tool execution error.
func ToolExecutionError(toolName string, message string) *ChuckyError {
	return NewChuckyError(ErrCodeToolExecution, message).WithDetails(map[string]any{
		"toolName": toolName,
	})
}

// TimeoutError creates a timeout error.
func TimeoutError(message string) *ChuckyError {
	return NewChuckyError(ErrCodeTimeout, message)
}

// ValidationError creates a validation error.
func ValidationError(message string) *ChuckyError {
	return NewChuckyError(ErrCodeValidation, message)
}

// ProtocolError creates a protocol error.
func ProtocolError(message string) *ChuckyError {
	return NewChuckyError(ErrCodeProtocol, message)
}
