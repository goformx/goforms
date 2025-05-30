package errors

import (
	"fmt"
)

// ErrorCode represents a unique error code for each error type
type ErrorCode string

const (
	// ErrCodeValidation represents a general validation error
	ErrCodeValidation ErrorCode = "VALIDATION_ERROR"
	// ErrCodeRequiredField represents a required field validation error
	ErrCodeRequiredField ErrorCode = "REQUIRED_FIELD"
	// ErrCodeInvalidFormat represents an invalid format error
	ErrCodeInvalidFormat ErrorCode = "INVALID_FORMAT"
	// ErrCodeInvalidValue represents an invalid value error
	ErrCodeInvalidValue ErrorCode = "INVALID_VALUE"

	// ErrCodeUnauthorized represents an unauthorized access error
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	// ErrCodeInvalidToken represents an invalid token error
	ErrCodeInvalidToken ErrorCode = "INVALID_TOKEN"
	// ErrCodeTokenExpired represents a token expiration error
	ErrCodeTokenExpired ErrorCode = "TOKEN_EXPIRED"
	// ErrCodeAuthentication represents a general authentication error
	ErrCodeAuthentication ErrorCode = "AUTHENTICATION_ERROR"

	// ErrCodeForbidden represents a forbidden access error
	ErrCodeForbidden ErrorCode = "FORBIDDEN"
	// ErrCodeInsufficientRole represents an insufficient role error
	ErrCodeInsufficientRole ErrorCode = "INSUFFICIENT_ROLE"

	// ErrCodeNotFound represents a resource not found error
	ErrCodeNotFound ErrorCode = "NOT_FOUND"
	// ErrCodeAlreadyExists represents a resource already exists error
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	// ErrCodeConflict represents a resource conflict error
	ErrCodeConflict ErrorCode = "CONFLICT"

	// ErrCodeInternal represents an internal server error
	ErrCodeInternal ErrorCode = "INTERNAL_ERROR"
	// ErrCodeDatabase represents a database error
	ErrCodeDatabase ErrorCode = "DATABASE_ERROR"
	// ErrCodeTimeout represents an operation timeout error
	ErrCodeTimeout ErrorCode = "TIMEOUT"

	// ErrCodeInvalidInput represents an invalid input error
	ErrCodeInvalidInput ErrorCode = "INVALID_INPUT"
)

// DomainError represents a domain-specific error
type DomainError struct {
	Code    ErrorCode
	Message string
	Err     error
	Context map[string]any
}

// New creates a new domain error
func New(code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Context: make(map[string]any),
	}
}

// Wrap wraps an existing error with domain context
func Wrap(err error, code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
		Context: make(map[string]any),
	}
}

// WithContext adds context to the error
func (e *DomainError) WithContext(key string, value any) *DomainError {
	e.Context[key] = value
	return e
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *DomainError) Unwrap() error {
	return e.Err
}

// Is checks if the error is of a specific type
func (e *DomainError) Is(target error) bool {
	if target == nil {
		return false
	}
	if err, ok := target.(*DomainError); ok {
		return e.Code == err.Code
	}
	return false
}

// Validation errors
var (
	ErrValidation    = New(ErrCodeValidation, "validation error")
	ErrRequiredField = New(ErrCodeRequiredField, "field is required")
	ErrInvalidFormat = New(ErrCodeInvalidFormat, "invalid format")
	ErrInvalidValue  = New(ErrCodeInvalidValue, "invalid value")
)

// Authentication errors
var (
	ErrUnauthorized   = New(ErrCodeUnauthorized, "unauthorized")
	ErrInvalidToken   = New(ErrCodeInvalidToken, "invalid token")
	ErrTokenExpired   = New(ErrCodeTokenExpired, "token expired")
	ErrAuthentication = New(ErrCodeAuthentication, "authentication error")
)

// Authorization errors
var (
	ErrForbidden        = New(ErrCodeForbidden, "forbidden")
	ErrInsufficientRole = New(ErrCodeInsufficientRole, "insufficient role")
)

// Resource errors
var (
	ErrNotFound      = New(ErrCodeNotFound, "resource not found")
	ErrAlreadyExists = New(ErrCodeAlreadyExists, "resource already exists")
	ErrConflict      = New(ErrCodeConflict, "resource conflict")
)

// System errors
var (
	ErrInternal = New(ErrCodeInternal, "internal error")
	ErrDatabase = New(ErrCodeDatabase, "database error")
	ErrTimeout  = New(ErrCodeTimeout, "operation timed out")
)
