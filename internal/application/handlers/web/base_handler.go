package web

import (
	"net/http"

	"github.com/goformx/goforms/internal/domain/form"
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

// FormService returns the form service instance
func (h *BaseHandler) FormService() form.Service {
	return h.formService
}
