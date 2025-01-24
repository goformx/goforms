// Package logging provides a unified logging interface using zap
package logging

import (
	"os"
	"strings"
	"time"

	forbidden_zap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger defines the interface for logging operations
type Logger interface {
	Info(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
}

// Field represents a logging field
type Field = forbidden_zap.Field

// String creates a string field
func String(key string, value string) Field { return forbidden_zap.String(key, value) }

// Int creates an integer field
func Int(key string, value int) Field { return forbidden_zap.Int(key, value) }

// Error creates an error field
func Error(err error) Field { return forbidden_zap.Error(err) }

// Duration creates a duration field
func Duration(key string, value time.Duration) Field { return forbidden_zap.Duration(key, value) }

// Any creates a field with any value
func Any(key string, value interface{}) Field { return forbidden_zap.Any(key, value) }

type zapLogger struct {
	*forbidden_zap.Logger
}

// GetLogger returns a singleton logger instance
func GetLogger() Logger {
	// Initialize logger with production configuration
	config := forbidden_zap.NewProductionConfig()

	// Configure log level from environment
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch level {
	case "debug":
		config.Level = forbidden_zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		config.Level = forbidden_zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		config.Level = forbidden_zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		config.Level = forbidden_zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		config.Level = forbidden_zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// Build the logger
	logger, _ := config.Build()
	return &zapLogger{logger}
}

// NewTestLogger creates a logger suitable for testing
func NewTestLogger() Logger {
	logger, _ := forbidden_zap.NewDevelopment()
	return &zapLogger{logger}
}

// NewMockLogger creates a mock logger for testing
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

// Info logs an info message
func (l *zapLogger) Info(msg string, fields ...Field) { l.Logger.Info(msg, fields...) }

// Error logs an error message
func (l *zapLogger) Error(msg string, fields ...Field) { l.Logger.Error(msg, fields...) }

// Debug logs a debug message
func (l *zapLogger) Debug(msg string, fields ...Field) { l.Logger.Debug(msg, fields...) }

// Warn logs a warning message
func (l *zapLogger) Warn(msg string, fields ...Field) { l.Logger.Warn(msg, fields...) }

// UnderlyingZap returns the underlying zap logger
func UnderlyingZap(l Logger) *forbidden_zap.Logger {
	if zl, ok := l.(*zapLogger); ok {
		return zl.Logger
	}
	return nil
}

// MockLogger is a mock implementation of Logger for testing
type MockLogger struct {
	InfoCalls []struct {
		Message string
		Fields  []Field
	}
	ErrorCalls []struct {
		Message string
		Fields  []Field
	}
	DebugCalls []struct {
		Message string
		Fields  []Field
	}
	WarnCalls []struct {
		Message string
		Fields  []Field
	}
}

// Info records an info message call
func (m *MockLogger) Info(msg string, fields ...Field) {
	m.InfoCalls = append(m.InfoCalls, struct {
		Message string
		Fields  []Field
	}{msg, fields})
}

// Error records an error message call
func (m *MockLogger) Error(msg string, fields ...Field) {
	m.ErrorCalls = append(m.ErrorCalls, struct {
		Message string
		Fields  []Field
	}{msg, fields})
}

// Debug records a debug message call
func (m *MockLogger) Debug(msg string, fields ...Field) {
	m.DebugCalls = append(m.DebugCalls, struct {
		Message string
		Fields  []Field
	}{msg, fields})
}

// Warn records a warning message call
func (m *MockLogger) Warn(msg string, fields ...Field) {
	m.WarnCalls = append(m.WarnCalls, struct {
		Message string
		Fields  []Field
	}{msg, fields})
}

// HasInfoLog checks if a specific info message was logged
func (m *MockLogger) HasInfoLog(msg string) bool {
	for _, call := range m.InfoCalls {
		if call.Message == msg {
			return true
		}
	}
	return false
}

// HasErrorLog checks if a specific error message was logged
func (m *MockLogger) HasErrorLog(msg string) bool {
	for _, call := range m.ErrorCalls {
		if call.Message == msg {
			return true
		}
	}
	return false
}
