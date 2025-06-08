// Package logging provides a unified logging interface using zap
package logging

import (
	"fmt"

	"github.com/goformx/goforms/internal/infrastructure/logging/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// LogEncodingConsole represents console encoding for logs
	LogEncodingConsole = "console"
	// LogEncodingJSON represents JSON encoding for logs
	LogEncodingJSON = "json"
)

// Config holds the configuration for creating a logger
type Config struct {
	Level   string
	AppName string
	Debug   bool
}

// FactoryConfig holds the configuration for creating a logger factory
type FactoryConfig struct {
	AppName     string
	Version     string
	Environment string
	Fields      map[string]any
}

// Factory creates loggers based on configuration
type Factory struct {
	initialFields map[string]any
	appName       string
	version       string
	environment   string
}

// NewFactory creates a new logger factory with the given configuration
func NewFactory(cfg FactoryConfig) *Factory {
	if cfg.Fields == nil {
		cfg.Fields = make(map[string]any)
	}

	// Ensure version is set
	if cfg.Version == "" {
		cfg.Version = "1.0.0"
	}

	// Add version to fields if not present
	if _, exists := cfg.Fields["version"]; !exists {
		cfg.Fields["version"] = cfg.Version
	}

	return &Factory{
		initialFields: cfg.Fields,
		appName:       cfg.AppName,
		version:       cfg.Version,
		environment:   cfg.Environment,
	}
}

// CreateLogger creates a new logger instance with the application name.
func (f *Factory) CreateLogger() (Logger, error) {
	// Create encoder config
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Create base logger
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.EncoderConfig = encoderConfig
	zapConfig.OutputPaths = []string{"stdout"}

	// Use console encoding for development, JSON for production
	if f.environment == "development" {
		zapConfig.Encoding = LogEncodingConsole
	} else {
		zapConfig.Encoding = LogEncodingJSON
	}

	zapLog, err := zapConfig.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(
			zap.String("app", f.appName),
			zap.String("version", f.version),
			zap.String("environment", f.environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &ZapLogger{log: zapLog}, nil
}

// CreateFromConfig creates a logger based on the provided configuration
func (f *Factory) CreateFromConfig(cfg *config.Config) (Logger, error) {
	// Create encoder config
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Parse log level
	var level zapcore.Level
	if cfg.Debug {
		level = zapcore.DebugLevel
	} else {
		levelErr := level.UnmarshalText([]byte(cfg.Level))
		if levelErr != nil {
			level = zapcore.InfoLevel // fallback
		}
	}

	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.EncoderConfig = encoderConfig
	zapConfig.OutputPaths = []string{"stdout"}

	// Use console encoding for development, JSON for production
	if f.environment == "development" {
		zapConfig.Encoding = LogEncodingConsole
	} else {
		zapConfig.Encoding = LogEncodingJSON
	}

	zapConfig.Level = zap.NewAtomicLevelAt(level)

	zapLog, err := zapConfig.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(
			zap.String("app", cfg.AppName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &ZapLogger{log: zapLog}, nil
}

// ZapLogger implements the Logger interface using zap
type ZapLogger struct {
	log *zap.Logger
}

// GetZapLogger returns the underlying zap logger
func (l *ZapLogger) GetZapLogger() *zap.Logger {
	return l.log
}

// Debug logs a debug message
func (l *ZapLogger) Debug(msg string, fields ...any) {
	l.log.Debug(msg, convertToZapFields(fields)...)
}

// Info logs an info message
func (l *ZapLogger) Info(msg string, fields ...any) {
	l.log.Info(msg, convertToZapFields(fields)...)
}

// Warn logs a warning message
func (l *ZapLogger) Warn(msg string, fields ...any) {
	l.log.Warn(msg, convertToZapFields(fields)...)
}

// Error logs an error message
func (l *ZapLogger) Error(msg string, fields ...any) {
	l.log.Error(msg, convertToZapFields(fields)...)
}

// Fatal logs a fatal message
func (l *ZapLogger) Fatal(msg string, fields ...any) {
	l.log.Fatal(msg, convertToZapFields(fields)...)
}

// With returns a new logger with the given fields
func (l *ZapLogger) With(fields ...any) Logger {
	zapFields := convertToZapFields(fields)

	return &ZapLogger{log: l.log.With(zapFields...)}
}

// WithComponent returns a new logger with the given component
func (l *ZapLogger) WithComponent(component string) Logger {
	return l.With(String("component", component))
}

// WithOperation returns a new logger with the given operation
func (l *ZapLogger) WithOperation(operation string) Logger {
	return l.With(String("operation", operation))
}

// WithRequestID returns a new logger with the given request ID
func (l *ZapLogger) WithRequestID(requestID string) Logger {
	return l.With(String("request_id", requestID))
}

// WithUserID returns a new logger with the given user ID
func (l *ZapLogger) WithUserID(userID string) Logger {
	return l.With(String("user_id", userID))
}

// WithError returns a new logger with the given error
func (l *ZapLogger) WithError(err error) Logger {
	return l.With(Error(err))
}

// WithFields adds multiple fields to the logger
func (l *ZapLogger) WithFields(fields map[string]any) Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &ZapLogger{log: l.log.With(zapFields...)}
}

func convertToZapFields(fields []any) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))

	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if !ok {
				continue
			}
			zapFields = append(zapFields, zap.Any(key, fields[i+1]))
		}
	}

	return zapFields
}
