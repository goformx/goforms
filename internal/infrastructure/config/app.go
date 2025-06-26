package config

import (
	"fmt"
	"strings"
	"time"
)

// AppConfig holds application-level configuration
type AppConfig struct {
	// Application Info
	Name     string `envconfig:"GOFORMS_APP_NAME" default:"GoFormX"`
	Env      string `envconfig:"GOFORMS_APP_ENV" default:"production"`
	Debug    bool   `envconfig:"GOFORMS_APP_DEBUG" default:"false"`
	LogLevel string `envconfig:"GOFORMS_APP_LOGLEVEL" default:"info"`

	// Server Settings
	URL            string        `envconfig:"GOFORMS_APP_URL" default:"http://localhost:8090"`
	Scheme         string        `envconfig:"GOFORMS_APP_SCHEME" default:"http"`
	Port           int           `envconfig:"GOFORMS_APP_PORT" default:"8090"`
	Host           string        `envconfig:"GOFORMS_APP_HOST" default:"0.0.0.0"`
	ReadTimeout    time.Duration `envconfig:"GOFORMS_APP_READ_TIMEOUT" default:"5s"`
	WriteTimeout   time.Duration `envconfig:"GOFORMS_APP_WRITE_TIMEOUT" default:"10s"`
	IdleTimeout    time.Duration `envconfig:"GOFORMS_APP_IDLE_TIMEOUT" default:"120s"`
	RequestTimeout time.Duration `envconfig:"GOFORMS_APP_REQUEST_TIMEOUT" default:"30s"`

	// Development Settings
	ViteDevHost string `envconfig:"GOFORMS_VITE_DEV_HOST" default:"localhost"`
	ViteDevPort string `envconfig:"GOFORMS_VITE_DEV_PORT" default:"5173"`
}

// IsDevelopment returns true if the application is running in development mode
func (c *AppConfig) IsDevelopment() bool {
	return strings.EqualFold(c.Env, "development")
}

// GetServerURL returns the server URL
func (c *AppConfig) GetServerURL() string {
	return c.URL
}

// GetServerPort returns the server port
func (c *AppConfig) GetServerPort() int {
	return c.Port
}

// validateAppConfig validates the application configuration
func (c *Config) validateAppConfig() error {
	var errs []string

	if c.App.Name == "" {
		errs = append(errs, "app name is required")
	}
	if c.App.GetServerPort() <= 0 || c.App.GetServerPort() > 65535 {
		errs = append(errs, "app port must be between 1 and 65535")
	}
	if c.App.ReadTimeout <= 0 {
		errs = append(errs, "read timeout must be positive")
	}
	if c.App.WriteTimeout <= 0 {
		errs = append(errs, "write timeout must be positive")
	}
	if c.App.IdleTimeout <= 0 {
		errs = append(errs, "idle timeout must be positive")
	}

	if len(errs) > 0 {
		return fmt.Errorf("app config validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}
