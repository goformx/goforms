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
	App       AppConfig
	Server    ServerConfig
	Database  DatabaseConfig
	Security  SecurityConfig
	RateLimit RateLimitConfig
	Static    StaticConfig
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name        string `envconfig:"GOFORMS_APP_NAME" default:"GoForms"`
	Env         string `envconfig:"GOFORMS_APP_ENV" default:"production"`
	Debug       bool   `envconfig:"GOFORMS_APP_DEBUG" default:"false"`
	LogLevel    string `envconfig:"GOFORMS_APP_LOGLEVEL" default:"info"`
	Port        int    `envconfig:"GOFORMS_APP_PORT" default:"8090"`
	Host        string `envconfig:"GOFORMS_APP_HOST" default:"localhost"`
	ViteDevHost string `envconfig:"GOFORMS_VITE_DEV_HOST" default:"localhost"`
	ViteDevPort string `envconfig:"GOFORMS_VITE_DEV_PORT" default:"3000"`
}

// IsDevelopment returns true if the application is running in development mode
func (c *AppConfig) IsDevelopment() bool {
	return strings.EqualFold(c.Env, "development")
}

// DatabaseConfig holds all database-related configuration
type DatabaseConfig struct {
	Host            string        `envconfig:"GOFORMS_DB_HOST" validate:"required"`
	Port            int           `envconfig:"GOFORMS_DB_PORT" validate:"required"`
	User            string        `envconfig:"GOFORMS_DB_USER" validate:"required"`
	Password        string        `envconfig:"GOFORMS_DB_PASSWORD" validate:"required"`
	Name            string        `envconfig:"GOFORMS_DB_NAME" validate:"required"`
	MaxOpenConns    int           `envconfig:"GOFORMS_DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int           `envconfig:"GOFORMS_DB_MAX_IDLE_CONNS" default:"25"`
	ConnMaxLifetime time.Duration `envconfig:"GOFORMS_DB_CONN_MAX_LIFETIME" default:"5m"`
}

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Host            string        `envconfig:"GOFORMS_APP_HOST" default:"localhost"`
	Port            int           `envconfig:"GOFORMS_APP_PORT" default:"8099"`
	ReadTimeout     time.Duration `envconfig:"GOFORMS_READ_TIMEOUT" default:"5s"`
	WriteTimeout    time.Duration `envconfig:"GOFORMS_WRITE_TIMEOUT" default:"10s"`
	IdleTimeout     time.Duration `envconfig:"GOFORMS_IDLE_TIMEOUT" default:"120s"`
	ShutdownTimeout time.Duration `envconfig:"GOFORMS_SHUTDOWN_TIMEOUT" default:"30s"`
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	JWTSecret            string `envconfig:"GOFORMS_JWT_SECRET" validate:"required"`
	CSRF                 CSRFConfig
	CorsAllowedOrigins   CORSOriginsDecoder `envconfig:"GOFORMS_CORS_ALLOWED_ORIGINS" default:"http://localhost:3000"`
	CorsAllowedMethods   CORSMethodsDecoder `envconfig:"GOFORMS_CORS_ALLOWED_METHODS"`
	CorsAllowedHeaders   CORSHeadersDecoder `envconfig:"GOFORMS_CORS_ALLOWED_HEADERS"`
	CorsMaxAge           int                `envconfig:"GOFORMS_CORS_MAX_AGE" default:"3600"`
	CorsAllowCredentials bool               `envconfig:"GOFORMS_CORS_ALLOW_CREDENTIALS" default:"true"`
	RequestTimeout       time.Duration      `envconfig:"GOFORMS_REQUEST_TIMEOUT" default:"30s"`

	// Form-specific CORS settings
	FormCorsAllowedOrigins CORSOriginsDecoder `envconfig:"GOFORMS_FORM_CORS_ALLOWED_ORIGINS"`
	FormCorsAllowedMethods CORSMethodsDecoder `envconfig:"GOFORMS_FORM_CORS_ALLOWED_METHODS"`
	FormCorsAllowedHeaders CORSHeadersDecoder `envconfig:"GOFORMS_FORM_CORS_ALLOWED_HEADERS"`
	FormRateLimit          int                `envconfig:"GOFORMS_FORM_RATE_LIMIT" default:"20"`
	FormRateLimitWindow    time.Duration      `envconfig:"GOFORMS_FORM_RATE_LIMIT_WINDOW" default:"1m"`
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	Enabled    bool          `envconfig:"GOFORMS_RATE_LIMIT_ENABLED" default:"true"`
	Rate       int           `envconfig:"GOFORMS_RATE_LIMIT" default:"100"`
	Burst      int           `envconfig:"GOFORMS_RATE_BURST" default:"5"`
	TimeWindow time.Duration `envconfig:"GOFORMS_RATE_LIMIT_TIME_WINDOW" default:"1m"`
	PerIP      bool          `envconfig:"GOFORMS_RATE_LIMIT_PER_IP" default:"true"`
}

// CSRFConfig holds CSRF-related configuration
type CSRFConfig struct {
	Enabled bool   `envconfig:"GOFORMS_CSRF_ENABLED" default:"true"`
	Secret  string `envconfig:"GOFORMS_CSRF_SECRET" validate:"required"`
}

// New creates a new Config instance with default values
func New() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name:        "GoForms",
			Env:         "production",
			Debug:       false,
			LogLevel:    "info",
			Port:        DefaultAppPort,
			Host:        "localhost",
			ViteDevHost: "localhost",
			ViteDevPort: "3000",
		},
		Server: ServerConfig{
			Host:            "localhost",
			Port:            DefaultServerPort,
			ReadTimeout:     DefaultReadTimeout,
			WriteTimeout:    DefaultWriteTimeout,
			IdleTimeout:     DefaultIdleTimeout,
			ShutdownTimeout: DefaultShutdownTimeout,
		},
		Security: SecurityConfig{
			CorsMaxAge:           DefaultCorsMaxAge,
			CorsAllowCredentials: true,
			RequestTimeout:       DefaultRequestTimeout,
		},
		Static: StaticConfig{
			DistDir: "dist",
		},
	}

	// Process environment variables
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("failed to process environment variables: %w", err)
	}

	return cfg, nil
}
