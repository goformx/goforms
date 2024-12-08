package database

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
)

var validate = validator.New()

// Config represents the complete database configuration
type Config struct {
	Host           string        `validate:"required" envconfig:"DB_HOST" default:"localhost"`
	Port           int           `validate:"required" envconfig:"DB_PORT" default:"3306"`
	User           string        `validate:"required" envconfig:"DB_USER"`
	Password       string        `validate:"required" envconfig:"DB_PASSWORD"`
	DBName         string        `validate:"required" envconfig:"DB_NAME"`
	MaxOpenConns   int           `validate:"required" envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns   int           `validate:"required" envconfig:"DB_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetme time.Duration `validate:"required" envconfig:"DB_CONN_MAX_LIFETIME" default:"5m"`
}

// NewConfig creates a new Config with default values
func NewConfig() Config {
	var cfg Config
	// Process environment variables and defaults using DB prefix
	envconfig.Process("DB", &cfg)
	return cfg
}
