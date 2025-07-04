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

	"github.com/goformx/goforms/internal/application"
)

//go:embed all:dist
var distFS embed.FS

// main is the entry point of the application.
// It initializes the dependency injection container, starts the application,
// and handles graceful shutdown on termination signals.
func main() {
	// Create the application using the extracted function
	app := application.NewApplication(distFS)

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

	// Attempt graceful shutdown with configurable timeout
	// The timeout is now handled within the application lifecycle
	if stopErr := app.Stop(context.Background()); stopErr != nil {
		fmt.Fprintf(os.Stderr, "Application shutdown failed: %v\n", stopErr)
		os.Exit(1)
	}
}
