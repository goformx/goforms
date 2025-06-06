package infrastructure

import (
	"context"

	"github.com/jmoiron/sqlx"
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

// HandlerModule provides HTTP handlers
var HandlerModule = fx.Options(
	// Web handlers
	fx.Provide(
		fx.Annotate(
			func(core CoreParams, services ServiceParams, middlewareManager *appmiddleware.Manager) (web.Handler, error) {
				baseHandler := web.NewBaseHandler(services.FormService, core.Logger)
				deps := web.HandlerDeps{
					BaseHandler:       baseHandler,
					UserService:       services.UserService,
					SessionManager:    middlewareManager.GetSessionManager(),
					Renderer:          core.Renderer,
					MiddlewareManager: middlewareManager,
					Config:            core.Config,
					Logger:            core.Logger,
				}
				return web.NewWebHandler(deps)
			},
			fx.ResultTags(`group:"web_handlers"`),
		),
		fx.Annotate(
			func(core CoreParams, services ServiceParams, middlewareManager *appmiddleware.Manager) (web.Handler, error) {
				baseHandler := web.NewBaseHandler(services.FormService, core.Logger)
				deps := web.HandlerDeps{
					BaseHandler:       baseHandler,
					UserService:       services.UserService,
					SessionManager:    middlewareManager.GetSessionManager(),
					Renderer:          core.Renderer,
					MiddlewareManager: middlewareManager,
					Config:            core.Config,
					Logger:            core.Logger,
				}
				return web.NewAuthHandler(deps)
			},
			fx.ResultTags(`group:"web_handlers"`),
		),
		fx.Annotate(
			func(core CoreParams, services ServiceParams, middlewareManager *appmiddleware.Manager) (web.Handler, error) {
				baseHandler := web.NewBaseHandler(services.FormService, core.Logger)
				deps := web.HandlerDeps{
					BaseHandler:       baseHandler,
					UserService:       services.UserService,
					SessionManager:    middlewareManager.GetSessionManager(),
					Renderer:          core.Renderer,
					MiddlewareManager: middlewareManager,
					Config:            core.Config,
					Logger:            core.Logger,
				}
				handler := web.NewFormHandler(deps, services.FormService)
				return handler, nil
			},
			fx.ResultTags(`group:"web_handlers"`),
		),
		fx.Annotate(
			func(core CoreParams, services ServiceParams, middlewareManager *appmiddleware.Manager) (web.Handler, error) {
				baseHandler := web.NewBaseHandler(services.FormService, core.Logger)
				deps := web.HandlerDeps{
					BaseHandler:       baseHandler,
					UserService:       services.UserService,
					SessionManager:    middlewareManager.GetSessionManager(),
					Renderer:          core.Renderer,
					MiddlewareManager: middlewareManager,
					Config:            core.Config,
					Logger:            core.Logger,
				}
				return web.NewDemoHandler(deps)
			},
			fx.ResultTags(`group:"web_handlers"`),
		),
	),
	// Provide the handlers as a group
	fx.Provide(
		fx.Annotate(
			func(handlers []web.Handler) []web.Handler {
				return handlers
			},
			fx.ParamTags(`group:"web_handlers"`),
		),
	),
)

// RootModule combines all infrastructure modules
var RootModule = fx.Options(
	InfrastructureModule,
	HandlerModule,
	fx.Invoke(func(
		lifecycle fx.Lifecycle,
		core CoreParams,
		services ServiceParams,
		handlers []web.Handler,
	) {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				// Register handlers with Echo
				core.Logger.Info("registering handlers with Echo")
				for _, handler := range handlers {
					handler.Register(core.Echo)
				}
				return nil
			},
		})
	}),
)
