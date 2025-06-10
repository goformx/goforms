package infrastructure

import (
	"errors"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	formstore "github.com/goformx/goforms/internal/infrastructure/persistence/store/form"
	formsubmissionstore "github.com/goformx/goforms/internal/infrastructure/persistence/store/form/submission"
	userstore "github.com/goformx/goforms/internal/infrastructure/persistence/store/user"
	"github.com/goformx/goforms/internal/infrastructure/server"
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

// Module provides core infrastructure dependencies
var Module = fx.Options(
	// Core infrastructure
	fx.Provide(
		logging.NewFactory,
		database.NewGormDB,
		view.NewRenderer,
	),
	// Stores
	fx.Provide(NewStores),
	// Base handler
	fx.Provide(
		web.NewBaseHandler,
	),
	// Middleware
	fx.Provide(
		func(logger logging.Logger, config *config.Config) (*middleware.SessionManager, error) {
			if config == nil {
				return nil, errors.New("config is required")
			}
			// In development, use secure cookies only if explicitly enabled
			// In production, always use secure cookies
			secureCookie := !config.App.Debug || config.Security.SecureCookie
			sessionManager := middleware.NewSessionManager(logger, secureCookie)
			if sessionManager == nil {
				return nil, errors.New("failed to create session manager")
			}
			return sessionManager, nil
		},
		func(
			logger logging.Logger,
			config *config.Config,
			userService user.Service,
			sessionManager *middleware.SessionManager,
		) (*middleware.Manager, error) {
			if config == nil || sessionManager == nil {
				return nil, errors.New("config and session manager are required")
			}
			manager := middleware.NewManager(&middleware.ManagerConfig{
				Logger:         logger,
				Security:       &config.Security,
				Config:         config,
				UserService:    userService,
				SessionManager: sessionManager,
			})
			if manager == nil {
				return nil, errors.New("failed to create middleware manager")
			}
			return manager, nil
		},
	),
	// Handlers
	fx.Provide(
		fx.Annotate(
			web.NewWebHandler,
			fx.ResultTags(`group:"handlers"`),
			fx.As(new(web.Handler)),
		),
		fx.Annotate(
			web.NewAuthHandler,
			fx.ResultTags(`group:"handlers"`),
			fx.As(new(web.Handler)),
		),
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
