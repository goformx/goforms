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

// HandlerProviders provides the application handlers
func HandlerProviders() []fx.Option {
	return []fx.Option{
		fx.Provide(
			// Base handler
			func(
				logger logging.Logger,
				authMiddleware *middleware.CookieAuthMiddleware,
				formService form.Service,
			) *handlers.BaseHandler {
				return handlers.NewBaseHandler(authMiddleware, formService, logger)
			},

			// Auth handler
			func(
				baseHandler *handlers.BaseHandler,
				userService user.Service,
				sessionManager *middleware.SessionManager,
				renderer *view.Renderer,
				middlewareManager *middleware.Manager,
				config *config.Config,
				logger logging.Logger,
			) *handler.AuthHandler {
				h := handler.NewAuthHandler(
					baseHandler,
					userService,
					sessionManager,
					renderer,
					middlewareManager,
					config,
					logger,
				)

				// Validate dependencies before returning
				if err := h.Validate(); err != nil {
					panic(fmt.Sprintf("failed to initialize auth handler: %v", err))
				}

				return h
			},

			// Page handler
			func(
				baseHandler *handlers.BaseHandler,
				userService user.Service,
				sessionManager *middleware.SessionManager,
				renderer *view.Renderer,
				middlewareManager *middleware.Manager,
				config *config.Config,
			) *handler.PageHandler {
				return handler.NewPageHandler(
					baseHandler,
					userService,
					sessionManager,
					renderer,
					middlewareManager,
					config,
				)
			},

			// Web handler
			func(
				baseHandler *handlers.BaseHandler,
				userService user.Service,
				sessionManager *middleware.SessionManager,
				renderer *view.Renderer,
				middlewareManager *middleware.Manager,
				config *config.Config,
				logger logging.Logger,
			) *handler.WebHandler {
				h := handler.NewWebHandler(
					baseHandler,
					userService,
					sessionManager,
					renderer,
					middlewareManager,
					config,
					logger,
				)

				// Validate dependencies before returning
				if err := h.Validate(); err != nil {
					panic(fmt.Sprintf("failed to initialize web handler: %v", err))
				}

				return h
			},
		),
	}
}
