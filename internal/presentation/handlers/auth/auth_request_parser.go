package auth

import (
	"fmt"
	"strings"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/labstack/echo/v4"
)

// AuthRequestParser parses authentication requests.
type AuthRequestParser struct{}

// NewAuthRequestParser creates a new auth request parser
func NewAuthRequestParser() *AuthRequestParser {
	return &AuthRequestParser{}
}

// ParseLogin parses login credentials from the request (JSON or form)
func (p *AuthRequestParser) ParseLogin(c echo.Context) (email, password string, err error) {
	contentType := c.Request().Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		var data struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if bindErr := c.Bind(&data); bindErr != nil {
			return "", "", fmt.Errorf("failed to bind login: %w", bindErr)
		}

		email = data.Email
		password = data.Password
	} else {
		email = c.FormValue("email")
		password = c.FormValue("password")
	}

	// Sanitize inputs
	email = strings.TrimSpace(strings.ToLower(email))

	return email, password, nil
}

// ParseSignup parses signup data from the request (JSON or form)
func (p *AuthRequestParser) ParseSignup(c echo.Context) (user.Signup, error) {
	contentType := c.Request().Header.Get("Content-Type")

	var signup user.Signup

	if strings.Contains(contentType, "application/json") {
		if err := c.Bind(&signup); err != nil {
			return signup, fmt.Errorf("failed to bind signup: %w", err)
		}
	} else {
		signup = user.Signup{
			Email:           c.FormValue("email"),
			Password:        c.FormValue("password"),
			ConfirmPassword: c.FormValue("confirm_password"),
		}
	}

	// Sanitize inputs
	signup.Email = strings.TrimSpace(strings.ToLower(signup.Email))

	return signup, nil
}

// ValidateLogin validates login credentials
func (p *AuthRequestParser) ValidateLogin(email, password string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	if password == "" {
		return fmt.Errorf("password is required")
	}

	return nil
}

// ValidateSignup validates signup data
func (p *AuthRequestParser) ValidateSignup(signup user.Signup) error {
	if signup.Email == "" {
		return fmt.Errorf("email is required")
	}

	if signup.Password == "" {
		return fmt.Errorf("password is required")
	}

	if signup.ConfirmPassword == "" {
		return fmt.Errorf("password confirmation is required")
	}

	if signup.Password != signup.ConfirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	return nil
}
