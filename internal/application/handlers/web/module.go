package web

import (
	"context"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// Module provides web handler dependencies
var Module = fx.Options(
	// Core dependencies
	fx.Provide(
		// Base handler for common functionality
		fx.Annotate(
			func(
				logger logging.Logger,
				cfg *config.Config,
				userService user.Service,
				formService form.Service,
				renderer view.Renderer,
				sessionManager *session.Manager,
			) *BaseHandler {
				return NewBaseHandler(logger, cfg, userService, formService, renderer, sessionManager)
			},
		),

		// Legacy HandlerDeps for backward compatibility
		fx.Annotate(
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
		),
	),

	// Handler providers
	fx.Provide(
		// Auth handler - public access
		fx.Annotate(
			func(base *BaseHandler) (Handler, error) {
				return NewAuthHandler(base)
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Web handler - public access
		fx.Annotate(
			func(base *BaseHandler) (Handler, error) {
				return NewWebHandler(base)
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Form Web handler - authenticated access
		fx.Annotate(
			func(base *BaseHandler, formService form.Service) (Handler, error) {
				return NewFormWebHandler(base, formService), nil
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Form API handler - authenticated access
		fx.Annotate(
			func(base *BaseHandler, formService form.Service, accessManager *access.AccessManager) (Handler, error) {
				return NewFormAPIHandler(base, formService, accessManager), nil
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Dashboard handler - authenticated access
		fx.Annotate(
			func(base *BaseHandler, accessManager *access.AccessManager) (Handler, error) {
				return NewDashboardHandler(base, accessManager), nil
			},
			fx.ResultTags(`group:"handlers"`),
		),
	),

	// Lifecycle hooks
	fx.Invoke(fx.Annotate(
		func(lc fx.Lifecycle, handlers []Handler, logger logging.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					for _, h := range handlers {
						if err := h.Start(ctx); err != nil {
							logger.Error("failed to start handler", "error", err)
							return err
						}
					}
					return nil
				},
				OnStop: func(ctx context.Context) error {
					for _, h := range handlers {
						if err := h.Stop(ctx); err != nil {
							logger.Error("failed to stop handler", "error", err)
							return err
						}
					}
					return nil
				},
			})
		},
		fx.ParamTags(``, `group:"handlers"`),
	)),
)

// RegisterHandlers registers all handlers with the Echo instance
func RegisterHandlers(
	e *echo.Echo,
	handlers []Handler,
	accessManager *access.AccessManager,
	logger logging.Logger,
) {
	for _, handler := range handlers {
		// Register routes with appropriate access control
		switch h := handler.(type) {
		case *AuthHandler:
			// Public routes
			e.GET("/login", h.Login)
			e.POST("/login", h.LoginPost)
			e.GET("/signup", h.Signup)
			e.POST("/signup", h.SignupPost)
			e.POST("/logout", h.Logout)

			// API routes with validation
			api := e.Group("/api/v1")
			validation := api.Group("/validation")
			validation.GET("/login", h.LoginValidation)
			validation.GET("/signup", h.SignupValidation)

		case *WebHandler:
			// Public routes
			e.GET("/", h.handleHome)
			e.GET("/demo", h.handleDemo)

		case *FormWebHandler:
			// Web UI routes with access control
			h.RegisterRoutes(e, accessManager)

		case *FormAPIHandler:
			// API routes
			h.RegisterRoutes(e)

		case *DashboardHandler:
			// Authenticated routes
			dashboard := e.Group("/dashboard")
			dashboard.Use(access.Middleware(accessManager, logger))
			dashboard.GET("", h.handleDashboard)
		}
	}
}
