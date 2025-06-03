package commands

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/urfave/cli/v2"
)

// ListRoutes prints all registered routes in the application
func ListRoutes(c *cli.Context) error {
	// Create a new Echo instance
	e := echo.New()

	// No handlers to register (handler.NewFormHandler was removed)

	// Print routes (will be empty unless you add routes elsewhere)
	fmt.Println("\nRegistered Routes:")
	fmt.Println("==================")

	for _, route := range e.Routes() {
		method := route.Method
		path := route.Path
		name := route.Name
		fmt.Printf("%-8s %-40s %s\n", method, path, name)
	}

	return nil
}
