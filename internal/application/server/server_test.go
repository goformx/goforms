package server_test

import (
	"context"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/application/server"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
)

// mockLifecycle captures hooks for testing
type mockLifecycle struct {
	startHook func(context.Context) error
	stopHook  func(context.Context) error
}

func (m *mockLifecycle) Append(hook fx.Hook) {
	m.startHook = hook.OnStart
	m.stopHook = hook.OnStop
}

func TestNewServer(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		App: config.AppConfig{
			Name:  "test-app",
			Env:   "test",
			Debug: true,
		},
	}

	// Create mock logger
	mockLogger := mocklogging.NewMockLogger()

	// Create Echo instance
	e := echo.New()

	// Create mock lifecycle
	lc := &mockLifecycle{}

	// Create server
	srv := server.New(lc, e, mockLogger, cfg)

	// Assert server is created
	assert.NotNil(t, srv)
	assert.NotNil(t, lc.startHook)
	assert.NotNil(t, lc.stopHook)

	// Verify logger calls
	if err := mockLogger.Verify(); err != nil {
		t.Errorf("logger expectations not met: %v", err)
	}
}

func TestServerStart(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		App: config.AppConfig{
			Name:  "test-app",
			Env:   "test",
			Debug: true,
		},
	}

	// Create mock logger
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectInfo("starting HTTP server",
		logging.String("host", ""),
		logging.Int("port", 0),
		logging.String("env", "test"),
	)

	// Create Echo instance
	e := echo.New()

	// Create mock lifecycle
	lc := &mockLifecycle{}

	// Create server
	server.New(lc, e, mockLogger, cfg)

	// Execute OnStart hook
	err := lc.startHook(context.Background())
	assert.NoError(t, err)

	// Verify logger calls
	if err := mockLogger.Verify(); err != nil {
		t.Errorf("logger expectations not met: %v", err)
	}
}

func TestServerStop(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		App: config.AppConfig{
			Name:  "test-app",
			Env:   "test",
			Debug: true,
		},
	}

	// Create mock logger
	mockLogger := mocklogging.NewMockLogger()
	mockLogger.ExpectInfo("stopping HTTP server")

	// Create Echo instance
	e := echo.New()

	// Create mock lifecycle
	lc := &mockLifecycle{}

	// Create server
	server.New(lc, e, mockLogger, cfg)

	// Execute OnStop hook
	err := lc.stopHook(context.Background())
	assert.NoError(t, err)

	// Verify logger calls
	if err := mockLogger.Verify(); err != nil {
		t.Errorf("logger expectations not met: %v", err)
	}
}
