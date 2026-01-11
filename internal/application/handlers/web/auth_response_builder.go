package web

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/presentation/inertia"
	"github.com/goformx/goforms/internal/presentation/view"
)

// AuthResponseBuilder handles authentication-related HTTP responses
type AuthResponseBuilder struct {
	Renderer view.Renderer
	Inertia  *inertia.Manager
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
// Note: With Inertia, errors are returned as JSON and handled by the frontend
func (b *AuthResponseBuilder) HTMLFormError(c echo.Context, page string, data *view.PageData, message string) error {
	// For Inertia-based apps, return JSON error
	return response.ErrorResponse(c, http.StatusBadRequest, message)
}

// Redirect returns a redirect response
func (b *AuthResponseBuilder) Redirect(c echo.Context, location string) error {
	return fmt.Errorf("redirect to location: %w", c.Redirect(http.StatusSeeOther, location))
}
