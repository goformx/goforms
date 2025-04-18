package server_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/server"
)

func TestNew(t *testing.T) {
	app := fxtest.New(t,
		fx.Provide(
			func() logging.Logger {
				return logging.NewTestLogger()
			},
			func() *config.Config {
				return &config.Config{
					App: config.AppConfig{
						Host: "localhost",
						Port: 0, // Use port 0 for testing
					},
					Server: config.ServerConfig{
						ReadTimeout:     100 * time.Millisecond,
						WriteTimeout:    100 * time.Millisecond,
						IdleTimeout:     100 * time.Millisecond,
						ShutdownTimeout: 100 * time.Millisecond,
					},
				}
			},
			server.New,
		),
		fx.StartTimeout(100*time.Millisecond),
		fx.StopTimeout(100*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()

	err := app.Start(ctx)
	if err != nil {
		t.Errorf("app.Start() error = %v, want nil", err)
	}

	err = app.Stop(ctx)
	if err != nil {
		t.Errorf("app.Stop() error = %v, want nil", err)
	}
}

func TestServerLifecycle(t *testing.T) {
	// Create a test app with a short timeout
	app := fxtest.New(t,
		fx.Provide(
			func() logging.Logger {
				return logging.NewTestLogger()
			},
			func() *config.Config {
				return &config.Config{
					App: config.AppConfig{
						Host: "localhost",
						Port: 0, // Use port 0 for testing (random available port)
					},
					Server: config.ServerConfig{
						ReadTimeout:     100 * time.Millisecond,
						WriteTimeout:    100 * time.Millisecond,
						IdleTimeout:     100 * time.Millisecond,
						ShutdownTimeout: 100 * time.Millisecond,
					},
				}
			},
			server.New,
		),
		fx.StartTimeout(100*time.Millisecond),
		fx.StopTimeout(100*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()

	// Test startup
	err := app.Start(ctx)
	if err != nil {
		t.Errorf("app.Start() error = %v, want nil", err)
	}

	// Test shutdown
	err = app.Stop(ctx)
	if err != nil {
		t.Errorf("app.Stop() error = %v, want nil", err)
	}
}
