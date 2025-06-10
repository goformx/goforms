package web

import (
	"errors"
	"fmt"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/services/auth"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/view"
	"go.uber.org/fx"
)

// HandlerDeps contains all dependencies required by web handlers
type HandlerDeps struct {
	fx.In

	UserService       user.Service
	FormService       form.Service
	AuthService       auth.Service
	SessionManager    *middleware.SessionManager
	Renderer          *view.Renderer
	MiddlewareManager *middleware.Manager
	Config            *config.Config
	Logger            logging.Logger
}

// validateField checks if a field is present in the handler dependencies
func (d *HandlerDeps) validateField(field string) error {
	switch field {
	case "UserService":
		if d.UserService == nil {
			return errors.New("user service is required")
		}
	case "FormService":
		if d.FormService == nil {
			return errors.New("form service is required")
		}
	case "AuthService":
		if d.AuthService == nil {
			return errors.New("auth service is required")
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

// HandlerParams contains dependencies for creating handlers
type HandlerParams struct {
	fx.In

	UserService       user.Service
	FormService       form.Service
	AuthService       auth.Service
	SessionManager    *middleware.SessionManager
	Renderer          *view.Renderer
	MiddlewareManager *middleware.Manager
	Config            *config.Config
	Logger            logging.Logger
}

// NewHandlerDeps creates a new HandlerDeps instance with proper error handling
func NewHandlerDeps(p HandlerParams) (*HandlerDeps, error) {
	deps := &HandlerDeps{
		UserService:       p.UserService,
		FormService:       p.FormService,
		AuthService:       p.AuthService,
		SessionManager:    p.SessionManager,
		Renderer:          p.Renderer,
		MiddlewareManager: p.MiddlewareManager,
		Config:            p.Config,
		Logger:            p.Logger,
	}

	// Validate all required dependencies
	if err := deps.Validate(
		"UserService",
		"FormService",
		"AuthService",
		"SessionManager",
		"Renderer",
		"MiddlewareManager",
		"Config",
		"Logger",
	); err != nil {
		return nil, fmt.Errorf("failed to create handler dependencies: %w", err)
	}

	return deps, nil
}
