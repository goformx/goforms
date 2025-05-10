package logging

import (
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
)

// Factory creates loggers based on configuration
type Factory struct{}

// NewFactory creates a new logger factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateFromConfig creates a logger from configuration
func (f *Factory) CreateFromConfig(cfg *config.Config) (Logger, error) {
	if cfg == nil {
		return NewLogger("info", "goforms")
	}
	return NewLogger(cfg.App.LogLevel, cfg.App.Name)
}

// CreateTestLogger creates a logger for testing
func (f *Factory) CreateTestLogger() (Logger, error) {
	return NewTestLogger()
}

// AuthLogger returns the shared logger for auth middleware
func (f *Factory) AuthLogger(logger Logger) Logger {
	return logger
}

// CookieAuthLogger returns the shared logger for cookie auth middleware
func (f *Factory) CookieAuthLogger(logger Logger) Logger {
	return logger
}
