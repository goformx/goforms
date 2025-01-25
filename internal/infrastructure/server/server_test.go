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
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
)

func TestNew(t *testing.T) {
	// Arrange
	logger := mocklogging.NewMockLogger()
	cfg := &config.Config{
		App: config.AppConfig{
			Host: "localhost",
			Port: 8090,
		},
		Server: config.ServerConfig{
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     30 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
	}

	var srv *server.Server
	app := fxtest.New(t,
		fx.Provide(
			func() logging.Logger { return logger },
			func() *config.Config { return cfg },
			server.New,
		),
		fx.Populate(&srv),
	)

	app.RequireStart()
	defer app.RequireStop()

	assert.NotNil(t, srv)
	assert.NotNil(t, srv.Echo())
}

func TestServerLifecycle(t *testing.T) {
	// Arrange
	logger := mocklogging.NewMockLogger()
	cfg := &config.Config{
		App: config.AppConfig{
			Host: "localhost",
			Port: 8090,
		},
		Server: config.ServerConfig{
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     30 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
	}

	// Create a test app
	var srv *server.Server
	app := fxtest.New(t,
		fx.Provide(
			func() logging.Logger { return logger },
			func() *config.Config { return cfg },
			server.New,
		),
		fx.Populate(&srv),
	)

	// Act & Assert
	// Starting the app will trigger our OnStart hooks
	assert.NoError(t, app.Start(context.Background()))

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stopping the app will trigger our OnStop hooks
	assert.NoError(t, app.Stop(context.Background()))
}
