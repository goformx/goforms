package middleware

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/goformx/goforms/internal/domain/common/interfaces"
	"github.com/goformx/goforms/internal/infrastructure/validation"
)

// EchoValidator wraps the infrastructure validator to implement Echo's Validator interface.
type EchoValidator struct {
	validator interfaces.Validator
}

// NewValidator creates a new Echo validator
func NewValidator() (*EchoValidator, error) {
	v, errNew := validation.New()
	if errNew != nil {
		return nil, fmt.Errorf("create validation instance: %w", errNew)
	}

	return &EchoValidator{
		validator: v,
	}, nil
}

// Validate implements echo.Validator interface.
func (v *EchoValidator) Validate(i any) error {
	if i == nil {
		return errors.New("validation failed: input is nil")
	}

	// Ensure the input is a struct before validating
	if reflect.TypeOf(i).Kind() != reflect.Struct {
		return errors.New("validation failed: input is not a struct")
	}

	err := v.validator.Struct(i)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	return nil
}
