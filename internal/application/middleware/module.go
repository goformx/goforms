package middleware

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Module provides middleware dependencies
var Module = fx.Options(
	fx.Provide(
		// Session manager
		func(logger logging.Logger, cfg *config.Config, lc fx.Lifecycle) *SessionManager {
			return NewSessionManager(logger, &cfg.Session, lc)
		},
		// Middleware manager
		func(
			logger logging.Logger,
			cfg *config.Config,
			userService user.Service,
			sessionManager *SessionManager,
		) *Manager {
			return NewManager(&ManagerConfig{
				Logger:         logger,
				Security:       &cfg.Security,
				UserService:    userService,
				SessionManager: sessionManager,
				Config:         cfg,
			})
		},
	),
)
