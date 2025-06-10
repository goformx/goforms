// Package infrastructure provides core infrastructure components and their dependency injection setup.
package infrastructure

import (
	"context"
	"errors"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/handlers/web"
	appmiddleware "github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/form"
	formevent "github.com/goformx/goforms/internal/domain/form/event"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/database"
	infraevent "github.com/goformx/goforms/internal/infrastructure/event"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	formstore "github.com/goformx/goforms/internal/infrastructure/repository/form"
	formsubmissionstore "github.com/goformx/goforms/internal/infrastructure/repository/form/submission"
	userstore "github.com/goformx/goforms/internal/infrastructure/repository/user"
	"github.com/goformx/goforms/internal/infrastructure/server"
	"github.com/goformx/goforms/internal/infrastructure/validation"
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

	UserStore           user.Repository
	FormStore           form.Repository
	FormSubmissionStore form.SubmissionStore
}

// CoreParams represents core infrastructure dependencies
type CoreParams struct {
	fx.In
	Logger   logging.Logger
	Config   *config.Config
	Renderer *view.Renderer
	Echo     *echo.Echo
}

// ServiceParams contains all service dependencies that handlers might need
type ServiceParams struct {
	fx.In
	UserService user.Service
	FormService form.Service
}

// EventPublisherParams contains dependencies for creating an event publisher
type EventPublisherParams struct {
	fx.In

	Logger logging.Logger
}

// NewEventPublisher creates a new event publisher with dependencies
func NewEventPublisher(p EventPublisherParams) (formevent.Publisher, error) {
	if p.Logger == nil {
		return nil, errors.New("logger is required for event publisher")
	}
	return infraevent.NewMemoryPublisher(p.Logger), nil
}

// AnnotateHandler is a helper function that simplifies the creation of handler providers
func AnnotateHandler(fn any) fx.Option {
	return fx.Provide(
		fx.Annotate(
			fn,
			fx.As(new(web.Handler)),
			fx.ResultTags(`group:"handlers"`),
		),
	)
}

// NewStores creates new stores with proper error handling and logging
func NewStores(db *database.GormDB, logger logging.Logger) (Stores, error) {
	if db == nil {
		return Stores{}, errors.New("database connection is required")
	}

	userStore := userstore.NewStore(db, logger)
	if userStore == nil {
		logger.Error("failed to create store",
			"operation", "store_initialization",
			"store_type", "user",
			"error_type", "nil_store",
		)
		return Stores{}, errors.New("failed to create user store")
	}

	formStore := formstore.NewStore(db, logger)
	if formStore == nil {
		logger.Error("failed to create store",
			"operation", "store_initialization",
			"store_type", "form",
			"error_type", "nil_store",
		)
		return Stores{}, errors.New("failed to create form store")
	}

	formSubmissionStore := formsubmissionstore.NewStore(db, logger)
	if formSubmissionStore == nil {
		logger.Error("failed to create store",
			"operation", "store_initialization",
			"store_type", "form_submission",
			"error_type", "nil_store",
		)
		return Stores{}, errors.New("failed to create form submission store")
	}

	logger.Info("stores initialized successfully",
		"operation", "store_initialization",
		"store_types", "user,form,form_submission",
	)

	return Stores{
		UserStore:           userStore,
		FormStore:           formStore,
		FormSubmissionStore: formSubmissionStore,
	}, nil
}

// Module provides all infrastructure dependencies
var Module = fx.Options(
	// Core infrastructure
	fx.Provide(
		// Configuration
		func() (*config.Config, error) {
			return config.New()
		},
		// Logger
		func(cfg *config.Config) (logging.Logger, error) {
			if cfg == nil {
				return nil, errors.New("config is required for logger setup")
			}
			factory := logging.NewFactory(logging.FactoryConfig{
				AppName:     cfg.App.Name,
				Version:     cfg.App.Version,
				Environment: cfg.App.Env,
				Fields:      map[string]any{},
			})
			return factory.CreateLogger()
		},
		// Echo instance
		echo.New,
		// Validation
		validation.New,
		// Database
		database.NewGormDB,
	),

	// Start connection pool monitoring
	fx.Invoke(func(db *database.GormDB, lc fx.Lifecycle) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go db.MonitorConnectionPool(ctx)
				return nil
			},
		})
	}),

	// Event system
	fx.Provide(
		fx.Annotate(
			NewEventPublisher,
			fx.As(new(formevent.Publisher)),
		),
	),

	// Repositories
	fx.Provide(
		fx.Annotate(
			userstore.NewStore,
			fx.As(new(user.Repository)),
		),
		fx.Annotate(
			formstore.NewStore,
			fx.As(new(form.Repository)),
		),
		fx.Annotate(
			formsubmissionstore.NewStore,
			fx.As(new(form.SubmissionStore)),
		),
	),

	// Middleware
	fx.Provide(
		func(logger logging.Logger, config *config.Config) *appmiddleware.SessionManager {
			// In development, use secure cookies only if explicitly enabled
			// In production, always use secure cookies
			secureCookie := !config.App.Debug || config.Security.SecureCookie
			return appmiddleware.NewSessionManager(logger, secureCookie)
		},
		func(
			logger logging.Logger,
			config *config.Config,
			userService user.Service,
			sessionManager *appmiddleware.SessionManager,
		) *appmiddleware.Manager {
			return appmiddleware.NewManager(&appmiddleware.ManagerConfig{
				Logger:         logger,
				Security:       &config.Security,
				UserService:    userService,
				SessionManager: sessionManager,
				Config:         config,
			})
		},
	),

	// Stores
	fx.Provide(NewStores),

	// Web handlers
	fx.Provide(
		web.NewWebHandler,
		web.NewAuthHandler,
	),

	// Handlers
	fx.Provide(
		fx.Annotate(
			web.NewFormHandler,
			fx.ResultTags(`group:"handlers"`),
			fx.As(new(web.Handler)),
		),
		fx.Annotate(
			web.NewDemoHandler,
			fx.ResultTags(`group:"handlers"`),
			fx.As(new(web.Handler)),
		),
	),

	// Server setup
	fx.Provide(server.New),
)
