package handler

import (
	"fmt"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// HandlerOption defines a handler option function
type HandlerOption func(*Base)

// WithLogger sets the logger for the handler
func WithLogger(logger logging.Logger) HandlerOption {
	return func(b *Base) {
		b.Logger = logger
	}
}

// Base provides common handler functionality
type Base struct {
	Logger logging.Logger
}

// NewBase creates a new base handler with options
func NewBase(opts ...HandlerOption) Base {
	var b Base

	for _, opt := range opts {
		opt(&b)
	}

	return b
}

// WrapResponseError provides consistent error wrapping
func (b *Base) WrapResponseError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// LogError provides consistent error logging
func (b *Base) LogError(msg string, err error, fields ...logging.Field) {
	b.Logger.Error(msg, append(fields, logging.Error(err))...)
}

// Validate validates that required dependencies are set
func (b *Base) Validate() error {
	if b.Logger == nil {
		return fmt.Errorf("logger is required")
	}
	return nil
}
