package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application"
	"github.com/goformx/goforms/internal/application/handlers/web"
	appmiddleware "github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/infrastructure"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/server"
	"github.com/goformx/goforms/internal/presentation"
	"github.com/labstack/echo/v4"
)

const DefaultShutdownTimeout = 30 * time.Second

type appParams struct {
	fx.In
	Lifecycle         fx.Lifecycle
	Echo              *echo.Echo
	Server            *server.Server
	Logger            logging.Logger
	Handlers          []web.Handler `group:"handlers"`
	MiddlewareManager *appmiddleware.Manager
	Config            *config.Config
}

func setupHandlers(handlers []web.Handler, e *echo.Echo) error {
	for i, handler := range handlers {
		if handler == nil {
			return fmt.Errorf("nil handler encountered at index %d", i)
		}
		handler.Register(e)
	}
	return nil
}

func setupApplication(params appParams) error {
	params.MiddlewareManager.Setup(params.Echo)
	return setupHandlers(params.Handlers, params.Echo)
}

func setupLifecycle(params appParams) {
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.Logger.Info("starting application",
				"app", params.Config.App.Name,
				"version", params.Config.App.Version,
				"environment", params.Config.App.Env,
			)

			// Start the server and handle errors properly
			go func() {
				if err := params.Server.Start(); err != nil {
					params.Logger.Fatal("server startup failed", "error", err)
					os.Exit(1)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info("shutting down application",
				"app", params.Config.App.Name,
				"version", params.Config.App.Version,
			)
			return nil
		},
	})
}

func main() {
	app := fx.New(
		fx.Provide(config.New),
		fx.Provide(func(cfg *config.Config) logging.Logger {
			factory := logging.NewFactory(logging.FactoryConfig{
				AppName:     cfg.App.Name,
				Version:     cfg.App.Version,
				Environment: cfg.App.Env,
				Fields:      map[string]any{},
			})
			logger, err := factory.CreateLogger()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
				os.Exit(1)
			}
			return logger
		}),
		infrastructure.Module,
		domain.Module,
		application.Module,
		presentation.Module,
		fx.Invoke(setupApplication),
		fx.Invoke(setupLifecycle),
	)

	if startErr := app.Start(context.Background()); startErr != nil {
		fmt.Fprintf(os.Stderr, "Application startup failed: %v\n", startErr)
		os.Exit(1)
	}

	// Handle termination signals gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	fmt.Printf("Received signal: %v, shutting down...\n", sig)

	if stopErr := app.Stop(context.Background()); stopErr != nil {
		fmt.Fprintf(os.Stderr, "Application shutdown failed: %v\n", stopErr)
		os.Exit(1)
	}
}
