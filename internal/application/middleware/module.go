// Package middleware provides HTTP middleware components.
package middleware

import (
	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/auth"
	"github.com/goformx/goforms/internal/application/middleware/session"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"go.uber.org/fx"
)

// Module provides all middleware dependencies (legacy Manager system only)
var Module = fx.Module("middleware",
	fx.Provide(
		// Path manager for centralized path management
		constants.NewPathManager,

		// Auth middleware
		auth.NewMiddleware,

		// Access manager using path manager
		fx.Annotate(
			func(_ logging.Logger, pathManager *constants.PathManager) *access.Manager {
				config := &access.Config{
					DefaultAccess: access.Authenticated,
					PublicPaths:   pathManager.PublicPaths,
					AdminPaths:    pathManager.AdminPaths,
				}
				rules := generateAccessRules(pathManager)

				return access.NewManager(config, rules)
			},
		),

		// Session manager using path manager
		fx.Annotate(
			func(
				logger logging.Logger,
				cfg *config.Config,
				lc fx.Lifecycle,
				accessManager *access.Manager,
				pathManager *constants.PathManager,
			) *session.Manager {
				sessionConfig := &session.Config{
					SessionConfig: &cfg.Session,
					Config:        cfg,
					PublicPaths:   pathManager.PublicPaths,
					StaticPaths:   pathManager.StaticPaths,
				}

				return session.NewManager(logger, sessionConfig, lc, accessManager)
			},
		),

		// Legacy Manager with simplified config - direct infrastructure config usage
		fx.Annotate(
			func(
				logger logging.Logger,
				cfg *config.Config,
				userService user.Service,
				formService formdomain.Service,
				sessionManager *session.Manager,
				accessManager *access.Manager,
				sanitizer sanitization.ServiceInterface,
			) *Manager {
				return NewManager(&ManagerConfig{
					Logger:         logger,
					Config:         cfg, // Single source of truth
					UserService:    userService,
					FormService:    formService,
					SessionManager: sessionManager,
					AccessManager:  accessManager,
					Sanitizer:      sanitizer,
				})
			},
		),
	),
)

// generateAccessRules creates access rules using the path manager
func generateAccessRules(pathManager *constants.PathManager) []access.Rule {
	rules := []access.Rule{}

	// Public routes
	for _, path := range pathManager.PublicPaths {
		rules = append(rules, access.Rule{
			Path:        path,
			AccessLevel: access.Public,
		})
	}

	// API validation endpoints
	for _, path := range pathManager.APIValidationPaths {
		rules = append(rules, access.Rule{
			Path:        path,
			AccessLevel: access.Public,
		})
	}

	// Static assets
	for _, path := range pathManager.StaticPaths {
		rules = append(rules, access.Rule{
			Path:        path,
			AccessLevel: access.Public,
		})
	}

	// Admin routes
	for _, path := range pathManager.AdminPaths {
		rules = append(rules, access.Rule{
			Path:        path,
			AccessLevel: access.Admin,
		})
	}

	// Add specific API rules
	apiPaths := []string{
		constants.PathAPIv1,
		constants.PathAPIForms,
		constants.PathAPIAdmin,
		constants.PathAPIAdminUsers,
		constants.PathAPIAdminForms,
	}

	for _, path := range apiPaths {
		rules = append(rules, access.Rule{
			Path:        path,
			AccessLevel: access.Authenticated,
		})
	}

	return rules
}
