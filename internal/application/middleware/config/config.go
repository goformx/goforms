package config

import (
	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/infrastructure/config"
)

// MiddlewareConfig holds all middleware configuration
type MiddlewareConfig struct {
	// Access control configuration
	Access *access.Config
	// Session configuration
	Session *session.SessionConfig
	// Security configuration (references infrastructure config)
	Security *config.SecurityConfig
	// Logging configuration (references infrastructure config)
	Logging *config.LoggingConfig
}

// NewMiddlewareConfig creates a new middleware configuration from the application config
func NewMiddlewareConfig(appConfig *config.Config) *MiddlewareConfig {
	return &MiddlewareConfig{
		Access:   newAccessConfig(),
		Session:  newSessionConfig(appConfig),
		Security: &appConfig.Security, // Direct reference to infrastructure config
		Logging:  &appConfig.Logging,  // Direct reference to infrastructure config
	}
}

// newAccessConfig creates the access control configuration
func newAccessConfig() *access.Config {
	return &access.Config{
		DefaultAccess: access.AuthenticatedAccess,
		PublicPaths: []string{
			constants.PathHome,
			constants.PathLogin,
			constants.PathSignup,
			constants.PathDemo,
			constants.PathHealth,
			constants.PathMetrics,
			constants.PathForgotPassword,
			constants.PathResetPassword,
			constants.PathVerifyEmail,
			constants.PathStatic,
			constants.PathAssets,
			constants.PathImages,
			constants.PathCSS,
			constants.PathJS,
			constants.PathFonts,
			constants.PathFavicon,
			constants.PathRobotsTxt,
		},
		AdminPaths: []string{
			constants.PathAdmin,
		},
	}
}

// newSessionConfig creates the session configuration
func newSessionConfig(appConfig *config.Config) *session.SessionConfig {
	return &session.SessionConfig{
		SessionConfig: &appConfig.Session,
		Config:        appConfig,
		PublicPaths: []string{
			constants.PathHome,
			constants.PathLogin,
			constants.PathSignup,
			constants.PathDemo,
			constants.PathHealth,
			constants.PathMetrics,
			constants.PathForgotPassword,
			constants.PathResetPassword,
		},
		ExemptPaths: []string{
			constants.PathAPIValidation,
			constants.PathAPIValidationLogin,
			constants.PathAPIValidationSignup,
			constants.PathAPIPublic,
			constants.PathAPIHealth,
			constants.PathAPIMetrics,
		},
		StaticPaths: []string{
			constants.PathStatic,
			constants.PathAssets,
			constants.PathImages,
			constants.PathCSS,
			constants.PathJS,
			constants.PathFonts,
			constants.PathFavicon,
		},
	}
}

// GetAccessRules returns the default access rules for the application
func (mc *MiddlewareConfig) GetAccessRules() []access.AccessRule {
	return []access.AccessRule{
		// Public routes
		{Path: constants.PathHome, AccessLevel: access.PublicAccess},
		{Path: constants.PathLogin, AccessLevel: access.PublicAccess},
		{Path: constants.PathSignup, AccessLevel: access.PublicAccess},
		{Path: constants.PathDemo, AccessLevel: access.PublicAccess},
		{Path: constants.PathHealth, AccessLevel: access.PublicAccess},
		{Path: constants.PathMetrics, AccessLevel: access.PublicAccess},

		// API validation endpoints
		{Path: constants.PathAPIValidation, AccessLevel: access.PublicAccess},
		{Path: constants.PathAPIValidationLogin, AccessLevel: access.PublicAccess},
		{Path: constants.PathAPIValidationSignup, AccessLevel: access.PublicAccess},

		// Public API endpoints
		{Path: constants.PathAPIPublic, AccessLevel: access.PublicAccess},

		// Public form endpoints (for embedded forms) - GET only
		{Path: constants.PathAPIForms + "/:id/schema", AccessLevel: access.PublicAccess, Methods: []string{"GET"}},

		// Static assets
		{Path: constants.PathStatic, AccessLevel: access.PublicAccess},
		{Path: constants.PathAssets, AccessLevel: access.PublicAccess},
		{Path: constants.PathImages, AccessLevel: access.PublicAccess},
		{Path: constants.PathCSS, AccessLevel: access.PublicAccess},
		{Path: constants.PathJS, AccessLevel: access.PublicAccess},
		{Path: constants.PathFonts, AccessLevel: access.PublicAccess},
		{Path: constants.PathFavicon, AccessLevel: access.PublicAccess},

		// Authenticated routes
		{Path: constants.PathDashboard, AccessLevel: access.AuthenticatedAccess},
		{Path: constants.PathForms, AccessLevel: access.AuthenticatedAccess},
		{Path: constants.PathForms + "/:id", AccessLevel: access.AuthenticatedAccess},
		{Path: constants.PathAPIForms, AccessLevel: access.AuthenticatedAccess},
		{Path: constants.PathAPIForms + "/:id", AccessLevel: access.AuthenticatedAccess},

		// Admin routes
		{Path: constants.PathAdmin, AccessLevel: access.AdminAccess},
		{Path: constants.PathAdminUsers, AccessLevel: access.AdminAccess},
		{Path: constants.PathAdminForms, AccessLevel: access.AdminAccess},
		{Path: constants.PathAPIAdmin, AccessLevel: access.AdminAccess},
		{Path: constants.PathAPIAdminUsers, AccessLevel: access.AdminAccess},
		{Path: constants.PathAPIAdminForms, AccessLevel: access.AdminAccess},
	}
}

// Validate validates the middleware configuration
func (mc *MiddlewareConfig) Validate() error {
	if err := mc.Access.Validate(); err != nil {
		return err
	}

	// Add additional validation as needed
	return nil
}
