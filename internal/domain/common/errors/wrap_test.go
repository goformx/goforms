package errors_test

import (
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goformx/goforms/internal/domain/common/errors"
)

func TestWrapErrorAndUnwrap(t *testing.T) {
	baseErr := stderrors.New("base error")
	domainErr := errors.WrapError(baseErr, errors.ErrCodeValidation, "validation failed")
	assert.True(t, errors.IsDomainError(domainErr))
	assert.Equal(t, errors.ErrCodeValidation, errors.GetErrorCode(domainErr))
	assert.Equal(t, "validation failed", errors.GetErrorMessage(domainErr))
	assert.Equal(t, baseErr, errors.UnwrapError(domainErr))
}

func TestWrapNotFoundError(t *testing.T) {
	baseErr := stderrors.New("not found")
	domainErr := errors.WrapNotFoundError(baseErr, "resource missing")
	assert.True(t, errors.IsDomainError(domainErr))
	assert.Equal(t, errors.ErrCodeNotFound, errors.GetErrorCode(domainErr))
	assert.Contains(t, errors.GetErrorMessage(domainErr), "resource missing")
}

func TestWrapError(t *testing.T) {
	baseErr := stderrors.New("invalid input")
	domainErr := errors.WrapValidationError(baseErr, "input is invalid")
	assert.True(t, errors.IsDomainError(domainErr))
	assert.Equal(t, errors.ErrCodeInvalid, errors.GetErrorCode(domainErr))
	assert.Contains(t, errors.GetErrorMessage(domainErr), "input is invalid")
}

func TestWrapAuthenticationError(t *testing.T) {
	baseErr := stderrors.New("bad credentials")
	domainErr := errors.WrapAuthenticationError(baseErr, "auth failed")
	assert.True(t, errors.IsDomainError(domainErr))
	assert.Equal(t, errors.ErrCodeAuthentication, errors.GetErrorCode(domainErr))
	assert.Contains(t, errors.GetErrorMessage(domainErr), "auth failed")
}

func TestWrapAuthorizationError(t *testing.T) {
	baseErr := stderrors.New("forbidden")
	domainErr := errors.WrapAuthorizationError(baseErr, "not allowed")
	assert.True(t, errors.IsDomainError(domainErr))
	assert.Equal(t, errors.ErrCodeInsufficientRole, errors.GetErrorCode(domainErr))
	assert.Contains(t, errors.GetErrorMessage(domainErr), "not allowed")
}

func TestGetDomainErrorAndContext(t *testing.T) {
	baseErr := stderrors.New("context error")
	domainErr := errors.WrapError(baseErr, errors.ErrCodeValidation, "context test")
	de := errors.GetDomainError(domainErr)
	assert.NotNil(t, de)
	assert.Equal(t, errors.ErrCodeValidation, de.Code)
	assert.Nil(t, errors.GetErrorContext(domainErr)) // No context set
}

func TestGetFullErrorMessage(t *testing.T) {
	baseErr := stderrors.New("deepest")
	domainErr := errors.WrapError(baseErr, errors.ErrCodeValidation, "middle")
	outer := errors.WrapError(domainErr, errors.ErrCodeInvalid, "outermost")
	msg := errors.GetFullErrorMessage(outer)
	assert.Contains(t, msg, "outermost")
	assert.Contains(t, msg, "middle")
	assert.Contains(t, msg, "deepest")
}
