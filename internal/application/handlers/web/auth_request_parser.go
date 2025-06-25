package web

import (
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/domain/user"
)

// AuthRequestParser parses authentication requests.
type AuthRequestParser struct{}

// NewAuthRequestParser creates a new AuthRequestParser.
func NewAuthRequestParser() *AuthRequestParser {
	return &AuthRequestParser{}
}

// ParseLogin parses login credentials from the request (JSON or form)
func (p *AuthRequestParser) ParseLogin(c echo.Context) (email, password string, err error) {
	contentType := c.Request().Header.Get("Content-Type")
	if contentType == "application/json" {
		var data struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if bindErr := c.Bind(&data); bindErr != nil {
			return "", "", bindErr
		}
		email = data.Email
		password = data.Password
	} else {
		email = c.FormValue("email")
		password = c.FormValue("password")
	}
	return email, password, nil
}

// ParseSignup parses signup data from the request (JSON or form)
func (p *AuthRequestParser) ParseSignup(c echo.Context) (user.Signup, error) {
	contentType := c.Request().Header.Get("Content-Type")
	var signup user.Signup
	if contentType == "application/json" {
		if err := c.Bind(&signup); err != nil {
			return signup, err
		}
	} else {
		signup = user.Signup{
			Email:           c.FormValue("email"),
			Password:        c.FormValue("password"),
			ConfirmPassword: c.FormValue("confirm_password"),
		}
	}
	return signup, nil
}
