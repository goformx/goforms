package web

import (
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
			accessManager *access.AccessManager,
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

		// Login handler - public access
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

		// Public web page handlers - public access
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

		// Form handler - authenticated access
		fx.Annotate(
			func(deps *HandlerDeps, accessManager *access.AccessManager) (Handler, error) {
				handler := NewFormHandler(*deps, deps.FormService, accessManager)
				return handler, nil
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Dashboard handler - authenticated access
		fx.Annotate(
			func(deps *HandlerDeps, accessManager *access.AccessManager) (Handler, error) {
				handler := NewDashboardHandler(*deps, accessManager)
				return handler, nil
			},
			fx.ResultTags(`group:"handlers"`),
		),
	),
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
