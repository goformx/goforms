package errors

import (
	"fmt"
	"strings"
)

// WrapError wraps an error with a domain error
func WrapError(err error, code ErrorCode, message string) *DomainError {
	if err == nil {
		return nil
	}

	// If the error is already a DomainError, preserve its context
	if domainErr, ok := err.(*DomainError); ok {
		return &DomainError{
			Code:    code,
			Message: message,
			Err:     domainErr,
			Context: domainErr.Context,
		}
	}

	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
		Context: make(map[string]any),
	}
}

// WrapErrorf wraps an error with a formatted message
func WrapErrorf(err error, code ErrorCode, format string, args ...any) *DomainError {
	return WrapError(err, code, fmt.Sprintf(format, args...))
}

// IsErrorCode checks if the error is of the given error code
func IsErrorCode(err error, code ErrorCode) bool {
	if err == nil {
		return false
	}

	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == code
	}

	return false
}

// GetErrorCode returns the error code if the error is a DomainError
func GetErrorCode(err error) ErrorCode {
	if err == nil {
		return ""
	}

	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code
	}

	return ""
}

// GetErrorMessage returns the error message
func GetErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Message
	}

	return err.Error()
}

// GetErrorContext returns the error context if the error is a DomainError
func GetErrorContext(err error) map[string]any {
	if err == nil {
		return nil
	}

	if domainErr, ok := err.(*DomainError); ok {
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
		if domainErr, ok := current.(*DomainError); ok {
			messages = append(messages, domainErr.Message)
			current = domainErr.Err
		} else {
			messages = append(messages, current.Error())
			break
		}
	}

	return strings.Join(messages, ": ")
}
