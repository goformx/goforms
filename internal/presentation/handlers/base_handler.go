package handlers

import (
	"net/http"

	amw "github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	authMiddleware *amw.CookieAuthMiddleware
	formService    form.Service
	logger         logging.Logger
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(
	authMiddleware *amw.CookieAuthMiddleware,
	formService form.Service,
	logger logging.Logger,
) *BaseHandler {
	return &BaseHandler{
		authMiddleware: authMiddleware,
		formService:    formService,
		logger:         logger,
	}
}

// SetupMiddleware sets up common middleware for a route group
func (h *BaseHandler) SetupMiddleware(group *echo.Group) {
	group.Use(h.authMiddleware.RequireAuth)
}

// getAuthenticatedUser retrieves and validates the authenticated user from the context
func (h *BaseHandler) getAuthenticatedUser(c echo.Context) (*user.User, error) {
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}
	return currentUser, nil
}

// getOwnedForm retrieves a form and verifies ownership
func (h *BaseHandler) getOwnedForm(c echo.Context, currentUser *user.User) (*form.Form, error) {
	formID := c.Param("id")
	if formID == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Form ID is required")
	}

	formObj, err := h.formService.GetForm(formID)
	if err != nil {
		h.LogError("Failed to get form", err)
		return nil, echo.NewHTTPError(http.StatusNotFound, "Form not found")
	}

	if formObj.UserID != currentUser.ID {
		return nil, echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	return formObj, nil
}

// handleError is a helper function to consistently handle and log errors
func (h *BaseHandler) handleError(err error, status int, message string) error {
	h.LogError(message, err)
	return echo.NewHTTPError(status, message)
}

// LogError logs an error with consistent formatting
func (h *BaseHandler) LogError(message string, err error) {
	h.logger.Error(message,
		logging.Error(err),
		logging.StringField("operation", "handler_error"),
	)
}
