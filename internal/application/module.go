// Package application provides the application layer components and their dependency injection setup.
package application

import (
	"errors"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/middleware/session"
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
	SessionManager    *session.Manager
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

// Module provides application dependencies
var Module = fx.Options(
	fx.Provide(
		// Session manager
		func(logger logging.Logger, cfg *config.Config, lc fx.Lifecycle) *session.Manager {
			sessionConfig := &session.SessionConfig{
				SessionConfig: &cfg.Session,
				PublicPaths: []string{
					"/",
					"/login",
					"/signup",
				},
				ExemptPaths: []string{
					"/api/validation/",
					"/forgot-password",
					"/contact",
				},
				StaticPaths: []string{
					"/static/",
					"/assets/",
					"/images/",
				},
			}
			return session.NewManager(logger, sessionConfig, lc)
		},
		// Middleware manager
		func(
			logger logging.Logger,
			cfg *config.Config,
			userService user.Service,
			sessionManager *session.Manager,
		) *middleware.Manager {
			return middleware.NewManager(&middleware.ManagerConfig{
				Logger:         logger,
				Security:       &cfg.Security,
				UserService:    userService,
				SessionManager: sessionManager,
				Config:         cfg,
			})
		},
	),
	web.Module,
)
