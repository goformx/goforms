// Package logging provides a unified logging interface using zap
package logging

import (
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger defines the interface for logging operations
//
// This interface abstracts the underlying logging implementation,
// allowing for easy mocking in tests and flexibility to change
// the logging backend without affecting application code.
//
// For testing, use test/mocks.Logger instead of implementing this interface directly.
type Logger interface {
	// Info logs a message at info level with optional fields
	Info(msg string, fields ...Field)
	// Error logs a message at error level with optional fields
	Error(msg string, fields ...Field)
	// Debug logs a message at debug level with optional fields
	Debug(msg string, fields ...Field)
	// Warn logs a message at warn level with optional fields
	Warn(msg string, fields ...Field)
}

// Field represents a logging field
type Field = zap.Field

// String creates a string field
func String(key string, value string) Field { return zap.String(key, value) }

// Int creates an integer field
func Int(key string, value int) Field { return zap.Int(key, value) }

// Error creates an error field
func Error(err error) Field { return zap.Error(err) }

// Duration creates a duration field
func Duration(key string, value time.Duration) Field { return zap.Duration(key, value) }

// Any creates a field with any value
func Any(key string, value interface{}) Field { return zap.Any(key, value) }

type zapLogger struct {
	*zap.Logger
}

// NewLogger creates a new production logger instance
func NewLogger() Logger {
	// Set up logger configuration
	logLevel := zapcore.InfoLevel
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		switch strings.ToLower(level) {
		case "debug":
			logLevel = zapcore.DebugLevel
		case "info":
			logLevel = zapcore.InfoLevel
		case "warn":
			logLevel = zapcore.WarnLevel
		case "error":
			logLevel = zapcore.ErrorLevel
		}
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(logLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return &zapLogger{logger}
}

// NewTestLogger creates a logger suitable for testing
func NewTestLogger() Logger {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	logger, _ := config.Build()
	return &zapLogger{logger}
}

func (l *zapLogger) Info(msg string, fields ...Field)  { l.Logger.Info(msg, fields...) }
func (l *zapLogger) Error(msg string, fields ...Field) { l.Logger.Error(msg, fields...) }
func (l *zapLogger) Debug(msg string, fields ...Field) { l.Logger.Debug(msg, fields...) }
func (l *zapLogger) Warn(msg string, fields ...Field)  { l.Logger.Warn(msg, fields...) }
