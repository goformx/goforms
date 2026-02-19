// Package infrastructure provides core infrastructure components and their dependency injection setup.
// This package serves as the foundation for the application's infrastructure layer,
// managing database connections, logging, event systems, and web server configuration.
package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"embed"

	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/domain/form"
	formevent "github.com/goformx/goforms/internal/domain/form/event"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/event"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/goformx/goforms/internal/infrastructure/server"
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

// CoreParams groups core infrastructure dependencies required for basic application functionality.
// These dependencies form the foundation of the application's runtime environment.
type CoreParams struct {
	fx.In
	Config *config.Config `validate:"required"`
	Logger logging.Logger `validate:"required"`
	Echo   *echo.Echo     `validate:"required"`
}

// Validate ensures all required core parameters are present
func (p CoreParams) Validate() error {
	if p.Config == nil {
		return ErrMissingConfig
	}

	if p.Logger == nil {
		return ErrMissingLogger
	}

	if p.Echo == nil {
		return errors.New("echo instance is required")
	}

	return nil
}

// ServiceParams groups business service dependencies.
// These represent the core business logic services of the application.
type ServiceParams struct {
	fx.In
	UserService user.Service `validate:"required"`
	FormService form.Service `validate:"required"`
}

// Validate ensures all required service parameters are present
func (p ServiceParams) Validate() error {
	if p.UserService == nil {
		return errors.New("user service is required")
	}

	if p.FormService == nil {
		return errors.New("form service is required")
	}

	return nil
}

// EventPublisherParams contains dependencies for creating an event publisher.
// The event publisher is responsible for distributing domain events throughout the application.
type EventPublisherParams struct {
	fx.In
	Logger logging.Logger `validate:"required"`
}

// LoggerFactoryParams contains dependencies for creating a logger factory
type LoggerFactoryParams struct {
	fx.In
	Config    *config.Config                `validate:"required"`
	Sanitizer sanitization.ServiceInterface `validate:"required"`
}

// AssetServerParams groups the dependencies for creating an asset server.
// The asset server handles static file serving with environment-specific optimizations.
type AssetServerParams struct {
	fx.In
	Config *config.Config `validate:"required"`
	Logger logging.Logger `validate:"required"`
	DistFS embed.FS
}

// AssetManagerParams contains dependencies for creating an asset manager
type AssetManagerParams struct {
	fx.In
	DistFS embed.FS
	Logger logging.Logger `validate:"required"`
	Config *config.Config `validate:"required"`
}

// NewEventPublisher creates a new event publisher with proper dependency validation.
// It returns an error if required dependencies are missing or invalid.
func NewEventPublisher(p EventPublisherParams) (formevent.Publisher, error) {
	if p.Logger == nil {
		return nil, fmt.Errorf("event publisher creation failed: %w", ErrMissingLogger)
	}

	publisher := event.NewMemoryPublisher(p.Logger)
	p.Logger.Info("Event publisher initialized successfully")

	return publisher, nil
}

// NewLoggerFactory creates a new logger factory with proper configuration and error handling.
func NewLoggerFactory(p LoggerFactoryParams) (*logging.Factory, error) {
	if p.Config == nil {
		return nil, fmt.Errorf("logger factory creation failed: %w", ErrMissingConfig)
	}

	if p.Sanitizer == nil {
		return nil, fmt.Errorf("logger factory creation failed: %w", ErrMissingSanitizer)
	}

	// Determine log level based on configuration
	logLevel := determineLogLevel(p.Config)

	// Set output paths based on environment
	var outputPaths []string
	if p.Config.App.IsDevelopment() {
		outputPaths = []string{"stdout"}
	} else {
		outputPaths = []string{"stdout", "/var/log/app.log"}
	}

	factoryConfig := logging.FactoryConfig{
		AppName:     p.Config.App.Name,
		Version:     version.Version,
		Environment: p.Config.App.Environment,
		Fields: map[string]any{
			"app":     p.Config.App.Name,
			"version": version.Version,
			"env":     p.Config.App.Environment,
		},
		LogLevel:         logLevel,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: []string{"stderr"},
	}

	factory, err := logging.NewFactory(&factoryConfig, p.Sanitizer)
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
func ProvideAssetServer(p AssetServerParams) (infraweb.AssetServer, error) {
	if p.Config == nil {
		return nil, fmt.Errorf("asset server creation failed: %w", ErrMissingConfig)
	}

	if p.Logger == nil {
		return nil, fmt.Errorf("asset server creation failed: %w", ErrMissingLogger)
	}

	if p.Config.App.IsDevelopment() {
		p.Logger.Info("Initializing development asset server for static files")

		return infraweb.NewDevelopmentAssetServer(p.Config, p.Logger), nil
	}

	p.Logger.Info("Initializing embedded asset server for production")

	return infraweb.NewEmbeddedAssetServer(p.Logger, p.DistFS), nil
}

// NewAssetManager creates a new asset manager with proper dependency validation.
// Returns the interface type for better dependency injection.
func NewAssetManager(p AssetManagerParams) (infraweb.AssetManagerInterface, error) {
	if p.Logger == nil {
		return nil, fmt.Errorf("asset manager creation failed: %w", ErrMissingLogger)
	}

	if p.Config == nil {
		return nil, fmt.Errorf("asset manager creation failed: %w", ErrMissingConfig)
	}

	manager, err := infraweb.NewAssetManager(p.Config, p.Logger, p.DistFS)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset manager: %w", err)
	}

	return manager, nil
}

// AnnotateHandler is a helper function that simplifies the creation of handler providers.
// It automatically registers handlers with the appropriate fx annotations and grouping.
func AnnotateHandler(fn any) fx.Option {
	return fx.Provide(
		fx.Annotate(
			fn,
			fx.As(new(web.Handler)),
			fx.ResultTags(`group:"handlers"`),
		),
	)
}

// ProvideEcho creates and configures a new Echo instance with sensible defaults.
func ProvideEcho() *echo.Echo {
	e := echo.New()

	// Configure Echo with production-ready settings
	e.HideBanner = true
	e.HidePort = true

	return e
}

// ProvideDatabase creates a new database connection with lifecycle management.
func ProvideDatabase(lc fx.Lifecycle, cfg *config.Config, logger logging.Logger) (database.DB, error) {
	if cfg == nil {
		return nil, ErrMissingConfig
	}

	if logger == nil {
		return nil, ErrMissingLogger
	}

	db, err := database.New(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// Register lifecycle hooks for graceful shutdown
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info("Database connection established")

			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info("Closing database connection")

			return db.Close()
		},
	})

	return db, nil
}

// ProvideSanitizationService creates a new sanitization service with proper annotations.
func ProvideSanitizationService() sanitization.ServiceInterface {
	return sanitization.NewService()
}

// Module provides comprehensive infrastructure dependencies with proper error handling,
// lifecycle management, and clear separation of concerns.
var Module = fx.Module("infrastructure",
	// Core infrastructure providers
	fx.Provide(
		// Echo web framework
		ProvideEcho,

		// Database with lifecycle management
		ProvideDatabase,

		// HTTP server
		server.New,

		// Sanitization service
		fx.Annotate(
			ProvideSanitizationService,
			fx.As(new(sanitization.ServiceInterface)),
		),

		// Logging system
		NewLoggerFactory,
		NewLogger,

		// Event system
		NewEventPublisher,
		event.NewMemoryEventBus,
	),

	// Lifecycle management
	fx.Invoke(func(lc fx.Lifecycle, logger logging.Logger, _ *config.Config) {
		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				logger.Info("Infrastructure module initialized")

				return nil
			},
			OnStop: func(_ context.Context) error {
				logger.Info("Infrastructure module shutting down")

				return nil
			},
		})
	}),
)
