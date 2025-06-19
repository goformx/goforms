// Package logging provides a unified logging interface
package logging

import (
	"fmt"
	"os"
	"strings"

	"github.com/goformx/goforms/internal/infrastructure/sanitization"
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
	// MaxStringLength represents the maximum length for string fields
	MaxStringLength = 1000
	// MaxPathLength represents the maximum length for path fields
	MaxPathLength = 500
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
	sanitizer     sanitization.ServiceInterface
	// Add testCore for test injection
	testCore zapcore.Core
}

// NewFactory creates a new logger factory with the given configuration
func NewFactory(cfg FactoryConfig, sanitizer sanitization.ServiceInterface) *Factory {
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
		sanitizer:     sanitizer,
	}
}

// WithTestCore allows tests to inject a zapcore.Core for capturing logs
func (f *Factory) WithTestCore(core zapcore.Core) *Factory {
	f.testCore = core
	return f
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

	// Determine log level from environment
	var level zapcore.Level
	switch strings.ToLower(f.environment) {
	case "development":
		level = zapcore.DebugLevel
	default:
		level = zapcore.InfoLevel
	}

	// Use testCore if set (for testing)
	var core zapcore.Core
	if f.testCore != nil {
		core = f.testCore
	} else {
		core = zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
	}

	// Create logger with options
	zapLogger := zap.New(core,
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
	zapLogger = zapLogger.With(fields...)

	// Create our logger implementation
	return newLogger(zapLogger, f.sanitizer), nil
}

// logger implements the Logger interface using zap
type logger struct {
	zapLogger *zap.Logger
	sanitizer sanitization.ServiceInterface
}

// newLogger creates a new logger instance
func newLogger(zapLogger *zap.Logger, sanitizer sanitization.ServiceInterface) Logger {
	return &logger{
		zapLogger: zapLogger,
		sanitizer: sanitizer,
	}
}

// sanitizeMessage sanitizes a log message to prevent log injection attacks
func sanitizeMessage(msg string, sanitizer sanitization.ServiceInterface) string {
	return sanitizer.SanitizeForLogging(msg)
}

// sanitizeError sanitizes an error for safe logging
func sanitizeError(err error, sanitizer sanitization.ServiceInterface) string {
	if err == nil {
		return ""
	}

	// Get the error message and sanitize it
	errMsg := err.Error()

	// Apply the same sanitization as regular messages
	return sanitizer.SanitizeForLogging(errMsg)
}

// Debug logs a debug message
func (l *logger) Debug(msg string, fields ...any) {
	l.zapLogger.Debug(sanitizeMessage(msg, l.sanitizer), convertToZapFields(fields, l.sanitizer)...)
}

// Info logs an info message
func (l *logger) Info(msg string, fields ...any) {
	l.zapLogger.Info(sanitizeMessage(msg, l.sanitizer), convertToZapFields(fields, l.sanitizer)...)
}

// Warn logs a warning message
func (l *logger) Warn(msg string, fields ...any) {
	l.zapLogger.Warn(sanitizeMessage(msg, l.sanitizer), convertToZapFields(fields, l.sanitizer)...)
}

// Error logs an error message
func (l *logger) Error(msg string, fields ...any) {
	l.zapLogger.Error(sanitizeMessage(msg, l.sanitizer), convertToZapFields(fields, l.sanitizer)...)
}

// Fatal logs a fatal message
func (l *logger) Fatal(msg string, fields ...any) {
	l.zapLogger.Fatal(sanitizeMessage(msg, l.sanitizer), convertToZapFields(fields, l.sanitizer)...)
}

// With returns a new logger with the given fields
func (l *logger) With(fields ...any) Logger {
	zapFields := convertToZapFields(fields, l.sanitizer)
	return newLogger(l.zapLogger.With(zapFields...), l.sanitizer)
}

// WithComponent returns a new logger with the given component
func (l *logger) WithComponent(component string) Logger {
	return l.With("component", component)
}

// WithOperation returns a new logger with the given operation
func (l *logger) WithOperation(operation string) Logger {
	return l.With("operation", operation)
}

// WithRequestID returns a new logger with the given request ID
func (l *logger) WithRequestID(requestID string) Logger {
	return l.With("request_id", requestID)
}

// WithUserID returns a new logger with the given user ID
func (l *logger) WithUserID(userID string) Logger {
	return l.With("user_id", l.SanitizeField("user_id", userID))
}

// WithError returns a new logger with the given error
func (l *logger) WithError(err error) Logger {
	return l.With("error", sanitizeError(err, l.sanitizer))
}

// WithFields adds multiple fields to the logger
func (l *logger) WithFields(fields map[string]any) Logger {
	zapFields := make([]zap.Field, 0, len(fields)/FieldPairSize)
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, l.SanitizeField(k, v)))
	}
	return newLogger(l.zapLogger.With(zapFields...), l.sanitizer)
}

var sensitiveKeys = map[string]struct{}{
	"password":           {},
	"token":              {},
	"secret":             {},
	"key":                {},
	"credential":         {},
	"authorization":      {},
	"cookie":             {},
	"session":            {},
	"api_key":            {},
	"access_token":       {},
	"refresh_token":      {},
	"private_key":        {},
	"public_key":         {},
	"certificate":        {},
	"ssn":                {},
	"credit_card":        {},
	"bank_account":       {},
	"phone":              {},
	"email":              {},
	"address":            {},
	"dob":                {},
	"birth_date":         {},
	"social_security":    {},
	"tax_id":             {},
	"driver_license":     {},
	"passport":           {},
	"national_id":        {},
	"health_record":      {},
	"medical_record":     {},
	"insurance":          {},
	"benefit":            {},
	"salary":             {},
	"compensation":       {},
	"bank_routing":       {},
	"bank_swift":         {},
	"iban":               {},
	"account_number":     {},
	"pin":                {},
	"cvv":                {},
	"cvc":                {},
	"security_code":      {},
	"verification_code":  {},
	"otp":                {},
	"mfa_code":           {},
	"2fa_code":           {},
	"recovery_code":      {},
	"backup_code":        {},
	"reset_token":        {},
	"activation_code":    {},
	"verification_token": {},
	"invite_code":        {},
	"referral_code":      {},
	"promo_code":         {},
	"discount_code":      {},
	"coupon_code":        {},
	"gift_card":          {},
	"voucher":            {},
	"license_key":        {},
	"product_key":        {},
	"serial_number":      {},
	"activation_key":     {},
	"registration_key":   {},
	"subscription_key":   {},
	"membership_key":     {},
	"access_code":        {},
	"security_key":       {},
	"encryption_key":     {},
	"decryption_key":     {},
	"signing_key":        {},
	"verification_key":   {},
	"authentication_key": {},
	"authorization_key":  {},
	"session_key":        {},
	"cookie_key":         {},
	"csrf_token":         {},
	"xsrf_token":         {},
	"jwt":                {},
	"jwe":                {},
	"jws":                {},
	"oauth_token":        {},
	"oauth_secret":       {},
	"oauth_verifier":     {},
	"oauth_code":         {},
	"oauth_state":        {},
	"oauth_nonce":        {},
	"oauth_scope":        {},
	"oauth_grant":        {},
	"oauth_refresh":      {},
	"oauth_access":       {},
	"oauth_id":           {},
	"oauth_key":          {},
}

// SanitizeField returns a masked version of a sensitive field value
func (l *logger) SanitizeField(key string, value any) string {
	// Check for sensitive keys
	if _, ok := sensitiveKeys[strings.ToLower(key)]; ok {
		return "****"
	}

	// Handle error values specially
	if err, ok := value.(error); ok {
		return sanitizeError(err, l.sanitizer)
	}

	// Handle path fields
	if key == "path" {
		if path, ok := value.(string); ok {
			if !validatePath(path) {
				return "[invalid path]"
			}
			return sanitizeString(truncateString(path, MaxPathLength), l.sanitizer)
		}
		return "[invalid path type]"
	}

	// Handle string values
	if str, ok := value.(string); ok {
		return sanitizeString(truncateString(str, MaxStringLength), l.sanitizer)
	}

	// For other types, convert to string and sanitize
	return sanitizeString(fmt.Sprintf("%v", value), l.sanitizer)
}

// convertToZapFields converts a slice of fields to zap fields
func convertToZapFields(fields []any, sanitizer sanitization.ServiceInterface) []zap.Field {
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
		// Sanitize the value based on its type
		sanitizedValue := sanitizeValue(key, value, sanitizer)

		// Always append as string since sanitizeValue returns string
		zapFields = append(zapFields, zap.String(key, sanitizedValue))
	}
	return zapFields
}

// sanitizeValue applies appropriate sanitization based on the field type
func sanitizeValue(key string, value any, sanitizer sanitization.ServiceInterface) string {
	// Check for sensitive keys
	if _, ok := sensitiveKeys[strings.ToLower(key)]; ok {
		return "****"
	}

	// Handle error values specially
	if err, ok := value.(error); ok {
		return sanitizeError(err, sanitizer)
	}

	// Handle path fields
	if key == "path" {
		if path, ok := value.(string); ok {
			if !validatePath(path) {
				return "[invalid path]"
			}
			return sanitizeString(truncateString(path, MaxPathLength), sanitizer)
		}
		return "[invalid path type]"
	}

	// Handle string values
	if str, ok := value.(string); ok {
		return sanitizeString(truncateString(str, MaxStringLength), sanitizer)
	}

	// For other types, convert to string and sanitize
	return sanitizeString(fmt.Sprintf("%v", value), sanitizer)
}

// sanitizeString sanitizes a string for safe logging
func sanitizeString(s string, sanitizer sanitization.ServiceInterface) string {
	return sanitizer.SanitizeForLogging(s)
}

// validatePath checks if a string is a valid URL path
func validatePath(path string) bool {
	if len(path) > MaxPathLength {
		return false
	}
	// Basic path validation - should start with / and contain only valid characters
	return path != "" && path[0] == '/' && !strings.ContainsAny(path, "\\<>\"'")
}

// truncateString truncates a string to the maximum allowed length
func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}
