package config

import (
	"fmt"
	"strings"
)

// MiddlewareConfig represents the complete middleware configuration
type MiddlewareConfig struct {
	// Global middleware settings
	Enabled bool `json:"enabled"`

	// Individual middleware configurations
	Recovery        RecoveryMiddlewareConfig        `json:"recovery"`
	CORS            CORSMiddlewareConfig            `json:"cors"`
	RequestID       RequestIDMiddlewareConfig       `json:"request_id"`
	Timeout         TimeoutMiddlewareConfig         `json:"timeout"`
	SecurityHeaders SecurityHeadersMiddlewareConfig `json:"security_headers"`
	CSRF            CSRFMiddlewareConfig            `json:"csrf"`
	RateLimit       RateLimitMiddlewareConfig       `json:"rate_limit"`
	InputValidation InputValidationMiddlewareConfig `json:"input_validation"`
	Logging         LoggingMiddlewareConfig         `json:"logging"`
	Session         SessionMiddlewareConfig         `json:"session"`
	Authentication  AuthenticationMiddlewareConfig  `json:"authentication"`
	Authorization   AuthorizationMiddlewareConfig   `json:"authorization"`

	// Chain configurations
	Chains ChainConfigs `json:"chains"`

	// Global middleware settings
	Global GlobalMiddlewareConfig `json:"global"`
}

// GlobalMiddlewareConfig contains global middleware settings
type GlobalMiddlewareConfig struct {
	// Default middleware to enable
	DefaultEnabled []string `json:"default_enabled"`

	// Environment-specific overrides
	Development []string `json:"development"`
	Production  []string `json:"production"`
	Staging     []string `json:"staging"`
	Test        []string `json:"test"`

	// Performance settings
	CacheEnabled bool `json:"cache_enabled"`
	CacheTTL     int  `json:"cache_ttl"` // seconds
}

// ChainConfigs defines configuration for different middleware chains
type ChainConfigs struct {
	Default ChainConfig `json:"default"`
	API     ChainConfig `json:"api"`
	Web     ChainConfig `json:"web"`
	Auth    ChainConfig `json:"auth"`
	Admin   ChainConfig `json:"admin"`
	Public  ChainConfig `json:"public"`
	Static  ChainConfig `json:"static"`
}

// ChainConfig defines configuration for a specific middleware chain
type ChainConfig struct {
	Enabled         bool           `json:"enabled"`
	MiddlewareNames []string       `json:"middleware_names"`
	Paths           []string       `json:"paths"`
	CustomConfig    map[string]any `json:"custom_config"`
}

// RecoveryMiddlewareConfig defines recovery middleware configuration
type RecoveryMiddlewareConfig struct {
	Enabled bool `json:"enabled"`
}

// CORSMiddlewareConfig defines CORS middleware configuration
type CORSMiddlewareConfig struct {
	Enabled bool `json:"enabled"`
}

// RequestIDMiddlewareConfig defines request ID middleware configuration
type RequestIDMiddlewareConfig struct {
	Enabled bool `json:"enabled"`
}

// TimeoutMiddlewareConfig defines timeout middleware configuration
type TimeoutMiddlewareConfig struct {
	Enabled        bool `json:"enabled"`
	TimeoutSeconds int  `json:"timeout_seconds"`
	GracePeriod    int  `json:"grace_period"`
}

// SecurityHeadersMiddlewareConfig defines security headers middleware configuration
type SecurityHeadersMiddlewareConfig struct {
	Enabled bool `json:"enabled"`
}

// CSRFMiddlewareConfig defines CSRF middleware configuration
type CSRFMiddlewareConfig struct {
	Enabled      bool     `json:"enabled"`
	TokenHeader  string   `json:"token_header"`
	CookieName   string   `json:"cookie_name"`
	ExpireTime   int      `json:"expire_time"` // seconds
	IncludePaths []string `json:"include_paths"`
	ExcludePaths []string `json:"exclude_paths"`
}

// RateLimitMiddlewareConfig defines rate limiting middleware configuration
type RateLimitMiddlewareConfig struct {
	Enabled           bool     `json:"enabled"`
	RequestsPerMinute int      `json:"requests_per_minute"`
	BurstSize         int      `json:"burst_size"`
	WindowSize        int      `json:"window_size"` // seconds
	IncludePaths      []string `json:"include_paths"`
	ExcludePaths      []string `json:"exclude_paths"`
}

// InputValidationMiddlewareConfig defines input validation middleware configuration
type InputValidationMiddlewareConfig struct {
	Enabled bool `json:"enabled"`
}

// LoggingMiddlewareConfig defines logging middleware configuration
type LoggingMiddlewareConfig struct {
	Enabled      bool     `json:"enabled"`
	LogLevel     string   `json:"log_level"`
	IncludeBody  bool     `json:"include_body"`
	MaskHeaders  []string `json:"mask_headers"`
	LogRequests  bool     `json:"log_requests"`
	LogResponses bool     `json:"log_responses"`
	IncludePaths []string `json:"include_paths"`
	ExcludePaths []string `json:"exclude_paths"`
}

// SessionMiddlewareConfig defines session middleware configuration
type SessionMiddlewareConfig struct {
	Enabled        bool `json:"enabled"`
	SessionTimeout int  `json:"session_timeout"` // seconds
	RefreshTimeout int  `json:"refresh_timeout"` // seconds
	SecureCookies  bool `json:"secure_cookies"`
	HTTPOnly       bool `json:"http_only"`
}

// AuthenticationMiddlewareConfig defines authentication middleware configuration
type AuthenticationMiddlewareConfig struct {
	Enabled       bool `json:"enabled"`
	TokenExpiry   int  `json:"token_expiry"`   // seconds
	RefreshExpiry int  `json:"refresh_expiry"` // seconds
}

// AuthorizationMiddlewareConfig defines authorization middleware configuration
type AuthorizationMiddlewareConfig struct {
	Enabled     bool   `json:"enabled"`
	DefaultRole string `json:"default_role"`
	AdminRole   string `json:"admin_role"`
	CacheTTL    int    `json:"cache_ttl"` // seconds
}

// Validate validates the middleware configuration
func (c *MiddlewareConfig) Validate() error {
	var errs []string

	// Validate chain configurations
	if err := c.Chains.Validate(); err != nil {
		errs = append(errs, err.Error())
	}

	// Validate global configuration
	if err := c.Global.Validate(); err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("middleware config validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// Validate validates chain configurations
func (c *ChainConfigs) Validate() error {
	var errs []string

	// Validate each chain configuration
	chains := map[string]ChainConfig{
		"default": c.Default,
		"api":     c.API,
		"web":     c.Web,
		"auth":    c.Auth,
		"admin":   c.Admin,
		"public":  c.Public,
		"static":  c.Static,
	}

	for name, chain := range chains {
		if err := chain.Validate(); err != nil {
			errs = append(errs, fmt.Sprintf("%s chain: %s", name, err.Error()))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("chain validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// Validate validates a single chain configuration
func (c *ChainConfig) Validate() error {
	if !c.Enabled {
		return nil // Skip validation for disabled chains
	}

	var errs []string

	// Validate middleware names are not empty
	for i, name := range c.MiddlewareNames {
		if strings.TrimSpace(name) == "" {
			errs = append(errs, fmt.Sprintf("middleware name at index %d is empty", i))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("chain config validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// Validate validates global middleware configuration
func (c *GlobalMiddlewareConfig) Validate() error {
	var errs []string

	// Validate cache TTL is positive if cache is enabled
	if c.CacheEnabled && c.CacheTTL <= 0 {
		errs = append(errs, "cache_ttl must be positive when cache is enabled")
	}

	if len(errs) > 0 {
		return fmt.Errorf("global middleware config validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}
