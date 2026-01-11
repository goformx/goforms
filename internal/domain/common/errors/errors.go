// Package errors provides domain-specific error types and utilities for
// consistent error handling across the application.
package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a specific type of error
type ErrorCode string

const (
	// ErrCodeValidation represents a validation error
	ErrCodeValidation ErrorCode = "VALIDATION_ERROR"
	// ErrCodeRequired represents a required field error
	ErrCodeRequired ErrorCode = "REQUIRED_FIELD"
	// ErrCodeInvalid represents an invalid value error
	ErrCodeInvalid ErrorCode = "INVALID_VALUE"
	// ErrCodeInvalidFormat represents an invalid format error
	ErrCodeInvalidFormat ErrorCode = "INVALID_FORMAT"
	// ErrCodeInvalidInput represents an invalid input error
	ErrCodeInvalidInput ErrorCode = "INVALID_INPUT"

	// ErrCodeUnauthorized represents an unauthorized access error
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	// ErrCodeForbidden represents a forbidden access error
	ErrCodeForbidden ErrorCode = "FORBIDDEN"
	// ErrCodeAuthentication represents an authentication error
	ErrCodeAuthentication ErrorCode = "AUTHENTICATION_ERROR"
	// ErrCodeInsufficientRole represents an insufficient role error
	ErrCodeInsufficientRole ErrorCode = "INSUFFICIENT_ROLE"

	// ErrCodeNotFound represents a resource not found error
	ErrCodeNotFound ErrorCode = "NOT_FOUND"
	// ErrCodeConflict represents a resource conflict error
	ErrCodeConflict ErrorCode = "CONFLICT"
	// ErrCodeBadRequest represents a bad request error
	ErrCodeBadRequest ErrorCode = "BAD_REQUEST"
	// ErrCodeServerError represents a server error
	ErrCodeServerError ErrorCode = "SERVER_ERROR"
	// ErrCodeAlreadyExists represents a resource already exists error
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"

	// ErrCodeStartup represents a startup error
	ErrCodeStartup ErrorCode = "STARTUP_ERROR"
	// ErrCodeShutdown represents a shutdown error
	ErrCodeShutdown ErrorCode = "SHUTDOWN_ERROR"
	// ErrCodeConfig represents a configuration error
	ErrCodeConfig ErrorCode = "CONFIG_ERROR"
	// ErrCodeDatabase represents a database error
	ErrCodeDatabase ErrorCode = "DB_ERROR"
	// ErrCodeTimeout represents a timeout error
	ErrCodeTimeout ErrorCode = "TIMEOUT"

	// ErrCodeFormValidation represents a form validation error
	ErrCodeFormValidation ErrorCode = "FORM_VALIDATION_ERROR"
	// ErrCodeFormNotFound represents a form not found error
	ErrCodeFormNotFound ErrorCode = "FORM_NOT_FOUND"
	// ErrCodeFormSubmission represents a form submission error
	ErrCodeFormSubmission ErrorCode = "FORM_SUBMISSION_ERROR"
	// ErrCodeFormAccessDenied represents a form access denied error
	ErrCodeFormAccessDenied ErrorCode = "FORM_ACCESS_DENIED"
	// ErrCodeFormInvalid represents an invalid form error
	ErrCodeFormInvalid ErrorCode = "FORM_INVALID"
	// ErrCodeFormExpired represents a form expired error
	ErrCodeFormExpired ErrorCode = "FORM_EXPIRED"

	// ErrCodeUserNotFound represents a user not found error
	ErrCodeUserNotFound ErrorCode = "USER_NOT_FOUND"
	// ErrCodeUserExists represents a user already exists error
	ErrCodeUserExists ErrorCode = "USER_EXISTS"
	// ErrCodeUserDisabled represents a user disabled error
	ErrCodeUserDisabled ErrorCode = "USER_DISABLED"
	// ErrCodeUserInvalid represents an invalid user error
	ErrCodeUserInvalid ErrorCode = "USER_INVALID"
	// ErrCodeUserUnauthorized represents a user unauthorized error
	ErrCodeUserUnauthorized ErrorCode = "USER_UNAUTHORIZED"
)

// DomainError represents a domain-specific error
type DomainError struct {
	Code    ErrorCode
	Message string
	Err     error
	Context map[string]any
}

// ErrorResponse represents a standardized error response for HTTP handlers
type ErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
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

// HTTPStatus returns the appropriate HTTP status code for the error
func (e *DomainError) HTTPStatus() int {
	return GetHTTPStatus(e.Code)
}

// GetHTTPStatus returns the appropriate HTTP status code for an error code
func GetHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrCodeValidation, ErrCodeRequired, ErrCodeInvalid, ErrCodeInvalidFormat, ErrCodeInvalidInput,
		ErrCodeBadRequest, ErrCodeFormValidation, ErrCodeFormInvalid, ErrCodeUserInvalid,
		ErrCodeFormSubmission, ErrCodeFormExpired, ErrCodeUserDisabled:
		return http.StatusBadRequest
	case ErrCodeUnauthorized, ErrCodeUserUnauthorized, ErrCodeAuthentication:
		return http.StatusUnauthorized
	case ErrCodeForbidden, ErrCodeFormAccessDenied, ErrCodeInsufficientRole:
		return http.StatusForbidden
	case ErrCodeNotFound, ErrCodeFormNotFound, ErrCodeUserNotFound:
		return http.StatusNotFound
	case ErrCodeConflict, ErrCodeAlreadyExists, ErrCodeUserExists:
		return http.StatusConflict
	case ErrCodeServerError, ErrCodeDatabase, ErrCodeConfig:
		return http.StatusInternalServerError
	case ErrCodeStartup, ErrCodeShutdown:
		return http.StatusServiceUnavailable
	case ErrCodeTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

// ToResponse converts the DomainError to a standardized ErrorResponse
func (e *DomainError) ToResponse() ErrorResponse {
	return ErrorResponse{
		Code:    string(e.Code),
		Message: e.Message,
		Details: e.Context,
	}
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

	// Form-specific errors
	ErrFormValidation   = New(ErrCodeFormValidation, "form validation error", nil)
	ErrFormNotFound     = New(ErrCodeFormNotFound, "form not found", nil)
	ErrFormSubmission   = New(ErrCodeFormSubmission, "form submission error", nil)
	ErrFormAccessDenied = New(ErrCodeFormAccessDenied, "form access denied", nil)
	ErrFormInvalid      = New(ErrCodeFormInvalid, "invalid form", nil)
	ErrFormExpired      = New(ErrCodeFormExpired, "form has expired", nil)

	// User-specific errors
	ErrUserNotFound     = New(ErrCodeUserNotFound, "user not found", nil)
	ErrUserExists       = New(ErrCodeUserExists, "user already exists", nil)
	ErrUserDisabled     = New(ErrCodeUserDisabled, "user is disabled", nil)
	ErrUserInvalid      = New(ErrCodeUserInvalid, "invalid user", nil)
	ErrUserUnauthorized = New(ErrCodeUserUnauthorized, "user is not authorized", nil)
)

// Wrap wraps an existing error with domain context
func Wrap(err error, code ErrorCode, message string) *DomainError {
	return New(code, message, err)
}

// IsNotFound checks if the error represents a "not found" error
func IsNotFound(err error) bool {
	return HasCategory(err, CategoryNotFound)
}

// IsValidation checks if the error represents a validation error
func IsValidation(err error) bool {
	return HasCategory(err, CategoryValidation)
}

// IsFormError checks if the error represents a form-related error
func IsFormError(err error) bool {
	return HasCategory(err, CategoryForm)
}

// IsUserError checks if the error represents a user-related error
func IsUserError(err error) bool {
	return HasCategory(err, CategoryUser)
}

// IsAuthenticationError checks if the error represents an authentication error
func IsAuthenticationError(err error) bool {
	return HasCategory(err, CategoryAuthentication)
}

// IsSystemError checks if the error represents a system error
func IsSystemError(err error) bool {
	return HasCategory(err, CategorySystem)
}

// IsConflictError checks if the error represents a conflict error
func IsConflictError(err error) bool {
	return HasCategory(err, CategoryConflict)
}

// IsForbiddenError checks if the error represents a forbidden error
func IsForbiddenError(err error) bool {
	return HasCategory(err, CategoryForbidden)
}
