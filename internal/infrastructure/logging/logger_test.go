package logging_test

import (
	"errors"
	"testing"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
)

func TestLogger(t *testing.T) {
	t.Run("creates logger with debug mode", func(t *testing.T) {
		logger := logging.NewLogger(true, "test-app")
		if logger == nil {
			t.Error("NewLogger() returned nil")
		}
	})

	t.Run("creates logger without debug mode", func(t *testing.T) {
		logger := logging.NewLogger(false, "test-app")
		if logger == nil {
			t.Error("NewLogger() returned nil")
		}
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
		if logger == nil {
			t.Error("NewLogger() returned nil")
		}
	})

	t.Run("creates logger with custom config", func(t *testing.T) {
		logger := logging.NewLogger(true, "custom-app")
		if logger == nil {
			t.Error("NewLogger() returned nil")
		}
	})
}

func TestLogLevels(t *testing.T) {
	logger := logging.NewLogger(false, "test-app")

	t.Run("logs at different levels", func(t *testing.T) {
		// These should not panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Logging panicked: %v", r)
				}
			}()
			logger.Info("info message", logging.String("key", "value"))
			logger.Error("error message", logging.Error(errors.New("test error")))
			logger.Debug("debug message")
			logger.Warn("warn message")
		}()
	})
}

func TestLoggerModes(t *testing.T) {
	t.Run("development mode", func(t *testing.T) {
		logger := logging.NewLogger(true, "debug-app")
		if logger == nil {
			t.Error("NewLogger() returned nil")
		}
	})

	t.Run("production mode", func(t *testing.T) {
		logger := logging.NewLogger(false, "prod-app")
		if logger == nil {
			t.Error("NewLogger() returned nil")
		}
	})
}

func TestLoggerFunctionality(t *testing.T) {
	logger := logging.NewLogger(true, "test-app")

	// Test logging methods don't panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Logging panicked: %v", r)
			}
		}()
		logger.Info("test info message",
			logging.String("key1", "value1"),
			logging.Int("key2", 123),
			logging.Error(errors.New("test error")),
		)
		logger.Error("test error message")
		logger.Debug("test debug message")
		logger.Warn("test warn message")
	}()
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
	if logger == nil {
		t.Error("NewTestLogger() returned nil")
	}

	// Test that logging methods don't panic in test mode
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Logging panicked: %v", r)
			}
		}()
		logger.Info("test message")
		logger.Error("test error")
		logger.Debug("test debug")
		logger.Warn("test warn")
	}()
}
