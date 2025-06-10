package web

import (
	"errors"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/auth"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/view"
)

// HandlerDeps contains all dependencies required by web handlers
type HandlerDeps struct {
	// Domain services
	UserService user.Service
	FormService form.Service
	AuthService auth.Service

	// Infrastructure
	Logger            logging.Logger
	Config            *config.Config
	SessionManager    *middleware.SessionManager
	MiddlewareManager *middleware.Manager
	Renderer          view.Renderer
}

// validateField checks if a field is nil and returns an error if it is
func (d *HandlerDeps) validateField(name string, value any) error {
	if value == nil {
		return errors.New(name + " is required")
	}
	return nil
}

// Validate checks if all required dependencies are present
func (d *HandlerDeps) Validate() error {
	required := []struct {
		name  string
		value any
	}{
		{"UserService", d.UserService},
		{"FormService", d.FormService},
		{"AuthService", d.AuthService},
		{"Logger", d.Logger},
		{"Config", d.Config},
		{"SessionManager", d.SessionManager},
		{"MiddlewareManager", d.MiddlewareManager},
		{"Renderer", d.Renderer},
	}

	for _, r := range required {
		if err := d.validateField(r.name, r.value); err != nil {
			return err
		}
	}
	return nil
}

// HandlerParams contains parameters for creating a handler
type HandlerParams struct {
	UserService       user.Service
	FormService       form.Service
	AuthService       auth.Service
	Logger            logging.Logger
	Config            *config.Config
	SessionManager    *middleware.SessionManager
	MiddlewareManager *middleware.Manager
	Renderer          view.Renderer
}

// NewHandlerDeps creates a new HandlerDeps instance
func NewHandlerDeps(params HandlerParams) (*HandlerDeps, error) {
	deps := &HandlerDeps{
		UserService:       params.UserService,
		FormService:       params.FormService,
		AuthService:       params.AuthService,
		Logger:            params.Logger,
		Config:            params.Config,
		SessionManager:    params.SessionManager,
		MiddlewareManager: params.MiddlewareManager,
		Renderer:          params.Renderer,
	}

	if err := deps.Validate(); err != nil {
		return nil, err
	}

	return deps, nil
}
