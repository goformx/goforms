// Package middleware provides HTTP middleware components.
package middleware

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/auth"
	middlewareconfig "github.com/goformx/goforms/internal/application/middleware/config"
	"github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/middleware/request"
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

		// Request utilities
		request.NewUtils,

		// Auth middleware
		fx.Annotate(
			auth.NewMiddleware,
		),

		// Access manager with centralized configuration
		fx.Annotate(
			func(logger logging.Logger, middlewareConfig *middlewareconfig.MiddlewareConfig) *access.AccessManager {
				manager := access.NewAccessManager(middlewareConfig.Access, middlewareConfig.GetAccessRules())
				logger.Debug("access manager initialized", "service", "access")
				return manager
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

		// Middleware manager
		// fx.Annotate(
		// 	func(
		// 		logger logging.Logger,
		// 		cfg *config.Config,
		// 		userService user.Service,
		// 		sessionManager *session.Manager,
		// 		accessManager *access.AccessManager,
		// 	) *Manager {
		// 		return NewManager(&ManagerConfig{
		// 			Logger:         logger,
		// 			Security:       &cfg.Security,
		// 			UserService:    userService,
		// 			Config:         cfg,
		// 			SessionManager: sessionManager,
		// 			AccessManager:  accessManager,
		// 		})
		// 	},
		// ),
	),
)
