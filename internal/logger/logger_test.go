package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLogger(t *testing.T) {
	// Get two instances of the logger
	logger1 := GetLogger()
	logger2 := GetLogger()

	// Test that they are not nil
	assert.NotNil(t, logger1)
	assert.NotNil(t, logger2)

	// Test that they are the same type
	assert.IsType(t, &zapLogger{}, logger1)
	assert.IsType(t, &zapLogger{}, logger2)
}

func TestLoggerFunctionality(t *testing.T) {
	logger := GetLogger()

	// Test logging methods don't panic
	assert.NotPanics(t, func() {
		logger.Info("This is a test log message")
		logger.Error("This is a test error message")
		logger.Debug("This is a test debug message")
		logger.Warn("This is a test warning message")
	})

	// Test with fields
	assert.NotPanics(t, func() {
		logger.Info("Test with fields",
			String("string_field", "value"),
			Int("int_field", 123),
			Error(assert.AnError),
		)
	})
}
