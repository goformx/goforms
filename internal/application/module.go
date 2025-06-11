// Package application provides the application layer components and their dependency injection setup.
package application

import (
	"context"
	"errors"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/server"
	"github.com/goformx/goforms/internal/presentation/view"
)

// Dependencies holds all application dependencies
type Dependencies struct {
	fx.In

	// Domain services
	UserService user.Service
	FormService form.Service

	// Infrastructure
	Logger            logging.Logger
	Config            *config.Config
	Server            *server.Server
	DomainModule      fx.Option
	Presentation      fx.Option
	MiddlewareModule  fx.Option
	SessionManager    *session.Manager
	Renderer          view.Renderer
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
		{"Server", d.Server},
		{"DomainModule", d.DomainModule},
		{"Presentation", d.Presentation},
		{"MiddlewareModule", d.MiddlewareModule},
		{"SessionManager", d.SessionManager},
		{"Renderer", d.Renderer},
		{"MiddlewareManager", d.MiddlewareManager},
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

// Module represents the application module
var Module = fx.Options(
	fx.Provide(
		New,
		provideMiddlewareManager,
	),
)

// provideMiddlewareManager creates a new middleware manager
func provideMiddlewareManager(
	logger logging.Logger,
	cfg *config.Config,
	userService user.Service,
	sessionManager *session.Manager,
) *middleware.Manager {
	return middleware.NewManager(&middleware.ManagerConfig{
		Logger:         logger,
		Security:       &cfg.Security,
		UserService:    userService,
		Config:         cfg,
		SessionManager: sessionManager,
	})
}

// New creates a new application instance
func New(lc fx.Lifecycle, deps Dependencies) *Application {
	app := &Application{
		logger:           deps.Logger,
		config:           deps.Config,
		server:           deps.Server,
		domainModule:     deps.DomainModule,
		presentation:     deps.Presentation,
		middlewareModule: deps.MiddlewareModule,
		sessionManager:   deps.SessionManager,
		renderer:         deps.Renderer,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return app.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return app.Stop(ctx)
		},
	})

	return app
}

// Application represents the main application
type Application struct {
	logger           logging.Logger
	config           *config.Config
	server           *server.Server
	domainModule     fx.Option
	presentation     fx.Option
	middlewareModule fx.Option
	sessionManager   *session.Manager
	renderer         view.Renderer
}

// Start starts the application
func (a *Application) Start(ctx context.Context) error {
	a.logger.Info("Starting application...")

	// Start the server
	if err := a.server.Start(); err != nil {
		return err
	}

	a.logger.Info("Application started successfully")
	return nil
}

// Stop stops the application
func (a *Application) Stop(ctx context.Context) error {
	a.logger.Info("Stopping application...")
	a.logger.Info("Application stopped successfully")
	return nil
}
