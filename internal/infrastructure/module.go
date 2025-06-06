package infrastructure

import (
	"errors"
	"fmt"
	"strings"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/handlers/health"
	"github.com/goformx/goforms/internal/application/handlers/web"
	appmiddleware "github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/form"
	healthdomain "github.com/goformx/goforms/internal/domain/services/health"
	"github.com/goformx/goforms/internal/domain/user"
	healthadapter "github.com/goformx/goforms/internal/infrastructure/adapters/health"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/view"
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
			fx.As(new(web.Handler)),
			fx.ResultTags(`group:"web_handlers"`),
		),
	)
}

// validateDatabaseConfig validates the database configuration
func validateDatabaseConfig(cfg *config.DatabaseConfig) error {
	if cfg.MaxOpenConns <= 0 {
		return errors.New("database max open connections must be a positive number")
	}
	if cfg.MaxIdleConns <= 0 {
		return errors.New("database max idle connections must be a positive number")
	}
	if cfg.ConnMaxLifetime <= 0 {
		return errors.New("database connection max lifetime must be a positive duration")
	}
	return nil
}

// validateSecurityConfig validates the security configuration
func validateSecurityConfig(cfg *config.SecurityConfig) error {
	if cfg == nil {
		return errors.New("security configuration cannot be nil")
	}

	// Validate CSRF configuration
	if cfg.CSRF.Enabled {
		if len(cfg.CSRF.Secret) < MinSecretLength {
			return fmt.Errorf("CSRF secret must be at least %d characters long", MinSecretLength)
		}
	}

	// Validate CORS configuration
	if cfg.CorsMaxAge <= 0 {
		return errors.New("CORS max age must be a positive number")
	}

	// Validate rate limiting configuration
	if cfg.FormRateLimit <= 0 {
		return errors.New("form rate limit must be a positive number")
	}
	if cfg.FormRateLimitWindow <= 0 {
		return errors.New("form rate limit window must be a positive duration")
	}

	// Validate request timeout
	if cfg.RequestTimeout <= 0 {
		return errors.New("request timeout must be a positive duration")
	}

	return nil
}

// validateServerConfig validates the server configuration
func validateServerConfig(cfg *config.ServerConfig) error {
	if cfg.Port <= 0 {
		return errors.New("server port must be a positive number")
	}
	if cfg.ReadTimeout <= 0 {
		return errors.New("server read timeout must be a positive duration")
	}
	if cfg.WriteTimeout <= 0 {
		return errors.New("server write timeout must be a positive duration")
	}
	return nil
}

// validateConfig checks if the configuration is valid
func validateConfig(cfg *config.Config, logger logging.Logger) error {
	logger.Info("validateConfig: Starting configuration validation...")
	logger.Info("validateConfig: Database config",
		logging.Int("MaxOpenConns", cfg.Database.MaxOpenConns),
		logging.Int("MaxIdleConns", cfg.Database.MaxIdleConns),
		logging.Duration("ConnMaxLifetime", cfg.Database.ConnMaxLifetime))

	var validationErrors []string

	// Validate database configuration
	logger.Info("validateConfig: Validating database configuration...")
	if dbErr := validateDatabaseConfig(&cfg.Database); dbErr != nil {
		logger.Error("validateConfig: Database validation failed",
			logging.Error(dbErr))
		validationErrors = append(validationErrors, dbErr.Error())
	}

	// Validate security configuration
	logger.Info("validateConfig: Validating security configuration...")
	if secErr := validateSecurityConfig(&cfg.Security); secErr != nil {
		logger.Error("validateConfig: Security validation failed",
			logging.Error(secErr))
		validationErrors = append(validationErrors, secErr.Error())
	}

	// Validate server configuration
	logger.Info("validateConfig: Validating server configuration...")
	if srvErr := validateServerConfig(&cfg.Server); srvErr != nil {
		logger.Error("validateConfig: Server validation failed",
			logging.Error(srvErr))
		validationErrors = append(validationErrors, srvErr.Error())
	}

	if len(validationErrors) > 0 {
		logger.Error("validateConfig: Validation failed",
			logging.StringField("errors", strings.Join(validationErrors, "; ")))
		return fmt.Errorf("configuration validation failed: %s", strings.Join(validationErrors, "; "))
	}

	logger.Info("validateConfig: Configuration validation successful")
	return nil
}

// InfrastructureModule provides core infrastructure dependencies.
var InfrastructureModule = fx.Options(
	fx.Provide(
		func(logger logging.Logger) (*config.Config, error) {
			logger.Info("InfrastructureModule: Starting configuration loading...")
			cfg, cfgErr := config.New(logger)
			if cfgErr != nil {
				logger.Error("InfrastructureModule: Error loading configuration",
					logging.Error(cfgErr))
				return nil, fmt.Errorf("failed to load configuration: %w", cfgErr)
			}

			logger.Info("InfrastructureModule: Configuration loaded, starting validation...")
			if validationErr := validateConfig(cfg, logger); validationErr != nil {
				logger.Error("InfrastructureModule: Validation failed",
					logging.Error(validationErr))
				return nil, validationErr
			}

			logger.Info("InfrastructureModule: Configuration validated successfully")
			return cfg, nil
		},
		func(cfg *config.Config, logger logging.Logger) (*database.Database, error) {
			logger.Info("InfrastructureModule: Creating database connection...")
			db, err := database.NewDB(cfg, logger)
			if err != nil {
				logger.Error("InfrastructureModule: Database connection failed",
					logging.Error(err),
					logging.StringField("host", cfg.Database.Host),
					logging.IntField("port", cfg.Database.Port),
					logging.StringField("database", cfg.Database.Name))
				return nil, err
			}
			logger.Info("InfrastructureModule: Database connection established")
			return db, nil
		},
	),
)

// HandlerModule provides all HTTP handlers for the application.
var HandlerModule = fx.Options(
	// Session manager provider
	fx.Provide(func(core CoreParams) *appmiddleware.SessionManager {
		return appmiddleware.NewSessionManager(core.Logger)
	}),
	// Web handlers
	AnnotateHandler(func(
		core CoreParams,
		services ServiceParams,
		middlewareManager *appmiddleware.Manager,
		sessionManager *appmiddleware.SessionManager,
	) (web.Handler, error) {
		baseHandler := web.NewBaseHandler(services.FormService, core.Logger)
		handler, err := web.NewWebHandler(web.HandlerDeps{
			BaseHandler:       baseHandler,
			UserService:       services.UserService,
			SessionManager:    sessionManager,
			Renderer:          core.Renderer,
			MiddlewareManager: middlewareManager,
			Config:            core.Config,
			Logger:            core.Logger,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create web handler: %w", err)
		}
		core.Logger.Debug("registered handler",
			logging.StringField("handler_name", "WebHandler"),
			logging.StringField("handler_type", fmt.Sprintf("%T", handler)),
			logging.StringField("operation", "handler_registration"),
		)
		return handler, nil
	}),
	// Auth handler
	AnnotateHandler(func(
		core CoreParams,
		services ServiceParams,
		middlewareManager *appmiddleware.Manager,
		sessionManager *appmiddleware.SessionManager,
	) (web.Handler, error) {
		baseHandler := web.NewBaseHandler(services.FormService, core.Logger)
		authHandler, err := web.NewAuthHandler(web.HandlerDeps{
			BaseHandler:       baseHandler,
			UserService:       services.UserService,
			SessionManager:    sessionManager,
			Renderer:          core.Renderer,
			MiddlewareManager: middlewareManager,
			Config:            core.Config,
			Logger:            core.Logger,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create auth handler: %w", err)
		}
		return authHandler, nil
	}),
	// Health handler
	AnnotateHandler(func(core CoreParams, db *database.Database) (web.Handler, error) {
		// Create repository
		repo := healthadapter.NewRepository(db.DB)

		// Create service
		svc := healthdomain.NewService(core.Logger, repo)

		// Create handler
		handler := health.NewHandler(svc, core.Logger)

		core.Logger.Debug("registered handler",
			logging.StringField("handler_name", "HealthHandler"),
			logging.StringField("handler_type", fmt.Sprintf("%T", handler)),
			logging.StringField("operation", "handler_registration"),
		)
		return handler, nil
	}),
	// Middleware manager provider
	fx.Provide(func(
		core CoreParams,
		services ServiceParams,
		sessionManager *appmiddleware.SessionManager,
	) *appmiddleware.Manager {
		return appmiddleware.New(&appmiddleware.ManagerConfig{
			Logger:         core.Logger,
			Security:       &core.Config.Security,
			UserService:    services.UserService,
			SessionManager: sessionManager,
			Config:         core.Config,
		})
	}),
)

// RootModule combines all infrastructure-level modules into a single module.
var RootModule = fx.Options(
	InfrastructureModule,
	HandlerModule,
)

// wrapCreatorSQLX and wrapAssigner are now in internal/infrastructure/store/module.go
