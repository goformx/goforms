package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Loader provides configuration loading capabilities
type Loader struct {
	envPrefix string
	envFiles  []string
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{
		envPrefix: "GOFORMS",
		envFiles:  []string{".env", ".env.local"},
	}
}

// WithEnvPrefix sets the environment variable prefix
func (l *Loader) WithEnvPrefix(prefix string) *Loader {
	l.envPrefix = prefix
	return l
}

// WithEnvFiles sets the environment files to load
func (l *Loader) WithEnvFiles(files ...string) *Loader {
	l.envFiles = files
	return l
}

// Load loads the configuration from environment variables and files
func (l *Loader) Load() (*Config, error) {
	// Load environment files if they exist
	if err := l.loadEnvFiles(); err != nil {
		return nil, fmt.Errorf("failed to load environment files: %w", err)
	}

	var config Config

	// Load environment variables into config struct
	if err := envconfig.Process(l.envPrefix, &config); err != nil {
		return nil, fmt.Errorf("failed to process environment variables: %w", err)
	}

	// Validate the configuration
	if err := config.validateConfig(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// loadEnvFiles loads environment variables from .env files
func (l *Loader) loadEnvFiles() error {
	for _, envFile := range l.envFiles {
		if _, err := os.Stat(envFile); os.IsNotExist(err) {
			// File doesn't exist, skip it
			continue
		}

		if err := godotenv.Load(envFile); err != nil {
			return fmt.Errorf("failed to load %s: %w", envFile, err)
		}
	}
	return nil
}

// LoadFromFile loads configuration from a specific file
func LoadFromFile(filename string) (*Config, error) {
	if !filepath.IsAbs(filename) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		filename = filepath.Join(wd, filename)
	}

	loader := NewLoader().WithEnvFiles(filename)
	return loader.Load()
}

// LoadForEnvironment loads configuration for a specific environment
func LoadForEnvironment(env string) (*Config, error) {
	envFiles := []string{
		".env",
		fmt.Sprintf(".env.%s", env),
		".env.local",
		fmt.Sprintf(".env.%s.local", env),
	}

	loader := NewLoader().WithEnvFiles(envFiles...)
	config, err := loader.Load()
	if err != nil {
		return nil, err
	}

	// Override the environment setting
	config.App.Env = env

	return config, nil
}

// MustLoad loads configuration and panics on error
func MustLoad() *Config {
	config, err := New()
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}
	return config
}

// GetConfigSummary returns a summary of the current configuration
func (c *Config) GetConfigSummary() map[string]interface{} {
	return map[string]interface{}{
		"app": map[string]interface{}{
			"name":        c.App.Name,
			"environment": c.App.Env,
			"debug":       c.App.Debug,
			"url":         c.App.GetServerURL(),
		},
		"database": map[string]interface{}{
			"connection": c.Database.Connection,
			"host":       c.Database.Host,
			"port":       c.Database.Port,
			"database":   c.Database.Database,
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
	return strings.ToLower(c.App.Env)
}

// IsProduction returns true if running in production
func (c *Config) IsProduction() bool {
	return c.GetEnvironment() == EnvProduction
}

// IsDevelopment returns true if running in development
func (c *Config) IsDevelopment() bool {
	return c.GetEnvironment() == EnvDevelopment
}

// IsStaging returns true if running in staging
func (c *Config) IsStaging() bool {
	return c.GetEnvironment() == EnvStaging
}

// IsTest returns true if running in test
func (c *Config) IsTest() bool {
	return c.GetEnvironment() == EnvTest
}
