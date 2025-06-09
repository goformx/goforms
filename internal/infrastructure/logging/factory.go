// Package logging provides a unified logging interface using zap
package logging

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// LogEncodingConsole represents console encoding for logs
	LogEncodingConsole = "console"
	// LogEncodingJSON represents JSON encoding for logs
	LogEncodingJSON = "json"
	// EnvironmentDevelopment represents the development environment
	EnvironmentDevelopment = "development"

	// Default environment variables
	envLogLevel = "GOFORMS_APP_LOGLEVEL"
)

// FactoryConfig holds the configuration for creating a logger factory
type FactoryConfig struct {
	AppName     string
	Version     string
	Environment string
	Fields      map[string]any
	// OutputPaths specifies where to write logs
	OutputPaths []string
	// ErrorOutputPaths specifies where to write error logs
	ErrorOutputPaths []string
}

// Factory creates loggers based on configuration
type Factory struct {
	initialFields map[string]any
	appName       string
	version       string
	environment   string
	outputPaths   []string
	errorPaths    []string
}

// NewFactory creates a new logger factory with the given configuration
func NewFactory(cfg FactoryConfig) *Factory {
	if cfg.Fields == nil {
		cfg.Fields = make(map[string]any)
	}

	// Set default output paths if not specified
	if len(cfg.OutputPaths) == 0 {
		cfg.OutputPaths = []string{"stdout"}
	}
	if len(cfg.ErrorOutputPaths) == 0 {
		cfg.ErrorOutputPaths = []string{"stderr"}
	}

	return &Factory{
		initialFields: cfg.Fields,
		appName:       cfg.AppName,
		version:       cfg.Version,
		environment:   cfg.Environment,
		outputPaths:   cfg.OutputPaths,
		errorPaths:    cfg.ErrorOutputPaths,
	}
}

// CreateLogger creates a new logger instance with the application name.
func (f *Factory) CreateLogger() (Logger, error) {
	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Parse log level from environment
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(getEnv(envLogLevel, "info"))); err != nil {
		level = zapcore.InfoLevel // fallback to info level
	}

	// Determine encoding based on environment
	encoding := LogEncodingConsole
	if f.environment != EnvironmentDevelopment {
		encoding = LogEncodingJSON
	}

	// Create base logger config
	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       f.environment == EnvironmentDevelopment,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil, // Disable sampling to show all logs
		Encoding:          encoding,
		EncoderConfig:     encoderConfig,
		OutputPaths:       f.outputPaths,
		ErrorOutputPaths:  f.errorPaths,
	}

	// Build initial fields from config only
	var initialFields []zap.Field
	for k, v := range f.initialFields {
		switch val := v.(type) {
		case string:
			initialFields = append(initialFields, zap.String(k, val))
		case int:
			initialFields = append(initialFields, zap.Int(k, val))
		case bool:
			initialFields = append(initialFields, zap.Bool(k, val))
		case float64:
			initialFields = append(initialFields, zap.Float64(k, val))
		default:
			initialFields = append(initialFields, zap.Any(k, val))
		}
	}

	// Build logger with options
	zapLog, err := zapConfig.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.WarnLevel), // Enable stack traces for warnings and above
		zap.Fields(initialFields...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &ZapLogger{log: zapLog}, nil
}

// getEnv gets an environment variable or returns the default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.ToLower(value)
	}
	return defaultValue
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
