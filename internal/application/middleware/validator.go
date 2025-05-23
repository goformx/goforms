package middleware

import (
	"github.com/goformx/goforms/internal/domain/common/interfaces"
	"github.com/goformx/goforms/internal/infrastructure/validation"
)

// EchoValidator wraps the infrastructure validator to implement Echo's Validator interface.
type EchoValidator struct {
	validator interfaces.Validator
}

// NewValidator creates a new Echo validator
func NewValidator() *EchoValidator {
	return &EchoValidator{
		validator: validation.New(),
	}
}

// Validate implements echo.Validator interface.
func (v *EchoValidator) Validate(i any) error {
	return v.validator.Struct(i)
}
