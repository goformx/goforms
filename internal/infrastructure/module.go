package infrastructure

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/handlers/web"
	appmiddleware "github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/form"
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

	UserStore user.Store
	FormStore form.Store
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

// Module provides core infrastructure dependencies
var Module = fx.Options(
	// Core infrastructure
	fx.Provide(
		logging.NewFactory,
		database.NewDB,
		view.NewRenderer,
	),
	// Stores
	fx.Provide(
		fx.Annotate(userstore.NewStore, fx.As(new(user.Store))),
		fx.Annotate(formstore.NewStore, fx.As(new(form.Store))),
	),
	// Base handler
	fx.Provide(
		func(formService form.Service, logger logging.Logger) *web.BaseHandler {
			return web.NewBaseHandler(formService, logger)
		},
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
			return appmiddleware.New(&appmiddleware.ManagerConfig{
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
