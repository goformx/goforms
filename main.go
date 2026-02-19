// Package main is the entry point for the GoForms application.
// It wires the system using Uber Fx, initializes middleware, registers handlers,
// and manages startup and graceful shutdown.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

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
)

// DefaultShutdownTimeout defines the maximum time to wait for graceful shutdown.
const DefaultShutdownTimeout = 30 * time.Second

// appParams collects all dependencies injected by Fx.
type appParams struct {
	fx.In

	Lifecycle         fx.Lifecycle
	Echo              *echo.Echo
	Server            *server.Server
	Logger            logging.Logger
	Handlers          []web.Handler `group:"handlers"`
	MiddlewareManager *appmiddleware.Manager
	AccessManager     *access.Manager
	Config            *config.Config

	// New middleware system components
	MigrationAdapter *appmiddleware.MigrationAdapter
}

// setupHandlers registers all HTTP handlers with Echo.
func setupHandlers(
	handlers []web.Handler,
	e *echo.Echo,
	accessManager *access.Manager,
	logger logging.Logger,
) error {
	for i, h := range handlers {
		if h == nil {
			return fmt.Errorf("nil handler encountered at index %d", i)
		}
	}

	web.RegisterHandlers(e, handlers, accessManager, logger)
	return nil
}

// setupApplication configures middleware and registers handlers.
func setupApplication(p appParams) error {
	if err := p.MigrationAdapter.SetupWithFallback(p.Echo, p.MiddlewareManager); err != nil {
		return fmt.Errorf("middleware setup failed: %w", err)
	}

	return setupHandlers(p.Handlers, p.Echo, p.AccessManager, p.Logger)
}

// setupLifecycle configures startup and shutdown hooks.
func setupLifecycle(p appParams) {
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			v := version.GetInfo()

			p.Logger.Info("starting application",
				"app", p.Config.App.Name,
				"version", v.Version,
				"environment", p.Config.App.Environment,
				"build_time", v.BuildTime,
				"git_commit", v.GitCommit,
			)

			status := p.MigrationAdapter.GetMigrationStatus()
			p.Logger.Info("middleware system status",
				"new_system_enabled", status.NewSystemEnabled,
				"registered_middleware_count", len(status.RegisteredMiddleware),
				"available_chains_count", len(status.AvailableChains),
			)

			// Start server asynchronously
			go func() {
				if err := p.Server.Start(ctx); err != nil {
					p.Logger.Fatal("server startup failed", "error", err)
					os.Exit(1)
				}
			}()

			return nil
		},

		OnStop: func(_ context.Context) error {
			v := version.GetInfo()

			p.Logger.Info("shutting down application",
				"app", p.Config.App.Name,
				"version", v.Version,
				"build_time", v.BuildTime,
				"git_commit", v.GitCommit,
			)

			return nil
		},
	})
}

// main initializes the Fx application and manages graceful shutdown.
func main() {
	app := fx.New(
		// Modules
		config.Module,
		infrastructure.Module,
		domain.Module,
		application.Module,
		appmiddleware.Module,
		web.Module,

		// Setup
		fx.Invoke(setupApplication),
		fx.Invoke(setupLifecycle),
	)

	if err := app.Start(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "application startup failed: %v\n", err)
		os.Exit(1)
	}

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	if err := app.Stop(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "application shutdown failed: %v\n", err)
		os.Exit(1)
	}
}
