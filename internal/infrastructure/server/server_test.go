package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := app.Start(ctx)
	assert.NoError(t, err)

	err = app.Stop(ctx)
	assert.NoError(t, err)
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Test startup
	err := app.Start(ctx)
	assert.NoError(t, err)

	// Test shutdown
	err = app.Stop(ctx)
	assert.NoError(t, err)
}
