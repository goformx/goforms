package infrastructure

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/application/handler"
	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/domain/user"
	h "github.com/jonesrussell/goforms/internal/handlers"
	wh "github.com/jonesrussell/goforms/internal/handlers/web"
	ah "github.com/jonesrussell/goforms/internal/handlers/web/admin"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/server"
	"github.com/jonesrussell/goforms/internal/infrastructure/store"
	formstore "github.com/jonesrussell/goforms/internal/infrastructure/store/form"
	"github.com/jonesrussell/goforms/internal/presentation/view"
)

// Stores groups all database store providers.
// This struct is used with fx.Out to provide multiple stores
// to the fx container in a single provider function.
type Stores struct {
	fx.Out

	ContactStore      contact.Store
	SubscriptionStore subscription.Store
	UserStore         user.Store
	FormStore         form.Store
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
	ContactService      contact.Service
	SubscriptionService subscription.Service
	UserService         user.Service
	FormService         form.Service
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

// validateConfig performs validation of critical configuration settings.
// It ensures all required settings are present and valid before initialization.
func validateConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration is nil")
	}

	var errors []string

	// Database configuration
	if cfg.Database.Host == "" {
		errors = append(errors, "database host is required")
	}
	if cfg.Database.Port <= 0 {
		errors = append(errors, "database port must be a positive number")
	}
	if cfg.Database.User == "" {
		errors = append(errors, "database user is required")
	}
	if cfg.Database.Password == "" {
		errors = append(errors, "database password is required")
	}
	if cfg.Database.Name == "" {
		errors = append(errors, "database name is required")
	}
	if cfg.Database.MaxOpenConns <= 0 {
		errors = append(errors, "database max open connections must be a positive number")
	}
	if cfg.Database.MaxIdleConns <= 0 {
		errors = append(errors, "database max idle connections must be a positive number")
	}
	if cfg.Database.ConnMaxLifetme <= 0 {
		errors = append(errors, "database connection max lifetime must be a positive duration")
	}

	// Security configuration
	if cfg.Security.JWTSecret == "" {
		errors = append(errors, "JWT secret is required")
	}
	if len(cfg.Security.JWTSecret) < 32 {
		errors = append(errors, "JWT secret must be at least 32 characters long")
	}
	if cfg.Security.CSRF.Enabled {
		if cfg.Security.CSRF.Secret == "" {
			errors = append(errors, "CSRF secret is required when CSRF is enabled")
		}
		if len(cfg.Security.CSRF.Secret) < 32 {
			errors = append(errors, "CSRF secret must be at least 32 characters long")
		}
	}

	// Server configuration
	if cfg.Server.Port <= 0 {
		errors = append(errors, "server port must be a positive number")
	}
	if cfg.Server.Host == "" {
		errors = append(errors, "server host is required")
	}
	if cfg.Server.ReadTimeout <= 0 {
		errors = append(errors, "server read timeout must be a positive duration")
	}
	if cfg.Server.WriteTimeout <= 0 {
		errors = append(errors, "server write timeout must be a positive duration")
	}
	if cfg.Server.IdleTimeout <= 0 {
		errors = append(errors, "server idle timeout must be a positive duration")
	}
	if cfg.Server.ShutdownTimeout <= 0 {
		errors = append(errors, "server shutdown timeout must be a positive duration")
	}

	// If any validation errors occurred, return them all
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// InfrastructureModule provides core infrastructure dependencies.
// This module includes configuration and database setup.
var InfrastructureModule = fx.Options(
	fx.Provide(
		func() (*config.Config, error) {
			cfg, err := config.New()
			if err != nil {
				return nil, fmt.Errorf("failed to load configuration: %w", err)
			}
			if err := validateConfig(cfg); err != nil {
				return nil, err
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
	// Web handlers
	AnnotateHandler(func(core CoreParams) (h.Handler, error) {
		handler := wh.NewHomeHandler(core.Logger, core.Renderer)
		if handler == nil {
			return nil, fmt.Errorf("failed to create home handler: renderer=%T", core.Renderer)
		}
		return handler, nil
	}),
	AnnotateHandler(func(core CoreParams, services ServiceParams) (h.Handler, error) {
		handler := wh.NewDemoHandler(core.Logger, core.Renderer, services.SubscriptionService)
		if handler == nil {
			return nil, fmt.Errorf("failed to create demo handler: renderer=%T, subscription_service=%T",
				core.Renderer, services.SubscriptionService)
		}
		return handler, nil
	}),
	AnnotateHandler(func(core CoreParams, services ServiceParams) (h.Handler, error) {
		handler := ah.NewDashboardHandler(core.Logger, core.Renderer, services.UserService, services.FormService)
		if handler == nil {
			return nil, fmt.Errorf("failed to create dashboard handler: renderer=%T, user_service=%T, form_service=%T",
				core.Renderer, services.UserService, services.FormService)
		}
		return handler, nil
	}),
	AnnotateHandler(func(core CoreParams, services ServiceParams) (h.Handler, error) {
		handler, err := handler.NewWebHandler(core.Logger,
			handler.WithRenderer(core.Renderer),
			handler.WithContactService(services.ContactService),
			handler.WithWebSubscriptionService(services.SubscriptionService),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create web handler: %w", err)
		}
		if handler == nil {
			return nil, fmt.Errorf("web handler is nil after creation: renderer=%T, contact_service=%T, subscription_service=%T",
				core.Renderer, services.ContactService, services.SubscriptionService)
		}
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

// NewStores creates all database stores.
// This function is responsible for initializing all database stores
// and providing them to the fx container.
func NewStores(db *database.Database, logger logging.Logger) (Stores, error) {
	if db == nil {
		logger.Error("database connection is nil",
			logging.String("operation", "store_initialization"),
			logging.String("error_type", "nil_database"),
		)
		return Stores{}, fmt.Errorf("database connection is nil")
	}

	dbInfo := map[string]any{
		"driver": db.DriverName(),
		"stats":  db.Stats(),
	}

	startTime := time.Now()

	logger.Debug("initializing database stores",
		logging.String("database_type", fmt.Sprintf("%T", db)),
		logging.String("operation", "store_initialization"),
		logging.Any("database_info", dbInfo),
	)

	// Define store creators
	storeCreators := map[string]struct {
		creator func(*database.Database, logging.Logger) any
		setter  func(*Stores, any)
	}{
		"contact": {
			creator: func(db *database.Database, l logging.Logger) any {
				return store.NewContactStore(db, l)
			},
			setter: func(s *Stores, store any) {
				s.ContactStore = store.(contact.Store)
			},
		},
		"subscription": {
			creator: func(db *database.Database, l logging.Logger) any {
				return store.NewSubscriptionStore(db, l)
			},
			setter: func(s *Stores, store any) {
				s.SubscriptionStore = store.(subscription.Store)
			},
		},
		"user": {
			creator: func(db *database.Database, l logging.Logger) any {
				return store.NewUserStore(db, l)
			},
			setter: func(s *Stores, store any) {
				s.UserStore = store.(user.Store)
			},
		},
		"form": {
			creator: func(db *database.Database, l logging.Logger) any {
				return formstore.NewStore(db, l)
			},
			setter: func(s *Stores, store any) {
				s.FormStore = store.(form.Store)
			},
		},
	}

	// Initialize stores
	var stores Stores
	createdStores := make(map[string]any)

	for name, creator := range storeCreators {
		storeInstance := creator.creator(db, logger)
		if storeInstance == nil {
			logger.Error("failed to create store",
				logging.String("operation", "store_initialization"),
				logging.String("store_type", name),
				logging.String("error_type", "nil_store"),
				logging.Any("database_info", dbInfo),
			)
			return Stores{}, fmt.Errorf("failed to create %s store: driver=%v, stats=%+v",
				name, db.DriverName(), db.Stats())
		}

		creator.setter(&stores, storeInstance)
		createdStores[name] = storeInstance
	}

	// Calculate initialization metrics
	initDuration := time.Since(startTime)
	totalStores := len(createdStores)

	// Log successful initialization with detailed metrics
	logger.Info("store initialization complete",
		logging.String("operation", "store_initialization"),
		logging.Int("total_stores_created", totalStores),
		logging.Duration("init_duration_ms", initDuration),
		logging.String("contact_store_type", fmt.Sprintf("%T", stores.ContactStore)),
		logging.String("subscription_store_type", fmt.Sprintf("%T", stores.SubscriptionStore)),
		logging.String("user_store_type", fmt.Sprintf("%T", stores.UserStore)),
		logging.String("form_store_type", fmt.Sprintf("%T", stores.FormStore)),
		logging.Bool("all_stores_initialized", totalStores == len(storeCreators)),
		logging.Any("database_info", dbInfo),
	)

	return stores, nil
}
