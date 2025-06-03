package bootstrap

import (
	"fmt"

	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
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
				formService form.Service,
			) *web.BaseHandler {
				return web.NewBaseHandler(formService, logger)
			},

			// Auth handler
			func(
				baseHandler *web.BaseHandler,
				userService user.Service,
				sessionManager *middleware.SessionManager,
				renderer *view.Renderer,
				middlewareManager *middleware.Manager,
				cfg *config.Config,
				logger logging.Logger,
			) *web.AuthHandler {
				h := web.NewAuthHandler(
					baseHandler,
					userService,
					sessionManager,
					renderer,
					middlewareManager,
					cfg,
					logger,
				)

				// Validate dependencies before returning
				if err := h.Validate(); err != nil {
					panic(fmt.Sprintf("failed to initialize auth handler: %v", err))
				}

				return h
			},

			// Page handler
			web.NewPageHandler,

			// Web handler
			func(
				baseHandler *web.BaseHandler,
				userService user.Service,
				sessionManager *middleware.SessionManager,
				renderer *view.Renderer,
				middlewareManager *middleware.Manager,
				cfg *config.Config,
				logger logging.Logger,
			) *web.WebHandler {
				h := web.NewWebHandler(
					baseHandler,
					userService,
					sessionManager,
					renderer,
					middlewareManager,
					cfg,
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
