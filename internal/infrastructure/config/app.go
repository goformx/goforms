package config

import (
	"fmt"
	"strings"
	"time"
)

// AppConfig holds application-level configuration
type AppConfig struct {
	// Application Info
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Debug       bool   `json:"debug"`
	LogLevel    string `json:"log_level"`

	// Server Settings
	URL            string        `json:"url"`
	Scheme         string        `json:"scheme"`
	Port           int           `json:"port"`
	Host           string        `json:"host"`
	ReadTimeout    time.Duration `json:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout"`
	IdleTimeout    time.Duration `json:"idle_timeout"`
	RequestTimeout time.Duration `json:"request_timeout"`

	// Development Settings
	ViteDevHost string `json:"vite_dev_host"`
	ViteDevPort string `json:"vite_dev_port"`
}

// IsDevelopment returns true if the application is running in development mode
func (c *AppConfig) IsDevelopment() bool {
	return strings.EqualFold(c.Environment, "development")
}

// GetServerURL returns the server URL
func (c *AppConfig) GetServerURL() string {
	return c.URL
}

// GetServerPort returns the server port
func (c *AppConfig) GetServerPort() int {
	return c.Port
}

// Validate validates the application configuration
func (c *AppConfig) Validate() error {
	var errs []string

	if c.Name == "" {
		errs = append(errs, "app name is required")
	}

	if c.GetServerPort() <= 0 || c.GetServerPort() > 65535 {
		errs = append(errs, "app port must be between 1 and 65535")
	}

	if c.ReadTimeout <= 0 {
		errs = append(errs, "read timeout must be positive")
	}

	if c.WriteTimeout <= 0 {
		errs = append(errs, "write timeout must be positive")
	}

	if c.IdleTimeout <= 0 {
		errs = append(errs, "idle timeout must be positive")
	}

	if len(errs) > 0 {
		return fmt.Errorf("app config validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}
