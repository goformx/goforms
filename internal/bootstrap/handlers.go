package bootstrap

import (
	"fmt"

	"github.com/goformx/goforms/internal/application/handler"
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/handlers"
	"github.com/goformx/goforms/internal/presentation/view"
	"go.uber.org/fx"
)

// HandlerProviders returns all the handler-related providers
func HandlerProviders() []fx.Option {
	return []fx.Option{
		fx.Provide(
			func(
				logger logging.Logger,
				authMiddleware *middleware.CookieAuthMiddleware,
				formService form.Service,
				userService user.Service,
			) *handler.AuthHandler {
				baseHandler := handlers.NewBaseHandler(authMiddleware, formService, logger)
				h := &handler.AuthHandler{
					BaseHandler: baseHandler,
					UserService: userService,
				}

				// Validate dependencies before returning
				if err := h.Validate(); err != nil {
					panic(fmt.Sprintf("failed to initialize auth handler: %v", err))
				}

				return h
			},
			handler.NewStaticHandler,
			handler.NewVersionHandler,
			func(
				logger logging.Logger,
				authMiddleware *middleware.CookieAuthMiddleware,
				formService form.Service,
				userService user.Service,
				sessionManager *middleware.SessionManager,
				renderer *view.Renderer,
				middlewareManager *middleware.Manager,
				cfg *config.Config,
			) *handler.WebHandler {
				baseHandler := handlers.NewBaseHandler(authMiddleware, formService, logger)
				h := handler.NewWebHandler(baseHandler, userService, sessionManager)

				// Set all required dependencies before returning the handler
				handler.WithRenderer(renderer)(h)
				handler.WithMiddlewareManager(middlewareManager)(h)
				handler.WithConfig(cfg)(h)

				// Validate dependencies before returning
				if err := h.Validate(); err != nil {
					panic(fmt.Sprintf("failed to initialize web handler: %v", err))
				}

				return h
			},
		),
	}
}
