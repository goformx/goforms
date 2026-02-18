// Package config provides enhanced security configuration with modern best practices
// This replaces/enhances your existing internal/infrastructure/config/security.go
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// MinAPIKeyLength is the minimum required length for API keys
const MinAPIKeyLength = 16

// SecurityConfig represents the enhanced security configuration
type SecurityConfig struct {
	CSRF            CSRFConfig            `json:"csrf"`
	CORS            CORSConfig            `json:"cors"`
	RateLimit       RateLimitConfig       `json:"rate_limit"`
	CSP             CSPConfig             `json:"csp"`
	TLS             TLSConfig             `json:"tls"`
	Encryption      EncryptionConfig      `json:"encryption"`
	SecurityHeaders SecurityHeadersConfig `json:"security_headers"`
	CookieSecurity  CookieSecurityConfig  `json:"cookie_security"`
	TrustProxy      TrustProxyConfig      `json:"trust_proxy"`
	Assertion       AssertionConfig       `json:"assertion"`
	APIKey          APIKeyConfig          `json:"api_key"`
	SecureCookie    bool                  `json:"secure_cookie"`
	Debug           bool                  `json:"debug"`
}

// CSRFConfig represents enhanced CSRF configuration
type CSRFConfig struct {
	Enabled        bool     `json:"enabled"`
	Secret         string   `json:"secret"`
	TokenName      string   `json:"token_name"`
	HeaderName     string   `json:"header_name"`
	TokenLength    int      `json:"token_length"`
	TokenLookup    string   `json:"token_lookup"`
	ContextKey     string   `json:"context_key"`
	CookieName     string   `json:"cookie_name"`
	CookiePath     string   `json:"cookie_path"`
	CookieDomain   string   `json:"cookie_domain"`
	CookieHTTPOnly bool     `json:"cookie_http_only"`
	CookieSameSite string   `json:"cookie_same_site"`
	CookieMaxAge   int      `json:"cookie_max_age"`
	CookieSecure   bool     `json:"cookie_secure"`
	ErrorHandler   string   `json:"error_handler"`
	SkipPaths      []string `json:"skip_paths"`
}

// CORSConfig represents enhanced CORS configuration
type CORSConfig struct {
	Enabled             bool     `json:"enabled"`
	AllowedOrigins      []string `json:"allowed_origins"`
	AllowedMethods      []string `json:"allowed_methods"`
	AllowedHeaders      []string `json:"allowed_headers"`
	ExposedHeaders      []string `json:"exposed_headers"`
	AllowCredentials    bool     `json:"allow_credentials"`
	MaxAge              int      `json:"max_age"`
	AllowOriginPatterns []string `json:"allow_origin_patterns"`
	AllowWildcardOrigin bool     `json:"allow_wildcard_origin"`
	OptionStatusCode    int      `json:"option_status_code"`
}

// RateLimitConfig represents enhanced rate limiting configuration
type RateLimitConfig struct {
	Enabled        bool                     `json:"enabled"`
	RPS            int                      `json:"rps"`
	Requests       int                      `json:"requests"` // Alias for RPS
	Burst          int                      `json:"burst"`
	Window         time.Duration            `json:"window"`
	PerIP          bool                     `json:"per_ip"`
	SkipPaths      []string                 `json:"skip_paths"`
	SkipMethods    []string                 `json:"skip_methods"`
	EndpointLimits map[string]EndpointLimit `json:"endpoint_limits"`
	Store          string                   `json:"store"` // memory, redis
	KeyGenerator   string                   `json:"key_generator"`
}

// EndpointLimit represents specific rate limits for endpoints
type EndpointLimit struct {
	RPS    int           `json:"rps"`
	Burst  int           `json:"burst"`
	Window time.Duration `json:"window"`
}

// CSPConfig represents enhanced Content Security Policy configuration
type CSPConfig struct {
	Enabled     bool   `json:"enabled"`
	DefaultSrc  string `json:"default_src"`
	ScriptSrc   string `json:"script_src"`
	StyleSrc    string `json:"style_src"`
	ImgSrc      string `json:"img_src"`
	ConnectSrc  string `json:"connect_src"`
	FontSrc     string `json:"font_src"`
	ObjectSrc   string `json:"object_src"`
	MediaSrc    string `json:"media_src"`
	FrameSrc    string `json:"frame_src"`
	FormAction  string `json:"form_action"`
	BaseURI     string `json:"base_uri"`
	ManifestSrc string `json:"manifest_src"`
	WorkerSrc   string `json:"worker_src"`
	ReportURI   string `json:"report_uri"`
	ReportOnly  bool   `json:"report_only"`
}

// TLSConfig represents enhanced TLS configuration
type TLSConfig struct {
	Enabled      bool     `json:"enabled"`
	CertFile     string   `json:"cert_file"`
	KeyFile      string   `json:"key_file"`
	MinVersion   string   `json:"min_version"`
	CipherSuites []string `json:"cipher_suites"`
	AutoCert     bool     `json:"auto_cert"`
	AutoCertHost string   `json:"auto_cert_host"`
}

// SecurityHeadersConfig represents security headers configuration
type SecurityHeadersConfig struct {
	Enabled                 bool   `json:"enabled"`
	XFrameOptions           string `json:"x_frame_options"`
	XContentTypeOptions     string `json:"x_content_type_options"`
	XXSSProtection          string `json:"xxss_protection"`
	ReferrerPolicy          string `json:"referrer_policy"`
	PermissionsPolicy       string `json:"permissions_policy"`
	StrictTransportSecurity string `json:"strict_transport_security"`
	ContentTypeNoSniff      bool   `json:"content_type_no_sniff"`
}

// CookieSecurityConfig represents default cookie security settings
type CookieSecurityConfig struct {
	Secure   bool   `json:"secure"`
	HTTPOnly bool   `json:"http_only"`
	SameSite string `json:"same_site"`
	Path     string `json:"path"`
	Domain   string `json:"domain"`
	MaxAge   int    `json:"max_age"`
}

// TrustProxyConfig represents proxy trust configuration
type TrustProxyConfig struct {
	Enabled        bool     `json:"enabled"`
	TrustedProxies []string `json:"trusted_proxies"`
	TrustedHeaders []string `json:"trusted_headers"`
}

// EncryptionConfig represents encryption configuration
type EncryptionConfig struct {
	Key            string `json:"key"`
	Algorithm      string `json:"algorithm"`
	KeySize        int    `json:"key_size"`
	SaltLength     int    `json:"salt_length"`
	Iterations     int    `json:"iterations"`
	EnableAES      bool   `json:"enable_aes"`
	EnableChaCha20 bool   `json:"enable_cha_cha20"`
}

// AssertionConfig represents Laravel signed assertion verification configuration
type AssertionConfig struct {
	Secret               string `json:"secret"`
	TimestampSkewSeconds int    `json:"timestamp_skew_seconds"`
}

// APIKeyConfig represents API key authentication configuration
type APIKeyConfig struct {
	Enabled     bool     `json:"enabled"`
	Keys        []string `json:"keys"`         // List of valid API keys
	HeaderName  string   `json:"header_name"`  // Header name to read API key from (default: X-API-Key)
	QueryParam  string   `json:"query_param"`  // Query parameter name (optional)
	SkipPaths   []string `json:"skip_paths"`   // Paths to skip API key validation
	SkipMethods []string `json:"skip_methods"` // HTTP methods to skip (e.g., OPTIONS)
}

// Validate validates the security configuration
func (s *SecurityConfig) Validate() error {
	var errs []string

	// Validate CSRF configuration
	if s.CSRF.Enabled {
		if err := s.validateCSRF(); err != nil {
			errs = append(errs, fmt.Sprintf("CSRF: %v", err))
		}
	}

	// Validate CORS configuration
	if s.CORS.Enabled {
		if err := s.validateCORS(); err != nil {
			errs = append(errs, fmt.Sprintf("CORS: %v", err))
		}
	}

	// Validate TLS configuration
	if s.TLS.Enabled {
		if err := s.validateTLS(); err != nil {
			errs = append(errs, fmt.Sprintf("TLS: %v", err))
		}
	}

	// Validate cookie security
	if err := s.validateCookieSecurity(); err != nil {
		errs = append(errs, fmt.Sprintf("Cookie Security: %v", err))
	}

	// Validate API key configuration
	if s.APIKey.Enabled {
		if err := s.validateAPIKey(); err != nil {
			errs = append(errs, fmt.Sprintf("API Key: %v", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("security validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// validateCSRF validates CSRF configuration
func (s *SecurityConfig) validateCSRF() error {
	if s.CSRF.Secret == "" {
		return errors.New("CSRF secret is required")
	}

	if len(s.CSRF.Secret) < MinSecretLength {
		return errors.New("CSRF secret must be at least 32 characters")
	}

	if s.CSRF.TokenLength < 16 || s.CSRF.TokenLength > 64 {
		return errors.New("CSRF token length must be between 16 and 64")
	}

	// Validate SameSite values
	validSameSite := []string{"Strict", "Lax", "None"}
	if !contains(validSameSite, s.CSRF.CookieSameSite) {
		return fmt.Errorf("invalid CSRF cookie SameSite value: %s", s.CSRF.CookieSameSite)
	}

	// If SameSite=None, Secure must be true (in production)
	if s.CSRF.CookieSameSite == "None" && !s.CSRF.CookieSecure {
		return errors.New("CSRF cookie with SameSite=None requires Secure=true")
	}

	return nil
}

// validateCORS validates CORS configuration
func (s *SecurityConfig) validateCORS() error {
	// Critical security check: prevent wildcard with credentials
	if s.CORS.AllowCredentials {
		for _, origin := range s.CORS.AllowedOrigins {
			if origin == "*" {
				return errors.New("CORS wildcard origin '*' cannot be used with AllowCredentials=true")
			}
		}
	}

	// Validate origins format
	for _, origin := range s.CORS.AllowedOrigins {
		if origin != "*" && !isValidOrigin(origin) {
			return fmt.Errorf("invalid CORS origin format: %s", origin)
		}
	}

	return nil
}

// validateTLS validates TLS configuration
func (s *SecurityConfig) validateTLS() error {
	if s.TLS.CertFile == "" || s.TLS.KeyFile == "" {
		if !s.TLS.AutoCert {
			return errors.New("TLS cert and key files are required when AutoCert is disabled")
		}
	}

	// Validate minimum TLS version
	validVersions := []string{"1.0", "1.1", "1.2", "1.3"}
	if !contains(validVersions, s.TLS.MinVersion) {
		return fmt.Errorf("invalid TLS minimum version: %s", s.TLS.MinVersion)
	}

	return nil
}

// validateCookieSecurity validates cookie security settings
func (s *SecurityConfig) validateCookieSecurity() error {
	validSameSite := []string{"Strict", "Lax", "None"}
	if !contains(validSameSite, s.CookieSecurity.SameSite) {
		return fmt.Errorf("invalid cookie SameSite value: %s", s.CookieSecurity.SameSite)
	}

	// If SameSite=None, Secure must be true
	if s.CookieSecurity.SameSite == "None" && !s.CookieSecurity.Secure {
		return errors.New("cookie with SameSite=None requires Secure=true")
	}

	return nil
}

// validateAPIKey validates API key configuration
func (s *SecurityConfig) validateAPIKey() error {
	if len(s.APIKey.Keys) == 0 {
		return errors.New("at least one API key must be configured when API key authentication is enabled")
	}

	// Validate that each API key is not empty and has minimum length
	for i, key := range s.APIKey.Keys {
		if key == "" {
			return fmt.Errorf("API key at index %d is empty", i)
		}
		if len(key) < MinAPIKeyLength {
			return fmt.Errorf("API key at index %d must be at least %d characters long", i, MinAPIKeyLength)
		}
	}

	// Validate header name
	if s.APIKey.HeaderName == "" {
		return errors.New("API key header name cannot be empty")
	}

	return nil
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}

func isValidOrigin(origin string) bool {
	// Basic origin validation - implement more comprehensive validation as needed
	return strings.HasPrefix(origin, "http://") || strings.HasPrefix(origin, "https://")
}

// GetCSRFSkipper returns a CSRF skipper function based on configuration
func (s *SecurityConfig) GetCSRFSkipper() func(c echo.Context) bool {
	if len(s.CSRF.SkipPaths) == 0 {
		return nil
	}

	return func(c echo.Context) bool {
		path := c.Request().URL.Path
		for _, skipPath := range s.CSRF.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				return true
			}
		}

		return false
	}
}

// GetRateLimitSkipper returns a rate limit skipper function based on configuration
func (s *SecurityConfig) GetRateLimitSkipper() func(c echo.Context) bool {
	if len(s.RateLimit.SkipPaths) == 0 && len(s.RateLimit.SkipMethods) == 0 {
		return nil
	}

	return func(c echo.Context) bool {
		path := c.Request().URL.Path
		method := c.Request().Method

		// Check skip paths
		for _, skipPath := range s.RateLimit.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				return true
			}
		}

		// Check skip methods
		for _, skipMethod := range s.RateLimit.SkipMethods {
			if method == skipMethod {
				return true
			}
		}

		return false
	}
}

// IsProduction returns true if the application should use production security settings
func (s *SecurityConfig) IsProduction() bool {
	return !s.Debug
}

// ShouldUseSecureCookies returns true if cookies should be marked as secure
func (s *SecurityConfig) ShouldUseSecureCookies() bool {
	return s.TLS.Enabled || s.CookieSecurity.Secure
}

// buildCSPPolicies builds CSP policies from the configuration
func (s *CSPConfig) buildCSPPolicies() []string {
	var policies []string

	cspDirectives := map[string]string{
		"default-src":  s.DefaultSrc,
		"script-src":   s.ScriptSrc,
		"style-src":    s.StyleSrc,
		"img-src":      s.ImgSrc,
		"connect-src":  s.ConnectSrc,
		"font-src":     s.FontSrc,
		"object-src":   s.ObjectSrc,
		"media-src":    s.MediaSrc,
		"frame-src":    s.FrameSrc,
		"form-action":  s.FormAction,
		"base-uri":     s.BaseURI,
		"manifest-src": s.ManifestSrc,
		"worker-src":   s.WorkerSrc,
		"report-uri":   s.ReportURI,
	}

	for directive, value := range cspDirectives {
		if value != "" {
			policies = append(policies, fmt.Sprintf("%s %s", directive, value))
		}
	}

	return policies
}

// GetCSPHeaderValue returns the complete CSP header value
func (s *CSPConfig) GetCSPHeaderValue() string {
	if !s.Enabled {
		return ""
	}

	policies := s.buildCSPPolicies()

	return strings.Join(policies, "; ")
}

// GetCSPHeaderName returns the appropriate CSP header name
func (s *CSPConfig) GetCSPHeaderName() string {
	if s.ReportOnly {
		return "Content-Security-Policy-Report-Only"
	}

	return "Content-Security-Policy"
}

// buildCSPDirectives builds CSP directives from the configuration
func (s *SecurityConfig) buildCSPDirectives() []string {
	var directives []string

	cspDirectives := map[string]string{
		"default-src":  s.CSP.DefaultSrc,
		"script-src":   s.CSP.ScriptSrc,
		"style-src":    s.CSP.StyleSrc,
		"img-src":      s.CSP.ImgSrc,
		"connect-src":  s.CSP.ConnectSrc,
		"font-src":     s.CSP.FontSrc,
		"object-src":   s.CSP.ObjectSrc,
		"media-src":    s.CSP.MediaSrc,
		"frame-src":    s.CSP.FrameSrc,
		"form-action":  s.CSP.FormAction,
		"base-uri":     s.CSP.BaseURI,
		"manifest-src": s.CSP.ManifestSrc,
		"worker-src":   s.CSP.WorkerSrc,
		"report-uri":   s.CSP.ReportURI,
	}

	for directive, value := range cspDirectives {
		if value != "" {
			directives = append(directives, fmt.Sprintf("%s %s", directive, value))
		}
	}

	return directives
}

// GetCSPDirectives returns the Content Security Policy directives as a string
func (s *SecurityConfig) GetCSPDirectives(appConfig *AppConfig) string {
	if !s.CSP.Enabled {
		return ""
	}

	directives := s.buildCSPDirectives()

	return strings.Join(directives, "; ")
}
