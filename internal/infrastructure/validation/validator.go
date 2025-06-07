package validation

import (
	"reflect"
	"strings"
	"sync"

	"errors"
	"fmt"

	validator "github.com/go-playground/validator/v10"
	"github.com/goformx/goforms/internal/domain/common/interfaces"
)

const (
	jsonTagSplitLimit = 2
)

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string
	Message string
}

// ValidationErrors represents a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, err := range e {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return sb.String()
}

// getFieldName returns the field name from the validation error
func getFieldName(e validator.FieldError) string {
	field := e.Field()
	if field == "" {
		return e.StructField()
	}
	return field
}

// getErrorMessage returns the error message from the validation error
func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s", e.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", e.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters", e.Param())
	case "oneof":
		return fmt.Sprintf("must be one of [%s]", e.Param())
	default:
		return fmt.Sprintf("failed on tag %s", e.Tag())
	}
}

// validatorImpl implements the Validator interface
type validatorImpl struct {
	*validator.Validate
}

//nolint:gochecknoglobals // singleton pattern requires global instance and once
var (
	instance *validatorImpl
	once     sync.Once
)

// New returns a singleton instance of the validator
func New() interfaces.Validator {
	once.Do(func() {
		v := validator.New()
		// Enable struct field validation
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", jsonTagSplitLimit)[0]
			if name == "-" {
				return ""
			}
			return name
		})
		instance = &validatorImpl{Validate: v}
	})
	return instance
}

// Struct implements validator.Struct
func (v *validatorImpl) Struct(s any) error {
	return v.Validate.Struct(s)
}

// Var implements validator.Var
func (v *validatorImpl) Var(field any, tag string) error {
	return v.Validate.Var(field, tag)
}

// RegisterValidation implements validator.RegisterValidation
func (v *validatorImpl) RegisterValidation(tag string, fn func(fl validator.FieldLevel) bool) error {
	return v.Validate.RegisterValidation(tag, fn)
}

// RegisterStructValidation implements validator.RegisterStructValidation
func (v *validatorImpl) RegisterStructValidation(fn func(sl validator.StructLevel), typ any) error {
	v.Validate.RegisterStructValidation(fn, typ)
	return nil
}

// RegisterCrossFieldValidation implements validator.RegisterCrossFieldValidation
func (v *validatorImpl) RegisterCrossFieldValidation(tag string, fn func(fl validator.FieldLevel) bool) error {
	return v.Validate.RegisterValidation(tag, fn)
}

// GetValidationErrors returns detailed validation errors
func (v *validatorImpl) GetValidationErrors(err error) map[string]string {
	if err == nil {
		return nil
	}

	validationErrors := make(map[string]string)
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			field := e.Field()
			switch e.Tag() {
			case "required":
				validationErrors[field] = "This field is required"
			case "email":
				validationErrors[field] = "Invalid email format"
			case "min":
				validationErrors[field] = "Value is too short"
			case "max":
				validationErrors[field] = "Value is too long"
			case "match":
				validationErrors[field] = "Fields do not match"
			default:
				validationErrors[field] = "Invalid value"
			}
		}
	} else {
		// Handle non-validation errors
		validationErrors["_error"] = err.Error()
	}
	return validationErrors
}

func (v *validatorImpl) ValidateStruct(s interface{}) error {
	err := v.Validate.Struct(s)
	if err == nil {
		return nil
	}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		validationErrors := make([]ValidationError, len(ve))
		for i, e := range ve {
			validationErrors[i] = ValidationError{
				Field:   getFieldName(e),
				Message: getErrorMessage(e),
			}
		}
		return ValidationErrors(validationErrors)
	}
	return err
}
