// Package web provides HTTP handlers for web-based functionality including
// authentication, form management, and user interface components.
package web

import (
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/form/model"
)

// AuthHelper handles common authentication and authorization patterns
type AuthHelper struct {
	baseHandler *FormBaseHandler
}

// NewAuthHelper creates a new AuthHelper instance
func NewAuthHelper(baseHandler *FormBaseHandler) *AuthHelper {
	return &AuthHelper{
		baseHandler: baseHandler,
	}
}

// RequireAuthenticatedUser ensures the user is authenticated and returns the user object
func (h *AuthHelper) RequireAuthenticatedUser(c echo.Context) (*entities.User, error) {
	return h.baseHandler.RequireAuthenticatedUser(c)
}

// GetFormWithOwnership gets a form and verifies ownership in one call
func (h *AuthHelper) GetFormWithOwnership(c echo.Context) (*model.Form, error) {
	return h.baseHandler.GetFormWithOwnership(c)
}

// RequireFormOwnership verifies the user owns the form
func (h *AuthHelper) RequireFormOwnership(c echo.Context, form *model.Form) error {
	return h.baseHandler.RequireFormOwnership(c, form)
}
