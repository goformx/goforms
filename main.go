// Package main is the entry point for the GoForms application.
// It sets up the application using the fx dependency injection framework
// and manages the application lifecycle including startup and graceful shutdown.
package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/fx"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application"
	"github.com/goformx/goforms/internal/application/handlers/web"
	appmiddleware "github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/infrastructure"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/server"
	"github.com/goformx/goforms/internal/infrastructure/version"
	"github.com/goformx/goforms/internal/presentation"
)

//go:embed all:dist
var distFS embed.FS

// DefaultShutdownTimeout defines the maximum time to wait for graceful shutdown
// before forcing termination.
const DefaultShutdownTimeout = 30 * time.Second

// appParams defines the dependency injection parameters for the application.
// It uses fx.In to automatically inject dependencies provided by the fx container.
// Note: This struct should be used by value, not as a pointer, when fx.In is embedded.
type appParams struct {
	fx.In
	Lifecycle         fx.Lifecycle           // Manages application lifecycle hooks
	Echo              *echo.Echo             // HTTP server framework instance
	Server            *server.Server         // Custom server implementation
	Logger            logging.Logger         // Application logger
	Handlers          []web.Handler          `group:"handlers"` // Web request handlers
	MiddlewareManager *appmiddleware.Manager // Legacy middleware management
	AccessManager     *access.Manager        // Access control management
	Config            *config.Config         // Application configuration

	// NEW: New middleware system components
	MigrationAdapter *appmiddleware.MigrationAdapter // Migration adapter for gradual transition
}

// setupHandlers registers all web handlers with the Echo server.
// It validates that no nil handlers are present and registers each handler
// with the Echo instance.
func setupHandlers(
	handlers []web.Handler,
	e *echo.Echo,
	accessManager *access.Manager,
	logger logging.Logger,
) error {
	for i, handler := range handlers {
		if handler == nil {
			return fmt.Errorf("nil handler encountered at index %d", i)
		}
	}

	// Use the RegisterHandlers function to properly register routes with access control
	web.RegisterHandlers(e, handlers, accessManager, logger)

	return nil
}

// setupApplication initializes the application by setting up middleware
// and registering all web handlers.
// Note: params is passed by value, not as a pointer
func setupApplication(params appParams) error {
	// Use the migration adapter to setup middleware with fallback capability
	if err := params.MigrationAdapter.SetupWithFallback(params.Echo, params.MiddlewareManager); err != nil {
		return fmt.Errorf("failed to setup middleware: %w", err)
	}

	return setupHandlers(params.Handlers, params.Echo, params.AccessManager, params.Logger)
}

// setupLifecycle configures the application lifecycle hooks for startup and shutdown.
// It logs application information and manages server startup in a goroutine.
// Note: params is passed by value, not as a pointer
func setupLifecycle(params appParams) {
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			versionInfo := version.GetInfo()
			// Log application startup information
			params.Logger.Info("starting application",
				"app", params.Config.App.Name,
				"version", versionInfo.Version,
				"environment", params.Config.App.Environment,
				"build_time", versionInfo.BuildTime,
				"git_commit", versionInfo.GitCommit,
			)

			// Log middleware system status
			migrationStatus := params.MigrationAdapter.GetMigrationStatus()
			params.Logger.Info("middleware system status",
				"new_system_enabled", migrationStatus.NewSystemEnabled,
				"registered_middleware_count", len(migrationStatus.RegisteredMiddleware),
				"available_chains_count", len(migrationStatus.AvailableChains),
			)

			// Start the server in a goroutine to prevent blocking
			go func() {
				if err := params.Server.Start(ctx); err != nil {
					params.Logger.Fatal("server startup failed", "error", err)
					os.Exit(1)
				}
			}()

			return nil
		},
		OnStop: func(_ context.Context) error {
			versionInfo := version.GetInfo()
			// Log application shutdown information
			params.Logger.Info("shutting down application",
				"app", params.Config.App.Name,
				"version", versionInfo.Version,
				"build_time", versionInfo.BuildTime,
				"git_commit", versionInfo.GitCommit,
			)

			return nil
		},
	})
}

// main is the entry point of the application.
// It initializes the dependency injection container, starts the application,
// and handles graceful shutdown on termination signals.
func main() {
	// Initialize the fx application container with all required modules and providers
	app := fx.New(
		// Provide embedded filesystem
		fx.Provide(func() embed.FS {
			return distFS
		}),
		// Include all application modules
		config.Module, // Config must come first as other modules depend on it
		infrastructure.Module,
		domain.Module,
		application.Module,
		appmiddleware.Module,
		presentation.Module,
		web.Module,
		// Invoke setup functions
		fx.Invoke(setupApplication),
		fx.Invoke(setupLifecycle),
	)

	// Start the application
	if startErr := app.Start(context.Background()); startErr != nil {
		fmt.Fprintf(os.Stderr, "Application startup failed: %v\n", startErr)
		os.Exit(1)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt signal
	<-sigChan

	// Attempt graceful shutdown
	if stopErr := app.Stop(context.Background()); stopErr != nil {
		fmt.Fprintf(os.Stderr, "Application shutdown failed: %v\n", stopErr)
		os.Exit(1)
	}
}
