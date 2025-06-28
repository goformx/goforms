// Package config provides configuration management for the GoForms application.
// It defines the configuration structures and validation logic used by the Viper-based configuration system.
package config

import (
	"fmt"
	"strings"
)

// Config represents the complete application configuration
type Config struct {
	App      AppConfig      `json:"app"`
	Database DatabaseConfig `json:"database"`
	Security SecurityConfig `json:"security"`
	Email    EmailConfig    `json:"email"`
	Storage  StorageConfig  `json:"storage"`
	Cache    CacheConfig    `json:"cache"`
	Logging  LoggingConfig  `json:"logging"`
	Session  SessionConfig  `json:"session"`
	Auth     AuthConfig     `json:"auth"`
	Form     FormConfig     `json:"form"`
	API      APIConfig      `json:"api"`
	Web      WebConfig      `json:"web"`
	User     UserConfig     `json:"user"`
}

// validateConfig validates the configuration
func (c *Config) validateConfig() error {
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

	// Validate Session config only if session type is not "none"
	if c.Session.Type != "none" && c.Session.Secret == "" {
		errs = append(errs, "session secret is required when session type is not 'none'")
	}

	// Validate Email config only if email host is set
	if c.Email.Host != "" {
		if c.Email.Username == "" {
			errs = append(errs, "Email username is required when email host is set")
		}

		if c.Email.Password == "" {
			errs = append(errs, "Email password is required when email host is set")
		}

		if c.Email.From == "" {
			errs = append(errs, "Email from address is required when email host is set")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// GetConfigSummary returns a summary of the current configuration
func (c *Config) GetConfigSummary() map[string]interface{} {
	return map[string]interface{}{
		"app": map[string]interface{}{
			"name":        c.App.Name,
			"environment": c.App.Environment,
			"debug":       c.App.Debug,
			"url":         c.App.GetServerURL(),
		},
		"database": map[string]interface{}{
			"driver": c.Database.Driver,
			"host":   c.Database.Host,
			"port":   c.Database.Port,
			"name":   c.Database.Name,
		},
		"security": map[string]interface{}{
			"csrf_enabled":       c.Security.CSRF.Enabled,
			"cors_enabled":       c.Security.CORS.Enabled,
			"rate_limit_enabled": c.Security.RateLimit.Enabled,
			"csp_enabled":        c.Security.CSP.Enabled,
		},
		"services": map[string]interface{}{
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
