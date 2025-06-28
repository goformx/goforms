// Package logging provides a unified logging interface
//
//go:generate mockgen -typed -source=types.go -destination=../../../test/mocks/logging/mock_logger.go -package=logging
package logging

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
	// MaxUserAgentLength represents the maximum length for user agent fields
	MaxUserAgentLength = 1000
	// UUIDLength represents the standard UUID length
	UUIDLength = 36
	// UUIDParts represents the number of parts in a UUID
	UUIDParts = 5
	// UUIDMinMaskLen represents the minimum length for UUID masking
	UUIDMinMaskLen = 8
	// UUIDMaskPrefixLen represents the prefix length for UUID masking
	UUIDMaskPrefixLen = 4
	// UUIDMaskSuffixLen represents the suffix length for UUID masking
	UUIDMaskSuffixLen = 4
)

// Logger interface defines the logging contract
type Logger interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
	Fatal(msg string, fields ...any)
	With(fields ...any) Logger
	WithComponent(component string) Logger
	WithOperation(operation string) Logger
	WithRequestID(requestID string) Logger
	WithUserID(userID string) Logger
	WithError(err error) Logger
	WithFields(fields map[string]any) Logger
	// New Field-based API methods
	WithFieldsStructured(fields ...Field) Logger
	DebugWithFields(msg string, fields ...Field)
	InfoWithFields(msg string, fields ...Field)
	WarnWithFields(msg string, fields ...Field)
	ErrorWithFields(msg string, fields ...Field)
	FatalWithFields(msg string, fields ...Field)
	SanitizeField(key string, value any) string
}

// FactoryConfig holds configuration for logger factory
type FactoryConfig struct {
	AppName          string
	Version          string
	Environment      string
	LogLevel         string
	OutputPaths      []string
	ErrorOutputPaths []string
	Fields           map[string]any
}

// LogLevel represents the severity of a log message
type LogLevel string

const (
	// LogLevelDebug represents debug level logging
	LogLevelDebug LogLevel = "debug"
	// LogLevelInfo represents info level logging
	LogLevelInfo LogLevel = "info"
	// LogLevelWarn represents warning level logging
	LogLevelWarn LogLevel = "warn"
	// LogLevelError represents error level logging
	LogLevelError LogLevel = "error"
	// LogLevelFatal represents fatal level logging
	LogLevelFatal LogLevel = "fatal"
)
