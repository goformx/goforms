// Package logging provides a unified logging interface
package logging

// Logger defines the interface for application logging
type Logger interface {
	// Basic logging methods
	// Debug is used for detailed information, typically useful only for diagnosing problems
	Debug(msg string, fields ...any)
	// Info is used for general operational entries about what's happening inside the application
	Info(msg string, fields ...any)
	// Warn is used for potentially harmful situations
	Warn(msg string, fields ...any)
	// Error is used for error events that might still allow the application to continue running
	Error(msg string, fields ...any)
	// Fatal is used for very severe error events that will presumably lead the application to abort
	Fatal(msg string, fields ...any)

	// Context methods
	// With adds fields to the logger context
	With(fields ...any) Logger
	// WithComponent adds a component name to the logger context
	WithComponent(component string) Logger
	// WithOperation adds an operation name to the logger context
	WithOperation(operation string) Logger
	// WithRequestID adds a request ID to the logger context
	WithRequestID(requestID string) Logger
	// WithUserID adds a user ID to the logger context (sanitized)
	WithUserID(userID string) Logger
	// WithError adds an error to the logger context
	WithError(err error) Logger
	// WithFields adds multiple fields to the logger context
	WithFields(fields map[string]any) Logger
	// SanitizeField returns a sanitized version of a field value
	SanitizeField(key string, value any) any
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
