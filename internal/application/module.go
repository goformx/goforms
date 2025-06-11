// Package application provides the application layer components and their dependency injection setup.
package application

import (
	"errors"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/application/middleware"
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
	SessionManager    *middleware.SessionManager
	MiddlewareManager *middleware.Manager
	Renderer          view.Renderer
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
		{"SessionManager", d.SessionManager},
		{"MiddlewareManager", d.MiddlewareManager},
		{"Renderer", d.Renderer},
	}

	for _, r := range required {
		if r.value == nil {
			return errors.New(r.name + " is required")
		}
	}
	return nil
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
		MiddlewareManager: deps.MiddlewareManager,
		Config:            deps.Config,
		Logger:            deps.Logger,
		Renderer:          deps.Renderer,
	}, nil
}

// Module provides all application layer dependencies
var Module = fx.Options(
	// Handler dependencies
	fx.Provide(
		NewHandlerDeps,
	),

	// Handlers
	fx.Provide(
		// Login handler
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

		// Public web page handlers
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

		// Dashboard handler
		fx.Annotate(
			func(deps *web.HandlerDeps) (web.Handler, error) {
				handler := web.NewDashboardHandler(*deps)
				return handler, nil
			},
			fx.ResultTags(`group:"handlers"`),
		),
	),
)
