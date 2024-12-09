package server

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string `validate:"required" envconfig:"SERVER_HOST" default:"localhost"`
	Port     int    `validate:"required" envconfig:"SERVER_PORT" default:"8090"`
	Timeouts TimeoutConfig
}

// TimeoutConfig contains server timeout settings
type TimeoutConfig struct {
	Read  time.Duration `validate:"required" envconfig:"SERVER_READ_TIMEOUT" default:"15s"`
	Write time.Duration `validate:"required" envconfig:"SERVER_WRITE_TIMEOUT" default:"15s"`
	Idle  time.Duration `validate:"required" envconfig:"SERVER_IDLE_TIMEOUT" default:"60s"`
}

// NewConfig creates a new Config with default values
func NewConfig() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("SERVER", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process server config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("server config validation failed: %w", err)
	}

	return &cfg, nil
}
