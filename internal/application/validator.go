package application

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// Validator provides validation functionality
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// ValidateStruct validates a struct
func (v *Validator) ValidateStruct(s interface{}) error {
	if err := v.validate.Struct(s); err != nil {
		return errors.New("invalid value")
	}
	return nil
}

// ValidateField validates a field
func (v *Validator) ValidateField(field interface{}, tag string) error {
	if err := v.validate.Var(field, tag); err != nil {
		return errors.New("field is required")
	}
	return nil
}

// CustomValidator implements echo.Validator interface
type CustomValidator struct{}

// NewCustomValidator creates a new custom validator
func NewCustomValidator() *CustomValidator {
	return &CustomValidator{}
}

// Validate validates the input using reflection
func (cv *CustomValidator) Validate(i any) error {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("validator: expected struct, got %s", v.Kind())
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		value := v.Field(i)

		if tag := f.Tag.Get("validate"); tag != "" {
			if err := cv.validateField(value, tag); err != nil {
				return fmt.Errorf("validator: %s: %w", f.Name, err)
			}
		}
	}

	return nil
}

// validateField validates a single field based on its tag
func (cv *CustomValidator) validateField(value reflect.Value, tag string) error {
	if !value.IsValid() {
		return fmt.Errorf("invalid value")
	}

	if tag == "required" && value.IsZero() {
		return fmt.Errorf("field is required")
	}

	return nil
}
