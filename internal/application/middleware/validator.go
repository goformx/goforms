package middleware

import (
	"github.com/jonesrussell/goforms/internal/domain/common/validation"
)

// echoValidator wraps the domain validator to implement Echo's Validator interface.
type echoValidator struct {
	validator *validation.Validator
}

// NewValidator returns an Echo-compatible validator using the domain validation system.
func NewValidator() *echoValidator {
	return &echoValidator{
		validator: validation.New(),
	}
}

// Validate implements echo.Validator interface.
func (v *echoValidator) Validate(i any) error {
	return v.validator.ValidateStruct(i)
}
