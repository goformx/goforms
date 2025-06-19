package web

import (
	"context"

	"go.uber.org/fx"

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
			func(
				base *BaseHandler,
				authMiddleware *auth.Middleware,
				requestUtils *request.Utils,
				schemaGenerator *validation.SchemaGenerator,
			) (Handler, error) {
				return NewAuthHandler(base, authMiddleware, requestUtils, schemaGenerator)
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Web handler - public access
		fx.Annotate(
			func(base *BaseHandler, authMiddleware *auth.Middleware) (Handler, error) {
				return NewWebHandler(base, authMiddleware)
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Form Web handler - authenticated access
		fx.Annotate(
			func(base *BaseHandler, formService form.Service, formValidator *validation.FormValidator) (Handler, error) {
				return NewFormWebHandler(base, formService, formValidator), nil
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Form API handler - authenticated access
		fx.Annotate(
			func(base *BaseHandler, formService form.Service, accessManager *access.AccessManager, formValidator *validation.FormValidator) (Handler, error) {
				return NewFormAPIHandler(base, formService, accessManager, formValidator), nil
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

// RouteRegistrar handles route registration for all handlers
type RouteRegistrar struct {
	handlers      []Handler
	accessManager *access.AccessManager
	logger        logging.Logger
}

// NewRouteRegistrar creates a new route registrar
func NewRouteRegistrar(
	handlers []Handler,
	accessManager *access.AccessManager,
	logger logging.Logger,
) *RouteRegistrar {
	return &RouteRegistrar{
		handlers:      handlers,
		accessManager: accessManager,
		logger:        logger,
	}
}

// RegisterAll registers all handler routes
func (rr *RouteRegistrar) RegisterAll(e *echo.Echo) {
	for _, handler := range rr.handlers {
		rr.registerHandlerRoutes(e, handler)
	}
}

// registerHandlerRoutes registers routes for a specific handler
func (rr *RouteRegistrar) registerHandlerRoutes(e *echo.Echo, handler Handler) {
	switch h := handler.(type) {
	case *AuthHandler:
		rr.registerAuthRoutes(e, h)
	case *WebHandler:
		rr.registerWebRoutes(e, h)
	case *FormWebHandler:
		rr.registerFormWebRoutes(e, h)
	case *FormAPIHandler:
		rr.registerFormAPIRoutes(e, h)
	case *DashboardHandler:
		rr.registerDashboardRoutes(e, h)
	}
}

// registerAuthRoutes registers authentication routes
func (rr *RouteRegistrar) registerAuthRoutes(e *echo.Echo, h *AuthHandler) {
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
}

// registerWebRoutes registers public web routes
func (rr *RouteRegistrar) registerWebRoutes(e *echo.Echo, h *WebHandler) {
	e.GET("/", h.handleHome)
	e.GET("/demo", h.handleDemo)
}

// registerFormWebRoutes registers form web UI routes
func (rr *RouteRegistrar) registerFormWebRoutes(e *echo.Echo, h *FormWebHandler) {
	h.RegisterRoutes(e, rr.accessManager)
}

// registerFormAPIRoutes registers form API routes
func (rr *RouteRegistrar) registerFormAPIRoutes(e *echo.Echo, h *FormAPIHandler) {
	h.RegisterRoutes(e)
}

// registerDashboardRoutes registers dashboard routes
func (rr *RouteRegistrar) registerDashboardRoutes(e *echo.Echo, h *DashboardHandler) {
	dashboard := e.Group("/dashboard")
	dashboard.Use(access.Middleware(rr.accessManager, rr.logger))
	dashboard.GET("", h.handleDashboard)
}

// RegisterHandlers registers all handlers with the Echo instance
func RegisterHandlers(
	e *echo.Echo,
	handlers []Handler,
	accessManager *access.AccessManager,
	logger logging.Logger,
) {
	registrar := NewRouteRegistrar(handlers, accessManager, logger)
	registrar.RegisterAll(e)
}
