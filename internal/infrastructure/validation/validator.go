// Package validation provides infrastructure-level validation utilities and interfaces.
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
	"github.com/goformx/goforms/internal/domain/common/interfaces"
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

// Error represents a single validation error
type Error struct {
	Field   string
	Message string
	Value   any // The invalid value that caused the error
}

// Errors represents a collection of validation errors
type Errors []Error

// Error implements the error interface
func (e Errors) Error() string {
	var sb strings.Builder

	for i, err := range e {
		if i > 0 {
			sb.WriteString("; ")
		}

		_, _ = fmt.Fprintf(&sb, "%s: %s", err.Field, err.Message)
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

	message, exists := errorMessages[e.Tag()]
	if !exists {
		return fmt.Sprintf("%s failed validation: %s", field, e.Tag())
	}

	return message(field, e.Param())
}

// errorMessages maps validation tags to their message generators
var errorMessages = map[string]func(field, param string) string{
	"required": func(field, _ string) string {
		return fmt.Sprintf("%s is required", field)
	},
	"email": func(field, _ string) string {
		return fmt.Sprintf("%s must be a valid email address", field)
	},
	"min": func(field, param string) string {
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	},
	"max": func(field, param string) string {
		return fmt.Sprintf("%s must be at most %s characters", field, param)
	},
	"len": func(field, param string) string {
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	},
	"oneof": func(field, param string) string {
		return fmt.Sprintf("%s must be one of [%s]", field, param)
	},
	"url": func(field, _ string) string {
		return fmt.Sprintf("%s must be a valid URL", field)
	},
	"phone": func(field, _ string) string {
		return fmt.Sprintf("%s must be a valid phone number", field)
	},
	"password": func(field, _ string) string {
		return fmt.Sprintf("%s must contain at least 8 characters, including uppercase, lowercase, "+
			"number and special character", field)
	},
	"date": func(field, _ string) string {
		return fmt.Sprintf("%s must be a valid date in format YYYY-MM-DD", field)
	},
	"datetime": func(field, _ string) string {
		return fmt.Sprintf("%s must be a valid datetime in format YYYY-MM-DD HH:mm:ss", field)
	},
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
		name := strings.SplitN(fld.Tag.Get("json"), ",", jsonTagSplitLimit)[0]
		if name == "-" {
			return ""
		}

		return name
	})

	// Register custom validations
	if err := v.RegisterValidation("url", validateURL); err != nil {
		return nil, fmt.Errorf("failed to register url validation: %w", err)
	}

	if err := v.RegisterValidation("phone", validatePhone); err != nil {
		return nil, fmt.Errorf("failed to register phone validation: %w", err)
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
			validationErrors := make([]Error, len(ve))
			for i, e := range ve {
				validationErrors[i] = Error{
					Field:   getFieldName(e),
					Message: getErrorMessage(e),
					Value:   e.Value(),
				}
			}

			err = domainerrors.New(
				domainerrors.ErrCodeValidation,
				"validation failed",
				Errors(validationErrors),
			)
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
	if err := v.validate.Var(i, tag); err != nil {
		return fmt.Errorf("validate variable: %w", err)
	}

	return nil
}

// RegisterValidation registers a custom validation function
func (v *validatorImpl) RegisterValidation(tag string, fn func(fl validator.FieldLevel) bool) error {
	if err := v.validate.RegisterValidation(tag, fn); err != nil {
		return fmt.Errorf("register validation: %w", err)
	}

	return nil
}

// RegisterCrossFieldValidation registers a cross-field validation function
func (v *validatorImpl) RegisterCrossFieldValidation(tag string, fn func(fl validator.FieldLevel) bool) error {
	if err := v.validate.RegisterValidation(tag, fn); err != nil {
		return fmt.Errorf("register cross-field validation: %w", err)
	}

	return nil
}

// RegisterStructValidation registers a struct validation function
func (v *validatorImpl) RegisterStructValidation(fn func(sl validator.StructLevel), types any) error {
	v.validate.RegisterStructValidation(fn, types)

	return nil
}

// GetErrors returns detailed validation errors
func (v *validatorImpl) GetErrors(err error) map[string]string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil
	}

	validationErrors := make(map[string]string)
	for _, e := range ve {
		validationErrors[getFieldName(e)] = getErrorMessage(e)
	}

	return validationErrors
}

// ValidateStruct validates a struct and returns any validation errors
func (v *validatorImpl) ValidateStruct(s any) error {
	return v.Struct(s)
}
