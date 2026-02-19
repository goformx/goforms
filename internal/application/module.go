// Package application provides the application layer components and their dependency injection setup.
package application

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/fx"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/request"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/application/validation"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/goformx/goforms/internal/infrastructure/server"
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
	MiddlewareModule  fx.Option
	SessionManager    *session.Manager
	MiddlewareManager *middleware.Manager
	AccessManager     *access.Manager
	Sanitizer         sanitization.ServiceInterface
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
		{"DomainModule", d.DomainModule},
		{"MiddlewareModule", d.MiddlewareModule},
		{"SessionManager", d.SessionManager},
		{"MiddlewareManager", d.MiddlewareManager},
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
	}, nil
}

// Module represents the application module
var Module = fx.Module("application",
	fx.Provide(
		New,
		provideRequestUtils,
		provideErrorHandler,
		provideRecoveryMiddleware,
	),
	validation.Module,
)

// provideRequestUtils creates a new request utils instance with sanitization service
func provideRequestUtils(sanitizer sanitization.ServiceInterface) *request.Utils {
	return request.NewUtils(sanitizer)
}

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
		middlewareModule: deps.MiddlewareModule,
		sessionManager:   deps.SessionManager,
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
	middlewareModule fx.Option
	sessionManager   *session.Manager
}

// Start starts the application
func (a *Application) Start(ctx context.Context) error {
	a.logger.Info("Starting application...")

	// Get the Echo instance
	e := a.server.Echo()

	// Register all handlers
	var handlers []web.Handler

	fx.Populate(&handlers)

	for _, handler := range handlers {
		handler.Register(e)
	}

	// Start the server
	if err := a.server.Start(ctx); err != nil {
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
