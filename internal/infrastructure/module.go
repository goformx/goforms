package infrastructure

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/application/middleware"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	formstore "github.com/goformx/goforms/internal/infrastructure/persistence/store/form"
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

	UserStore user.Repository
	FormStore formdomain.Repository
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
	FormService formdomain.Service
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

// InfrastructureModule provides infrastructure dependencies
var InfrastructureModule = fx.Options(
	fx.Provide(
		config.New,
		database.NewDB,
		func(db *database.Database) *sqlx.DB {
			return db.DB
		},
		fx.Annotate(
			userstore.NewStore,
			fx.As(new(user.Repository)),
		),
		fx.Annotate(
			formstore.NewFormStore,
			fx.As(new(formdomain.Repository)),
		),
		func(logger logging.Logger, config *config.Config) *middleware.SessionManager {
			// In development, use secure cookies only if explicitly enabled
			// In production, always use secure cookies
			secureCookie := !config.App.Debug || config.Security.SecureCookie
			return middleware.NewSessionManager(logger, secureCookie)
		},
		func(
			core CoreParams,
			services ServiceParams,
			sessionManager *middleware.SessionManager,
		) *middleware.Manager {
			return middleware.NewManager(&middleware.ManagerConfig{
				Logger:         core.Logger,
				Security:       &core.Config.Security,
				UserService:    services.UserService,
				SessionManager: sessionManager,
				Config:         core.Config,
			})
		},
	),
)

// Module provides core infrastructure dependencies
var Module = fx.Options(
	// Core infrastructure
	fx.Provide(
		logging.NewFactory,
		database.NewDB,
		func(db *database.Database) *sqlx.DB {
			return db.DB
		},
		view.NewRenderer,
	),
	// Stores
	fx.Provide(
		fx.Annotate(userstore.NewStore, fx.As(new(user.Repository))),
		fx.Annotate(formstore.NewFormStore, fx.As(new(formdomain.Repository))),
	),
	// Base handler
	fx.Provide(
		web.NewBaseHandler,
	),
	// Middleware
	fx.Provide(
		func(logger logging.Logger, config *config.Config) *middleware.SessionManager {
			// In development, use secure cookies only if explicitly enabled
			// In production, always use secure cookies
			secureCookie := !config.App.Debug || config.Security.SecureCookie
			return middleware.NewSessionManager(logger, secureCookie)
		},
		func(
			logger logging.Logger,
			config *config.Config,
			userService user.Service,
			sessionManager *middleware.SessionManager,
		) *middleware.Manager {
			return middleware.NewManager(&middleware.ManagerConfig{
				Logger:         logger,
				Security:       &config.Security,
				Config:         config,
				UserService:    userService,
				SessionManager: sessionManager,
			})
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
