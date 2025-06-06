// Package main is the entry point for the GoFormX application.
// It sets up the application using dependency injection (via fx),
// configures the server, and manages the application lifecycle.
package web

import (
	"errors"
	"fmt"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/view"
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
			return errors.New("base handler is required")
		}
	case "UserService":
		if d.UserService == nil {
			return errors.New("user service is required")
		}
	case "SessionManager":
		if d.SessionManager == nil {
			return errors.New("session manager is required")
		}
	case "Renderer":
		if d.Renderer == nil {
			return errors.New("renderer is required")
		}
	case "MiddlewareManager":
		if d.MiddlewareManager == nil {
			return errors.New("middleware manager is required")
		}
	case "Config":
		if d.Config == nil {
			return errors.New("config is required")
		}
	case "Logger":
		if d.Logger == nil {
			return errors.New("logger is required")
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
