package config

import (
	"fmt"
	"strings"
	"time"
)

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	// JWT configuration
	JWTSecret string `json:"jwt_secret"`

	// CSRF protection
	CSRF CSRFConfig `json:"csrf"`

	// CORS configuration
	CORS CORSConfig `json:"cors"`

	// Rate limiting configuration
	RateLimit RateLimitConfig `json:"rate_limit"`

	// Security headers configuration
	Headers SecurityHeadersConfig `json:"headers"`

	// Content Security Policy configuration
	CSP CSPConfig `json:"csp"`

	// TLS configuration
	TLS TLSConfig `json:"tls"`

	// Encryption configuration
	Encryption EncryptionConfig `json:"encryption"`

	// Cookie security
	SecureCookie bool `json:"secure_cookie"`

	// Debug mode
	Debug bool `json:"debug"`
}

// CSRFConfig holds CSRF-related configuration
type CSRFConfig struct {
	Enabled        bool   `json:"enabled"`
	Secret         string `json:"secret"`
	TokenName      string `json:"token_name"`
	HeaderName     string `json:"header_name"`
	TokenLength    int    `json:"token_length"`
	TokenLookup    string `json:"token_lookup"`
	ContextKey     string `json:"context_key"`
	CookieName     string `json:"cookie_name"`
	CookiePath     string `json:"cookie_path"`
	CookieDomain   string `json:"cookie_domain"`
	CookieHTTPOnly bool   `json:"cookie_http_only"`
	CookieSameSite string `json:"cookie_same_site"`
	CookieMaxAge   int    `json:"cookie_max_age"`
}

// CORSConfig holds CORS-related configuration
type CORSConfig struct {
	Enabled          bool     `json:"enabled"`
	AllowedOrigins   []string `json:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool          `json:"enabled"`
	RPS         int           `json:"rps"`
	Burst       int           `json:"burst"`
	Requests    int           `json:"requests"`
	Window      time.Duration `json:"window"`
	PerIP       bool          `json:"per_ip"`
	SkipPaths   []string      `json:"skip_paths"`
	SkipMethods []string      `json:"skip_methods"`
}

// SecurityHeadersConfig holds security headers configuration
type SecurityHeadersConfig struct {
	XFrameOptions           string `json:"x_frame_options"`
	XContentTypeOptions     string `json:"x_content_type_options"`
	XXSSProtection          string `json:"x_xss_protection"`
	ReferrerPolicy          string `json:"referrer_policy"`
	StrictTransportSecurity string `json:"strict_transport_security"`
}

// CSPConfig holds Content Security Policy configuration
type CSPConfig struct {
	Enabled    bool   `json:"enabled"`
	Directives string `json:"directives"`
	DefaultSrc string `json:"default_src"`
	ScriptSrc  string `json:"script_src"`
	StyleSrc   string `json:"style_src"`
	ImgSrc     string `json:"img_src"`
	ConnectSrc string `json:"connect_src"`
	FontSrc    string `json:"font_src"`
	ObjectSrc  string `json:"object_src"`
	MediaSrc   string `json:"media_src"`
	FrameSrc   string `json:"frame_src"`
	ReportURI  string `json:"report_uri"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	Key string `json:"key"`
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

// Validate validates the security configuration
func (c *SecurityConfig) Validate() error {
	var errs []string

	if c.JWTSecret == "" {
		errs = append(errs, "JWT secret is required")
	}

	if c.CSRF.Enabled && c.CSRF.Secret == "" {
		errs = append(errs, "CSRF secret is required when CSRF is enabled")
	}

	if len(errs) > 0 {
		return fmt.Errorf("security config validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}
