// Package logging provides a unified logging interface using zap
package logging

import (
	"fmt"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds the configuration for creating a logger
type Config struct {
	Level   string
	AppName string
}

// Factory creates loggers based on configuration
type Factory struct {
	initialFields map[string]any
}

// NewFactory creates a new logger factory
func NewFactory() *Factory {
	return &Factory{
		initialFields: map[string]any{
			"version": "1.0.0",
		},
	}
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
	levelErr := level.UnmarshalText([]byte(cfg.App.LogLevel))
	if levelErr != nil {
		level = zapcore.InfoLevel // fallback
	}

	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.EncoderConfig = encoderConfig
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.Encoding = "console"
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Use JSON encoding for production
	if level >= zapcore.WarnLevel {
		zapConfig.Encoding = "json"
	}

	zapLog, err := zapConfig.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(
			zap.String("app", cfg.App.Name),
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

func (l *ZapLogger) Info(msg string, fields ...LogField)  { l.log.Info(msg, convertFields(fields)...) }
func (l *ZapLogger) Error(msg string, fields ...LogField) { l.log.Error(msg, convertFields(fields)...) }
func (l *ZapLogger) Debug(msg string, fields ...LogField) { l.log.Debug(msg, convertFields(fields)...) }
func (l *ZapLogger) Warn(msg string, fields ...LogField)  { l.log.Warn(msg, convertFields(fields)...) }
func (l *ZapLogger) Fatal(msg string, fields ...LogField) { l.log.Fatal(msg, convertFields(fields)...) }

// With returns a new logger with the given fields
func (l *ZapLogger) With(fields ...LogField) Logger {
	return &ZapLogger{log: l.log.With(convertFields(fields)...)}
}

// convertFields converts our LogField to zap.Field
func convertFields(fields []LogField) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		switch f.Type {
		case StringType:
			zapFields[i] = zap.String(f.Key, f.String)
		case IntType:
			zapFields[i] = zap.Int(f.Key, f.Int)
		case ErrorType:
			zapFields[i] = zap.Error(f.Error)
		case DurationType:
			zapFields[i] = zap.String(f.Key, f.String)
		case BoolType:
			zapFields[i] = zap.String(f.Key, f.String)
		case AnyType:
			zapFields[i] = zap.String(f.Key, f.String)
		case UintType:
			zapFields[i] = zap.Uint(f.Key, f.Uint)
		}
	}
	return zapFields
}

// CreateTestLogger creates a logger for testing
func (f *Factory) CreateTestLogger() (Logger, error) {
	return NewTestLogger()
}

// AuthLogger returns the shared logger for auth middleware
func (f *Factory) AuthLogger(logger Logger) Logger {
	return logger
}

// CookieAuthLogger returns the shared logger for cookie auth middleware
func (f *Factory) CookieAuthLogger(logger Logger) Logger {
	return logger
}
