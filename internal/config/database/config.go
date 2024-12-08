package database

import "time"

// Config represents the complete database configuration
type Config struct {
	Host           string        `validate:"required" envconfig:"HOSTNAME" default:"localhost"`
	Port           int           `validate:"required" envconfig:"PORT" default:"3306"`
	User           string        `validate:"required" envconfig:"USER"`
	Password       string        `validate:"required" envconfig:"PASSWORD"`
	DBName         string        `validate:"required" envconfig:"DATABASE"`
	MaxOpenConns   int           `validate:"required" envconfig:"MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns   int           `validate:"required" envconfig:"MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetme time.Duration `validate:"required" envconfig:"CONN_MAX_LIFETIME" default:"5m"`
}
