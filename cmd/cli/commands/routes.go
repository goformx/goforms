package commands

import (
	"fmt"

	"github.com/goformx/goforms/internal/application/handler"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
	"github.com/urfave/cli/v2"
)

// ListRoutes prints all registered routes in the application
func ListRoutes(c *cli.Context) error {
	// Create a new Echo instance
	e := echo.New()

	// Initialize logger
	logger, err := logging.NewLogger("info", "goforms-cli")
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	// Initialize handlers (you'll need to provide the actual services)
	// For now, we'll just create empty handlers to get the routes
	formHandler, err := handler.NewFormHandler(logger, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create form handler: %w", err)
	}

	// Register routes
	formHandler.Register(e)

	// Print routes
	fmt.Println("\nRegistered Routes:")
	fmt.Println("==================")

	for _, route := range e.Routes() {
		method := route.Method
		path := route.Path
		name := route.Name

		// Format the output similar to Laravel's artisan route:list
		fmt.Printf("%-8s %-40s %s\n", method, path, name)
	}

	return nil
}
