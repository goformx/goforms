package validation

import (
	"reflect"
	"strings"
	"sync"

	validator "github.com/go-playground/validator/v10"
	"github.com/goformx/goforms/internal/domain/common/interfaces"
)

const (
	jsonTagSplitLimit = 2
)

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
