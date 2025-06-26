package web

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/view"
)

// AuthResponseBuilder handles authentication-related HTTP responses
type AuthResponseBuilder struct {
	Renderer view.Renderer
}

// NewAuthResponseBuilder creates a new AuthResponseBuilder
func NewAuthResponseBuilder(renderer view.Renderer) *AuthResponseBuilder {
	return &AuthResponseBuilder{Renderer: renderer}
}

// AJAXError returns a JSON error response for AJAX requests
func (b *AuthResponseBuilder) AJAXError(c echo.Context, status int, message string) error {
	return response.ErrorResponse(c, status, message)
}

// HTMLFormError renders the form page with an error message
func (b *AuthResponseBuilder) HTMLFormError(c echo.Context, page string, data *view.PageData, message string) error {
	data.Message = &view.Message{
		Type: "error",
		Text: message,
	}
	switch page {
	case "login":
		if err := b.Renderer.Render(c, pages.Login(*data)); err != nil {
			return fmt.Errorf("render login page: %w", err)
		}
		return nil
	case "signup":
		if err := b.Renderer.Render(c, pages.Signup(*data)); err != nil {
			return fmt.Errorf("render signup page: %w", err)
		}
		return nil
	default:
		if err := b.Renderer.Render(c, pages.Error(*data)); err != nil {
			return fmt.Errorf("render error page: %w", err)
		}
		return nil
	}
}

// Redirect returns a redirect response
func (b *AuthResponseBuilder) Redirect(c echo.Context, location string) error {
	return fmt.Errorf("redirect to location: %w", c.Redirect(http.StatusSeeOther, location))
}
