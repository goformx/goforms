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
	logger.Info("Initializing base handler")
	handler := web.NewBaseHandler(formService, logger)
	logger.Info("Base handler initialized successfully")
	return handler
}

// provideAuthHandler provides the auth handler
func provideAuthHandler(
	deps web.HandlerDeps,
) (*web.AuthHandler, error) {
	deps.Logger.Info("Initializing auth handler")
	handler, err := web.NewAuthHandler(deps)
	if err != nil {
		deps.Logger.Error("Failed to initialize auth handler", logging.Error(err))
		return nil, err
	}
	deps.Logger.Info("Auth handler initialized successfully")
	return handler, nil
}

// providePageHandler provides the page handler
func providePageHandler(
	deps web.HandlerDeps,
	formService form.Service,
) (*web.PageHandler, error) {
	deps.Logger.Info("Initializing page handler")
	handler, err := web.NewPageHandler(deps, formService)
	if err != nil {
		deps.Logger.Error("Failed to initialize page handler", logging.Error(err))
		return nil, err
	}
	deps.Logger.Info("Page handler initialized successfully")
	return handler, nil
}

// provideWebHandler provides the web handler
func provideWebHandler(
	deps web.HandlerDeps,
) (*web.WebHandler, error) {
	deps.Logger.Info("Initializing web handler")
	handler, err := web.NewWebHandler(deps)
	if err != nil {
		deps.Logger.Error("Failed to initialize web handler", logging.Error(err))
		return nil, err
	}
	deps.Logger.Info("Web handler initialized successfully")
	return handler, nil
}

// provideDemoHandler provides the demo handler
func provideDemoHandler(
	deps web.HandlerDeps,
) *web.DemoHandler {
	deps.Logger.Info("Initializing demo handler")
	handler := web.NewDemoHandler(deps)
	deps.Logger.Info("Demo handler initialized successfully")
	return handler
}

// provideFormHandler provides the form handler
func provideFormHandler(
	deps web.HandlerDeps,
	formService form.Service,
) *web.FormHandler {
	deps.Logger.Info("Initializing form handler")
	handler := web.NewFormHandler(deps, formService)
	deps.Logger.Info("Form handler initialized successfully")
	return handler
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
				func(deps web.HandlerDeps) web.Handler {
					deps.Logger.Info("Registering demo handler in web_handlers group")
					return provideDemoHandler(deps)
				},
				fx.ResultTags(`group:"web_handlers"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				func(deps web.HandlerDeps, formService form.Service) web.Handler {
					deps.Logger.Info("Registering form handler in web_handlers group")
					return provideFormHandler(deps, formService)
				},
				fx.ResultTags(`group:"web_handlers"`),
			),
		),
	}
}
