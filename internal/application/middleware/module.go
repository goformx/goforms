// Package middleware provides HTTP middleware components.
package middleware

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/auth"
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
		// Context middleware
		context.NewMiddleware,

		// Request utilities
		request.NewUtils,

		// Auth middleware
		fx.Annotate(
			auth.NewMiddleware,
		),

		// Access manager
		fx.Annotate(
			func(logger logging.Logger) *access.AccessManager {
				config := access.DefaultConfig()
				manager := access.NewAccessManager(config, access.DefaultRules())
				logger.Debug("access manager initialized", "service", "access")
				return manager
			},
		),

		// Session manager
		fx.Annotate(
			func(
				logger logging.Logger,
				cfg *config.Config,
				lc fx.Lifecycle,
				accessManager *access.AccessManager,
			) *session.Manager {
				sessionConfig := &session.SessionConfig{
					Config:        cfg,
					SessionConfig: &cfg.Session,
					PublicPaths: []string{
						"/login",
						"/signup",
						"/forgot-password",
						"/reset-password",
						"/health",
						"/metrics",
						"/demo",
					},
					ExemptPaths: []string{
						"/api/v1/validation",
						"/api/v1/validation/login",
						"/api/v1/validation/signup",
						"/api/v1/public",
						"/api/v1/health",
						"/api/v1/metrics",
					},
					StaticPaths: []string{
						"/static",
						"/assets",
						"/images",
						"/css",
						"/js",
						"/favicon.ico",
					},
				}
				return session.NewManager(logger, sessionConfig, lc, accessManager)
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
