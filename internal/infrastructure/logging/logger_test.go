package logging_test

import (
	"errors"
	"testing"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	mocklogging "github.com/goformx/goforms/test/mocks/logging"
	"go.uber.org/mock/gomock"
)

func TestLogger(t *testing.T) {
	t.Run("creates logger with debug mode", func(t *testing.T) {
		logger, err := logging.NewLogger("debug", "test-app")
		if err != nil {
			t.Fatalf("NewLogger() returned error: %v", err)
		}
		if logger == nil {
			t.Error("NewLogger() returned nil")
		}
	})

	t.Run("creates logger without debug mode", func(t *testing.T) {
		logger, err := logging.NewLogger("info", "test-app")
		if err != nil {
			t.Fatalf("NewLogger() returned error: %v", err)
		}
		if logger == nil {
			t.Error("NewLogger() returned nil")
		}
	})

	t.Run("logs messages at different levels", func(t *testing.T) {
		logger, err := logging.NewLogger("debug", "test-app")
		if err != nil {
			t.Fatalf("NewLogger() returned error: %v", err)
		}

		// Just verify no panics
		logger.Info("info message", logging.StringField("key", "value"))
		logger.Error("error message", logging.ErrorField("error", errors.New("test error")))
		logger.Debug("debug message")
		logger.Warn("warn message")
	})
}

func TestNewLogger(t *testing.T) {
	t.Run("creates logger with default config", func(t *testing.T) {
		logger, err := logging.NewLogger("info", "test-app")
		if err != nil {
			t.Fatalf("NewLogger() returned error: %v", err)
		}
		if logger == nil {
			t.Error("NewLogger() returned nil")
		}
	})

	t.Run("creates logger with custom config", func(t *testing.T) {
		logger, err := logging.NewLogger("debug", "custom-app")
		if err != nil {
			t.Fatalf("NewLogger() returned error: %v", err)
		}
		if logger == nil {
			t.Error("NewLogger() returned nil")
		}
	})
}

func TestLogLevels(t *testing.T) {
	logger, err := logging.NewLogger("info", "test-app")
	if err != nil {
		t.Fatalf("NewLogger() returned error: %v", err)
	}

	t.Run("logs at different levels", func(t *testing.T) {
		// These should not panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Logging panicked: %v", r)
				}
			}()
			logger.Info("info message", logging.StringField("key", "value"))
			logger.Error("error message", logging.ErrorField("error", errors.New("test error")))
			logger.Debug("debug message")
			logger.Warn("warn message")
		}()
	})
}

func TestLoggerModes(t *testing.T) {
	t.Run("development mode", func(t *testing.T) {
		logger, err := logging.NewLogger("debug", "debug-app")
		if err != nil {
			t.Fatalf("NewLogger() returned error: %v", err)
		}
		if logger == nil {
			t.Error("NewLogger() returned nil")
		}
	})

	t.Run("production mode", func(t *testing.T) {
		logger, err := logging.NewLogger("info", "prod-app")
		if err != nil {
			t.Fatalf("NewLogger() returned error: %v", err)
		}
		if logger == nil {
			t.Error("NewLogger() returned nil")
		}
	})
}

func TestLoggerFunctionality(t *testing.T) {
	logger, err := logging.NewLogger("debug", "test-app")
	if err != nil {
		t.Fatalf("NewLogger() returned error: %v", err)
	}

	// Test logging methods don't panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Logging panicked: %v", r)
			}
		}()
		logger.Info("test info message",
			logging.StringField("key1", "value1"),
			logging.IntField("key2", 123),
			logging.ErrorField("error", errors.New("test error")),
		)
		logger.Error("test error message", logging.ErrorField("error", errors.New("test error")))
		logger.Debug("test debug message")
		logger.Warn("test warn message")
	}()
}

func TestMockLogger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocklogging.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info("info message", gomock.Any()).Times(1)
	mockLogger.EXPECT().Error("error message", gomock.Any()).Times(1)
	mockLogger.EXPECT().Debug("debug message", gomock.Any()).Times(1)
	mockLogger.EXPECT().Warn("warn message", gomock.Any()).Times(1)

	mockLogger.Info("info message")
	mockLogger.Error("error message")
	mockLogger.Debug("debug message")
	mockLogger.Warn("warn message")
}

func TestNewTestLogger(t *testing.T) {
	logger, err := logging.NewTestLogger()
	if err != nil {
		t.Fatalf("NewTestLogger() returned error: %v", err)
	}
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
