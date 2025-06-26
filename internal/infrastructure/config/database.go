package config

import (
	"fmt"
	"strings"
	"time"
)

// DatabaseConfig holds all database-related configuration
type DatabaseConfig struct {
	// Common database settings
	Connection      string        `envconfig:"GOFORMS_DB_CONNECTION" default:"mariadb"`
	Host            string        `envconfig:"GOFORMS_DB_HOST" validate:"required"`
	Port            int           `envconfig:"GOFORMS_DB_PORT" default:"3306"`
	Database        string        `envconfig:"GOFORMS_DB_DATABASE" validate:"required"`
	Username        string        `envconfig:"GOFORMS_DB_USERNAME" validate:"required"`
	Password        string        `envconfig:"GOFORMS_DB_PASSWORD" validate:"required"`
	MaxOpenConns    int           `envconfig:"GOFORMS_DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int           `envconfig:"GOFORMS_DB_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime time.Duration `envconfig:"GOFORMS_DB_CONN_MAX_LIFETIME" default:"5m"`

	// PostgreSQL specific settings
	SSLMode string `envconfig:"GOFORMS_DB_SSLMODE" default:"disable"`

	// MariaDB specific settings
	RootPassword string `envconfig:"GOFORMS_DB_ROOT_PASSWORD"`

	// Logging configuration
	Logging DatabaseLoggingConfig `envconfig:"GOFORMS_DB_LOGGING"`
}

// DatabaseLoggingConfig holds database logging configuration
type DatabaseLoggingConfig struct {
	// SlowThreshold is the threshold for logging slow queries
	SlowThreshold time.Duration `envconfig:"GOFORMS_DB_SLOW_THRESHOLD" default:"1s"`
	// Parameterized enables logging of query parameters
	Parameterized bool `envconfig:"GOFORMS_DB_LOG_PARAMETERS" default:"false"`
	// IgnoreNotFound determines whether to ignore record not found errors
	IgnoreNotFound bool `envconfig:"GOFORMS_DB_IGNORE_NOT_FOUND" default:"false"`
	// LogLevel determines the verbosity of database logging
	// Valid values: "silent", "error", "warn", "info"
	LogLevel string `envconfig:"GOFORMS_DB_LOG_LEVEL" default:"warn"`
}

// validateDatabaseConfig validates database configuration
func (c *Config) validateDatabaseConfig() error {
	var errs []string

	if c.Database.Host == "" {
		errs = append(errs, "database host is required")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		errs = append(errs, "database port must be between 1 and 65535")
	}
	if c.Database.Username == "" {
		errs = append(errs, "database username is required")
	}
	if c.Database.Password == "" {
		errs = append(errs, "database password is required")
	}
	if c.Database.Database == "" {
		errs = append(errs, "database name is required")
	}

	// Validate database-specific settings
	switch c.Database.Connection {
	case "postgres":
		if c.Database.SSLMode == "" {
			errs = append(errs, "PostgreSQL SSL mode is required")
		}
	case "mariadb":
		if c.Database.RootPassword == "" {
			errs = append(errs, "MariaDB root password is required")
		}
	default:
		errs = append(errs, "unsupported database connection type")
	}

	if len(errs) > 0 {
		return fmt.Errorf("database config validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}
