// Package middleware provides HTTP middleware components.
package middleware

import (
	"fmt"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/auth"
	"github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/goformx/goforms/internal/application/services"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/session"
)

// Module provides all middleware dependencies
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

				// Debug logging to show what rules are being created
				fmt.Printf("DEBUG: Creating access manager with %d rules\n", len(rules))
				for _, rule := range rules {
					fmt.Printf("DEBUG: Access rule - Path: %s, AccessLevel: %d\n", rule.Path, rule.AccessLevel)
				}

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

		// Session manager adapter for services.SessionManager interface
		fx.Annotate(
			func(manager *session.Manager) services.SessionManager {
				return &sessionManagerAdapter{manager: manager}
			},
			fx.As(new(services.SessionManager)),
		),

		// NEW ARCHITECTURE: Core middleware components
		// Middleware configuration provider
		fx.Annotate(
			NewViperMiddlewareConfig,
			fx.As(new(MiddlewareConfig)),
		),

		// Registry provider
		fx.Annotate(
			NewRegistry,
			fx.As(new(core.Registry)),
		),

		// Orchestrator provider
		fx.Annotate(
			NewOrchestrator,
			fx.As(new(core.Orchestrator)),
		),

		// Echo integration adapter
		fx.Annotate(
			NewEchoOrchestratorAdapter,
		),
	),
)

// RegisterAllMiddleware registers all middleware with the registry
func RegisterAllMiddleware(registry core.Registry, logger logging.Logger) error {
	logger.Debug("Starting middleware registration")

	// Register basic middleware
	basicMiddleware := []struct {
		name string
		mw   core.Middleware
	}{
		{"recovery", NewRecoveryMiddleware()},
		{"cors", NewCORSMiddleware()},
		{"security-headers", NewSecurityHeadersMiddleware()},
		{"request-id", NewRequestIDMiddleware()},
		{"timeout", NewTimeoutMiddleware()},
		{"logging", NewLoggingMiddleware()},
	}

	for _, m := range basicMiddleware {
		logger.Debug("Registering basic middleware", "name", m.name)
		if err := registry.Register(m.name, m.mw); err != nil {
			logger.Error("Failed to register basic middleware", "name", m.name, "error", err)
			return fmt.Errorf("failed to register basic middleware %s: %w", m.name, err)
		}

		logger.Debug("registered middleware", "name", m.name)
	}

	// Register security middleware
	securityMiddleware := []struct {
		name string
		mw   core.Middleware
	}{
		{"csrf", NewCSRFMiddleware()},
		{"rate-limit", NewRateLimitMiddleware()},
		{"input-validation", NewInputValidationMiddleware()},
	}

	for _, m := range securityMiddleware {
		logger.Debug("Registering security middleware", "name", m.name)
		if err := registry.Register(m.name, m.mw); err != nil {
			logger.Error("Failed to register security middleware", "name", m.name, "error", err)
			return fmt.Errorf("failed to register security middleware %s: %w", m.name, err)
		}

		logger.Debug("registered security middleware", "name", m.name)
	}

	// Register auth middleware
	authMiddleware := []struct {
		name string
		mw   core.Middleware
	}{
		{"session", NewSessionMiddleware()},
		{"authentication", NewAuthenticationMiddleware()},
		{"authorization", NewAuthorizationMiddleware()},
	}

	for _, m := range authMiddleware {
		logger.Debug("Registering auth middleware", "name", m.name)
		if err := registry.Register(m.name, m.mw); err != nil {
			logger.Error("Failed to register auth middleware", "name", m.name, "error", err)
			return fmt.Errorf("failed to register auth middleware %s: %w", m.name, err)
		}

		logger.Debug("registered auth middleware", "name", m.name)
	}

	logger.Debug("Middleware registration completed", "total_count", registry.Count())
	return nil
}

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

	// Add authenticated routes
	authenticatedPaths := []string{
		constants.PathDashboard,
		constants.PathForms,
		"/forms/new",
		"/forms/:id",
		"/forms/:id/edit",
		"/forms/:id/preview",
		"/forms/:id/submissions",
	}

	for _, path := range authenticatedPaths {
		rules = append(rules, access.Rule{
			Path:        path,
			AccessLevel: access.Authenticated,
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

	// Debug logging to show what rules are being generated
	fmt.Printf("DEBUG: Generated %d access rules\n", len(rules))
	for _, rule := range rules {
		fmt.Printf("DEBUG: Rule - Path: %s, AccessLevel: %d\n", rule.Path, rule.AccessLevel)
	}

	return rules
}

// sessionManagerAdapter adapts session.Manager to services.SessionManager interface
type sessionManagerAdapter struct {
	manager *session.Manager
}

func (a *sessionManagerAdapter) CreateSession(userID, email, role string) (string, error) {
	sessionID, err := a.manager.CreateSessionApp(userID, email, role)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionID, nil
}

func (a *sessionManagerAdapter) DeleteSession(sessionID string) {
	a.manager.DeleteSessionApp(sessionID)
}

func (a *sessionManagerAdapter) GetSession(sessionID string) (services.SessionData, bool) {
	return a.manager.GetSessionApp(sessionID)
}
