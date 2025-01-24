// Package logging provides a unified logging interface using zap
package logging

import (
	"os"
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

type zapLogger struct {
	*zap.Logger
}

// NewLogger creates a new logger instance
func NewLogger(cfg *config.AppConfig) Logger {
	// Set log level based on APP_DEBUG
	var level zapcore.Level
	if cfg.Debug {
		level = zapcore.DebugLevel
	} else {
		level = zapcore.InfoLevel
	}

	logFormat := os.Getenv("LOG_FORMAT")
	if logFormat == "" {
		logFormat = "json"
	}

	// Configure encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "ts"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if logFormat == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	logger := zap.New(core, zap.Fields(
		zap.String("app", cfg.Name),
		zap.String("env", cfg.Env),
	))
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
