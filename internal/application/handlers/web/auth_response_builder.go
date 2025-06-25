package web

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/view"
)

type AuthResponseBuilder struct {
	Renderer view.Renderer
}

func NewAuthResponseBuilder(renderer view.Renderer) *AuthResponseBuilder {
	return &AuthResponseBuilder{Renderer: renderer}
}

// AJAXError returns a JSON error response for AJAX requests
func (b *AuthResponseBuilder) AJAXError(c echo.Context, status int, message string) error {
	return c.JSON(status, map[string]string{
		"message": message,
	})
}

// HTMLFormError renders the form page with an error message
func (b *AuthResponseBuilder) HTMLFormError(c echo.Context, page string, data *view.PageData, message string) error {
	data.Message = &view.Message{
		Type: "error",
		Text: message,
	}
	switch page {
	case "login":
		return b.Renderer.Render(c, pages.Login(*data))
	case "signup":
		return b.Renderer.Render(c, pages.Signup(*data))
	default:
		return b.Renderer.Render(c, pages.Error(*data))
	}
}

// Redirect returns a redirect response
func (b *AuthResponseBuilder) Redirect(c echo.Context, location string) error {
	return c.Redirect(http.StatusSeeOther, location)
}
