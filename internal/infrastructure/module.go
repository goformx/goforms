// Package infrastructure provides core infrastructure components.
// This package serves as the foundation of the infrastructure layer.
package infrastructure

import (
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"

	"embed"

	formevent "github.com/goformx/goforms/internal/domain/form/event"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/event"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/goformx/goforms/internal/infrastructure/version"
	infraweb "github.com/goformx/goforms/internal/infrastructure/web"
)

const (
	// MinSecretLength is the minimum length required for security secrets
	MinSecretLength = 32

	// DefaultLogLevel is the default log level for production
	DefaultLogLevel = "info"
	// DevelopmentLogLevel is the default log level for development
	DevelopmentLogLevel = "debug"
	// ProductionLogLevel is the log level for production
	ProductionLogLevel = "warn"
)

var (
	// ErrMissingLogger is returned when a required logger dependency is nil
	ErrMissingLogger = errors.New("logger is required")

	// ErrMissingConfig is returned when a required config dependency is nil
	ErrMissingConfig = errors.New("config is required")

	// ErrInvalidLogLevel is returned when an invalid log level is specified
	ErrInvalidLogLevel = errors.New("invalid log level specified")

	// ErrMissingSanitizer is returned when a required sanitizer dependency is nil
	ErrMissingSanitizer = errors.New("sanitizer is required for logger factory")
)

// NewEventPublisher creates a new event publisher with proper dependency validation.
// It returns an error if required dependencies are missing or invalid.
func NewEventPublisher(logger logging.Logger) (formevent.Publisher, error) {
	if logger == nil {
		return nil, fmt.Errorf("event publisher creation failed: %w", ErrMissingLogger)
	}

	publisher := event.NewMemoryPublisher(logger)
	logger.Info("Event publisher initialized successfully")

	return publisher, nil
}

// NewLoggerFactory creates a new logger factory with proper configuration and error handling.
func NewLoggerFactory(
	cfg *config.Config,
	sanitizer sanitization.ServiceInterface,
) (*logging.Factory, error) {
	if cfg == nil {
		return nil, fmt.Errorf("logger factory creation failed: %w", ErrMissingConfig)
	}

	if sanitizer == nil {
		return nil, fmt.Errorf("logger factory creation failed: %w", ErrMissingSanitizer)
	}

	// Determine log level based on configuration
	logLevel := determineLogLevel(cfg)

	// Set output paths based on environment
	var outputPaths []string
	if cfg.App.IsDevelopment() {
		outputPaths = []string{"stdout"}
	} else {
		outputPaths = []string{"stdout", "/var/log/app.log"}
	}

	factoryConfig := logging.FactoryConfig{
		AppName:     cfg.App.Name,
		Version:     version.Version,
		Environment: cfg.App.Environment,
		Fields: map[string]any{
			"app":     cfg.App.Name,
			"version": version.Version,
			"env":     cfg.App.Environment,
		},
		LogLevel:         logLevel,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: []string{"stderr"},
	}

	factory, err := logging.NewFactory(&factoryConfig, sanitizer)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger factory: %w", err)
	}

	return factory, nil
}

// determineLogLevel determines the appropriate log level based on configuration and environment.
// Priority: explicit LogLevel > Debug flag > Environment > default
func determineLogLevel(cfg *config.Config) string {
	// Explicit log level takes highest priority
	if cfg.App.LogLevel != "" {
		return cfg.App.LogLevel
	}

	// Debug flag overrides environment-based defaults
	if cfg.App.Debug {
		return DevelopmentLogLevel
	}

	// Environment-based defaults
	switch cfg.App.Environment {
	case "development", "dev":
		return DevelopmentLogLevel
	case "production", "prod":
		return ProductionLogLevel
	default:
		return DefaultLogLevel
	}
}

// NewLogger creates a logger instance from the factory with proper error handling.
func NewLogger(factory *logging.Factory) (logging.Logger, error) {
	if factory == nil {
		return nil, errors.New("logger factory is required")
	}

	logger, err := factory.CreateLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return logger, nil
}

// ProvideAssetServer creates an appropriate asset server based on the environment.
// In development, it serves static files from public directory while Vite handles JS/CSS.
// In production, it serves from embedded filesystem for optimal performance.
func ProvideAssetServer(
	cfg *config.Config,
	logger logging.Logger,
	distFS embed.FS,
) (infraweb.AssetServer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("asset server creation failed: %w", ErrMissingConfig)
	}

	if logger == nil {
		return nil, fmt.Errorf("asset server creation failed: %w", ErrMissingLogger)
	}

	if cfg.App.IsDevelopment() {
		logger.Info("Initializing development asset server for static files")

		return infraweb.NewDevelopmentAssetServer(cfg, logger), nil
	}

	logger.Info("Initializing embedded asset server for production")

	return infraweb.NewEmbeddedAssetServer(logger, distFS), nil
}

// NewAssetManager creates a new asset manager with proper dependency validation.
// Returns the interface type for better dependency injection.
func NewAssetManager(
	distFS embed.FS,
	logger logging.Logger,
	cfg *config.Config,
) (infraweb.AssetManagerInterface, error) {
	if logger == nil {
		return nil, fmt.Errorf("asset manager creation failed: %w", ErrMissingLogger)
	}

	if cfg == nil {
		return nil, fmt.Errorf("asset manager creation failed: %w", ErrMissingConfig)
	}

	manager, err := infraweb.NewAssetManager(cfg, logger, distFS)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset manager: %w", err)
	}

	return manager, nil
}

// ProvideEcho creates and configures a new Echo instance with sensible defaults.
func ProvideEcho() *echo.Echo {
	e := echo.New()

	// Configure Echo with production-ready settings
	e.HideBanner = true
	e.HidePort = true

	return e
}

// ProvideSanitizationService creates a new sanitization service with proper annotations.
func ProvideSanitizationService() sanitization.ServiceInterface {
	return sanitization.NewService()
}
