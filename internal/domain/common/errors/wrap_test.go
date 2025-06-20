package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapErrorAndUnwrap(t *testing.T) {
	baseErr := errors.New("base error")
	domainErr := WrapError(baseErr, ErrCodeValidation, "validation failed")
	assert.True(t, IsDomainError(domainErr))
	assert.Equal(t, ErrCodeValidation, GetErrorCode(domainErr))
	assert.Equal(t, "validation failed", GetErrorMessage(domainErr))
	assert.Equal(t, baseErr, UnwrapError(domainErr))
}

func TestWrapNotFoundError(t *testing.T) {
	baseErr := errors.New("not found")
	domainErr := WrapNotFoundError(baseErr, "resource missing")
	assert.True(t, IsDomainError(domainErr))
	assert.Equal(t, ErrCodeNotFound, GetErrorCode(domainErr))
	assert.Contains(t, GetErrorMessage(domainErr), "resource missing")
}

func TestWrapValidationError(t *testing.T) {
	baseErr := errors.New("invalid input")
	domainErr := WrapValidationError(baseErr, "input is invalid")
	assert.True(t, IsDomainError(domainErr))
	assert.Equal(t, ErrCodeInvalid, GetErrorCode(domainErr))
	assert.Contains(t, GetErrorMessage(domainErr), "input is invalid")
}

func TestWrapAuthenticationError(t *testing.T) {
	baseErr := errors.New("bad credentials")
	domainErr := WrapAuthenticationError(baseErr, "auth failed")
	assert.True(t, IsDomainError(domainErr))
	assert.Equal(t, ErrCodeAuthentication, GetErrorCode(domainErr))
	assert.Contains(t, GetErrorMessage(domainErr), "auth failed")
}

func TestWrapAuthorizationError(t *testing.T) {
	baseErr := errors.New("forbidden")
	domainErr := WrapAuthorizationError(baseErr, "not allowed")
	assert.True(t, IsDomainError(domainErr))
	assert.Equal(t, ErrCodeInsufficientRole, GetErrorCode(domainErr))
	assert.Contains(t, GetErrorMessage(domainErr), "not allowed")
}

func TestGetDomainErrorAndContext(t *testing.T) {
	baseErr := errors.New("context error")
	domainErr := WrapError(baseErr, ErrCodeValidation, "context test")
	de := GetDomainError(domainErr)
	assert.NotNil(t, de)
	assert.Equal(t, ErrCodeValidation, de.Code)
	assert.Nil(t, GetErrorContext(domainErr)) // No context set
}

func TestGetFullErrorMessage(t *testing.T) {
	baseErr := errors.New("deepest")
	domainErr := WrapError(baseErr, ErrCodeValidation, "middle")
	outer := WrapError(domainErr, ErrCodeInvalid, "outermost")
	msg := GetFullErrorMessage(outer)
	assert.Contains(t, msg, "outermost")
	assert.Contains(t, msg, "middle")
	assert.Contains(t, msg, "deepest")
} 