// Package handler provides HTTP request handlers following a consistent pattern for
// dependency injection and configuration. Each handler type follows these patterns:
//
// 1. Options Pattern:
//   - Each handler accepts functional options for configuration
//   - Options are type-safe and immutable
//   - Dependencies are explicitly declared and validated
//
// 2. Base Handler:
//   - Provides common functionality (logging, error handling)
//   - Embedded in all specific handlers
//   - Requires explicit logger configuration
//
// 3. Validation:
//   - All handlers validate their dependencies before use
//   - Required dependencies are checked at startup
//   - Clear error messages for missing dependencies
//
// Example Usage:
//
//	handler := NewContactHandler(logger,
//	    WithContactServiceOpt(contactService),
//	)
//
// For testing:
//
//	handler := NewContactHandler(testLogger,
//	    WithContactServiceOpt(mockService),
//	)
package handler

import (
	"errors"
	"fmt"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Option configures a Base handler. It follows the functional
// options pattern for clean and type-safe dependency injection.
type Option func(*Base)

// WithLogger sets the logger for the handler.
// This is a required dependency for all handlers.
func WithLogger(logger logging.Logger) Option {
	return func(b *Base) {
		b.Logger = logger
	}
}

// Base provides common handler functionality that is embedded in all specific
// handlers. It enforces consistent logging and error handling patterns across
// all handlers.
type Base struct {
	Logger logging.Logger
}

// NewBase creates a new base handler with the provided options. The logger must
// be explicitly provided using WithLogger option. There is no default logger to
// ensure proper configuration.
func NewBase(opts ...Option) Base {
	var b Base

	for _, opt := range opts {
		opt(&b)
	}

	return b
}

// WrapResponseError provides consistent error wrapping for HTTP responses.
// It ensures errors include proper context and maintain error chain for debugging.
func (b *Base) WrapResponseError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// LogError provides consistent error logging across all handlers.
// It ensures errors are logged with proper context and additional fields.
func (b *Base) LogError(msg string, err error, fields ...logging.Field) {
	if err != nil {
		fields = append(fields, logging.Error(err))
	}
	b.Logger.Error(msg, fields...)
}

// Validate ensures all required dependencies are properly set.
// This is called during handler initialization and route registration
// to fail fast if configuration is incomplete.
func (b *Base) Validate() error {
	if b.Logger == nil {
		return errors.New("logger is required")
	}
	return nil
}
