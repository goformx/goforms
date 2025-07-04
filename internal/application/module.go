// Package application provides the application layer components and their dependency injection setup.
package application

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/fx"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/application/services"
	"github.com/goformx/goforms/internal/application/validation"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/adapters/http"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
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
	Logger           logging.Logger
	Config           *config.Config
	Server           *server.Server
	DB               database.DB
	DomainModule     fx.Option
	Presentation     fx.Option
	MiddlewareModule fx.Option
	SessionManager   services.SessionManager
	Renderer         view.Renderer
	AccessManager    *access.Manager
	Sanitizer        sanitization.ServiceInterface
}

// Validate checks if all required dependencies are present
func (d Dependencies) Validate() error {
	required := []struct {
		name  string
		value any
	}{
		{"UserService", d.UserService},
		{"FormService", d.FormService},
		{"Logger", d.Logger},
		{"Config", d.Config},
		{"Server", d.Server},
		{"DB", d.DB},
		{"DomainModule", d.DomainModule},
		{"Presentation", d.Presentation},
		{"MiddlewareModule", d.MiddlewareModule},
		{"SessionManager", d.SessionManager},
		{"Renderer", d.Renderer},
		{"AccessManager", d.AccessManager},
		{"Sanitizer", d.Sanitizer},
	}

	for _, r := range required {
		if r.value == nil {
			return errors.New(r.name + " is required")
		}
	}

	return nil
}

// Module represents the application module
var Module = fx.Module("application",
	fx.Provide(
		New,
		provideErrorHandler,
		provideRecoveryMiddleware,
		// Application services
		services.NewAuthUseCaseService,
		services.NewFormUseCaseService,
		// HTTP adapters
		fx.Annotate(
			http.NewEchoRequestAdapter,
			fx.As(new(http.RequestAdapter)),
		),
		fx.Annotate(
			http.NewEchoResponseAdapter,
			fx.As(new(http.ResponseAdapter)),
		),
	),
	validation.Module,
)

// provideErrorHandler creates a new error handler with sanitization service
func provideErrorHandler(
	logger logging.Logger,
	sanitizer sanitization.ServiceInterface,
) response.ErrorHandlerInterface {
	return response.NewErrorHandler(logger, sanitizer)
}

// provideRecoveryMiddleware creates a new recovery middleware with sanitization service
func provideRecoveryMiddleware(logger logging.Logger, sanitizer sanitization.ServiceInterface) echo.MiddlewareFunc {
	return middleware.Recovery(logger, sanitizer)
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
	sessionManager   services.SessionManager
	renderer         view.Renderer
}

// Start starts the application
func (a *Application) Start(_ context.Context) error {
	a.logger.Info("Starting application...")

	// Handler registration is now handled by the presentation module

	// Start the server
	if err := a.server.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	a.logger.Info("Application started successfully")

	return nil
}

// Stop stops the application
func (a *Application) Stop(_ context.Context) error {
	a.logger.Info("Stopping application...")
	a.logger.Info("Application stopped successfully")

	return nil
}
