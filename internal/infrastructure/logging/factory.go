package logging

import "github.com/jonesrussell/goforms/internal/infrastructure/config"

// Factory creates loggers based on configuration
type Factory struct{}

// NewFactory creates a new logger factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateFromConfig creates a logger from configuration
func (f *Factory) CreateFromConfig(cfg *config.Config) Logger {
	return NewLogger(cfg.App.Debug, cfg.App.Name)
}

// CreateTestLogger creates a logger for testing
func (f *Factory) CreateTestLogger() Logger {
	return NewTestLogger()
}
