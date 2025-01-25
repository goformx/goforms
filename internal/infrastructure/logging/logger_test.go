package logging_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
)

func TestNewLogger(t *testing.T) {
	t.Run("creates logger with default config", func(t *testing.T) {
		cfg := &config.Config{
			App: config.AppConfig{
				Name:  "test-app",
				Env:   "development",
				Debug: false,
				Port:  8080,
				Host:  "localhost",
			},
		}

		logger := logging.NewLogger(cfg)
		assert.NotNil(t, logger)
	})

	t.Run("creates logger with custom config", func(t *testing.T) {
		cfg := &config.Config{
			App: config.AppConfig{
				Name:  "custom-app",
				Env:   "production",
				Debug: true,
				Port:  9090,
				Host:  "custom-host",
			},
		}

		logger := logging.NewLogger(cfg)
		assert.NotNil(t, logger)
	})
}

func TestLogLevels(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:  "test-app",
			Env:   "test",
			Debug: false,
		},
	}

	logger := logging.NewLogger(cfg)

	t.Run("logs at different levels", func(t *testing.T) {
		// These should not panic
		logger.Info("info message", logging.String("key", "value"))
		logger.Error("error message", logging.Error(assert.AnError))
		logger.Debug("debug message")
		logger.Warn("warn message")
	})
}

func TestLoggerModes(t *testing.T) {
	t.Run("development mode", func(t *testing.T) {
		debugCfg := &config.Config{
			App: config.AppConfig{
				Name:  "debug-app",
				Env:   "development",
				Debug: true,
			},
		}
		logger := logging.NewLogger(debugCfg)
		assert.NotNil(t, logger)
	})

	t.Run("production mode", func(t *testing.T) {
		nonDebugCfg := &config.Config{
			App: config.AppConfig{
				Name:  "prod-app",
				Env:   "production",
				Debug: false,
			},
		}
		logger := logging.NewLogger(nonDebugCfg)
		assert.NotNil(t, logger)
	})
}

func TestLoggerFunctionality(t *testing.T) {
	// Create test config with full Config structure
	cfg := &config.Config{
		App: config.AppConfig{
			Name:  "test-app",
			Env:   "test",
			Debug: true,
		},
	}

	logger := logging.NewLogger(cfg)

	// Test logging methods don't panic
	assert.NotPanics(t, func() {
		logger.Info("test info message",
			logging.String("key1", "value1"),
			logging.Int("key2", 123),
			logging.Error(errors.New("test error")),
		)
		logger.Error("test error message")
		logger.Debug("test debug message")
		logger.Warn("test warn message")
	})
}

func TestMockLogger(t *testing.T) {
	mockLogger := mocklogging.NewMockLogger()

	mockLogger.ExpectInfo("info message")
	mockLogger.ExpectError("error message")
	mockLogger.ExpectDebug("debug message")
	mockLogger.ExpectWarn("warn message")

	mockLogger.Info("info message")
	mockLogger.Error("error message")
	mockLogger.Debug("debug message")
	mockLogger.Warn("warn message")

	if err := mockLogger.Verify(); err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
}

func TestNewTestLogger(t *testing.T) {
	logger := logging.NewTestLogger()
	assert.NotNil(t, logger, "Test logger should not be nil")

	// Test that logging methods don't panic in test mode
	assert.NotPanics(t, func() {
		logger.Info("test message")
		logger.Error("test error")
		logger.Debug("test debug")
		logger.Warn("test warn")
	})
}
