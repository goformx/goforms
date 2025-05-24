package errors

import (
	"errors"
	"fmt"
)

// ErrorCode represents a specific type of error
type ErrorCode string

const (
	// Validation errors
	ErrCodeValidation    ErrorCode = "VALIDATION_ERROR"
	ErrCodeRequired      ErrorCode = "REQUIRED_FIELD"
	ErrCodeInvalid       ErrorCode = "INVALID_VALUE"
	ErrCodeInvalidFormat ErrorCode = "INVALID_FORMAT"
	ErrCodeInvalidInput  ErrorCode = "INVALID_INPUT"

	// Authentication errors
	ErrCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrCodeInvalidToken     ErrorCode = "INVALID_TOKEN"
	ErrCodeAuthentication   ErrorCode = "AUTHENTICATION_ERROR"
	ErrCodeInsufficientRole ErrorCode = "INSUFFICIENT_ROLE"

	// Resource errors
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeConflict      ErrorCode = "CONFLICT"
	ErrCodeBadRequest    ErrorCode = "BAD_REQUEST"
	ErrCodeServerError   ErrorCode = "SERVER_ERROR"
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"

	// Application lifecycle errors
	ErrCodeStartup  ErrorCode = "STARTUP_ERROR"
	ErrCodeShutdown ErrorCode = "SHUTDOWN_ERROR"
	ErrCodeConfig   ErrorCode = "CONFIG_ERROR"
	ErrCodeDatabase ErrorCode = "DATABASE_ERROR"
	ErrCodeTimeout  ErrorCode = "TIMEOUT"
)

// DomainError represents a domain-specific error
type DomainError struct {
	Code    ErrorCode
	Message string
	Err     error
	Context map[string]any
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// New creates a new domain error
func New(code ErrorCode, message string, err error) *DomainError {
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

// Common error constructors
func NewValidationError(message string, err error) *DomainError {
	return New(ErrCodeValidation, message, err)
}

func NewNotFoundError(message string, err error) *DomainError {
	return New(ErrCodeNotFound, message, err)
}

func NewForbiddenError(message string, err error) *DomainError {
	return New(ErrCodeForbidden, message, err)
}

func NewStartupError(message string, err error) *DomainError {
	return New(ErrCodeStartup, message, err)
}

func NewShutdownError(message string, err error) *DomainError {
	return New(ErrCodeShutdown, message, err)
}

// IsDomainError checks if an error is a domain error
func IsDomainError(err error) bool {
	var de *DomainError
	return errors.As(err, &de)
}

// GetDomainError returns the domain error if the error is a domain error
func GetDomainError(err error) *DomainError {
	var de *DomainError
	if errors.As(err, &de) {
		return de
	}
	return nil
}

// Common error instances
var (
	// Validation errors
	ErrValidation    = New(ErrCodeValidation, "validation error", nil)
	ErrRequiredField = New(ErrCodeRequired, "field is required", nil)
	ErrInvalidFormat = New(ErrCodeInvalidFormat, "invalid format", nil)
	ErrInvalidValue  = New(ErrCodeInvalid, "invalid value", nil)
	ErrInvalidInput  = New(ErrCodeInvalidInput, "invalid input", nil)

	// Authentication errors
	ErrUnauthorized     = New(ErrCodeUnauthorized, "unauthorized", nil)
	ErrForbidden        = New(ErrCodeForbidden, "forbidden", nil)
	ErrInvalidToken     = New(ErrCodeInvalidToken, "invalid token", nil)
	ErrAuthentication   = New(ErrCodeAuthentication, "authentication error", nil)
	ErrInsufficientRole = New(ErrCodeInsufficientRole, "insufficient role", nil)

	// Resource errors
	ErrNotFound      = New(ErrCodeNotFound, "resource not found", nil)
	ErrConflict      = New(ErrCodeConflict, "resource conflict", nil)
	ErrBadRequest    = New(ErrCodeBadRequest, "bad request", nil)
	ErrServerError   = New(ErrCodeServerError, "internal server error", nil)
	ErrAlreadyExists = New(ErrCodeAlreadyExists, "resource already exists", nil)

	// System errors
	ErrDatabase = New(ErrCodeDatabase, "database error", nil)
	ErrTimeout  = New(ErrCodeTimeout, "operation timed out", nil)
	ErrConfig   = New(ErrCodeConfig, "configuration error", nil)
)

// Wrap wraps an existing error with domain context
func Wrap(err error, code ErrorCode, message string) *DomainError {
	return New(code, message, err)
}
