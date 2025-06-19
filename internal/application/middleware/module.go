// Package middleware provides HTTP middleware components.
package middleware

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/auth"
	middlewareconfig "github.com/goformx/goforms/internal/application/middleware/config"
	"github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Module provides middleware dependencies
var Module = fx.Options(
	// Core middleware components
	fx.Provide(
		// Centralized middleware configuration
		middlewareconfig.NewMiddlewareConfig,

		// Context middleware
		context.NewMiddleware,

		// Auth middleware
		fx.Annotate(
			auth.NewMiddleware,
		),

		// Access manager with centralized configuration
		fx.Annotate(
			func(logger logging.Logger, middlewareConfig *middlewareconfig.MiddlewareConfig) *access.AccessManager {
				return access.NewAccessManager(middlewareConfig.Access, middlewareConfig.GetAccessRules())
			},
		),

		// Session manager with centralized configuration
		fx.Annotate(
			func(
				logger logging.Logger,
				cfg *config.Config,
				lc fx.Lifecycle,
				accessManager *access.AccessManager,
				middlewareConfig *middlewareconfig.MiddlewareConfig,
			) *session.Manager {
				return session.NewManager(logger, middlewareConfig.Session, lc, accessManager)
			},
		),
	),
)
