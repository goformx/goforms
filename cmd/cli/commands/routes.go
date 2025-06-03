package commands

import (
	"context"
	"fmt"

	"github.com/goformx/goforms/internal/application/handler"
	"github.com/goformx/goforms/internal/bootstrap"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/infrastructure"
	"github.com/labstack/echo/v4"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

// ListRoutes prints all registered routes in the application
func ListRoutes(c *cli.Context) error {
	// Create a minimal app with just what we need for routes
	options := []fx.Option{
		infrastructure.RootModule,
		domain.Module,
	}
	options = append(options, bootstrap.Providers()...)
	options = append(options, bootstrap.ServerProviders()...)
	options = append(options, bootstrap.HandlerProviders()...)
	options = append(options, fx.Invoke(func(e *echo.Echo, handlers []handler.Handler) {
		// Register all handlers
		for _, h := range handlers {
			h.Register(e)
		}

		// Print routes
		fmt.Println("\nRegistered Routes:")
		fmt.Println("==================")

		for _, route := range e.Routes() {
			method := route.Method
			path := route.Path
			name := route.Name
			fmt.Printf("%-8s %-40s %s\n", method, path, name)
		}
	}))

	app := fx.New(options...)

	// Start the application
	if err := app.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}
	defer app.Stop(context.Background())

	return nil
}
