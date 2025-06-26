package config

import (
	"fmt"
	"strings"
	"time"
)

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	// CSRF protection
	CSRF CSRFConfig `envconfig:"CSRF"`

	// CORS configuration
	CORS CORSConfig `envconfig:"CORS"`

	// Rate limiting configuration
	RateLimit RateLimitConfig `envconfig:"RATE_LIMIT"`

	// Security headers configuration
	Headers SecurityHeadersConfig `envconfig:"HEADERS"`

	// Content Security Policy configuration
	CSP CSPConfig `envconfig:"CSP"`

	// Cookie security
	SecureCookie bool `envconfig:"GOFORMS_SECURITY_SECURE_COOKIE" default:"true"`

	// Debug mode
	Debug bool `envconfig:"GOFORMS_SECURITY_DEBUG" default:"false"`
}

// CSRFConfig holds CSRF-related configuration
type CSRFConfig struct {
	Enabled        bool   `envconfig:"GOFORMS_SECURITY_CSRF_ENABLED" default:"true"`
	Secret         string `envconfig:"GOFORMS_SECURITY_CSRF_SECRET" validate:"required"`
	TokenLength    int    `envconfig:"GOFORMS_SECURITY_CSRF_TOKEN_LENGTH" default:"32"`
	TokenLookup    string `envconfig:"GOFORMS_SECURITY_CSRF_TOKEN_LOOKUP" default:"header:X-Csrf-Token"`
	ContextKey     string `envconfig:"GOFORMS_SECURITY_CSRF_CONTEXT_KEY" default:"csrf"`
	CookieName     string `envconfig:"GOFORMS_SECURITY_CSRF_COOKIE_NAME" default:"_csrf"`
	CookiePath     string `envconfig:"GOFORMS_SECURITY_CSRF_COOKIE_PATH" default:"/"`
	CookieDomain   string `envconfig:"GOFORMS_SECURITY_CSRF_COOKIE_DOMAIN" default:""`
	CookieHTTPOnly bool   `envconfig:"GOFORMS_SECURITY_CSRF_COOKIE_HTTP_ONLY" default:"true"`
	CookieSameSite string `envconfig:"GOFORMS_SECURITY_CSRF_COOKIE_SAME_SITE" default:"Lax"`
	CookieMaxAge   int    `envconfig:"GOFORMS_SECURITY_CSRF_COOKIE_MAX_AGE" default:"86400"`
}

// CORSConfig holds CORS-related configuration
type CORSConfig struct {
	Enabled        bool     `envconfig:"GOFORMS_SECURITY_CORS_ENABLED" default:"true"`
	AllowedOrigins []string `envconfig:"GOFORMS_SECURITY_CORS_ORIGINS" default:"http://localhost:5173"`
	AllowedMethods []string `envconfig:"GOFORMS_SECURITY_CORS_METHODS" default:"GET,POST,PUT,DELETE,OPTIONS"`
	//nolint:lll // This is a valid header
	AllowedHeaders   []string `envconfig:"GOFORMS_SECURITY_CORS_HEADERS" default:"Content-Type,Authorization,X-Csrf-Token,X-Requested-With"`
	AllowCredentials bool     `envconfig:"GOFORMS_SECURITY_CORS_CREDENTIALS" default:"true"`
	MaxAge           int      `envconfig:"GOFORMS_SECURITY_CORS_MAX_AGE" default:"3600"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool          `envconfig:"GOFORMS_SECURITY_RATE_LIMIT_ENABLED" default:"true"`
	Requests    int           `envconfig:"GOFORMS_SECURITY_RATE_LIMIT_REQUESTS" default:"100"`
	Window      time.Duration `envconfig:"GOFORMS_SECURITY_RATE_LIMIT_WINDOW" default:"1m"`
	Burst       int           `envconfig:"GOFORMS_SECURITY_RATE_LIMIT_BURST" default:"20"`
	PerIP       bool          `envconfig:"GOFORMS_SECURITY_RATE_LIMIT_PER_IP" default:"true"`
	SkipPaths   []string      `envconfig:"GOFORMS_SECURITY_RATE_LIMIT_SKIP_PATHS"`
	SkipMethods []string      `envconfig:"GOFORMS_SECURITY_RATE_LIMIT_SKIP_METHODS" default:"GET,HEAD,OPTIONS"`
}

// SecurityHeadersConfig holds security headers configuration
type SecurityHeadersConfig struct {
	XFrameOptions           string `envconfig:"GOFORMS_SECURITY_X_FRAME_OPTIONS" default:"DENY"`
	XContentTypeOptions     string `envconfig:"GOFORMS_SECURITY_X_CONTENT_TYPE_OPTIONS" default:"nosniff"`
	XXSSProtection          string `envconfig:"GOFORMS_SECURITY_X_XSS_PROTECTION" default:"1; mode=block"`
	ReferrerPolicy          string `envconfig:"GOFORMS_SECURITY_REFERRER_POLICY" default:"strict-origin-when-cross-origin"`
	StrictTransportSecurity string `envconfig:"GOFORMS_SECURITY_HSTS" default:"max-age=31536000; includeSubDomains"`
}

// CSPConfig holds Content Security Policy configuration
type CSPConfig struct {
	Enabled    bool   `envconfig:"GOFORMS_SECURITY_CSP_ENABLED" default:"true"`
	Directives string `envconfig:"GOFORMS_SECURITY_CSP_DIRECTIVES"`
}

// GetCSPDirectives returns the Content Security Policy directives based on environment
func (s *SecurityConfig) GetCSPDirectives(appConfig *AppConfig) string {
	if appConfig.IsDevelopment() {
		return "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' http://localhost:5173 https://cdn.form.io blob:; " +
			"worker-src 'self' blob:; " +
			"style-src 'self' 'unsafe-inline' http://localhost:5173 https://cdn.form.io; " +
			"img-src 'self' data:; " +
			"font-src 'self' http://localhost:5173; " +
			"connect-src 'self' http://localhost:5173 ws://localhost:5173; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"
	}

	// If custom CSP directives are provided via environment, use them
	if s.CSP.Directives != "" {
		return s.CSP.Directives
	}

	// Generate CSP directives based on environment
	return "default-src 'self'; " +
		"script-src 'self' 'unsafe-inline'; " +
		"style-src 'self' 'unsafe-inline'; " +
		"img-src 'self' data:; " +
		"font-src 'self'; " +
		"connect-src 'self'; " +
		"frame-ancestors 'none'; " +
		"base-uri 'self'; " +
		"form-action 'self'"
}

// validateSecurityConfig validates security configuration
func (c *Config) validateSecurityConfig() error {
	var errs []string

	if c.Security.CSRF.Enabled && c.Security.CSRF.Secret == "" {
		errs = append(errs, "CSRF secret is required when CSRF is enabled")
	}

	if len(errs) > 0 {
		return fmt.Errorf("security config validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}
