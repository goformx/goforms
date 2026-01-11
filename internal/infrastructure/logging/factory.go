// Package logging provides a unified logging interface
package logging

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// Factory creates loggers based on configuration
type Factory struct {
	initialFields  map[string]any
	appName        string
	version        string
	environment    string
	outputPaths    []string
	errorPaths     []string
	sanitizer      sanitization.ServiceInterface
	fieldSanitizer *Sanitizer
	// Add testCore for test injection
	testCore  zapcore.Core
	LogLevel  string
	zapLogger *zap.Logger // Store the created zap logger for slog adapter
}

// NewFactory creates a new logger factory with the given configuration
func NewFactory(cfg *FactoryConfig, sanitizer sanitization.ServiceInterface) (*Factory, error) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid factory configuration: %w", err)
	}

	if cfg.Fields == nil {
		cfg.Fields = make(map[string]any)
	}

	// Set default paths using config helper
	setDefaultPaths(cfg)

	return &Factory{
		initialFields:  cfg.Fields,
		appName:        cfg.AppName,
		version:        cfg.Version,
		environment:    cfg.Environment,
		outputPaths:    cfg.OutputPaths,
		errorPaths:     cfg.ErrorOutputPaths,
		sanitizer:      sanitizer,
		fieldSanitizer: NewSanitizer(),
		LogLevel:       cfg.LogLevel,
	}, nil
}

// WithTestCore allows tests to inject a zapcore.Core for capturing logs
func (f *Factory) WithTestCore(core zapcore.Core) *Factory {
	f.testCore = core

	return f
}

// CreateLogger creates a new logger instance with the application name.
func (f *Factory) CreateLogger() (Logger, error) {
	// Parse log level using config helper
	level := parseLogLevel(f.LogLevel, f.environment)

	// Create zap core using config helper
	var core zapcore.Core
	if f.testCore != nil {
		core = f.testCore
	} else if f.environment == "production" {
		core = createProductionCore(level)
	} else {
		core = createZapCore(level, f.testCore)
	}

	// Create logger with options
	zapLogger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Development(),
	)

	// Create initial fields
	fields := make([]zap.Field, 0, len(f.initialFields))
	for k, v := range f.initialFields {
		fields = append(fields, zap.String(k, fmt.Sprintf("%v", v)))
	}

	// Create logger with initial fields
	zapLogger = zapLogger.With(fields...)

	// Store the zap logger for slog adapter access
	f.zapLogger = zapLogger

	// Create our logger implementation
	return newLogger(zapLogger, f.sanitizer, f.fieldSanitizer), nil
}

// GetZapLogger returns the underlying zap logger for slog integration
func (f *Factory) GetZapLogger() *zap.Logger {
	return f.zapLogger
}
