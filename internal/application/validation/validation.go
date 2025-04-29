package validation

import (
	"errors"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	minPasswordLength = 8
)

var (
	validate = validator.New()
	// UsernameRegex is a regex for valid usernames
	UsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`)
	// EmailRegex is a regex for valid email addresses
	EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// RegisterCustomValidations registers custom validation functions
func RegisterCustomValidations() error {
	if err := validate.RegisterValidation("username", validateUsername); err != nil {
		return err
	}
	if err := validate.RegisterValidation("password", validatePassword); err != nil {
		return err
	}
	return nil
}

// validateUsername checks if a username is valid
func validateUsername(fl validator.FieldLevel) bool {
	return UsernameRegex.MatchString(fl.Field().String())
}

// validatePassword checks if a password is valid
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < minPasswordLength {
		return false
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case strings.ContainsAny(string(char), "!@#$%^&*()_+-=[]{}|;:,.<>?"):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// SignupRequest represents the signup form data
type SignupRequest struct {
	Username        string `json:"username" validate:"required,username"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,password"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// LoginRequest represents the login form data
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// ValidateSignup validates the signup request
func ValidateSignup(req *SignupRequest) error {
	return validate.Struct(req)
}

// ValidateLogin validates the login request
func ValidateLogin(req *LoginRequest) error {
	return validate.Struct(req)
}

// GetValidationErrors returns a map of validation errors
func GetValidationErrors(err error) map[string]string {
	errs := make(map[string]string)
	if err == nil {
		return errs
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return errs
	}

	for _, e := range validationErrors {
		field := strings.ToLower(e.Field())
		switch e.Tag() {
		case "required":
			errs[field] = "This field is required"
		case "username":
			errs[field] = "Username must be 3-50 characters and can only contain letters, numbers, and underscores"
		case "email":
			errs[field] = "Please enter a valid email address"
		case "password":
			errs[field] = "Password must be at least 8 characters and " +
				"contain uppercase, lowercase, number, and special character"
		case "eqfield":
			errs[field] = "Passwords do not match"
		default:
			errs[field] = "Invalid value"
		}
	}
	return errs
}
 