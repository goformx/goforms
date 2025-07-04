package web

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/fx"

	"github.com/labstack/echo/v4"

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
	"github.com/goformx/goforms/internal/presentation/view"

	"github.com/goformx/goforms/internal/application/constants"
)

// Module provides web handler dependencies
var Module = fx.Module("web-handlers",
	// Core dependencies
	fx.Provide(
		// Base handler for common functionality
		fx.Annotate(
			NewBaseHandler,
			fx.ParamTags(``, ``, ``, ``, ``, ``, ``, ``),
		),

		// Auth components for SRP compliance
		NewAuthRequestParser,
		fx.Annotate(
			NewAuthResponseBuilder,
			fx.ParamTags(``),
		),
		NewAuthService,

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
				requestParser *AuthRequestParser,
				responseBuilder *AuthResponseBuilder,
				authService *AuthService,
				sanitizer sanitization.ServiceInterface,
			) (Handler, error) {
				return NewAuthHandler(
					base, authMiddleware, requestUtils, schemaGenerator,
					requestParser, responseBuilder, authService, sanitizer,
				)
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Web handler - public access
		fx.Annotate(
			func(base *BaseHandler, authMiddleware *auth.Middleware) (Handler, error) {
				return NewPageHandler(base, authMiddleware)
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Form Web handler - authenticated access
		fx.Annotate(
			func(
				base *BaseHandler,
				formService form.Service,
				formValidator *validation.FormValidator,
				sanitizer sanitization.ServiceInterface,
			) (Handler, error) {
				return NewFormWebHandler(base, formService, formValidator, sanitizer), nil
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Form API handler - authenticated access
		fx.Annotate(
			func(
				base *BaseHandler,
				formService form.Service,
				accessManager *access.Manager,
				formValidator *validation.FormValidator,
				sanitizer sanitization.ServiceInterface,
			) (Handler, error) {
				return NewFormAPIHandler(base, formService, accessManager, formValidator, sanitizer), nil
			},
			fx.ResultTags(`group:"handlers"`),
		),

		// Dashboard handler - authenticated access
		fx.Annotate(
			func(base *BaseHandler, accessManager *access.Manager, authMiddleware *auth.Middleware) (Handler, error) {
				return NewDashboardHandler(base, accessManager, authMiddleware), nil
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
		},
		fx.ParamTags(``, `group:"handlers"`),
	)),
)

// RouteRegistrar handles route registration for all handlers
type RouteRegistrar struct {
	handlers      []Handler
	accessManager *access.Manager
	logger        logging.Logger
}

// NewRouteRegistrar creates a new route registrar
func NewRouteRegistrar(
	handlers []Handler,
	accessManager *access.Manager,
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
	for i, handler := range rr.handlers {
		rr.logger.Info("Registering handler",
			"index", i,
			"handler_type", fmt.Sprintf("%T", handler))
		rr.registerHandlerRoutes(e, handler)
	}
}

// registerHandlerRoutes registers routes for a specific handler
func (rr *RouteRegistrar) registerHandlerRoutes(e *echo.Echo, handler Handler) {
	switch h := handler.(type) {
	case *AuthHandler:
		rr.registerAuthRoutes(e, h)
	case *PageHandler:
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
	e.GET(constants.PathLogin, h.Login)
	e.POST(constants.PathLoginPost, h.LoginPost)
	e.GET(constants.PathSignup, h.Signup)
	e.POST(constants.PathSignupPost, h.SignupPost)
	e.POST(constants.PathLogout, h.Logout)

	// API routes with validation
	api := e.Group(constants.PathAPIV1)
	validationGroup := api.Group(constants.PathValidation)
	validationGroup.GET("/user-login", h.LoginValidation)
	validationGroup.GET("/user-signup", h.SignupValidation)
}

// registerWebRoutes registers public web routes
func (rr *RouteRegistrar) registerWebRoutes(e *echo.Echo, h *PageHandler) {
	e.GET(constants.PathHome, h.handleHome)
	e.GET(constants.PathDemo, h.handleDemo)
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
	dashboard := e.Group(constants.PathDashboard)
	// Apply auth middleware to fetch and store user in context
	dashboard.Use(h.AuthMiddleware.RequireAuth)
	dashboard.GET("", h.handleDashboard)
}

// RegisterHandlers registers all handlers with the Echo instance
func RegisterHandlers(
	e *echo.Echo,
	handlers []Handler,
	accessManager *access.Manager,
	logger logging.Logger,
) {
	registrar := NewRouteRegistrar(handlers, accessManager, logger)
	registrar.RegisterAll(e)

	// Log route count for debugging with breakdown
	if logger != nil {
		allRoutes := e.Routes()
		httpRoutes := 0
		assetRoutes := 0

		// Categorize routes and log them for debugging
		logger.Debug("Route breakdown:")

		for _, route := range allRoutes {
			path := route.Path
			method := route.Method

			if strings.HasPrefix(path, "/assets/") ||
				strings.HasPrefix(path, "/fonts/") ||
				path == "/favicon.ico" ||
				path == "/robots.txt" {
				assetRoutes++

				logger.Debug("  Asset route", "method", method, "path", path)
			} else {
				httpRoutes++

				logger.Debug("  HTTP route", "method", method, "path", path)
			}
		}

		logger.Info("Route registration completed",
			"total_routes", len(allRoutes),
			"http_routes", httpRoutes,
			"asset_routes", assetRoutes,
			"handlers_registered", len(handlers))
	}
}
