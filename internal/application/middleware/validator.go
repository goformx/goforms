package middleware

import (
	"github.com/goformx/goforms/internal/domain/common/validation"
)

// EchoValidator wraps the domain validator to implement Echo's Validator interface.
type EchoValidator struct {
	validator *validation.Validator
}

// NewValidator creates a new Echo validator
func NewValidator() *EchoValidator {
	return &EchoValidator{
		validator: validation.New(),
	}
}

// Validate implements echo.Validator interface.
func (v *EchoValidator) Validate(i any) error {
	return v.validator.ValidateStruct(i)
}
