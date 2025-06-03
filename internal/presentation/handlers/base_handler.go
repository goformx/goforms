package handlers

import (
	"net/http"

	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	formService form.Service
	logger      logging.Logger
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(
	formService form.Service,
	logger logging.Logger,
) *BaseHandler {
	return &BaseHandler{
		formService: formService,
		logger:      logger,
	}
}

// RegisterRoute is a helper method to register routes with middleware
func (h *BaseHandler) RegisterRoute(
	e *echo.Echo,
	method, path string,
	handler echo.HandlerFunc,
	middleware ...echo.MiddlewareFunc,
) {
	switch method {
	case "GET":
		e.GET(path, handler, middleware...)
	case "POST":
		e.POST(path, handler, middleware...)
	case "PUT":
		e.PUT(path, handler, middleware...)
	case "DELETE":
		e.DELETE(path, handler, middleware...)
	}
	h.logger.Debug("registered route",
		logging.StringField("method", method),
		logging.StringField("path", path),
	)
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

// LogDebug logs a debug message with consistent formatting
func (h *BaseHandler) LogDebug(message string, fields ...any) {
	h.logger.Debug(message, fields...)
}

// LogInfo logs an info message with consistent formatting
func (h *BaseHandler) LogInfo(message string, fields ...any) {
	h.logger.Info(message, fields...)
}

// WrapResponseError provides consistent error wrapping for HTTP responses
func (h *BaseHandler) WrapResponseError(err error, msg string) error {
	if err == nil {
		return nil
	}
	h.LogError(msg, err)
	return echo.NewHTTPError(http.StatusInternalServerError, msg)
}

// Validate ensures all required dependencies are properly set
func (h *BaseHandler) Validate() error {
	if h.logger == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "logger is required")
	}
	if h.formService == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "form service is required")
	}
	return nil
}

// Logger returns the logger instance
func (h *BaseHandler) Logger() logging.Logger {
	return h.logger
}
