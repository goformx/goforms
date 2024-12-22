package app

import (
	"github.com/go-playground/validator/v10"
)

// CustomValidator for request validation
type CustomValidator struct {
	validator *validator.Validate
}

// Validate implements echo.Validator interface
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// NewValidator creates a new validator instance
func NewValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}
