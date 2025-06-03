package commands

import (
	"context"
	"fmt"

	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/bootstrap"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/infrastructure"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

// ListRoutes prints all registered routes in the application
func ListRoutes(c *cli.Context) error {
	// Create logger
	logger, logErr := logging.NewFactory().CreateLogger()
	if logErr != nil {
		return logErr
	}

	// Create a minimal app with just what we need for routes
	options := []fx.Option{
		infrastructure.RootModule,
		domain.Module,
	}
	options = append(options, bootstrap.Providers()...)
	options = append(options, bootstrap.ServerProviders()...)
	options = append(options, bootstrap.HandlerProviders()...)
	options = append(options, fx.Invoke(func(e *echo.Echo, handlers []web.Handler) {
		// Register all handlers
		for _, h := range handlers {
			h.Register(e)
		}

		// Print routes
		logger.Info("Registered Routes:")
		logger.Info("==================")

		for _, route := range e.Routes() {
			method := route.Method
			path := route.Path
			name := route.Name
			logger.Info("Route details",
				logging.StringField("method", method),
				logging.StringField("path", path),
				logging.StringField("name", name),
			)
		}
	}))

	app := fx.New(options...)

	// Start the application
	if err := app.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}
	defer func() {
		if err := app.Stop(context.Background()); err != nil {
			logger.Error("Error stopping application", logging.ErrorField("error", err))
		}
	}()

	return nil
}
