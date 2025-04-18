package logging

import (
	"fmt"

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
	logger, err := NewLogger(cfg.App.Debug, cfg.App.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger from config: %w", err)
	}
	return logger, nil
}

// CreateTestLogger creates a logger for testing
func (f *Factory) CreateTestLogger() (Logger, error) {
	logger, err := NewTestLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to create test logger: %w", err)
	}
	return logger, nil
}
