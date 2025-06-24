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
	// UUIDLength represents the standard UUID length
	UUIDLength = 36
	// UUIDParts represents the number of parts in a UUID
	UUIDParts = 5
	// UUIDMinMaskLen represents the minimum length for UUID masking
	UUIDMinMaskLen = 8
	// UUIDMaskPrefixLen represents the prefix length for UUID masking
	UUIDMaskPrefixLen = 8
	// UUIDMaskSuffixLen represents the suffix length for UUID masking
	UUIDMaskSuffixLen = 4
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
	LogLevel         string
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
	LogLevel string
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
		LogLevel:      cfg.LogLevel,
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

	// Determine log level from config
	var level zapcore.Level
	if f.LogLevel != "" {
		switch strings.ToLower(f.LogLevel) {
		case "debug":
			level = zapcore.DebugLevel
		case "info":
			level = zapcore.InfoLevel
		case "warn":
			level = zapcore.WarnLevel
		case "error":
			level = zapcore.ErrorLevel
		case "fatal":
			level = zapcore.FatalLevel
		default:
			level = zapcore.InfoLevel
		}
	} else {
		switch strings.ToLower(f.environment) {
		case "development":
			level = zapcore.DebugLevel
		default:
			level = zapcore.InfoLevel
		}
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

// SanitizeField returns a masked version of a sensitive field value
func (l *logger) SanitizeField(key string, value any) string {
	// Handle error values specially
	if err, ok := value.(error); ok {
		return sanitizeError(err, l.sanitizer)
	}

	// Handle path fields
	if key == "path" {
		return sanitizePathField(value, l.sanitizer)
	}

	// Handle user agent fields
	if key == "user_agent" {
		return sanitizeUserAgentField(value, l.sanitizer)
	}

	// Handle UUID-like fields (form_id, user_id, etc.)
	if isUUIDField(key) {
		return sanitizeUUIDField(value)
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

// sanitizeRequestID handles request_id field validation
func sanitizeRequestID(value any) string {
	if id, ok := value.(string); ok {
		if !validateUUID(id) {
			return "[invalid request id]"
		}
		return id
	}
	return "[invalid request id type]"
}

// sanitizePathField handles path field validation and sanitization
func sanitizePathField(value any, sanitizer sanitization.ServiceInterface) string {
	if path, ok := value.(string); ok {
		if !validatePath(path) {
			return "[invalid path]"
		}
		return sanitizeString(truncateString(path, MaxPathLength), sanitizer)
	}
	return "[invalid path type]"
}

// sanitizeUserAgentField handles user agent field validation and sanitization
func sanitizeUserAgentField(value any, sanitizer sanitization.ServiceInterface) string {
	if ua, ok := value.(string); ok {
		if !validateUserAgent(ua) {
			return "[invalid user agent]"
		}
		return sanitizeString(truncateString(ua, MaxStringLength), sanitizer)
	}
	return "[invalid user agent type]"
}

// sanitizeUUIDField handles UUID-like field validation and masking
func sanitizeUUIDField(value any) string {
	if id, ok := value.(string); ok {
		if !validateUUID(id) {
			return "[invalid uuid format]"
		}
		// For UUIDs, we return a masked version for security
		if len(id) >= UUIDMinMaskLen {
			return id[:UUIDMaskPrefixLen] + "..." + id[len(id)-UUIDMaskSuffixLen:]
		}
		return "[invalid uuid length]"
	}
	return "[invalid uuid type]"
}

// isUUIDField checks if a field key represents a UUID field that should be masked
func isUUIDField(key string) bool {
	return strings.Contains(strings.ToLower(key), "id") &&
		!strings.Contains(strings.ToLower(key), "length") &&
		key != "request_id"
}

// sanitizeValue applies appropriate sanitization based on the field type
func sanitizeValue(key string, value any, sanitizer sanitization.ServiceInterface) string {
	// Special handling for request_id
	if key == "request_id" {
		return sanitizeRequestID(value)
	}

	// Handle error values specially
	if err, ok := value.(error); ok {
		return sanitizeError(err, sanitizer)
	}

	// Handle path fields
	if key == "path" {
		return sanitizePathField(value, sanitizer)
	}

	// Handle user agent fields
	if key == "user_agent" {
		return sanitizeUserAgentField(value, sanitizer)
	}

	// Handle UUID-like fields (form_id, user_id, etc.)
	if isUUIDField(key) {
		return sanitizeUUIDField(value)
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
	if path == "" || path[0] != '/' {
		return false
	}

	// Check for potentially dangerous characters
	dangerousChars := []string{"\\", "<", ">", "\"", "'", "\x00", "\n", "\r"}
	for _, char := range dangerousChars {
		if strings.Contains(path, char) {
			return false
		}
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") || strings.Contains(path, "//") {
		return false
	}

	return true
}

// truncateString truncates a string to the maximum allowed length
func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

// validateUUID checks if a string is a valid UUID format
func validateUUID(uuidStr string) bool {
	if len(uuidStr) != UUIDLength { // Standard UUID length
		return false
	}

	// Check for valid UUID characters (hex + hyphens)
	validChars := "0123456789abcdefABCDEF-"
	for _, char := range uuidStr {
		if !strings.ContainsRune(validChars, char) {
			return false
		}
	}

	// Check UUID format (8-4-4-4-12)
	parts := strings.Split(uuidStr, "-")
	if len(parts) != UUIDParts {
		return false
	}

	// Check each part length
	expectedLengths := []int{8, 4, 4, 4, 12}
	for i, part := range parts {
		if len(part) != expectedLengths[i] {
			return false
		}
	}

	return true
}

// validateUserAgent checks if a string is a valid user agent
func validateUserAgent(userAgent string) bool {
	if len(userAgent) > MaxStringLength {
		return false
	}

	// Check for potentially dangerous characters in user agent
	dangerousChars := []string{"\x00", "\n", "\r", "<", ">", "\"", "'"}
	for _, char := range dangerousChars {
		if strings.Contains(userAgent, char) {
			return false
		}
	}

	// Check for suspicious patterns
	suspiciousPatterns := []string{"<script", "javascript:", "vbscript:", "onload=", "onerror="}
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(strings.ToLower(userAgent), pattern) {
			return false
		}
	}

	return true
}
