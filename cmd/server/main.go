package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonesrussell/goforms/internal/app"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		app.Module(),
		fx.Invoke(registerHooks),
	)

	startCtx := context.Background()
	if err := app.Start(startCtx); err != nil {
		os.Exit(1)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	stopCtx := context.Background()
	if err := app.Stop(stopCtx); err != nil {
		os.Exit(1)
	}
}

func registerHooks(lifecycle fx.Lifecycle) {
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
