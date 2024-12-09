package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jonesrussell/goforms/internal/config/database"
	"github.com/jonesrussell/goforms/internal/config/server"
	"github.com/kelseyhightower/envconfig"
)

// Config represents the complete application configuration
type Config struct {
	App       AppConfig       `validate:"required" envconfig:"APP"`
	Server    server.Config   `validate:"required" envconfig:"SERVER"`
	Database  database.Config `validate:"required" envconfig:"DB"`
	Security  SecurityConfig  `validate:"required" envconfig:"SECURITY"`
	RateLimit RateLimitConfig `validate:"required" envconfig:"RATE_LIMIT"`
	Logging   LoggingConfig   `validate:"required" envconfig:"LOG"`
}

// AppConfig contains basic application settings
type AppConfig struct {
	Name string `validate:"required" envconfig:"NAME" default:"goforms"`
	Env  string `validate:"required,oneof=development staging production" envconfig:"ENV" default:"development"`
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	CorsAllowedOrigins   []string      `validate:"required" envconfig:"CORS_ALLOWED_ORIGINS" default:"http://localhost:3000"`
	CorsAllowedMethods   []string      `validate:"required" envconfig:"CORS_ALLOWED_METHODS" default:"GET,POST,PUT,DELETE,OPTIONS"`
	CorsAllowedHeaders   []string      `validate:"required" envconfig:"CORS_ALLOWED_HEADERS" default:"Origin,Content-Type,Accept,Authorization,X-Requested-With"`
	CorsMaxAge           int           `validate:"required" envconfig:"CORS_MAX_AGE" default:"3600"`
	CorsAllowCredentials bool          `validate:"required" envconfig:"CORS_ALLOW_CREDENTIALS" default:"true"`
	TrustedProxies       []string      `validate:"required" envconfig:"TRUSTED_PROXIES" default:"127.0.0.1,::1"`
	RequestTimeout       time.Duration `validate:"required" envconfig:"REQUEST_TIMEOUT" default:"30s"`
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	Enabled     bool          `validate:"required" envconfig:"ENABLED" default:"true"`
	Rate        int           `validate:"required_if=Enabled true,min=1" envconfig:"RATE" default:"100"`
	Burst       int           `validate:"required_if=Enabled true,min=1" envconfig:"BURST" default:"5"`
	TimeWindow  time.Duration `validate:"required_if=Enabled true" envconfig:"TIME_WINDOW" default:"1m"`
	PerIP       bool          `validate:"required" envconfig:"PER_IP" default:"true"`
	ExemptPaths []string      `validate:"required" envconfig:"EXEMPT_PATHS" default:"/health,/metrics"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `validate:"required,oneof=debug info warn error" envconfig:"LEVEL" default:"debug"`
	Format string `validate:"required,oneof=json console" envconfig:"FORMAT" default:"json"`
}

// New provides the application configuration for fx
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
