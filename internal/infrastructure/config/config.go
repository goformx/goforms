package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

const (
	// DefaultAppPort is the default port number for the application server
	DefaultAppPort = 8090
	// DefaultServerPort is the default port number for the API server
	DefaultServerPort = 8099

	// DefaultReadTimeout is the default timeout for reading the entire request
	DefaultReadTimeout = 5 * time.Second
	// DefaultWriteTimeout is the default timeout for writing the response
	DefaultWriteTimeout = 10 * time.Second
	// DefaultIdleTimeout is the default timeout for idle connections
	DefaultIdleTimeout = 120 * time.Second
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 30 * time.Second
	// DefaultRequestTimeout is the default timeout for processing requests
	DefaultRequestTimeout = 30 * time.Second

	// DefaultCorsMaxAge is the default maximum age for CORS preflight requests
	DefaultCorsMaxAge = 3600

	// DefaultCorsOrigins is the default allowed CORS origins
	DefaultCorsOrigins = "http://localhost:3000,http://localhost:5173"

	// DefaultPort is the default port for the server
	DefaultPort = 8090
	// DefaultStartupTimeout is the default timeout for application startup
	DefaultStartupTimeout = 30 * time.Second
)

// CORSOriginsDecoder handles parsing of CORS allowed origins
type CORSOriginsDecoder []string

// Decode decodes the CORS origins configuration
func (c *CORSOriginsDecoder) Decode(value string) error {
	if value == "" {
		return nil
	}
	*c = strings.Split(value, ",")
	return nil
}

// CORSMethodsDecoder handles parsing of CORS allowed methods
type CORSMethodsDecoder []string

// Decode decodes the CORS methods configuration
func (c *CORSMethodsDecoder) Decode(value string) error {
	if value == "" {
		*c = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		return nil
	}
	*c = strings.Split(value, ",")
	return nil
}

// CORSHeadersDecoder handles parsing of CORS allowed headers
type CORSHeadersDecoder []string

// Decode decodes the CORS headers configuration
func (c *CORSHeadersDecoder) Decode(value string) error {
	if value == "" {
		*c = []string{"Origin", "Content-Type", "Accept"}
		return nil
	}
	*c = strings.Split(value, ",")
	return nil
}

// StaticConfig holds static file serving configuration
type StaticConfig struct {
	DistDir string `envconfig:"GOFORMS_STATIC_DIST_DIR" default:"dist"`
}

// Config represents the complete application configuration
type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Security SecurityConfig
	Static   StaticConfig
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name        string `envconfig:"GOFORMS_APP_NAME" default:"GoFormX"`
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
	LogLevel            string        `envconfig:"GOFORMS_LOG_LEVEL" default:"debug"`
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
		Secret  string `envconfig:"GOFORMS_CSRF_SECRET" required:"true"`
	} `envconfig:"GOFORMS_CSRF"`
}

// CSRFConfig holds CSRF-related configuration
type CSRFConfig struct {
	Enabled bool   `envconfig:"GOFORMS_CSRF_ENABLED" default:"true"`
	Secret  string `envconfig:"GOFORMS_CSRF_SECRET" validate:"required"`
}

// New creates a new configuration
func New() (*Config, error) {
	cfg := &Config{}

	// Load environment variables
	if err := cfg.loadEnv(); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Validate database configuration
	if err := cfg.validateDatabaseConfig(); err != nil {
		return nil, fmt.Errorf("invalid database configuration: %w", err)
	}

	// Validate security configuration
	if err := cfg.validateSecurityConfig(); err != nil {
		return nil, fmt.Errorf("invalid security configuration: %w", err)
	}

	return cfg, nil
}

func (c *Config) loadEnv() error {
	if err := envconfig.Process("", c); err != nil {
		return fmt.Errorf("failed to process environment variables: %w", err)
	}
	return nil
}

func (c *Config) validateDatabaseConfig() error {
	switch c.Database.Driver {
	case "mariadb":
		return c.validateMariaDBConfig()
	case "postgres":
		return c.validatePostgresConfig()
	default:
		return fmt.Errorf("unsupported database driver: %s", c.Database.Driver)
	}
}

func (c *Config) validateMariaDBConfig() error {
	if c.Database.MariaDB.Port == 0 {
		return fmt.Errorf("MariaDB port is required")
	}
	if c.Database.MariaDB.User == "" {
		return fmt.Errorf("MariaDB user is required")
	}
	if c.Database.MariaDB.Password == "" {
		return fmt.Errorf("MariaDB password is required")
	}
	if c.Database.MariaDB.Name == "" {
		return fmt.Errorf("MariaDB database name is required")
	}
	return nil
}

func (c *Config) validatePostgresConfig() error {
	if c.Database.Postgres.Port == 0 {
		return fmt.Errorf("PostgreSQL port is required")
	}
	if c.Database.Postgres.User == "" {
		return fmt.Errorf("PostgreSQL user is required")
	}
	if c.Database.Postgres.Password == "" {
		return fmt.Errorf("PostgreSQL password is required")
	}
	if c.Database.Postgres.Name == "" {
		return fmt.Errorf("PostgreSQL database name is required")
	}
	return nil
}

func (c *Config) validateSecurityConfig() error {
	if c.Security.CSRFConfig.Secret == "" {
		return fmt.Errorf("CSRF secret is required")
	}
	return nil
}
