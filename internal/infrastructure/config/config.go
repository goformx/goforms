package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
)

// Config represents the complete application configuration
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Security  SecurityConfig
	RateLimit RateLimitConfig
}

// DatabaseConfig holds all database-related configuration
type DatabaseConfig struct {
	Host           string        `envconfig:"DB_HOST" default:"localhost"`
	Port           int           `envconfig:"DB_PORT" default:"3306"`
	User           string        `envconfig:"DB_USER" default:"root"`
	Password       string        `envconfig:"DB_PASSWORD" default:""`
	Name           string        `envconfig:"DB_NAME" default:"goforms"`
	MaxOpenConns   int           `envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns   int           `envconfig:"DB_MAX_IDLE_CONNS" default:"25"`
	ConnMaxLifetme time.Duration `envconfig:"DB_CONN_MAX_LIFETIME" default:"5m"`
}

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Port int    `envconfig:"SERVER_PORT" default:"8080"`
	Host string `envconfig:"SERVER_HOST" default:"localhost"`
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	CorsAllowedOrigins   []string      `envconfig:"SECURITY_CORS_ALLOWED_ORIGINS" default:"http://localhost:3000"`
	CorsAllowedMethods   []string      `envconfig:"SECURITY_CORS_ALLOWED_METHODS" default:"GET,POST,PUT,DELETE,OPTIONS"`
	CorsAllowedHeaders   []string      `envconfig:"SECURITY_CORS_ALLOWED_HEADERS" default:"Origin,Content-Type,Accept,Authorization"`
	CorsMaxAge           int           `envconfig:"SECURITY_CORS_MAX_AGE" default:"3600"`
	CorsAllowCredentials bool          `envconfig:"SECURITY_CORS_ALLOW_CREDENTIALS" default:"true"`
	RequestTimeout       time.Duration `envconfig:"SECURITY_REQUEST_TIMEOUT" default:"30s"`
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	Enabled    bool          `envconfig:"RATE_LIMIT_ENABLED" default:"true"`
	Rate       int           `envconfig:"RATE_LIMIT_RATE" default:"100"`
	Burst      int           `envconfig:"RATE_LIMIT_BURST" default:"5"`
	TimeWindow time.Duration `envconfig:"RATE_LIMIT_TIME_WINDOW" default:"1m"`
	PerIP      bool          `envconfig:"RATE_LIMIT_PER_IP" default:"true"`
}

// New creates a new Config with default values
func New() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
