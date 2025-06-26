// Package config provides configuration management for the GoForms application.
// It uses environment variables to configure various aspects of the application
// including database connections, security settings, logging, and more.
package config

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// Config represents the complete application configuration
type Config struct {
	App      AppConfig      `envconfig:"APP"`
	Database DatabaseConfig `envconfig:"DATABASE"`
	Security SecurityConfig `envconfig:"SECURITY"`
	Email    EmailConfig    `envconfig:"EMAIL"`
	Storage  StorageConfig  `envconfig:"STORAGE"`
	Cache    CacheConfig    `envconfig:"CACHE"`
	Logging  LoggingConfig  `envconfig:"LOGGING"`
	Session  SessionConfig  `envconfig:"SESSION"`
	Auth     AuthConfig     `envconfig:"AUTH"`
	Form     FormConfig     `envconfig:"FORM"`
	API      APIConfig      `envconfig:"API"`
	Web      WebConfig      `envconfig:"WEB"`
	User     UserConfig     `envconfig:"USER"`
}

// New creates a new Config instance
func New() (*Config, error) {
	var config Config

	// Load environment variables
	if err := envconfig.Process("GOFORMS", &config); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Validate required fields
	if err := config.validateConfig(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// validateConfig validates the configuration
func (c *Config) validateConfig() error {
	var errs []string

	// Validate App config
	if err := c.validateAppConfig(); err != nil {
		errs = append(errs, err.Error())
	}

	// Validate Database config
	if err := c.validateDatabaseConfig(); err != nil {
		errs = append(errs, err.Error())
	}

	// Validate Security config
	if err := c.validateSecurityConfig(); err != nil {
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
