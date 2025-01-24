package server

import (
	"context"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
)

// mockLifecycle implements fx.Lifecycle for testing
type mockLifecycle struct{}

func (m *mockLifecycle) Append(hook fx.Hook) {
	// No-op for testing
}

func TestNewServer(t *testing.T) {
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectInfo("initializing server")

	cfg := &config.Config{
		App: config.AppConfig{
			Host: "localhost",
			Port: 8090,
		},
	}

	e := echo.New()
	lc := &mockLifecycle{}

	srv := New(lc, e, mockLogger, cfg)
	assert.NotNil(t, srv)
	mockLogger.AssertExpectations(t)
}

func TestServerStart(t *testing.T) {
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectInfo("starting HTTP server",
		logging.String("host", "localhost"),
		logging.Int("port", 8090),
		logging.String("env", ""),
	)

	cfg := &config.Config{
		App: config.AppConfig{
			Host: "localhost",
			Port: 8090,
		},
	}

	e := echo.New()
	lc := &mockLifecycle{}

	srv := New(lc, e, mockLogger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := srv.Start(ctx)
	assert.NoError(t, err)
	mockLogger.AssertExpectations(t)
}

func TestServerStop(t *testing.T) {
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectInfo("initializing server")
	mockLogger.ExpectInfo("stopping HTTP server")

	cfg := &config.Config{
		App: config.AppConfig{
			Host: "localhost",
			Port: 8090,
		},
	}

	e := echo.New()
	lc := &mockLifecycle{}

	srv := New(lc, e, mockLogger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := srv.Stop(ctx)
	assert.NoError(t, err)
	mockLogger.AssertExpectations(t)
}
