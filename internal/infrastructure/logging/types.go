// Package logging provides a unified logging interface
package logging

// Logger defines the interface for application logging
type Logger interface {
	// Basic logging methods
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
	Fatal(msg string, fields ...any)

	// Context methods
	With(fields ...any) Logger
	WithComponent(component string) Logger
	WithOperation(operation string) Logger
	WithRequestID(requestID string) Logger
	WithUserID(userID string) Logger
	WithError(err error) Logger
	WithFields(fields map[string]any) Logger
}
