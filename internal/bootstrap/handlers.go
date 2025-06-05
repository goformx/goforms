package bootstrap

import (
	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"go.uber.org/fx"
)

// provideBaseHandler provides the base handler
func provideBaseHandler(
	formService form.Service,
	logger logging.Logger,
) *web.BaseHandler {
	return web.NewBaseHandler(formService, logger)
}

// provideAuthHandler provides the auth handler
func provideAuthHandler(
	deps web.HandlerDeps,
) (*web.AuthHandler, error) {
	return web.NewAuthHandler(deps)
}

// providePageHandler provides the page handler
func providePageHandler(
	deps web.HandlerDeps,
	formService form.Service,
) (*web.PageHandler, error) {
	return web.NewPageHandler(deps, formService)
}

// provideWebHandler provides the web handler
func provideWebHandler(
	deps web.HandlerDeps,
) (*web.WebHandler, error) {
	return web.NewWebHandler(deps)
}

// provideDemoHandler provides the demo handler
func provideDemoHandler(
	deps web.HandlerDeps,
) *web.DemoHandler {
	return web.NewDemoHandler(deps)
}

// provideFormHandler provides the form handler
func provideFormHandler(
	deps web.HandlerDeps,
	formService form.Service,
) *web.FormHandler {
	return web.NewFormHandler(deps, formService)
}

// HandlerProviders provides all web handlers
func HandlerProviders() []fx.Option {
	return []fx.Option{
		fx.Provide(
			provideBaseHandler,
			provideAuthHandler,
			providePageHandler,
			provideWebHandler,
			provideDemoHandler,
			provideFormHandler,
		),
		fx.Provide(
			fx.Annotate(
				provideDemoHandler,
				fx.ResultTags(`group:"web_handlers"`),
				fx.As(new(web.Handler)),
			),
		),
		fx.Provide(
			fx.Annotate(
				provideFormHandler,
				fx.ResultTags(`group:"web_handlers"`),
				fx.As(new(web.Handler)),
			),
		),
	}
}
