package config

import (
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
)

// AppConfig holds application-level configuration
type AppConfig struct {
	Name  string `envconfig:"APP_NAME" default:"goforms"`
	Env   string `envconfig:"APP_ENV" default:"development"`
	Debug bool   `envconfig:"APP_DEBUG" default:"false"`
	Port  int    `envconfig:"APP_PORT" default:"9009"`
	Host  string `envconfig:"APP_HOST" default:"localhost"`
}

// Config represents the complete application configuration
type Config struct {
	App       AppConfig
	Server    ServerConfig
	Database  DatabaseConfig
	Security  SecurityConfig
	RateLimit RateLimitConfig
}

// DatabaseConfig holds all database-related configuration
type DatabaseConfig struct {
	Host           string        `envconfig:"DB_HOST" validate:"required" default:"localhost"`
	Port           int           `envconfig:"DB_PORT" validate:"required" default:"3306"`
	User           string        `envconfig:"DB_USER" validate:"required" default:"goforms"`
	Password       string        `envconfig:"DB_PASSWORD" validate:"required" default:"goforms"`
	Name           string        `envconfig:"DB_NAME" validate:"required" default:"goforms"`
	MaxOpenConns   int           `envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns   int           `envconfig:"DB_MAX_IDLE_CONNS" default:"25"`
	ConnMaxLifetme time.Duration `envconfig:"DB_CONN_MAX_LIFETIME" default:"5m"`
}

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Host            string        `env:"SERVER_HOST" envDefault:"localhost"`
	Port            int           `env:"SERVER_PORT" envDefault:"8080"`
	ReadTimeout     time.Duration `envconfig:"READ_TIMEOUT" default:"5s"`
	WriteTimeout    time.Duration `envconfig:"WRITE_TIMEOUT" default:"10s"`
	IdleTimeout     time.Duration `envconfig:"IDLE_TIMEOUT" default:"120s"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	JWTSecret            string `env:"JWT_SECRET" envDefault:"your-secret-key"`
	CSRF                 CSRFConfig
	CorsAllowedOrigins   []string      `envconfig:"CORS_ALLOWED_ORIGINS" default:"http://localhost:3000"`
	CorsAllowedMethods   []string      `envconfig:"CORS_ALLOWED_METHODS" default:"GET,POST,PUT,DELETE,OPTIONS"`
	CorsAllowedHeaders   []string      `envconfig:"CORS_ALLOWED_HEADERS" default:"Origin,Content-Type,Accept,Authorization"`
	CorsMaxAge           int           `envconfig:"CORS_MAX_AGE" default:"3600"`
	CorsAllowCredentials bool          `envconfig:"CORS_ALLOW_CREDENTIALS" default:"true"`
	RequestTimeout       time.Duration `envconfig:"REQUEST_TIMEOUT" default:"30s"`
}

// CSRFConfig holds CSRF-related configuration
type CSRFConfig struct {
	Enabled bool   `env:"CSRF_ENABLED" envDefault:"true"`
	Secret  string `env:"CSRF_SECRET" envDefault:"csrf-secret-key"`
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	Enabled    bool          `envconfig:"RATE_LIMIT_ENABLED" default:"true"`
	Rate       int           `envconfig:"RATE_LIMIT" default:"100"`
	Burst      int           `envconfig:"RATE_BURST" default:"5"`
	TimeWindow time.Duration `envconfig:"RATE_LIMIT_TIME_WINDOW" default:"1m"`
	PerIP      bool          `envconfig:"RATE_LIMIT_PER_IP" default:"true"`
}

// New creates a new Config with default values
func New() (*Config, error) {
	var cfg Config

	// Simple debug output without logger dependency
	if os.Getenv("APP_DEBUG") == "true" {
		fmt.Fprintln(os.Stdout, "Loading configuration...")
	}

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
