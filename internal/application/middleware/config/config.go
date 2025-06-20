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
	// Security configuration
	Security *SecurityConfig
	// Rate limiting configuration
	RateLimit *RateLimitConfig
	// Logging configuration
	Logging *LoggingConfig
}

// SecurityConfig holds security-related middleware configuration
type SecurityConfig struct {
	// CSRF protection
	CSRFEnabled bool
	CSRFSecret  string
	// CORS configuration
	CORSEnabled bool
	CORSOrigins []string
	// Security headers
	SecurityHeaders map[string]string
	// Content Security Policy
	CSPEnabled    bool
	CSPDirectives string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool
	Requests    int
	Window      int // seconds
	Burst       int
	SkipPaths   []string
	SkipMethods []string
}

// LoggingConfig holds logging middleware configuration
type LoggingConfig struct {
	Enabled      bool
	LogRequests  bool
	LogResponses bool
	LogErrors    bool
	LogUserAgent bool
	LogIPAddress bool
	SkipPaths    []string
	SkipMethods  []string
}

// NewMiddlewareConfig creates a new middleware configuration from the application config
func NewMiddlewareConfig(appConfig *config.Config) *MiddlewareConfig {
	return &MiddlewareConfig{
		Access:    newAccessConfig(),
		Session:   newSessionConfig(appConfig),
		Security:  newSecurityConfig(appConfig),
		RateLimit: newRateLimitConfig(),
		Logging:   newLoggingConfig(),
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

// newSecurityConfig creates the security configuration
func newSecurityConfig(appConfig *config.Config) *SecurityConfig {
	return &SecurityConfig{
		CSRFEnabled: appConfig.Security.CSRFConfig.Enabled,
		CSRFSecret:  appConfig.Security.CSRFConfig.Secret,
		CORSEnabled: true,
		CORSOrigins: []string{"*"}, // Configure based on environment
		SecurityHeaders: map[string]string{
			"X-Frame-Options":           "DENY",
			"X-Content-Type-Options":    "nosniff",
			"X-XSS-Protection":          "1; mode=block",
			"Referrer-Policy":           "strict-origin-when-cross-origin",
			"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		},
		CSPEnabled: appConfig.App.Env == constants.EnvProduction,
		CSPDirectives: "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'",
	}
}

// newRateLimitConfig creates the rate limiting configuration
func newRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Enabled:  true,
		Requests: constants.DefaultRateLimit,
		Window:   constants.RateLimitWindow,
		Burst:    constants.RateLimitBurst,
		SkipPaths: []string{
			constants.PathHealth,
			constants.PathMetrics,
			constants.PathAPIHealth,
			constants.PathAPIMetrics,
		},
		SkipMethods: []string{"GET", "HEAD", "OPTIONS"},
	}
}

// newLoggingConfig creates the logging configuration
func newLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		Enabled:      true,
		LogRequests:  true,
		LogResponses: false, // Only log responses for errors
		LogErrors:    true,
		LogUserAgent: true,
		LogIPAddress: true,
		SkipPaths: []string{
			constants.PathHealth,
			constants.PathMetrics,
			constants.PathAPIHealth,
			constants.PathAPIMetrics,
			constants.PathFavicon,
			constants.PathRobotsTxt,
		},
		SkipMethods: []string{"HEAD", "OPTIONS"},
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
