package middleware

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator holds the validator instance
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() echo.Validator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

// Validate validates the provided struct
func (cv *CustomValidator) Validate(i any) error {
	return cv.validator.Struct(i)
} 