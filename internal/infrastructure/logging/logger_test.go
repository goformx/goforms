package logging_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/test/mocks"
)

func TestNewLogger(t *testing.T) {
	// Create a new logger
	logger := logging.NewLogger()

	// Test that it's not nil
	assert.NotNil(t, logger, "Logger should not be nil")

	// Test that each new instance is independent
	logger2 := logging.NewLogger()
	assert.NotNil(t, logger2, "Second logger should not be nil")
}

func TestLoggerFunctionality(t *testing.T) {
	logger := logging.NewLogger()

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
	mock := mocks.NewLogger()

	// Test info logging
	testMsg := "test info message"
	mock.Info(testMsg, logging.String("key", "value"))
	assert.True(t, mock.HasInfoLog(testMsg), "Mock should record info message")

	// Test error logging
	errMsg := "test error message"
	mock.Error(errMsg)
	assert.True(t, mock.HasErrorLog(errMsg), "Mock should record error message")

	// Test debug logging
	mock.Debug("test debug")
	assert.Len(t, mock.DebugCalls, 1, "Mock should record debug call")

	// Test warn logging
	mock.Warn("test warn")
	assert.Len(t, mock.WarnCalls, 1, "Mock should record warn call")
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
