package framework

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/auth"
	"github.com/goformx/goforms/internal/application/middleware/request"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/application/validation"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/goformx/goforms/internal/presentation/inertia"
)

func handlersModule() fx.Option {
	return fx.Module(
		"handlers",
		fx.Provide(
			web.NewBaseHandler,
			web.NewAuthRequestParser,
			web.NewAuthResponseBuilder,
			web.NewAuthService,
			provideHandlerDeps,
		),
		fx.Provide(
			fx.Annotate(
				func(
					base *web.BaseHandler,
					authMiddleware *auth.Middleware,
					requestUtils *request.Utils,
					schemaGenerator *validation.SchemaGenerator,
					requestParser *web.AuthRequestParser,
					responseBuilder *web.AuthResponseBuilder,
					authService *web.AuthService,
					sanitizer sanitization.ServiceInterface,
				) (web.Handler, error) {
					return web.NewAuthHandler(
						base,
						authMiddleware,
						requestUtils,
						schemaGenerator,
						requestParser,
						responseBuilder,
						authService,
						sanitizer,
					)
				},
				fx.As(new(web.Handler)),
				fx.ResultTags(`group:"handlers"`),
			),
			fx.Annotate(
				func(base *web.BaseHandler, authMiddleware *auth.Middleware) (web.Handler, error) {
					return web.NewPageHandler(base, authMiddleware)
				},
				fx.As(new(web.Handler)),
				fx.ResultTags(`group:"handlers"`),
			),
			fx.Annotate(
				func(
					base *web.BaseHandler,
					formService form.Service,
					formValidator *validation.FormValidator,
					sanitizer sanitization.ServiceInterface,
				) (web.Handler, error) {
					return web.NewFormWebHandler(base, formService, formValidator, sanitizer), nil
				},
				fx.As(new(web.Handler)),
				fx.ResultTags(`group:"handlers"`),
			),
			fx.Annotate(
				func(
					base *web.BaseHandler,
					formService form.Service,
					accessManager *access.Manager,
					formValidator *validation.FormValidator,
					sanitizer sanitization.ServiceInterface,
				) (web.Handler, error) {
					return web.NewFormAPIHandler(base, formService, accessManager, formValidator, sanitizer), nil
				},
				fx.As(new(web.Handler)),
				fx.ResultTags(`group:"handlers"`),
			),
			fx.Annotate(
				func(
					base *web.BaseHandler,
					accessManager *access.Manager,
					authMiddleware *auth.Middleware,
				) (web.Handler, error) {
					return web.NewDashboardHandler(base, accessManager, authMiddleware), nil
				},
				fx.As(new(web.Handler)),
				fx.ResultTags(`group:"handlers"`),
			),
		),
		fx.Invoke(registerHandlerLifecycle),
	)
}

func provideHandlerDeps(
	logger logging.Logger,
	cfg *config.Config,
	sessionManager *session.Manager,
	middlewareManager *middleware.Manager,
	inertiaManager *inertia.Manager,
	userService user.Service,
	formService form.Service,
) (*web.HandlerDeps, error) {
	deps := &web.HandlerDeps{
		Logger:            logger,
		Config:            cfg,
		SessionManager:    sessionManager,
		MiddlewareManager: middlewareManager,
		Inertia:           inertiaManager,
		UserService:       userService,
		FormService:       formService,
	}

	if err := deps.Validate(); err != nil {
		return nil, err
	}

	return deps, nil
}

func registerHandlerLifecycle(
	lc fx.Lifecycle,
	handlers []web.Handler,
	logger logging.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			for _, h := range handlers {
				if err := h.Start(ctx); err != nil {
					logger.Error("failed to start handler", "error", err)
					return fmt.Errorf("start handler: %w", err)
				}
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			for _, h := range handlers {
				if err := h.Stop(ctx); err != nil {
					logger.Error("failed to stop handler", "error", err)
					return fmt.Errorf("stop handler: %w", err)
				}
			}

			return nil
		},
	})
}
