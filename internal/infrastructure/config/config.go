// Package config provides configuration management for the GoForms application.
// It defines the configuration structures and validation logic used by the Viper-based configuration system.
package config

import (
	"fmt"
	"strings"
)

// Config represents the complete application configuration
type Config struct {
	App        AppConfig        `json:"app"`
	Database   DatabaseConfig   `json:"database"`
	Security   SecurityConfig   `json:"security"`
	Email      EmailConfig      `json:"email"`
	Storage    StorageConfig    `json:"storage"`
	Cache      CacheConfig      `json:"cache"`
	Logging    LoggingConfig    `json:"logging"`
	Session    SessionConfig    `json:"session"`
	Auth       AuthConfig       `json:"auth"`
	Form       FormConfig       `json:"form"`
	API        APIConfig        `json:"api"`
	Web        WebConfig        `json:"web"`
	User       UserConfig       `json:"user"`
	Middleware MiddlewareConfig `json:"middleware"`
}

// Ensure Config implements ConfigInterface
var _ ConfigInterface = (*Config)(nil)

// App returns the application configuration
func (c *Config) GetApp() AppConfig {
	return c.App
}

// Database returns the database configuration
func (c *Config) GetDatabase() DatabaseConfig {
	return c.Database
}

// Security returns the security configuration
func (c *Config) GetSecurity() SecurityConfig {
	return c.Security
}

// Email returns the email configuration
func (c *Config) GetEmail() EmailConfig {
	return c.Email
}

// Storage returns the storage configuration
func (c *Config) GetStorage() StorageConfig {
	return c.Storage
}

// Cache returns the cache configuration
func (c *Config) GetCache() CacheConfig {
	return c.Cache
}

// Logging returns the logging configuration
func (c *Config) GetLogging() LoggingConfig {
	return c.Logging
}

// Session returns the session configuration
func (c *Config) GetSession() SessionConfig {
	return c.Session
}

// Auth returns the auth configuration
func (c *Config) GetAuth() AuthConfig {
	return c.Auth
}

// Form returns the form configuration
func (c *Config) GetForm() FormConfig {
	return c.Form
}

// API returns the API configuration
func (c *Config) GetAPI() APIConfig {
	return c.API
}

// Web returns the web configuration
func (c *Config) GetWeb() WebConfig {
	return c.Web
}

// User returns the user configuration
func (c *Config) GetUser() UserConfig {
	return c.User
}

// Middleware returns the middleware configuration
func (c *Config) GetMiddleware() MiddlewareConfig {
	return c.Middleware
}

// validateConfig validates the configuration
func (c *Config) validateConfig() error {
	var errs []string

	// Validate core config sections
	if err := c.validateCoreConfig(); err != nil {
		errs = append(errs, err.Error())
	}

	// Validate conditional config sections
	if err := c.validateConditionalConfig(); err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// validateCoreConfig validates the core configuration sections
func (c *Config) validateCoreConfig() error {
	var errs []string

	// Validate App config
	if err := c.App.Validate(); err != nil {
		errs = append(errs, err.Error())
	}

	// Validate Database config
	if err := c.Database.Validate(); err != nil {
		errs = append(errs, err.Error())
	}

	// Validate Security config
	if err := c.Security.Validate(); err != nil {
		errs = append(errs, err.Error())
	}

	// Validate Middleware config
	if err := c.Middleware.Validate(); err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	return nil
}

// validateConditionalConfig validates configuration sections that depend on other settings
func (c *Config) validateConditionalConfig() error {
	var errs []string

	// Validate Session config only if session type is not "none"
	if err := c.validateSessionConfig(); err != nil {
		errs = append(errs, err.Error())
	}

	// Validate Email config only if email host is set
	if err := c.validateEmailConfig(); err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	return nil
}

// validateSessionConfig validates session configuration
func (c *Config) validateSessionConfig() error {
	if c.Session.Type != "none" && c.Session.Secret == "" {
		return fmt.Errorf("session secret is required when session type is not 'none'")
	}

	return nil
}

// validateEmailConfig validates email configuration
func (c *Config) validateEmailConfig() error {
	if c.Email.Host == "" {
		return nil // Email is optional
	}

	var errs []string

	if c.Email.Username == "" {
		errs = append(errs, "Email username is required when email host is set")
	}

	if c.Email.Password == "" {
		errs = append(errs, "Email password is required when email host is set")
	}

	if c.Email.From == "" {
		errs = append(errs, "Email from address is required when email host is set")
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	return nil
}

// GetConfigSummary returns a summary of the current configuration
func (c *Config) GetConfigSummary() map[string]any {
	return map[string]any{
		"app": map[string]any{
			"name":        c.App.Name,
			"environment": c.App.Environment,
			"debug":       c.App.Debug,
			"url":         c.App.GetServerURL(),
		},
		"database": map[string]any{
			"driver": c.Database.Driver,
			"host":   c.Database.Host,
			"port":   c.Database.Port,
			"name":   c.Database.Name,
		},
		"security": map[string]any{
			"csrf_enabled":       c.Security.CSRF.Enabled,
			"cors_enabled":       c.Security.CORS.Enabled,
			"rate_limit_enabled": c.Security.RateLimit.Enabled,
			"csp_enabled":        c.Security.CSP.Enabled,
		},
		"services": map[string]any{
			"email_configured": c.Email.Host != "",
			"cache_type":       c.Cache.Type,
			"storage_type":     c.Storage.Type,
			"session_type":     c.Session.Type,
		},
	}
}

// IsValid checks if the configuration is valid
func (c *Config) IsValid() bool {
	return c.validateConfig() == nil
}

// GetEnvironment returns the current environment
func (c *Config) GetEnvironment() string {
	return strings.ToLower(c.App.Environment)
}

// IsProduction returns true if running in production
func (c *Config) IsProduction() bool {
	return c.GetEnvironment() == "production"
}

// IsDevelopment returns true if running in development
func (c *Config) IsDevelopment() bool {
	return c.GetEnvironment() == "development"
}

// IsStaging returns true if running in staging
func (c *Config) IsStaging() bool {
	return c.GetEnvironment() == "staging"
}

// IsTest returns true if running in test
func (c *Config) IsTest() bool {
	return c.GetEnvironment() == "test"
}
