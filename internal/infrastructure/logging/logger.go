// Package logging provides a unified logging interface using zap
package logging

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
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

// logger implements the Logger interface using zap
type logger struct {
	log *zap.Logger
}

// NewLogger creates a new logger instance
func NewLogger(cfg *config.Config) Logger {
	// Create encoder config
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	var zapLog *zap.Logger
	if cfg.App.Debug {
		// Development mode with colored output
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig = encoderConfig
		config.OutputPaths = []string{"stdout"}
		config.Encoding = "console"

		zapLog, _ = config.Build(
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel),
			zap.Fields(
				zap.String("app", cfg.App.Name),
				zap.String("host", cfg.App.Host),
				zap.Int("port", cfg.App.Port),
				zap.String("env", cfg.App.Env),
			),
		)
	} else {
		// Production mode with JSON output
		zapLog, _ = zap.NewProduction(
			zap.Fields(
				zap.String("app", cfg.App.Name),
				zap.String("host", cfg.App.Host),
				zap.Int("port", cfg.App.Port),
				zap.String("env", cfg.App.Env),
			),
		)
	}

	return &logger{
		log: zapLog,
	}
}

// NewTestLogger creates a logger suitable for testing
func NewTestLogger() Logger {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	zapLog, _ := config.Build()
	return &logger{log: zapLog}
}

func (l *logger) Info(msg string, fields ...Field)  { l.log.Info(msg, fields...) }
func (l *logger) Error(msg string, fields ...Field) { l.log.Error(msg, fields...) }
func (l *logger) Debug(msg string, fields ...Field) { l.log.Debug(msg, fields...) }
func (l *logger) Warn(msg string, fields ...Field)  { l.log.Warn(msg, fields...) }
