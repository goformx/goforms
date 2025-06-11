// Package middleware provides HTTP middleware components.
package middleware

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Module provides middleware dependencies
var Module = fx.Options(
	fx.Provide(
		// Access manager
		func(logger logging.Logger) *access.AccessManager {
			return access.NewAccessManager(access.DefaultRules())
		},

		// Session manager
		func(
			logger logging.Logger,
			cfg *config.Config,
			lc fx.Lifecycle,
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
			return session.NewManager(logger, sessionConfig, lc)
		},
	),
)
