package infrastructure

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/application/handler"
	"github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/user"
	h "github.com/jonesrussell/goforms/internal/handlers"
	wh "github.com/jonesrussell/goforms/internal/handlers/web"
	wh_auth "github.com/jonesrussell/goforms/internal/handlers/web/auth"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/server"
	"github.com/jonesrussell/goforms/internal/infrastructure/store"
	formstore "github.com/jonesrussell/goforms/internal/infrastructure/store/form"
	ph "github.com/jonesrussell/goforms/internal/presentation/handlers"
	"github.com/jonesrussell/goforms/internal/presentation/view"
)

const (
	// MinSecretLength is the minimum length required for security secrets
	MinSecretLength = 32
)

// Stores groups all database store providers.
// This struct is used with fx.Out to provide multiple stores
// to the fx container in a single provider function.
type Stores struct {
	fx.Out

	UserStore user.Store
	FormStore form.Store
}

// CoreParams contains core infrastructure dependencies that are commonly needed by handlers.
// These are typically required for basic handler functionality like logging and rendering.
type CoreParams struct {
	fx.In
	Logger   logging.Logger
	Renderer *view.Renderer
	Config   *config.Config
}

// ServiceParams contains all service dependencies that handlers might need.
// This separation makes it easier to manage service dependencies and allows for
// more granular dependency injection.
type ServiceParams struct {
	fx.In
	UserService user.Service
	FormService form.Service
}

// AnnotateHandler is a helper function that simplifies the creation of handler providers.
// It wraps the common fx.Provide and fx.Annotate pattern used for handlers.
func AnnotateHandler(fn any) fx.Option {
	return fx.Provide(
		fx.Annotate(
			fn,
			fx.As(new(h.Handler)),
			fx.ResultTags(`group:"handlers"`),
		),
	)
}

// validateDatabaseConfig validates database configuration
func validateDatabaseConfig(cfg *config.DatabaseConfig) []string {
	var validationErrors []string
	if cfg.MaxOpenConns <= 0 {
		validationErrors = append(validationErrors, "database max open connections must be a positive number")
	}
	if cfg.MaxIdleConns <= 0 {
		validationErrors = append(validationErrors, "database max idle connections must be a positive number")
	}
	if cfg.ConnMaxLifetme <= 0 {
		validationErrors = append(validationErrors, "database connection max lifetime must be a positive duration")
	}
	return validationErrors
}

// validateSecurityConfig validates security configuration
func validateSecurityConfig(cfg *config.SecurityConfig) []string {
	var validationErrors []string
	if len(cfg.JWTSecret) < MinSecretLength {
		validationErrors = append(validationErrors, "JWT secret must be at least 32 characters long")
	}
	if cfg.CSRF.Enabled && len(cfg.CSRF.Secret) < MinSecretLength {
		validationErrors = append(validationErrors, "CSRF secret must be at least 32 characters long")
	}
	return validationErrors
}

// validateServerConfig validates server configuration
func validateServerConfig(cfg *config.ServerConfig) []string {
	var validationErrors []string
	if cfg.Port <= 0 {
		validationErrors = append(validationErrors, "server port must be a positive number")
	}
	if cfg.ReadTimeout <= 0 {
		validationErrors = append(validationErrors, "server read timeout must be a positive duration")
	}
	if cfg.WriteTimeout <= 0 {
		validationErrors = append(validationErrors, "server write timeout must be a positive duration")
	}
	if cfg.IdleTimeout <= 0 {
		validationErrors = append(validationErrors, "server idle timeout must be a positive duration")
	}
	if cfg.ShutdownTimeout <= 0 {
		validationErrors = append(validationErrors, "server shutdown timeout must be a positive duration")
	}
	return validationErrors
}

// validateConfig performs validation of critical configuration settings
func validateConfig(cfg *config.Config) error {
	if cfg == nil {
		return errors.New("configuration is nil")
	}

	var validationErrors []string

	// Validate database configuration
	validationErrors = append(validationErrors, validateDatabaseConfig(&cfg.Database)...)

	// Validate security configuration
	validationErrors = append(validationErrors, validateSecurityConfig(&cfg.Security)...)

	// Validate server configuration
	validationErrors = append(validationErrors, validateServerConfig(&cfg.Server)...)

	// If any validation errors occurred, return them all
	if len(validationErrors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// InfrastructureModule provides core infrastructure dependencies.
// This module includes configuration and database setup.
var InfrastructureModule = fx.Options(
	fx.Provide(
		func() (*config.Config, error) {
			cfg, cfgErr := config.New()
			if cfgErr != nil {
				return nil, fmt.Errorf("failed to load configuration: %w", cfgErr)
			}
			if validationErr := validateConfig(cfg); validationErr != nil {
				return nil, validationErr
			}
			return cfg, nil
		},
		database.NewDB,
	),
)

// StoreModule provides all database store implementations.
// This module is responsible for creating and managing database stores.
var StoreModule = fx.Options(
	fx.Provide(NewStores),
)

// HandlerModule provides all HTTP handlers for the application.
// This module is responsible for setting up route handlers and their dependencies.
var HandlerModule = fx.Options(
	// Static file handler (must be first)
	AnnotateHandler(func(core CoreParams) (h.Handler, error) {
		handler := handler.NewStaticHandler(core.Logger, core.Config)
		if handler == nil {
			return nil, errors.New("failed to create static handler")
		}
		core.Logger.Debug("registered handler",
			logging.String("handler_name", "StaticHandler"),
			logging.String("handler_type", fmt.Sprintf("%T", handler)),
			logging.String("operation", "handler_registration"),
		)
		return handler, nil
	}),
	// Web handlers
	AnnotateHandler(func(core CoreParams, middlewareManager *middleware.Manager) (h.Handler, error) {
		handler := wh.NewHomeHandler(core.Logger, core.Renderer)
		if handler == nil {
			return nil, fmt.Errorf("failed to create home handler: renderer=%T", core.Renderer)
		}
		core.Logger.Debug("registered handler",
			logging.String("handler_name", "HomeHandler"),
			logging.String("handler_type", fmt.Sprintf("%T", handler)),
			logging.String("operation", "handler_registration"),
		)
		return handler, nil
	}),
	AnnotateHandler(func(core CoreParams, middlewareManager *middleware.Manager) (h.Handler, error) {
		handler := wh_auth.NewWebLoginHandler(core.Logger, core.Renderer)
		if handler == nil {
			return nil, fmt.Errorf("failed to create web login handler: renderer=%T", core.Renderer)
		}
		core.Logger.Debug("registered handler",
			logging.String("handler_name", "WebLoginHandler"),
			logging.String("handler_type", fmt.Sprintf("%T", handler)),
			logging.String("operation", "handler_registration"),
		)
		return handler, nil
	}),
	AnnotateHandler(
		func(
			core CoreParams,
			services ServiceParams,
			middlewareManager *middleware.Manager,
		) (h.Handler, error) {
			handler, err := handler.NewWebHandler(
				core.Logger,
				handler.WithRenderer(core.Renderer),
				handler.WithMiddlewareManager(middlewareManager),
				handler.WithConfig(core.Config),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create web handler: %w", err)
			}
			core.Logger.Debug(
				"registered handler",
				logging.String("handler_name", "WebHandler"),
				logging.String("handler_type", fmt.Sprintf("%T", handler)),
				logging.String("operation", "handler_registration"),
			)
			return handler, nil
		},
	),
	// Auth handler
	AnnotateHandler(func(core CoreParams, services ServiceParams) (h.Handler, error) {
		handler := handler.NewAuthHandler(core.Logger, handler.WithUserService(services.UserService))
		if handler == nil {
			return nil, fmt.Errorf("failed to create auth handler: user_service=%T", services.UserService)
		}
		core.Logger.Debug("registered handler",
			logging.String("handler_name", "AuthHandler"),
			logging.String("handler_type", fmt.Sprintf("%T", handler)),
			logging.String("operation", "handler_registration"),
		)
		return handler, nil
	}),
	// Dashboard handler
	AnnotateHandler(func(core CoreParams, services ServiceParams) (h.Handler, error) {
		handler, err := ph.NewDashboardHandler(services.UserService, services.FormService)
		if err != nil {
			return nil, fmt.Errorf("failed to create dashboard handler: %w", err)
		}
		core.Logger.Debug("registered handler",
			logging.String("handler_name", "DashboardHandler"),
			logging.String("handler_type", fmt.Sprintf("%T", handler)),
			logging.String("operation", "handler_registration"),
		)
		return handler, nil
	}),
)

// ServerModule provides the HTTP server setup.
// This module is responsible for creating and configuring the Echo server.
var ServerModule = fx.Options(
	fx.Provide(server.New),
)

// Module combines all infrastructure-level modules into a single module.
// This is the main entry point for infrastructure dependencies.
var Module = fx.Options(
	InfrastructureModule,
	StoreModule,
	ServerModule,
	HandlerModule,
)

// wrapCreator creates a type-safe wrapper for store creation functions
func wrapCreator[T any](
	creator func(*database.Database, logging.Logger) T,
) func(*database.Database, logging.Logger) any {
	return func(db *database.Database, logger logging.Logger) any {
		return creator(db, logger)
	}
}

// wrapAssigner creates a type-safe wrapper for store assignment functions
func wrapAssigner[T any](assigner func(*Stores, T)) func(*Stores, any) {
	return func(s *Stores, instance any) {
		if s == nil {
			panic(errors.New("database connection is nil"))
		}
		typedInstance, ok := instance.(T)
		if !ok {
			panic(errors.New("invalid instance type"))
		}
		assigner(s, typedInstance)
	}
}

// NewStores creates all database stores.
// This function is responsible for initializing all database stores
// and providing them to the fx container.
func NewStores(db *database.Database, logger logging.Logger) (Stores, error) {
	if db == nil {
		logger.Error("database connection is nil",
			logging.String("operation", "store_initialization"),
			logging.String("error_type", "nil_database"),
		)
		return Stores{}, errors.New("database connection is nil")
	}

	startTime := time.Now()

	// Map of store creators
	storeCreators := map[string]struct {
		create func(*database.Database, logging.Logger) any
		assign func(*Stores, any)
	}{
		"user": {
			create: wrapCreator(store.NewUserStore),
			assign: wrapAssigner(func(s *Stores, v user.Store) { s.UserStore = v }),
		},
		"form": {
			create: wrapCreator(formstore.NewStore),
			assign: wrapAssigner(func(s *Stores, v form.Store) { s.FormStore = v }),
		},
	}

	var stores Stores
	var wg sync.WaitGroup
	var mu sync.Mutex
	failedStores := make(chan string, len(storeCreators))
	results := make(chan struct {
		name     string
		instance any
	}, len(storeCreators))

	// Initialize stores concurrently
	for name, creator := range storeCreators {
		wg.Add(1)
		go func(name string, creator struct {
			create func(*database.Database, logging.Logger) any
			assign func(*Stores, any)
		}) {
			defer wg.Done()

			// Create store instance
			instance := creator.create(db, logger)
			if instance == nil {
				logger.Error("store creation failed",
					logging.String("store_type", name),
					logging.String("operation", "store_initialization"),
					logging.String("error_type", "nil_instance"),
				)
				failedStores <- name
				return
			}

			results <- struct {
				name     string
				instance any
			}{name, instance}
		}(name, creator)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(failedStores)
	close(results)

	// Collect failed stores
	var failedStoreNames []string
	for name := range failedStores {
		failedStoreNames = append(failedStoreNames, name)
	}

	// Handle initialization errors
	if len(failedStoreNames) > 0 {
		return Stores{}, fmt.Errorf("failed to initialize the following stores: %s", strings.Join(failedStoreNames, ", "))
	}

	// Assign successful stores
	for result := range results {
		mu.Lock()
		storeCreators[result.name].assign(&stores, result.instance)
		mu.Unlock()
	}

	// Log successful initialization metrics
	logger.Info("all database stores initialized successfully",
		logging.String("operation", "store_initialization"),
		logging.Duration("init_duration", time.Since(startTime)),
		logging.Int("total_stores_initialized", len(storeCreators)),
	)

	return stores, nil
}
