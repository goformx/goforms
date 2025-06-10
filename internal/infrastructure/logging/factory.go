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
	// MaxPartsLength represents the maximum number of parts in a log message
	MaxPartsLength = 2
	// FieldPairSize represents the number of elements in a key-value pair
	FieldPairSize = 2
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
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("15:04:05.000"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller: func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			// Show only the last two parts of the file path
			parts := strings.Split(caller.File, "/")
			if len(parts) > MaxPartsLength {
				parts = parts[len(parts)-MaxPartsLength:]
			}
			file := strings.Join(parts, "/")
			enc.AppendString(fmt.Sprintf("%s:%d", file, caller.Line))
		},
	}

	// Create console encoder for better readability in development mode
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// Create core with console encoder
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	)

	// Create logger with options
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Development(),
	)

	// Create initial fields
	fields := make([]zap.Field, 0, len(f.initialFields))
	for k, v := range f.initialFields {
		fields = append(fields, zap.String(k, fmt.Sprintf("%v", v)))
	}

	// Create logger with initial fields
	logger = logger.With(fields...)

	return &ZapLogger{
		logger: logger,
	}, nil
}

// ZapLogger implements the Logger interface using zap
type ZapLogger struct {
	logger *zap.Logger
}

// GetZapLogger returns the underlying zap logger
func (l *ZapLogger) GetZapLogger() *zap.Logger {
	return l.logger
}

// Debug logs a debug message
func (l *ZapLogger) Debug(msg string, fields ...any) {
	l.logger.Debug(msg, convertToZapFields(fields)...)
}

// Info logs an info message
func (l *ZapLogger) Info(msg string, fields ...any) {
	l.logger.Info(msg, convertToZapFields(fields)...)
}

// Warn logs a warning message
func (l *ZapLogger) Warn(msg string, fields ...any) {
	l.logger.Warn(msg, convertToZapFields(fields)...)
}

// Error logs an error message
func (l *ZapLogger) Error(msg string, fields ...any) {
	l.logger.Error(msg, convertToZapFields(fields)...)
}

// Fatal logs a fatal message
func (l *ZapLogger) Fatal(msg string, fields ...any) {
	l.logger.Fatal(msg, convertToZapFields(fields)...)
}

// With returns a new logger with the given fields
func (l *ZapLogger) With(fields ...any) Logger {
	zapFields := convertToZapFields(fields)

	return &ZapLogger{logger: l.logger.With(zapFields...)}
}

// WithComponent returns a new logger with the given component
func (l *ZapLogger) WithComponent(component string) Logger {
	return l.With("component", component)
}

// WithOperation returns a new logger with the given operation
func (l *ZapLogger) WithOperation(operation string) Logger {
	return l.With("operation", operation)
}

// WithRequestID returns a new logger with the given request ID
func (l *ZapLogger) WithRequestID(requestID string) Logger {
	return l.With("request_id", requestID)
}

// WithUserID returns a new logger with the given user ID
func (l *ZapLogger) WithUserID(userID string) Logger {
	return l.With("user_id", userID)
}

// WithError returns a new logger with the given error
func (l *ZapLogger) WithError(err error) Logger {
	return l.With("error", err)
}

// WithFields adds multiple fields to the logger
func (l *ZapLogger) WithFields(fields map[string]any) Logger {
	zapFields := make([]zap.Field, 0, len(fields)/FieldPairSize)
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &ZapLogger{logger: l.logger.With(zapFields...)}
}

// convertToZapFields converts a slice of fields to zap fields
func convertToZapFields(fields []any) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields)/FieldPairSize)
	for i := 0; i < len(fields); i += FieldPairSize {
		if i+1 >= len(fields) {
			break
		}

		key, ok := fields[i].(string)
		if !ok {
			continue
		}

		value := fields[i+1]
		switch v := value.(type) {
		case string:
			zapFields = append(zapFields, zap.String(key, v))
		case int:
			zapFields = append(zapFields, zap.Int(key, v))
		case int64:
			zapFields = append(zapFields, zap.Int64(key, v))
		case uint:
			zapFields = append(zapFields, zap.Uint(key, v))
		case uint64:
			zapFields = append(zapFields, zap.Uint64(key, v))
		case float64:
			zapFields = append(zapFields, zap.Float64(key, v))
		case bool:
			zapFields = append(zapFields, zap.Bool(key, v))
		case error:
			zapFields = append(zapFields, zap.Error(v), zap.String(key+"_details", fmt.Sprintf("%+v", v)))
		default:
			zapFields = append(zapFields, zap.Any(key, v))
		}
	}
	return zapFields
}
