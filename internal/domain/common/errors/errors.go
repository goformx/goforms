package errors

import (
	"errors"
	"fmt"
	"net/http"
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

	// Form-specific errors
	ErrCodeFormValidation   ErrorCode = "FORM_VALIDATION_ERROR"
	ErrCodeFormNotFound     ErrorCode = "FORM_NOT_FOUND"
	ErrCodeFormSubmission   ErrorCode = "FORM_SUBMISSION_ERROR"
	ErrCodeFormAccessDenied ErrorCode = "FORM_ACCESS_DENIED"
	ErrCodeFormInvalid      ErrorCode = "FORM_INVALID"
	ErrCodeFormExpired      ErrorCode = "FORM_EXPIRED"

	// User-specific errors
	ErrCodeUserNotFound     ErrorCode = "USER_NOT_FOUND"
	ErrCodeUserExists       ErrorCode = "USER_EXISTS"
	ErrCodeUserDisabled     ErrorCode = "USER_DISABLED"
	ErrCodeUserInvalid      ErrorCode = "USER_INVALID"
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

// Error type checking utilities
func IsNotFound(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case ErrCodeNotFound, ErrCodeFormNotFound, ErrCodeUserNotFound:
			return true
		case ErrCodeValidation, ErrCodeRequired, ErrCodeInvalid, ErrCodeInvalidFormat, ErrCodeInvalidInput,
			ErrCodeUnauthorized, ErrCodeForbidden, ErrCodeAuthentication,
			ErrCodeInsufficientRole, ErrCodeConflict, ErrCodeBadRequest, ErrCodeServerError,
			ErrCodeAlreadyExists, ErrCodeStartup, ErrCodeShutdown, ErrCodeConfig, ErrCodeDatabase,
			ErrCodeTimeout, ErrCodeFormValidation, ErrCodeFormSubmission, ErrCodeFormAccessDenied,
			ErrCodeFormInvalid, ErrCodeFormExpired, ErrCodeUserExists, ErrCodeUserDisabled,
			ErrCodeUserInvalid, ErrCodeUserUnauthorized:
			return false
		}
	}
	return false
}

func IsValidation(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case ErrCodeValidation, ErrCodeRequired, ErrCodeInvalid, ErrCodeInvalidFormat, ErrCodeInvalidInput,
			ErrCodeBadRequest, ErrCodeFormValidation, ErrCodeFormInvalid, ErrCodeUserInvalid,
			ErrCodeFormSubmission, ErrCodeFormExpired, ErrCodeUserDisabled:
			return true
		case ErrCodeUnauthorized, ErrCodeForbidden, ErrCodeAuthentication,
			ErrCodeNotFound, ErrCodeConflict, ErrCodeServerError,
			ErrCodeAlreadyExists, ErrCodeStartup, ErrCodeShutdown, ErrCodeConfig, ErrCodeDatabase,
			ErrCodeTimeout, ErrCodeFormNotFound, ErrCodeFormAccessDenied, ErrCodeUserNotFound,
			ErrCodeUserExists, ErrCodeUserUnauthorized:
			return false
		}
	}
	return false
}

func IsFormError(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case ErrCodeFormValidation, ErrCodeFormNotFound, ErrCodeFormSubmission,
			ErrCodeFormAccessDenied, ErrCodeFormInvalid, ErrCodeFormExpired:
			return true
		case ErrCodeValidation, ErrCodeRequired, ErrCodeInvalid, ErrCodeInvalidFormat, ErrCodeInvalidInput,
			ErrCodeUnauthorized, ErrCodeForbidden, ErrCodeAuthentication,
			ErrCodeInsufficientRole, ErrCodeNotFound, ErrCodeConflict, ErrCodeBadRequest,
			ErrCodeServerError, ErrCodeAlreadyExists, ErrCodeStartup, ErrCodeShutdown,
			ErrCodeConfig, ErrCodeDatabase, ErrCodeTimeout, ErrCodeUserNotFound,
			ErrCodeUserExists, ErrCodeUserDisabled, ErrCodeUserInvalid, ErrCodeUserUnauthorized:
			return false
		}
	}
	return false
}

func IsUserError(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case ErrCodeUserNotFound, ErrCodeUserExists, ErrCodeUserDisabled,
			ErrCodeUserInvalid, ErrCodeUserUnauthorized:
			return true
		case ErrCodeValidation, ErrCodeRequired, ErrCodeInvalid, ErrCodeInvalidFormat, ErrCodeInvalidInput,
			ErrCodeUnauthorized, ErrCodeForbidden, ErrCodeAuthentication,
			ErrCodeInsufficientRole, ErrCodeNotFound, ErrCodeConflict, ErrCodeBadRequest,
			ErrCodeServerError, ErrCodeAlreadyExists, ErrCodeStartup, ErrCodeShutdown,
			ErrCodeConfig, ErrCodeDatabase, ErrCodeTimeout, ErrCodeFormValidation,
			ErrCodeFormNotFound, ErrCodeFormSubmission, ErrCodeFormAccessDenied,
			ErrCodeFormInvalid, ErrCodeFormExpired:
			return false
		}
	}
	return false
}

func IsAuthenticationError(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case ErrCodeUnauthorized, ErrCodeUserUnauthorized, ErrCodeAuthentication,
			ErrCodeInsufficientRole:
			return true
		case ErrCodeValidation, ErrCodeRequired, ErrCodeInvalid, ErrCodeInvalidFormat, ErrCodeInvalidInput,
			ErrCodeForbidden, ErrCodeNotFound, ErrCodeConflict, ErrCodeBadRequest,
			ErrCodeServerError, ErrCodeAlreadyExists, ErrCodeStartup, ErrCodeShutdown,
			ErrCodeConfig, ErrCodeDatabase, ErrCodeTimeout, ErrCodeFormValidation,
			ErrCodeFormNotFound, ErrCodeFormSubmission, ErrCodeFormAccessDenied,
			ErrCodeFormInvalid, ErrCodeFormExpired, ErrCodeUserNotFound,
			ErrCodeUserExists, ErrCodeUserDisabled, ErrCodeUserInvalid:
			return false
		}
	}
	return false
}

func IsSystemError(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case ErrCodeServerError, ErrCodeDatabase, ErrCodeConfig,
			ErrCodeStartup, ErrCodeShutdown, ErrCodeTimeout:
			return true
		case ErrCodeValidation, ErrCodeRequired, ErrCodeInvalid, ErrCodeInvalidFormat, ErrCodeInvalidInput,
			ErrCodeUnauthorized, ErrCodeForbidden, ErrCodeAuthentication,
			ErrCodeInsufficientRole, ErrCodeNotFound, ErrCodeConflict, ErrCodeBadRequest,
			ErrCodeAlreadyExists, ErrCodeFormValidation, ErrCodeFormNotFound,
			ErrCodeFormSubmission, ErrCodeFormAccessDenied, ErrCodeFormInvalid,
			ErrCodeFormExpired, ErrCodeUserNotFound, ErrCodeUserExists,
			ErrCodeUserDisabled, ErrCodeUserInvalid, ErrCodeUserUnauthorized:
			return false
		}
	}
	return false
}

func IsConflictError(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case ErrCodeConflict, ErrCodeAlreadyExists, ErrCodeUserExists:
			return true
		case ErrCodeValidation, ErrCodeRequired, ErrCodeInvalid, ErrCodeInvalidFormat, ErrCodeInvalidInput,
			ErrCodeUnauthorized, ErrCodeForbidden, ErrCodeAuthentication,
			ErrCodeInsufficientRole, ErrCodeNotFound, ErrCodeBadRequest, ErrCodeServerError,
			ErrCodeStartup, ErrCodeShutdown, ErrCodeConfig, ErrCodeDatabase, ErrCodeTimeout,
			ErrCodeFormValidation, ErrCodeFormNotFound, ErrCodeFormSubmission,
			ErrCodeFormAccessDenied, ErrCodeFormInvalid, ErrCodeFormExpired,
			ErrCodeUserNotFound, ErrCodeUserDisabled, ErrCodeUserInvalid,
			ErrCodeUserUnauthorized:
			return false
		}
	}
	return false
}

func IsForbiddenError(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case ErrCodeForbidden, ErrCodeFormAccessDenied, ErrCodeInsufficientRole:
			return true
		case ErrCodeValidation, ErrCodeRequired, ErrCodeInvalid, ErrCodeInvalidFormat, ErrCodeInvalidInput,
			ErrCodeUnauthorized, ErrCodeAuthentication, ErrCodeNotFound,
			ErrCodeConflict, ErrCodeBadRequest, ErrCodeServerError, ErrCodeAlreadyExists,
			ErrCodeStartup, ErrCodeShutdown, ErrCodeConfig, ErrCodeDatabase, ErrCodeTimeout,
			ErrCodeFormValidation, ErrCodeFormNotFound, ErrCodeFormSubmission,
			ErrCodeFormInvalid, ErrCodeFormExpired, ErrCodeUserNotFound,
			ErrCodeUserExists, ErrCodeUserDisabled, ErrCodeUserInvalid,
			ErrCodeUserUnauthorized:
			return false
		}
	}
	return false
}
