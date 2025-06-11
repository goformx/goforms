package web

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/view"
)

// Module provides web handler dependencies
var Module = fx.Options(
	fx.Provide(
		// Handler dependencies
		func(
			logger logging.Logger,
			cfg *config.Config,
			sessionManager *session.Manager,
			middlewareManager *middleware.Manager,
			renderer view.Renderer,
			userService user.Service,
			formService form.Service,
		) *HandlerDeps {
			return &HandlerDeps{
				Logger:            logger,
				Config:            cfg,
				SessionManager:    sessionManager,
				MiddlewareManager: middlewareManager,
				Renderer:          renderer,
				UserService:       userService,
				FormService:       formService,
			}
		},

		// Login handler
		fx.Annotate(
			func(deps *HandlerDeps) (Handler, error) {
				handler, err := NewAuthHandler(*deps)
				if err != nil {
					return nil, err
				}
				return handler, nil
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Public web page handlers
		fx.Annotate(
			func(deps *HandlerDeps) (Handler, error) {
				handler, err := NewWebHandler(*deps)
				if err != nil {
					return nil, err
				}
				return handler, nil
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Form handler
		fx.Annotate(
			func(deps *HandlerDeps) (Handler, error) {
				handler := NewFormHandler(*deps, deps.FormService)
				return handler, nil
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Dashboard handler
		fx.Annotate(
			func(deps *HandlerDeps) (Handler, error) {
				handler := NewDashboardHandler(*deps)
				return handler, nil
			},
			fx.ResultTags(`group:"handlers"`),
		),
	),
)
