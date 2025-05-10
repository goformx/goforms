// Package logging provides a unified logging interface using zap
package logging

import (
	"fmt"
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
	// Int64 adds an int64 field to the log entry
	Int64(key string, value int64) Field
	// Int adds an int field to the log entry
	Int(key string, value int) Field
	// Int32 adds an int32 field to the log entry
	Int32(key string, value int32) Field
	// Uint64 adds a uint64 field to the log entry
	Uint64(key string, value uint64) Field
	// Uint adds a uint field to the log entry
	Uint(key string, value uint) Field
	// Uint32 adds a uint32 field to the log entry
	Uint32(key string, value uint32) Field
}

// Field represents a logging field
type Field = zap.Field

// String creates a string field
func String(key, value string) Field { return zap.String(key, value) }

// Int creates an integer field
func Int(key string, value int) Field { return zap.Int(key, value) }

// Int64 creates a 64-bit integer field
func Int64(key string, value int64) Field { return zap.Int64(key, value) }

// Uint creates an unsigned integer field
func Uint(key string, value uint) Field { return zap.Uint(key, value) }

// Bool creates a boolean field
func Bool(key string, value bool) Field { return zap.Bool(key, value) }

// Error creates an error field
func Error(err error) Field { return zap.Error(err) }

// Duration creates a duration field
func Duration(key string, value time.Duration) Field { return zap.Duration(key, value) }

// Any creates a field with any value
func Any(key string, value any) Field { return zap.Any(key, value) }

// logger implements the Logger interface using zap
type logger struct {
	log *zap.Logger
}

// NoopLogger is a logger that does nothing
type noopLogger struct{}

// NewNoopLogger creates a new no-op logger
func NewNoopLogger() Logger {
	return &noopLogger{}
}

// NewLogger creates a new logger instance
func NewLogger(logLevel, appName string) (Logger, error) {
	// Create encoder config
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	var zapLog *zap.Logger

	// Parse log level
	var level zapcore.Level
	levelErr := level.UnmarshalText([]byte(logLevel))
	if levelErr != nil {
		level = zapcore.InfoLevel // fallback
	}

	config := zap.NewDevelopmentConfig()
	config.EncoderConfig = encoderConfig
	config.OutputPaths = []string{"stdout"}
	config.Encoding = "console"
	config.Level = zap.NewAtomicLevelAt(level)

	// Use JSON encoding for production
	if level >= zapcore.WarnLevel {
		config.Encoding = "json"
	}

	zapLog, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(
			zap.String("app", appName),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &logger{log: zapLog}, nil
}

// NewTestLogger creates a logger suitable for testing
func NewTestLogger() (Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	zapLog, err := config.Build(
		zap.Fields(
			zap.String("app", "test"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create test logger: %w", err)
	}
	return &logger{log: zapLog}, nil
}

// NewZapLogger creates a new logger from a zap logger
func NewZapLogger(zapLog *zap.Logger) Logger {
	return &logger{log: zapLog}
}

func (l *logger) Info(msg string, fields ...Field)  { l.log.Info(msg, fields...) }
func (l *logger) Error(msg string, fields ...Field) { l.log.Error(msg, fields...) }
func (l *logger) Debug(msg string, fields ...Field) { l.log.Debug(msg, fields...) }
func (l *logger) Warn(msg string, fields ...Field)  { l.log.Warn(msg, fields...) }

// Int64 adds an int64 field to the log entry
func (l *logger) Int64(key string, value int64) Field {
	return zap.Int64(key, value)
}

// Int adds an int field to the log entry
func (l *logger) Int(key string, value int) Field {
	return zap.Int(key, value)
}

// Int32 adds an int32 field to the log entry
func (l *logger) Int32(key string, value int32) Field {
	return zap.Int32(key, value)
}

// Uint64 adds a uint64 field to the log entry
func (l *logger) Uint64(key string, value uint64) Field {
	return zap.Uint64(key, value)
}

// Uint adds a uint field to the log entry
func (l *logger) Uint(key string, value uint) Field {
	return zap.Uint(key, value)
}

// Uint32 adds a uint32 field to the log entry
func (l *logger) Uint32(key string, value uint32) Field {
	return zap.Uint32(key, value)
}

func (l *noopLogger) Info(msg string, fields ...Field)  {}
func (l *noopLogger) Error(msg string, fields ...Field) {}
func (l *noopLogger) Debug(msg string, fields ...Field) {}
func (l *noopLogger) Warn(msg string, fields ...Field)  {}

func (l *noopLogger) Int64(key string, value int64) Field {
	return zap.Int64(key, value)
}

func (l *noopLogger) Int(key string, value int) Field {
	return zap.Int(key, value)
}

func (l *noopLogger) Int32(key string, value int32) Field {
	return zap.Int32(key, value)
}

func (l *noopLogger) Uint64(key string, value uint64) Field {
	return zap.Uint64(key, value)
}

func (l *noopLogger) Uint(key string, value uint) Field {
	return zap.Uint(key, value)
}

func (l *noopLogger) Uint32(key string, value uint32) Field {
	return zap.Uint32(key, value)
}
