package web

import (
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

// Validate checks for required dependencies by name
func (d *HandlerDeps) Validate(required ...string) error {
	missing := []string{}
	for _, dep := range required {
		switch dep {
		case "BaseHandler":
			if d.BaseHandler == nil {
				missing = append(missing, "BaseHandler")
			}
		case "UserService":
			if d.UserService == nil {
				missing = append(missing, "UserService")
			}
		case "SessionManager":
			if d.SessionManager == nil {
				missing = append(missing, "SessionManager")
			}
		case "Renderer":
			if d.Renderer == nil {
				missing = append(missing, "Renderer")
			}
		case "MiddlewareManager":
			if d.MiddlewareManager == nil {
				missing = append(missing, "MiddlewareManager")
			}
		case "Config":
			if d.Config == nil {
				missing = append(missing, "Config")
			}
		case "Logger":
			if d.Logger == nil {
				missing = append(missing, "Logger")
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required dependencies: %v", missing)
	}
	return nil
}

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
