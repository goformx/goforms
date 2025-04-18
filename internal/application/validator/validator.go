package validator

import (
	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps the validator.Validate instance
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

// Validate validates the provided struct using the validator instance
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

type Validator interface {
	Struct(any) error
	Var(any, string) error
	RegisterValidation(string, func(any) bool) error
}
