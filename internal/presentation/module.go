// Package presentation provides the presentation layer components and their dependency injection setup.
package presentation

import (
	"context"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/services"
	"github.com/goformx/goforms/internal/infrastructure/adapters/http"
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
	AuthService     *services.AuthUseCaseService
	RequestAdapter  http.RequestAdapter
	ResponseAdapter http.ResponseAdapter
	Renderer        view.Renderer
	Config          *config.Config
	AssetManager    web.AssetManagerInterface
	Logger          logging.Logger
}

// DashboardHandlerParams contains dependencies for creating a DashboardHandler
type DashboardHandlerParams struct {
	fx.In
	FormService     *services.FormUseCaseService
	RequestAdapter  http.RequestAdapter
	ResponseAdapter http.ResponseAdapter
	Renderer        view.Renderer
	Config          *config.Config
	AssetManager    web.AssetManagerInterface
	Logger          logging.Logger
}

// FormHandlerParams contains dependencies for creating a FormHandler
type FormHandlerParams struct {
	fx.In
	FormService     *services.FormUseCaseService
	RequestAdapter  http.RequestAdapter
	ResponseAdapter http.ResponseAdapter
	Renderer        view.Renderer
	Config          *config.Config
	AssetManager    web.AssetManagerInterface
	Logger          logging.Logger
}

// PageHandlerParams contains dependencies for creating a PageHandler
type PageHandlerParams struct {
	fx.In
	Renderer     view.Renderer
	Cfg          *config.Config
	AssetManager web.AssetManagerInterface
	Logger       logging.Logger
}

// NewAuthHandlerWithDeps creates a new AuthHandler with injected dependencies
func NewAuthHandlerWithDeps(params AuthHandlerParams) *auth.AuthHandler {
	return auth.NewAuthHandler(
		params.AuthService,
		params.RequestAdapter,
		params.ResponseAdapter,
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
		params.RequestAdapter,
		params.ResponseAdapter,
		params.Renderer,
		params.Config,
		params.AssetManager,
		params.Logger,
	)
}

// NewFormHandlerWithDeps creates a new FormHandler with injected dependencies
func NewFormHandlerWithDeps(params FormHandlerParams) *forms.FormHandler {
	return forms.NewFormHandler(
		params.FormService,
		params.RequestAdapter,
		params.ResponseAdapter,
		params.Renderer,
		params.Config,
		params.AssetManager,
		params.Logger,
	)
}

// NewPageHandlerWithDeps creates a new PageHandler with injected dependencies
func NewPageHandlerWithDeps(params PageHandlerParams) *pages.PageHandler {
	return pages.NewPageHandler(
		params.Renderer,
		params.Cfg,
		params.AssetManager,
		params.Logger,
	)
}

var Module = fx.Module("presentation",
	fx.Provide(
		fx.Annotate(
			NewPageHandlerWithDeps,
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
			NewFormHandlerWithDeps,
			fx.As(new(httpiface.Handler)),
			fx.ResultTags(`group:"handlers"`),
		),
		fx.Annotate(
			api.NewAPIHandler,
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
