package logging_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

func TestGetLogger(t *testing.T) {
	// Get two instances of the logger
	logger1 := logging.GetLogger()
	logger2 := logging.GetLogger()

	// Test that they are not nil
	assert.NotNil(t, logger1)
	assert.NotNil(t, logger2)

	// Test that they are the same instance (singleton)
	assert.Equal(t, logger1, logger2)
}

func TestLoggerFunctionality(t *testing.T) {
	logger := logging.GetLogger()

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
