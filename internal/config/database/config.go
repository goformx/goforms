package database

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
)

// Config represents the complete database configuration
type Config struct {
	Host           string        `validate:"required" envconfig:"DB_HOST" default:"localhost"`
	Port           int           `validate:"required" envconfig:"DB_PORT" default:"3306"`
	User           string        `validate:"required" envconfig:"DB_USER"`
	Password       string        `validate:"required" envconfig:"DB_PASSWORD"`
	Name           string        `validate:"required" envconfig:"DB_NAME"`
	MaxOpenConns   int           `validate:"required" envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns   int           `validate:"required" envconfig:"DB_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetme time.Duration `validate:"required" envconfig:"DB_CONN_MAX_LIFETIME" default:"5m"`
}

// NewConfig creates a new Config with default values
func NewConfig() (*Config, error) {
	var cfg Config

	// Process environment variables and defaults using DB prefix
	if err := envconfig.Process("DB", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process database config: %w", err)
	}

	// Create validator instance locally instead of globally
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("database config validation failed: %w", err)
	}

	return &cfg, nil
}
