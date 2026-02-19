// Package middleware provides HTTP middleware components.
package middleware

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/auth"
	"github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/goformx/goforms/internal/application/middleware/session"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
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
					// Laravel assertion auth: no session cookie; auth via X-User-Id/X-Signature
					ExemptPaths: []string{constants.PathAPIFormsLaravel},
				}

				return session.NewManager(logger, sessionConfig, lc, accessManager)
			},
		),

		// NEW ARCHITECTURE: Core middleware components
		// Middleware configuration provider
		fx.Annotate(
			NewMiddlewareConfig,
			fx.As(new(MiddlewareConfig)),
		),

		// Registry provider
		fx.Annotate(
			func(logger logging.Logger, config MiddlewareConfig) core.Registry {
				return NewRegistry(logger, config)
			},
			fx.As(new(core.Registry)),
		),

		// Orchestrator provider
		fx.Annotate(
			func(registry core.Registry, config MiddlewareConfig, logger logging.Logger) core.Orchestrator {
				return NewOrchestrator(registry, config, logger)
			},
			fx.As(new(core.Orchestrator)),
		),

		// Echo integration adapter
		fx.Annotate(
			NewEchoOrchestratorAdapter,
		),

		// Migration adapter for gradual transition
		fx.Annotate(
			NewMigrationAdapter,
		),

		// LEGACY: Manager with simplified config - direct infrastructure config usage
		// This will be removed after migration is complete
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

	// Lifecycle hooks for middleware initialization
	fx.Invoke(func(
		lc fx.Lifecycle,
		registry core.Registry,
		orchestrator core.Orchestrator,
		logger logging.Logger,
	) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				// Register all middleware with the registry
				if err := registerAllMiddleware(registry, logger); err != nil {
					return err
				}

				// Validate orchestrator configuration
				if err := orchestrator.ValidateConfiguration(); err != nil {
					return fmt.Errorf("failed to validate orchestrator configuration: %w", err)
				}

				logger.Info("middleware system initialized successfully")

				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Info("middleware system shutting down")

				return nil
			},
		})
	}),
)

// registerAllMiddleware registers all middleware with the registry
func registerAllMiddleware(registry core.Registry, logger logging.Logger) error {
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
		if err := registry.Register(m.name, m.mw); err != nil {
			return fmt.Errorf("failed to register basic middleware %s: %w", m.name, err)
		}

		logger.Info("registered middleware", "name", m.name)
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
		if err := registry.Register(m.name, m.mw); err != nil {
			return fmt.Errorf("failed to register security middleware %s: %w", m.name, err)
		}

		logger.Info("registered security middleware", "name", m.name)
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
		if err := registry.Register(m.name, m.mw); err != nil {
			return fmt.Errorf("failed to register auth middleware %s: %w", m.name, err)
		}

		logger.Info("registered auth middleware", "name", m.name)
	}

	return nil
}

// generateAccessRules creates access rules using the path manager
func generateAccessRules(pathManager *constants.PathManager) []access.Rule {
	// Preallocate with estimated capacity based on typical path counts
	rules := make([]access.Rule, 0, len(pathManager.PublicPaths)+len(pathManager.APIValidationPaths)+
		len(pathManager.AdminPaths)+len(pathManager.StaticPaths))

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

	// Public form embed routes at /forms/:id/... for cross-origin embedding
	publicFormRules := []access.Rule{
		{Path: constants.PathFormsPublic + "/:id/schema", AccessLevel: access.Public},
		{Path: constants.PathFormsPublic + "/:id/validation", AccessLevel: access.Public},
		{Path: constants.PathFormsPublic + "/:id/submit", AccessLevel: access.Public},
		{Path: constants.PathFormsPublic + "/:id/embed", AccessLevel: access.Public},
	}
	rules = append(rules, publicFormRules...)

	return rules
}
