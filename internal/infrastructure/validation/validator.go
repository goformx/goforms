package validation

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	validator "github.com/go-playground/validator/v10"
	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
)

const (
	jsonTagSplitLimit = 2
	// Common validation patterns
	phoneRegex = `^\+?[1-9]\d{1,14}$`
)

var (
	// Common validation regexes
	phonePattern = regexp.MustCompile(phoneRegex)
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

// Validator defines the interface for validation
type Validator interface {
	// Struct validates a struct against its validation tags
	Struct(any) error
	// Var validates a single variable against the provided tag
	Var(any, string) error
	// RegisterValidation registers a custom validation function
	RegisterValidation(string, func(fl validator.FieldLevel) bool) error
	// GetValidationErrors returns a map of field names to error messages
	GetValidationErrors(error) map[string]string
	// RegisterCrossFieldValidation registers a cross-field validation function
	RegisterCrossFieldValidation(string, func(fl validator.FieldLevel) bool) error
	// RegisterStructValidation registers a struct validation function
	RegisterStructValidation(func(sl validator.StructLevel), any) error
}

// validatorImpl implements the Validator interface
type validatorImpl struct {
	validate *validator.Validate
	cache    sync.Map // Cache for validation results
}

//nolint:gochecknoglobals // singleton pattern requires global instance and once
var (
	instance *validatorImpl
	once     sync.Once
)

// New creates a new validator instance with common validation rules
func New() Validator {
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

		// Register custom validations
		_ = v.RegisterValidation("url", validateURL)
		_ = v.RegisterValidation("phone", validatePhone)
		_ = v.RegisterValidation("password", validatePassword)
		_ = v.RegisterValidation("date", validateDate)
		_ = v.RegisterValidation("datetime", validateDateTime)

		instance = &validatorImpl{validate: v}
	})
	return instance
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

// validatePhone validates if a string is a valid phone number
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true // Empty phone numbers are handled by required tag
	}
	return phonePattern.MatchString(phone)
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
		if cachedErr, isErr := cached.(error); isErr && cachedErr != nil {
			return cachedErr
		}
		return nil
	}

	err := v.validate.Struct(i)
	if err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			validationErrors := make([]ValidationError, len(ve))
			for i, e := range ve {
				validationErrors[i] = ValidationError{
					Field:   getFieldName(e),
					Message: getErrorMessage(e),
					Value:   e.Value(),
				}
			}
			err = domainerrors.New(domainerrors.ErrCodeValidation, "validation failed", ValidationErrors(validationErrors))
		}
		// Cache the error
		v.cache.Store(cacheKey, err)
		return err
	}

	// Cache successful validation
	v.cache.Store(cacheKey, nil)
	return nil
}

// Var validates a single variable
func (v *validatorImpl) Var(i any, tag string) error {
	return v.validate.Var(i, tag)
}

// RegisterValidation registers a custom validation function
func (v *validatorImpl) RegisterValidation(tag string, fn func(fl validator.FieldLevel) bool) error {
	return v.validate.RegisterValidation(tag, fn)
}

// RegisterCrossFieldValidation registers a cross-field validation function
func (v *validatorImpl) RegisterCrossFieldValidation(tag string, fn func(fl validator.FieldLevel) bool) error {
	return v.validate.RegisterValidation(tag, fn)
}

// RegisterStructValidation registers a struct validation function
func (v *validatorImpl) RegisterStructValidation(fn func(sl validator.StructLevel), types any) error {
	v.validate.RegisterStructValidation(fn, types)
	return nil
}

// GetValidationErrors returns detailed validation errors
func (v *validatorImpl) GetValidationErrors(err error) map[string]string {
	if err == nil {
		return nil
	}

	validationErrors := make(map[string]string)
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, e := range ve {
			field := e.Field()
			validationErrors[field] = getErrorMessage(e)
		}
	} else {
		// Handle non-validation errors
		validationErrors["_error"] = err.Error()
	}
	return validationErrors
}

// ValidateStruct validates a struct and returns validation errors
func (v *validatorImpl) ValidateStruct(s any) error {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		validationErrors := make([]ValidationError, len(ve))
		for i, err := range ve {
			validationErrors[i] = ValidationError{
				Field:   err.Field(),
				Message: getErrorMessage(err),
				Value:   err.Value(),
			}
		}
		return ValidationErrors(validationErrors)
	}
	return err
}
