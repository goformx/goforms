package errors

import (
	"errors"
	"fmt"
	"strings"
)

// WrapError wraps an error with a domain error
func WrapError(err error, code ErrorCode, message string) error {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
		Context: make(map[string]any),
	}
}

// WrapErrorf wraps an error with a formatted message
func WrapErrorf(err error, code ErrorCode, format string, args ...any) *DomainError {
	return &DomainError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
		Context: make(map[string]any),
	}
}

// WrapNotFoundError wraps an error with a not found error
func WrapNotFoundError(err error, message string) error {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr
	}

	return &DomainError{
		Code:    ErrCodeNotFound,
		Message: message,
		Err:     err,
		Context: make(map[string]any),
	}
}

// WrapValidationError wraps an error with a validation error
func WrapValidationError(err error, message string) error {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr
	}

	return &DomainError{
		Code:    ErrCodeInvalid,
		Message: message,
		Err:     err,
		Context: make(map[string]any),
	}
}

// WrapAuthenticationError wraps an error with an authentication error
func WrapAuthenticationError(err error, message string) error {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr
	}

	return &DomainError{
		Code:    ErrCodeAuthentication,
		Message: message,
		Err:     err,
		Context: make(map[string]any),
	}
}

// WrapAuthorizationError wraps an error with an authorization error
func WrapAuthorizationError(err error, message string) error {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr
	}

	return &DomainError{
		Code:    ErrCodeInsufficientRole,
		Message: message,
		Err:     err,
		Context: make(map[string]any),
	}
}

// UnwrapError unwraps an error to its original error
func UnwrapError(err error) error {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Err
	}

	return err
}

// IsDomainError checks if the error is a domain error
func IsDomainError(err error) bool {
	var domainErr *DomainError

	return errors.As(err, &domainErr)
}

// GetDomainError returns the domain error if the error is a DomainError
func GetDomainError(err error) *DomainError {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr
	}

	return nil
}

// GetErrorCode returns the error code if the error is a DomainError
func GetErrorCode(err error) ErrorCode {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code
	}

	return ErrCodeServerError
}

// GetErrorMessage returns the error message
func GetErrorMessage(err error) string {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Message
	}

	return err.Error()
}

// GetErrorDetails returns the error details if the error is a DomainError
func GetErrorDetails(err error) map[string]any {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Context
	}

	return nil
}

// GetErrorStack returns the error stack if the error is a DomainError
func GetErrorStack(err error) []error {
	var stack []error

	current := err

	for current != nil {
		stack = append(stack, current)

		var domainErr *DomainError
		if errors.As(current, &domainErr) {
			current = domainErr.Err
		} else {
			break
		}
	}

	return stack
}

// GetErrorContext returns the error context if the error is a DomainError
func GetErrorContext(err error) map[string]any {
	if err == nil {
		return nil
	}

	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		if len(domainErr.Context) == 0 {
			return nil
		}

		return domainErr.Context
	}

	return nil
}

// GetFullErrorMessage returns the full error message including wrapped errors
func GetFullErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	var messages []string

	current := err

	for current != nil {
		var domainErr *DomainError
		if errors.As(current, &domainErr) {
			messages = append(messages, domainErr.Message)
			current = domainErr.Err
		} else {
			messages = append(messages, current.Error())

			break
		}
	}

	return strings.Join(messages, ": ")
}
