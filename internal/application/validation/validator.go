package validation

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/goformx/goforms/internal/domain/common/interfaces"
)

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string
	Message string
	Value   any // The invalid value that caused the error
}

// ValidationErrors represents a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (e ValidationErrors) Error() string {
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

// getErrorMessage returns a user-friendly error message for the validation error
func getErrorMessage(e validator.FieldError) string {
	field := getFieldName(e)

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", field, e.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number", field)
	case "password":
		return fmt.Sprintf("%s must contain at least 8 characters, including uppercase, lowercase, "+
			"number and special character", field)
	case "date":
		return fmt.Sprintf("%s must be a valid date in format YYYY-MM-DD", field)
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime in format YYYY-MM-DD HH:mm:ss", field)
	default:
		return fmt.Sprintf("%s failed validation: %s", field, e.Tag())
	}
}

// validatorImpl implements the interfaces.Validator interface
type validatorImpl struct {
	validate *validator.Validate
	cache    sync.Map // Cache for validation results
}

// New creates a new validator instance with common validation rules
func New() (interfaces.Validator, error) {
	v := validator.New()

	// Enable struct field validation
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		// Get the first part of the JSON tag (before any comma)
		tag := fld.Tag.Get("json")
		if tag == "" || tag == "-" {
			return ""
		}
		if idx := strings.Index(tag, ","); idx != -1 {
			tag = tag[:idx]
		}
		return tag
	})

	// Register custom validations
	if err := v.RegisterValidation("url", validateURL); err != nil {
		return nil, fmt.Errorf("failed to register url validation: %w", err)
	}
	if err := v.RegisterValidation("password", validatePassword); err != nil {
		return nil, fmt.Errorf("failed to register password validation: %w", err)
	}
	if err := v.RegisterValidation("date", validateDate); err != nil {
		return nil, fmt.Errorf("failed to register date validation: %w", err)
	}
	if err := v.RegisterValidation("datetime", validateDateTime); err != nil {
		return nil, fmt.Errorf("failed to register datetime validation: %w", err)
	}

	return &validatorImpl{validate: v}, nil
}

// validateURL validates if a string is a valid URL
func validateURL(fl validator.FieldLevel) bool {
	urlStr := fl.Field().String()
	if urlStr == "" {
		return true // Empty URLs are handled by required tag
	}
	_, err := url.ParseRequestURI(urlStr)
	return err == nil
}

// validatePassword validates if a string meets password requirements
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if password == "" {
		return true // Empty passwords are handled by required tag
	}

	// Password requirements:
	// - At least 8 characters
	// - At least one uppercase letter
	// - At least one lowercase letter
	// - At least one number
	// - At least one special character
	hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
	hasNumber := strings.ContainsAny(password, "0123456789")
	hasSpecial := strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?")

	return len(password) >= 8 && hasUpper && hasLower && hasNumber && hasSpecial
}

// validateDate validates if a string is a valid date in YYYY-MM-DD format
func validateDate(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	if dateStr == "" {
		return true // Empty dates are handled by required tag
	}
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

// validateDateTime validates if a string is a valid datetime in YYYY-MM-DD HH:mm:ss format
func validateDateTime(fl validator.FieldLevel) bool {
	datetimeStr := fl.Field().String()
	if datetimeStr == "" {
		return true // Empty datetimes are handled by required tag
	}
	_, err := time.Parse("2006-01-02 15:04:05", datetimeStr)
	return err == nil
}

// Struct validates a struct and caches the result
func (v *validatorImpl) Struct(i any) error {
	// Generate cache key
	cacheKey := fmt.Sprintf("%T", i)

	// Check cache first
	if cached, ok := v.cache.Load(cacheKey); ok {
		if err, ok := cached.(error); ok {
			return err
		}
	}

	// Validate the struct
	err := v.validate.Struct(i)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			domainErrs := make(ValidationErrors, 0, len(validationErrors))
			for _, e := range validationErrors {
				domainErrs = append(domainErrs, ValidationError{
					Field:   getFieldName(e),
					Message: getErrorMessage(e),
					Value:   e.Value(),
				})
			}
			err = domainErrs
		}
	}

	// Cache the result
	v.cache.Store(cacheKey, err)

	return err
}

// Var validates a single variable
func (v *validatorImpl) Var(i any, tag string) error {
	return v.validate.Var(i, tag)
}

// RegisterValidation registers a custom validation function
func (v *validatorImpl) RegisterValidation(tag string, fn func(fl validator.FieldLevel) bool) error {
	return v.validate.RegisterValidation(tag, fn)
}

// RegisterCrossFieldValidation registers a custom cross-field validation function
func (v *validatorImpl) RegisterCrossFieldValidation(tag string, fn func(fl validator.FieldLevel) bool) error {
	return v.validate.RegisterValidation(tag, fn)
}

// RegisterStructValidation registers a custom struct validation function
func (v *validatorImpl) RegisterStructValidation(fn func(sl validator.StructLevel), types any) error {
	v.validate.RegisterStructValidation(fn, types)
	return nil
}

// GetValidationErrors returns a map of field names to error messages
func (v *validatorImpl) GetValidationErrors(err error) map[string]string {
	if err == nil {
		return nil
	}

	// Handle our custom ValidationErrors type
	var validationErrors ValidationErrors
	if errors.As(err, &validationErrors) {
		errors := make(map[string]string)
		for _, e := range validationErrors {
			errors[e.Field] = e.Message
		}
		return errors
	}

	// Handle validator.ValidationErrors
	var validatorErrors validator.ValidationErrors
	if errors.As(err, &validatorErrors) {
		errors := make(map[string]string)
		for _, e := range validatorErrors {
			errors[getFieldName(e)] = getErrorMessage(e)
		}
		return errors
	}

	// Handle other errors
	return map[string]string{
		"error": err.Error(),
	}
}

// ValidateStruct validates a struct and returns domain errors
func (v *validatorImpl) ValidateStruct(s any) error {
	if err := v.Struct(s); err != nil {
		var validationErrors ValidationErrors
		if errors.As(err, &validationErrors) {
			return validationErrors
		}
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}
