package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLogger(t *testing.T) {
	logger1 := GetLogger()
	assert.NotNil(t, logger1, "Logger instance should not be nil")

	logger2 := GetLogger()
	assert.NotNil(t, logger2, "Logger instance should not be nil")

	assert.Equal(t, logger1, logger2, "Logger instances should be the same")
}

func TestLoggerFunctionality(t *testing.T) {
	logger := GetLogger()

	// Test logging a simple message
	assert.NotPanics(t, func() {
		logger.Info("This is a test log message")
	}, "Logging should not panic")
}
