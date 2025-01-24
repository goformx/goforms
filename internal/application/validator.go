package application

import (
	"github.com/jonesrussell/goforms/internal/domain/common/interfaces"
	"github.com/jonesrussell/goforms/internal/infrastructure/validation"
)

// CustomValidator for request validation
type CustomValidator struct {
	validator interfaces.Validator
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
