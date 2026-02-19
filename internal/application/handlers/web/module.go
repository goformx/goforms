package web

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/fx"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/validation"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// Module provides web handler dependencies
var Module = fx.Module("web-handlers",
	// Core dependencies
	fx.Provide(NewBaseHandler),

	// Handler providers
	fx.Provide(
		// Form API handler - authenticated access
		fx.Annotate(
			func(
				base *BaseHandler,
				formService form.Service,
				accessManager *access.Manager,
				formValidator *validation.FormValidator,
				sanitizer sanitization.ServiceInterface,
				userEnsurer user.UserEnsurer,
			) (Handler, error) {
				return NewFormAPIHandler(base, formService, accessManager, formValidator, sanitizer, userEnsurer), nil
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
	case *FormAPIHandler:
		rr.registerFormAPIRoutes(e, h)
	default:
		// Unknown handler type - skip
		_ = h
	}
}

// registerFormAPIRoutes registers form API routes
func (rr *RouteRegistrar) registerFormAPIRoutes(e *echo.Echo, h *FormAPIHandler) {
	h.RegisterRoutes(e)
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
