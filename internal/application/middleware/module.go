// Package middleware provides HTTP middleware components.
package middleware

import (
	"fmt"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// RegisterAllMiddleware registers all middleware with the registry.
func RegisterAllMiddleware(registry core.Registry, logger logging.Logger) error {
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

// GenerateAccessRules creates access rules using the path manager.
func GenerateAccessRules(pathManager *constants.PathManager) []access.Rule {
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

	return rules
}
