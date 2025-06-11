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
			fx.ResultTags(`group:"handler_deps"`),
		),
	),

	// Handler providers
	fx.Provide(
		// Auth handler - public access
		fx.Annotate(
			func(deps *HandlerDeps) (Handler, error) {
				return NewAuthHandler(*deps)
			},
			fx.ResultTags(`group:"handlers"`),
			fx.As(new(Handler)),
		),

		// Web handler - public access
		fx.Annotate(
			func(deps *HandlerDeps) (Handler, error) {
				return NewWebHandler(*deps)
			},
			fx.ResultTags(`group:"handlers"`),
			fx.As(new(Handler)),
		),

		// Form handler - authenticated access
		fx.Annotate(
			func(deps *HandlerDeps, accessManager *access.AccessManager) (Handler, error) {
				return NewFormHandler(*deps, deps.FormService, accessManager), nil
			},
			fx.ResultTags(`group:"handlers"`),
			fx.As(new(Handler)),
		),

		// Dashboard handler - authenticated access
		fx.Annotate(
			func(deps *HandlerDeps, accessManager *access.AccessManager) (Handler, error) {
				return NewDashboardHandler(*deps, accessManager), nil
			},
			fx.ResultTags(`group:"handlers"`),
			fx.As(new(Handler)),
		),
	),

	// Lifecycle hooks
	fx.Invoke(func(lc fx.Lifecycle, handlers []Handler, logger logging.Logger) {
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
	}),
)

// RegisterHandlers registers all handlers with the Echo instance
func RegisterHandlers(e *echo.Echo, handlers []Handler, accessManager *access.AccessManager, logger logging.Logger) {
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

		case *FormHandler:
			// Authenticated routes
			forms := e.Group("/forms")
			forms.Use(access.Middleware(accessManager, logger))
			forms.GET("/new", h.handleFormNew)
			forms.POST("", h.handleFormCreate)
			forms.GET("/:id/edit", h.handleFormEdit)
			forms.PUT("/:id", h.handleFormUpdate)
			forms.DELETE("/:id", h.handleFormDelete)
			forms.GET("/:id/submissions", h.handleFormSubmissions)

			// API routes
			api := e.Group("/api/v1")
			formsAPI := api.Group("/forms")
			formsAPI.GET("/:id/schema", h.handleFormSchema)
			formsAPI.PUT("/:id/schema", h.handleFormSchemaUpdate)

		case *DashboardHandler:
			// Authenticated routes
			dashboard := e.Group("/dashboard")
			dashboard.Use(access.Middleware(accessManager, logger))
			dashboard.GET("", h.handleDashboard)
		}
	}
}
