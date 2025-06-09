package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents the complete application configuration
type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Security SecurityConfig
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name        string `envconfig:"GOFORMS_APP_NAME" default:"GoFormX"`
	Version     string `envconfig:"GOFORMS_APP_VERSION" default:"1.0.0"`
	Env         string `envconfig:"GOFORMS_APP_ENV" default:"production"`
	Debug       bool   `envconfig:"GOFORMS_APP_DEBUG" default:"false"`
	LogLevel    string `envconfig:"GOFORMS_APP_LOGLEVEL" default:"info"`
	Port        int    `envconfig:"GOFORMS_APP_PORT" default:"8090"`
	Host        string `envconfig:"GOFORMS_APP_HOST" default:"0.0.0.0"`
	ViteDevHost string `envconfig:"GOFORMS_VITE_DEV_HOST" default:"localhost"`
	ViteDevPort string `envconfig:"GOFORMS_VITE_DEV_PORT" default:"3000"`
}

// IsDevelopment returns true if the application is running in development mode
func (c *AppConfig) IsDevelopment() bool {
	return strings.EqualFold(c.Env, "development")
}

// DatabaseConfig holds all database-related configuration
type DatabaseConfig struct {
	// MariaDB Configuration
	MariaDB struct {
		Host            string        `envconfig:"GOFORMS_MARIADB_HOST" default:"mariadb"`
		Port            int           `envconfig:"GOFORMS_MARIADB_PORT" default:"3306"`
		User            string        `envconfig:"GOFORMS_MARIADB_USER" default:"goforms"`
		Password        string        `envconfig:"GOFORMS_MARIADB_PASSWORD" default:"goforms"`
		Name            string        `envconfig:"GOFORMS_MARIADB_NAME" default:"goforms"`
		MaxOpenConns    int           `envconfig:"GOFORMS_MARIADB_MAX_OPEN_CONNS" default:"25"`
		MaxIdleConns    int           `envconfig:"GOFORMS_MARIADB_MAX_IDLE_CONNS" default:"5"`
		ConnMaxLifetime time.Duration `envconfig:"GOFORMS_MARIADB_CONN_MAX_LIFETIME" default:"5m"`
	} `envconfig:"GOFORMS_MARIADB"`

	// PostgreSQL Configuration
	Postgres struct {
		Host            string        `envconfig:"GOFORMS_POSTGRES_HOST" default:"postgres"`
		Port            int           `envconfig:"GOFORMS_POSTGRES_PORT" default:"5432"`
		User            string        `envconfig:"GOFORMS_POSTGRES_USER" default:"goforms"`
		Password        string        `envconfig:"GOFORMS_POSTGRES_PASSWORD" default:"goforms"`
		Name            string        `envconfig:"GOFORMS_POSTGRES_DB" default:"goforms"`
		SSLMode         string        `envconfig:"GOFORMS_POSTGRES_SSLMODE" default:"disable"`
		MaxOpenConns    int           `envconfig:"GOFORMS_POSTGRES_MAX_OPEN_CONNS" default:"25"`
		MaxIdleConns    int           `envconfig:"GOFORMS_POSTGRES_MAX_IDLE_CONNS" default:"5"`
		ConnMaxLifetime time.Duration `envconfig:"GOFORMS_POSTGRES_CONN_MAX_LIFETIME" default:"5m"`
	} `envconfig:"GOFORMS_POSTGRES"`

	// Active database driver
	Driver string `envconfig:"GOFORMS_DB_DRIVER" default:"mariadb"`
}

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Host            string        `envconfig:"GOFORMS_APP_HOST" default:"0.0.0.0"`
	Port            int           `envconfig:"GOFORMS_APP_PORT" default:"8090"`
	ReadTimeout     time.Duration `envconfig:"GOFORMS_READ_TIMEOUT" default:"5s"`
	WriteTimeout    time.Duration `envconfig:"GOFORMS_WRITE_TIMEOUT" default:"10s"`
	IdleTimeout     time.Duration `envconfig:"GOFORMS_IDLE_TIMEOUT" default:"120s"`
	ShutdownTimeout time.Duration `envconfig:"GOFORMS_SHUTDOWN_TIMEOUT" default:"30s"`
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	Debug               bool          `envconfig:"GOFORMS_DEBUG" default:"true"`
	FormRateLimit       float64       `envconfig:"GOFORMS_FORM_RATE_LIMIT" default:"20"`
	FormRateLimitWindow time.Duration `envconfig:"GOFORMS_FORM_RATE_LIMIT_WINDOW" default:"1s"`
	SecureCookie        bool          `envconfig:"GOFORMS_SECURE_COOKIE" default:"true"`

	// CORS settings
	CorsAllowedOrigins   []string `envconfig:"GOFORMS_CORS_ALLOWED_ORIGINS" default:"http://localhost:3000"`
	CorsAllowedMethods   []string `envconfig:"GOFORMS_CORS_ALLOWED_METHODS" default:"GET,POST,PUT,DELETE,OPTIONS"`
	CorsAllowedHeaders   []string `envconfig:"GOFORMS_CORS_ALLOWED_HEADERS" default:"Content-Type,Authorization"`
	CorsAllowCredentials bool     `envconfig:"GOFORMS_CORS_ALLOW_CREDENTIALS" default:"true"`
	CorsMaxAge           int      `envconfig:"GOFORMS_CORS_MAX_AGE" default:"3600"`

	// Form-specific CORS settings
	FormCorsAllowedOrigins []string `envconfig:"GOFORMS_FORM_CORS_ALLOWED_ORIGINS" default:"*"`
	FormCorsAllowedMethods []string `envconfig:"GOFORMS_FORM_CORS_ALLOWED_METHODS" default:"GET,POST,OPTIONS"`
	FormCorsAllowedHeaders []string `envconfig:"GOFORMS_FORM_CORS_ALLOWED_HEADERS" default:"Content-Type"`

	// CSRF settings
	CSRFConfig struct {
		Enabled bool   `envconfig:"GOFORMS_CSRF_ENABLED" default:"true"`
		Secret  string `envconfig:"GOFORMS_CSRF_SECRET" validate:"required"`
	} `envconfig:"GOFORMS_CSRF"`
}

// Validation errors
var (
	ErrMissingAppName    = errors.New("application name is required")
	ErrInvalidPort       = errors.New("port must be between 1 and 65535")
	ErrMissingDBDriver   = errors.New("database driver is required")
	ErrMissingDBHost     = errors.New("database host is required")
	ErrMissingDBUser     = errors.New("database user is required")
	ErrMissingDBPassword = errors.New("database password is required")
	ErrMissingDBName     = errors.New("database name is required")
	ErrMissingCSRFSecret = errors.New("CSRF secret is required when CSRF is enabled")
	ErrInvalidTimeout    = errors.New("timeout duration must be positive")
	ErrInvalidRateLimit  = errors.New("rate limit must be positive")
	ErrInvalidMaxConns   = errors.New("max connections must be positive")
)

// validateAppConfig validates application-level configuration
func (c *Config) validateAppConfig() error {
	if c.App.Name == "" {
		return ErrMissingAppName
	}

	if c.App.Port <= 0 || c.App.Port > 65535 {
		return ErrInvalidPort
	}

	return nil
}

// validateServerConfig validates server-related configuration
func (c *Config) validateServerConfig() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return ErrInvalidPort
	}

	if c.Server.ReadTimeout <= 0 {
		return ErrInvalidTimeout
	}

	if c.Server.WriteTimeout <= 0 {
		return ErrInvalidTimeout
	}

	if c.Server.IdleTimeout <= 0 {
		return ErrInvalidTimeout
	}

	if c.Server.ShutdownTimeout <= 0 {
		return ErrInvalidTimeout
	}

	return nil
}

// validateMariaDBConfig validates MariaDB-specific configuration
func (c *Config) validateMariaDBConfig() error {
	if c.Database.MariaDB.Host == "" {
		return ErrMissingDBHost
	}
	if c.Database.MariaDB.User == "" {
		return ErrMissingDBUser
	}
	if c.Database.MariaDB.Password == "" {
		return ErrMissingDBPassword
	}
	if c.Database.MariaDB.Name == "" {
		return ErrMissingDBName
	}
	if c.Database.MariaDB.MaxOpenConns <= 0 {
		return ErrInvalidMaxConns
	}
	if c.Database.MariaDB.MaxIdleConns <= 0 {
		return ErrInvalidMaxConns
	}
	if c.Database.MariaDB.ConnMaxLifetime <= 0 {
		return ErrInvalidTimeout
	}
	return nil
}

// validatePostgresConfig validates PostgreSQL-specific configuration
func (c *Config) validatePostgresConfig() error {
	if c.Database.Postgres.Host == "" {
		return ErrMissingDBHost
	}
	if c.Database.Postgres.User == "" {
		return ErrMissingDBUser
	}
	if c.Database.Postgres.Password == "" {
		return ErrMissingDBPassword
	}
	if c.Database.Postgres.Name == "" {
		return ErrMissingDBName
	}
	if c.Database.Postgres.MaxOpenConns <= 0 {
		return ErrInvalidMaxConns
	}
	if c.Database.Postgres.MaxIdleConns <= 0 {
		return ErrInvalidMaxConns
	}
	if c.Database.Postgres.ConnMaxLifetime <= 0 {
		return ErrInvalidTimeout
	}
	return nil
}

// validateDatabaseConfig validates database-related configuration
func (c *Config) validateDatabaseConfig() error {
	if c.Database.Driver == "" {
		return ErrMissingDBDriver
	}

	switch c.Database.Driver {
	case "mariadb":
		return c.validateMariaDBConfig()
	case "postgres":
		return c.validatePostgresConfig()
	default:
		return fmt.Errorf("unsupported database driver: %s", c.Database.Driver)
	}
}

// validateSecurityConfig validates security-related configuration
func (c *Config) validateSecurityConfig() error {
	if c.Security.FormRateLimit <= 0 {
		return ErrInvalidRateLimit
	}

	if c.Security.FormRateLimitWindow <= 0 {
		return ErrInvalidTimeout
	}

	if c.Security.CSRFConfig.Enabled && c.Security.CSRFConfig.Secret == "" {
		return ErrMissingCSRFSecret
	}

	return nil
}

// validateConfig performs comprehensive validation of the entire configuration
func (c *Config) validateConfig() error {
	validations := []struct {
		name string
		fn   func() error
	}{
		{"app", c.validateAppConfig},
		{"server", c.validateServerConfig},
		{"database", c.validateDatabaseConfig},
		{"security", c.validateSecurityConfig},
	}

	for _, v := range validations {
		if err := v.fn(); err != nil {
			return fmt.Errorf("invalid %s configuration: %w", v.name, err)
		}
	}

	return nil
}

// New creates a new configuration
func New() (*Config, error) {
	cfg := &Config{}

	// Load environment variables
	if err := cfg.loadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Validate all configuration
	if err := cfg.validateConfig(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) loadEnv() error {
	if err := envconfig.Process("", c); err != nil {
		return fmt.Errorf("failed to process environment variables: %w", err)
	}
	return nil
}
