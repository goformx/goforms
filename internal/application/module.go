// Package application provides the application layer components and their dependency injection setup.
package application

import (
	"errors"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/services/auth"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/view"
)

// Dependencies contains all application layer dependencies
type Dependencies struct {
	fx.In

	// Domain services
	UserService user.Service
	FormService form.Service

	// Infrastructure
	Logger            logging.Logger
	Config            *config.Config
	Renderer          *view.Renderer
	SessionManager    *middleware.SessionManager
	MiddlewareManager *middleware.Manager
}

// Validate checks if all required dependencies are present
func (d *Dependencies) Validate() error {
	required := []struct {
		name  string
		value any
	}{
		{"UserService", d.UserService},
		{"FormService", d.FormService},
		{"Logger", d.Logger},
		{"Config", d.Config},
		{"Renderer", d.Renderer},
		{"SessionManager", d.SessionManager},
		{"MiddlewareManager", d.MiddlewareManager},
	}

	for _, r := range required {
		if r.value == nil {
			return errors.New(r.name + " is required")
		}
	}
	return nil
}

// NewAuthService creates a new auth service
func NewAuthService(deps Dependencies) (auth.Service, error) {
	if err := deps.Validate(); err != nil {
		return nil, err
	}
	return auth.NewService(deps.UserService, deps.Logger), nil
}

// NewHandlerDeps creates handler dependencies
func NewHandlerDeps(deps Dependencies) (*web.HandlerDeps, error) {
	if err := deps.Validate(); err != nil {
		return nil, err
	}

	return &web.HandlerDeps{
		UserService:       deps.UserService,
		FormService:       deps.FormService,
		SessionManager:    deps.SessionManager,
		Renderer:          deps.Renderer,
		MiddlewareManager: deps.MiddlewareManager,
		Config:            deps.Config,
		Logger:            deps.Logger,
	}, nil
}

// Module provides all application layer dependencies
var Module = fx.Options(
	// Services
	fx.Provide(
		fx.Annotate(
			NewAuthService,
			fx.As(new(auth.Service)),
		),
	),

	// Handler dependencies
	fx.Provide(
		NewHandlerDeps,
	),

	// Handlers
	fx.Provide(
		// Auth handler
		fx.Annotate(
			func(deps *web.HandlerDeps) (web.Handler, error) {
				handler, err := web.NewAuthHandler(*deps)
				if err != nil {
					return nil, err
				}
				return handler, nil
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Web handler
		fx.Annotate(
			func(deps *web.HandlerDeps) (web.Handler, error) {
				handler, err := web.NewWebHandler(*deps)
				if err != nil {
					return nil, err
				}
				return handler, nil
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Form handler
		fx.Annotate(
			func(deps *web.HandlerDeps) (web.Handler, error) {
				handler := web.NewFormHandler(*deps, deps.FormService)
				return handler, nil
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Demo handler
		fx.Annotate(
			func(deps *web.HandlerDeps) (web.Handler, error) {
				handler, err := web.NewDemoHandler(*deps)
				if err != nil {
					return nil, err
				}
				return handler, nil
			},
			fx.ResultTags(`group:"handlers"`),
		),
	),
)
