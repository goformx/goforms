package logging_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
)

func TestLogger(t *testing.T) {
	t.Run("creates logger with debug mode", func(t *testing.T) {
		logger := logging.NewLogger(true, "test-app")
		assert.NotNil(t, logger)
	})

	t.Run("creates logger without debug mode", func(t *testing.T) {
		logger := logging.NewLogger(false, "test-app")
		assert.NotNil(t, logger)
	})

	t.Run("logs messages at different levels", func(t *testing.T) {
		logger := logging.NewLogger(true, "test-app")

		// Just verify no panics
		logger.Info("info message")
		logger.Error("error message")
		logger.Debug("debug message")
		logger.Warn("warn message")
	})
}

func TestNewLogger(t *testing.T) {
	t.Run("creates logger with default config", func(t *testing.T) {
		logger := logging.NewLogger(false, "test-app")
		assert.NotNil(t, logger)
	})

	t.Run("creates logger with custom config", func(t *testing.T) {
		logger := logging.NewLogger(true, "custom-app")
		assert.NotNil(t, logger)
	})
}

func TestLogLevels(t *testing.T) {
	logger := logging.NewLogger(false, "test-app")

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
		logger := logging.NewLogger(true, "debug-app")
		assert.NotNil(t, logger)
	})

	t.Run("production mode", func(t *testing.T) {
		logger := logging.NewLogger(false, "prod-app")
		assert.NotNil(t, logger)
	})
}

func TestLoggerFunctionality(t *testing.T) {
	logger := logging.NewLogger(true, "test-app")

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
