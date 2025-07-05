// Package application provides application creation and lifecycle management.
package application

import (
	"context"
	"embed"
	"fmt"

	"go.uber.org/fx"

	appmiddleware "github.com/goformx/goforms/internal/application/middleware"
	middlewarecore "github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/goformx/goforms/internal/application/providers"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/infrastructure"
	"github.com/goformx/goforms/internal/infrastructure/adapters/http"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/server"
	"github.com/goformx/goforms/internal/presentation"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
	echosrv "github.com/labstack/echo/v4"
)

// NewApplication creates a new fx application with the provided modules and options
func NewApplication(distFS embed.FS, modules ...fx.Option) *fx.App {
	baseModules := []fx.Option{
		// Provide embedded filesystem
		fx.Provide(func() embed.FS {
			return distFS
		}),
		// Include all application modules in dependency order
		config.Module,         // Config must come first as other modules depend on it
		infrastructure.Module, // Infrastructure (database, logging, etc.)
		domain.Module,         // Domain services and repositories
		Module,                // Application layer services (correct reference)
		appmiddleware.Module,  // Middleware orchestration
		presentation.Module,   // Presentation layer (handlers, templates)
		// Include OpenAPI validation provider
		providers.OpenAPIValidationProvider(),
		// Infrastructure lifecycle logging (moved from infrastructure module)
		fx.Invoke(func(lc fx.Lifecycle, logger logging.Logger, _ config.ConfigInterface) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					logger.Info("Infrastructure module initialized")

					return nil
				},
				OnStop: func(_ context.Context) error {
					logger.Info("Infrastructure module shutting down")

					return nil
				},
			})
		}),
		// Middleware lifecycle logging and registration (moved from middleware module)
		fx.Invoke(func(
			lc fx.Lifecycle,
			registry middlewarecore.Registry,
			orchestrator middlewarecore.Orchestrator,
			logger logging.Logger,
		) {
			logger.Debug("Setting up middleware registration hook")
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Debug("Middleware registration hook starting")
					// Register all middleware with the registry
					if err := appmiddleware.RegisterAllMiddleware(registry, logger); err != nil {
						logger.Error("Failed to register middleware", "error", err)
						return fmt.Errorf("failed to register middleware: %w", err)
					}

					// Validate orchestrator configuration
					if err := orchestrator.ValidateConfiguration(); err != nil {
						logger.Error("Failed to validate orchestrator configuration", "error", err)
						return fmt.Errorf("failed to validate orchestrator configuration: %w", err)
					}

					logger.Info("middleware system initialized successfully")

					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("middleware system shutting down")

					return nil
				},
			})
		}),
		// Presentation route registration (moved from presentation module)
		fx.Invoke(func(
			lc fx.Lifecycle,
			e *echosrv.Echo,
			adapter *http.EchoAdapter,
			orchestrator *appmiddleware.EchoOrchestratorAdapter,
			handlers struct {
				fx.In
				Handlers []httpiface.Handler `group:"handlers"`
			},
		) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					// Set the middleware orchestrator on the adapter
					adapter.SetMiddlewareOrchestrator(orchestrator)

					for _, h := range handlers.Handlers {
						if err := adapter.RegisterHandler(h); err != nil {
							return fmt.Errorf("failed to register handler %s: %w", h.Name(), err)
						}
					}

					return nil
				},
			})
		}),
		// Application setup functions (moved to end to ensure all dependencies are available)
		fx.Invoke(setupApplication),
		fx.Invoke(setupLifecycle),
	}

	// Combine base modules with additional modules
	baseModules = append(baseModules, modules...)

	return fx.New(baseModules...)
}

// setupApplication initializes the application using the ApplicationSetup service.
func setupApplication(params appParams) error {
	return params.SetupService.Setup()
}

// setupLifecycle configures the application lifecycle using the LifecycleManager.
func setupLifecycle(params appParams) {
	lifecycleManager := NewLifecycleManager(LifecycleParams{
		Lifecycle: params.Lifecycle,
		Logger:    params.Logger,
		Server:    params.Server,
		Config:    params.Config,
	})
	lifecycleManager.SetupLifecycle()
}

// appParams defines the dependency injection parameters for the application.
type appParams struct {
	fx.In
	SetupService *ApplicationSetup      // Application setup service
	Lifecycle    fx.Lifecycle           // Manages application lifecycle hooks
	Logger       logging.Logger         // Application logger
	Server       *server.Server         // Server concrete type
	Config       config.ConfigInterface // Config interface instead of concrete type
}
