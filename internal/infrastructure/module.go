package infrastructure

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
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
	formstore "github.com/goformx/goforms/internal/infrastructure/persistence/store/form"
	userstore "github.com/goformx/goforms/internal/infrastructure/persistence/store/user"
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

// InfrastructureModule provides core infrastructure dependencies
var InfrastructureModule = fx.Options(
	fx.Provide(
		config.New,
		database.NewDB,
		func(db *database.Database) *sqlx.DB {
			return db.DB
		},
		fx.Annotate(
			userstore.NewStore,
			fx.As(new(user.Store)),
		),
		fx.Annotate(
			formstore.NewStore,
			fx.As(new(form.Store)),
		),
		func(logger logging.Logger) *appmiddleware.SessionManager {
			logger.Info("InfrastructureModule: Creating session manager...")
			sm := appmiddleware.NewSessionManager(logger)
			logger.Info("InfrastructureModule: Session manager created")
			return sm
		},
		func(
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
		},
	),
)

// HandlerModule provides all HTTP handlers for the application.
var HandlerModule = fx.Options(
	// Web handlers
	AnnotateHandler(func(
		core CoreParams,
		services ServiceParams,
		middlewareManager *appmiddleware.Manager,
	) (web.Handler, error) {
		baseHandler := web.NewBaseHandler(services.FormService, core.Logger)
		handler, err := web.NewWebHandler(web.HandlerDeps{
			BaseHandler:       baseHandler,
			UserService:       services.UserService,
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
)

// RootModule provides core infrastructure dependencies
var RootModule = fx.Options(
	// Core infrastructure
	fx.Provide(
		config.New,
		database.NewDB,
		func(db *database.Database) *sqlx.DB {
			return db.DB
		},
		// Store implementations
		fx.Annotate(
			userstore.NewStore,
			fx.As(new(user.Store)),
		),
		fx.Annotate(
			formstore.NewStore,
			fx.As(new(form.Store)),
		),
		// Middleware
		func(logger logging.Logger) *appmiddleware.SessionManager {
			logger.Info("Creating session manager...")
			return appmiddleware.NewSessionManager(logger)
		},
		func(
			logger logging.Logger,
			config *config.Config,
			userService user.Service,
			sessionManager *appmiddleware.SessionManager,
		) *appmiddleware.Manager {
			logger.Info("Creating middleware manager...")
			return appmiddleware.New(&appmiddleware.ManagerConfig{
				Logger:         logger,
				Security:       &config.Security,
				UserService:    userService,
				SessionManager: sessionManager,
				Config:         config,
			})
		},
	),
	// Register handlers
	HandlerModule,
	// Lifecycle hooks
	fx.Invoke(
		func(lc fx.Lifecycle, db *database.Database, logger logging.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Info("Verifying database connection...")
					if err := db.Ping(); err != nil {
						logger.Error("Failed to verify database connection",
							logging.ErrorField("error", err))
						return fmt.Errorf("failed to verify database connection: %w", err)
					}
					logger.Info("Database connection verified successfully")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("Closing database connection...")
					if err := db.Close(); err != nil {
						logger.Error("Failed to close database connection",
							logging.ErrorField("error", err))
						return fmt.Errorf("failed to close database connection: %w", err)
					}
					logger.Info("Database connection closed successfully")
					return nil
				},
			})
		},
	),
)
