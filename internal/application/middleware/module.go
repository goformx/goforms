// Package middleware provides HTTP middleware components.
package middleware

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Module provides middleware dependencies
var Module = fx.Options(
	fx.Provide(
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
					"/",
					"/login",
					"/signup",
					"/forgot-password",
					"/reset-password",
					"/health",
					"/metrics",
					"/demo",
				},
				ExemptPaths: []string{
					"/api/public",
					"/api/health",
					"/api/metrics",
					"/api/validation",
					"/api/validation/signup",
					"/api/validation/login",
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
