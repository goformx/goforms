// Package presentation provides the presentation layer components and their dependency injection setup.
package presentation

import (
	"context"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/adapters/echo"
	"github.com/goformx/goforms/internal/presentation/handlers/api"
	"github.com/goformx/goforms/internal/presentation/handlers/auth"
	"github.com/goformx/goforms/internal/presentation/handlers/dashboard"
	"github.com/goformx/goforms/internal/presentation/handlers/forms"
	"github.com/goformx/goforms/internal/presentation/handlers/openapi"
	"github.com/goformx/goforms/internal/presentation/handlers/pages"
	"github.com/goformx/goforms/internal/presentation/handlers/validation"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
	"github.com/goformx/goforms/internal/presentation/view"
	echosrv "github.com/labstack/echo/v4"
)

// AuthHandlerParams contains dependencies for creating an AuthHandler
type AuthHandlerParams struct {
	fx.In
	UserService    user.Service
	SessionManager *session.Manager
	Renderer       view.Renderer
	Config         *config.Config
	AssetManager   web.AssetManagerInterface
	Logger         logging.Logger
}

// DashboardHandlerParams contains dependencies for creating a DashboardHandler
type DashboardHandlerParams struct {
	fx.In
	FormService    form.Service
	SessionManager *session.Manager
	Renderer       view.Renderer
	Config         *config.Config
	AssetManager   web.AssetManagerInterface
	Logger         logging.Logger
}

// NewAuthHandlerWithDeps creates a new AuthHandler with injected dependencies
func NewAuthHandlerWithDeps(params AuthHandlerParams) *auth.AuthHandler {
	return auth.NewAuthHandler(
		params.UserService,
		params.SessionManager,
		params.Renderer,
		params.Config,
		params.AssetManager,
		params.Logger,
	)
}

// NewDashboardHandlerWithDeps creates a new DashboardHandler with injected dependencies
func NewDashboardHandlerWithDeps(params DashboardHandlerParams) *dashboard.DashboardHandler {
	return dashboard.NewDashboardHandler(
		params.FormService,
		params.SessionManager,
		params.Renderer,
		params.Config,
		params.AssetManager,
		params.Logger,
	)
}

var Module = fx.Module("presentation",
	fx.Provide(
		fx.Annotate(
			pages.NewPageHandler,
			fx.As(new(httpiface.Handler)),
			fx.ResultTags(`group:"handlers"`),
		),
		fx.Annotate(
			NewAuthHandlerWithDeps,
			fx.As(new(httpiface.Handler)),
			fx.ResultTags(`group:"handlers"`),
		),
		fx.Annotate(
			NewDashboardHandlerWithDeps,
			fx.As(new(httpiface.Handler)),
			fx.ResultTags(`group:"handlers"`),
		),
		fx.Annotate(
			forms.NewFormHandler,
			fx.As(new(httpiface.Handler)),
			fx.ResultTags(`group:"handlers"`),
		),
		fx.Annotate(
			api.NewApiHandler,
			fx.As(new(httpiface.Handler)),
			fx.ResultTags(`group:"handlers"`),
		),
		fx.Annotate(
			openapi.NewOpenAPIHandler,
			fx.As(new(httpiface.Handler)),
			fx.ResultTags(`group:"handlers"`),
		),
		fx.Annotate(
			validation.NewValidationHandler,
			fx.As(new(httpiface.Handler)),
			fx.ResultTags(`group:"handlers"`),
		),
		echo.NewEchoAdapter,
	),
	fx.Invoke(RegisterRoutes),
)

// RegisterRoutes registers all handlers with the EchoAdapter
func RegisterRoutes(
	lc fx.Lifecycle,
	e *echosrv.Echo,
	adapter *echo.EchoAdapter,
	handlers struct {
		fx.In
		Handlers []httpiface.Handler `group:"handlers"`
	},
) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			for _, h := range handlers.Handlers {
				adapter.RegisterHandler(h)
			}

			return nil
		},
	})
}
