package handler

import (
	"fmt"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Base provides common handler functionality
type Base struct {
	Logger logging.Logger
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
