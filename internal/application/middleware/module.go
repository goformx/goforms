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

// Module provides all middleware dependencies
var Module = fx.Options(
	fx.Provide(
		// Path manager for centralized path management
		constants.NewPathManager,

		// Auth middleware
		auth.NewMiddleware,

		// Access manager using path manager
		fx.Annotate(
			func(logger logging.Logger, pathManager *constants.PathManager) *access.AccessManager {
				config := &access.Config{
					DefaultAccess: access.AuthenticatedAccess,
					PublicPaths:   pathManager.PublicPaths,
					AdminPaths:    pathManager.AdminPaths,
				}
				rules := generateAccessRules(pathManager)
				return access.NewAccessManager(config, rules)
			},
		),

		// Session manager using path manager
		fx.Annotate(
			func(
				logger logging.Logger,
				cfg *config.Config,
				lc fx.Lifecycle,
				accessManager *access.AccessManager,
				pathManager *constants.PathManager,
			) *session.Manager {
				sessionConfig := &session.SessionConfig{
					SessionConfig: &cfg.Session,
					Config:        cfg,
					PublicPaths:   pathManager.PublicPaths,
					StaticPaths:   pathManager.StaticPaths,
				}
				return session.NewManager(logger, sessionConfig, lc, accessManager)
			},
		),

		// Manager with simplified config - direct infrastructure config usage
		fx.Annotate(
			func(
				logger logging.Logger,
				cfg *config.Config,
				userService user.Service,
				formService formdomain.Service,
				sessionManager *session.Manager,
				accessManager *access.AccessManager,
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
func generateAccessRules(pathManager *constants.PathManager) []access.AccessRule {
	rules := []access.AccessRule{}

	// Public routes
	for _, path := range pathManager.PublicPaths {
		rules = append(rules, access.AccessRule{
			Path:        path,
			AccessLevel: access.PublicAccess,
		})
	}

	// API validation endpoints
	for _, path := range pathManager.APIValidationPaths {
		rules = append(rules, access.AccessRule{
			Path:        path,
			AccessLevel: access.PublicAccess,
		})
	}

	// Static assets
	for _, path := range pathManager.StaticPaths {
		rules = append(rules, access.AccessRule{
			Path:        path,
			AccessLevel: access.PublicAccess,
		})
	}

	// Admin routes
	for _, path := range pathManager.AdminPaths {
		rules = append(rules, access.AccessRule{
			Path:        path,
			AccessLevel: access.AdminAccess,
		})
	}

	// Add specific API rules
	rules = append(rules, []access.AccessRule{
		// Public form endpoints (for embedded forms) - GET only
		{Path: constants.PathAPIForms + "/:id/schema", AccessLevel: access.PublicAccess, Methods: []string{"GET"}},

		// Authenticated routes
		{Path: constants.PathDashboard, AccessLevel: access.AuthenticatedAccess},
		{Path: constants.PathForms, AccessLevel: access.AuthenticatedAccess},
		{Path: constants.PathForms + "/:id", AccessLevel: access.AuthenticatedAccess},
		{Path: constants.PathAPIForms, AccessLevel: access.AuthenticatedAccess},
		{Path: constants.PathAPIForms + "/:id", AccessLevel: access.AuthenticatedAccess},

		// Admin API routes
		{Path: constants.PathAPIAdmin, AccessLevel: access.AdminAccess},
		{Path: constants.PathAPIAdminUsers, AccessLevel: access.AdminAccess},
		{Path: constants.PathAPIAdminForms, AccessLevel: access.AdminAccess},
	}...)

	return rules
}
