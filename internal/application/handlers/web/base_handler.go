package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// HandlerDeps centralizes common handler dependencies and validation
// Add new dependencies here as needed
// Usage: pass HandlerDeps to each handler's constructor
// and call Validate with required fields

type HandlerDeps struct {
	BaseHandler       *BaseHandler
	UserService       domain.UserService
	SessionManager    *middleware.SessionManager
	Renderer          *view.Renderer
	MiddlewareManager *middleware.Manager
	Config            *config.Config
	Logger            logging.Logger
}

// validateField checks if a field is present in the handler dependencies
func (d *HandlerDeps) validateField(field string) error {
	switch field {
	case "BaseHandler":
		if d.BaseHandler == nil {
			return errors.New("BaseHandler is required")
		}
	case "UserService":
		if d.UserService == nil {
			return errors.New("UserService is required")
		}
	case "SessionManager":
		if d.SessionManager == nil {
			return errors.New("SessionManager is required")
		}
	case "Renderer":
		if d.Renderer == nil {
			return errors.New("renderer is required")
		}
	case "MiddlewareManager":
		if d.MiddlewareManager == nil {
			return errors.New("MiddlewareManager is required")
		}
	case "Config":
		if d.Config == nil {
			return errors.New("config is required")
		}
	case "Logger":
		if d.Logger == nil {
			return errors.New("Logger is required")
		}
	default:
		return fmt.Errorf("unknown required field: %s", field)
	}
	return nil
}

// Validate checks if all required fields are present
func (d *HandlerDeps) Validate(required ...string) error {
	for _, field := range required {
		if err := d.validateField(field); err != nil {
			return err
		}
	}
	return nil
}

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	formService form.Service
	logger      logging.Logger
	middlewares []echo.MiddlewareFunc
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
	middlewares ...echo.MiddlewareFunc,
) {
	switch method {
	case "GET":
		e.GET(path, handler, middlewares...)
	case "POST":
		e.POST(path, handler, middlewares...)
	case "PUT":
		e.PUT(path, handler, middlewares...)
	case "DELETE":
		e.DELETE(path, handler, middlewares...)
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

// WithMiddleware adds middleware to the handler
func (h *BaseHandler) WithMiddleware(
	mwFuncs ...echo.MiddlewareFunc,
) *BaseHandler {
	h.middlewares = append(h.middlewares, mwFuncs...)
	return h
}
