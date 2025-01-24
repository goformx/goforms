package app

import (
	"github.com/jonesrussell/goforms/internal/validation"
)

// CustomValidator for request validation
type CustomValidator struct {
	validator validation.Validator
}

// Validate implements echo.Validator interface
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// NewValidator creates a new validator instance
func NewValidator() *CustomValidator {
	return &CustomValidator{
		validator: validation.New(),
	}
}
