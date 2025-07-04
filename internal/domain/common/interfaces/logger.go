// Package interfaces provides domain-specific interfaces that define contracts
// for external dependencies without creating coupling to specific implementations.
package interfaces

import (
	"context"
)

// Logger defines the interface for logging functionality that domain layer can use
// This interface should be implemented by the infrastructure layer
type Logger interface {
	// Info logs an informational message
	Info(msg string, fields ...any)

	// Error logs an error message
	Error(msg string, fields ...any)

	// Debug logs a debug message
	Debug(msg string, fields ...any)

	// Warn logs a warning message
	Warn(msg string, fields ...any)

	// WithContext returns a logger with context
	WithContext(ctx context.Context) Logger

	// WithFields returns a logger with additional fields
	WithFields(fields ...any) Logger
}

// LoggerFactory defines the interface for creating loggers
type LoggerFactory interface {
	// NewLogger creates a new logger instance
	NewLogger() (Logger, error)

	// NewLoggerWithContext creates a new logger with context
	NewLoggerWithContext(ctx context.Context) (Logger, error)
}
