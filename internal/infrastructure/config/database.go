package config

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// DatabaseConfig holds all database-related configuration
type DatabaseConfig struct {
	// Common database settings
	Driver          string        `json:"driver"`
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	Name            string        `json:"name"`
	Username        string        `json:"username"`
	Password        string        `json:"password"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`

	// PostgreSQL specific settings
	SSLMode string `json:"ssl_mode"`

	// MariaDB specific settings
	RootPassword string `json:"root_password"`

	// Logging configuration
	Logging DatabaseLoggingConfig `json:"logging"`
}

// DatabaseLoggingConfig holds database logging configuration
type DatabaseLoggingConfig struct {
	// SlowThreshold is the threshold for logging slow queries
	SlowThreshold time.Duration `json:"slow_threshold"`
	// Parameterized enables logging of query parameters
	Parameterized bool `json:"parameterized"`
	// IgnoreNotFound determines whether to ignore record not found errors
	IgnoreNotFound bool `json:"ignore_not_found"`
	// LogLevel determines the verbosity of database logging
	// Valid values: "silent", "error", "warn", "info"
	LogLevel string `json:"log_level"`
}

// Validate validates the database configuration
func (c *DatabaseConfig) Validate() error {
	var errs []string

	// Validate common fields
	if err := c.validateCommonFields(); err != nil {
		errs = append(errs, err.Error())
	}

	// Validate driver-specific fields
	if err := c.validateDriverSpecificFields(); err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("database config validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// validateCommonFields validates common database configuration fields
func (c *DatabaseConfig) validateCommonFields() error {
	var errs []string

	if c.Host == "" {
		errs = append(errs, "database host is required")
	}

	if c.Port <= 0 || c.Port > 65535 {
		errs = append(errs, "database port must be between 1 and 65535")
	}

	if c.Username == "" {
		errs = append(errs, "database username is required")
	}

	if c.Password == "" {
		errs = append(errs, "database password is required")
	}

	if c.Name == "" {
		errs = append(errs, "database name is required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	return nil
}

// validateDriverSpecificFields validates driver-specific configuration fields
func (c *DatabaseConfig) validateDriverSpecificFields() error {
	switch c.Driver {
	case "postgres":
		return c.validatePostgresFields()
	case "mariadb":
		return c.validateMariaDBFields()
	default:
		return errors.New("unsupported database driver type")
	}
}

// validatePostgresFields validates PostgreSQL-specific fields
func (c *DatabaseConfig) validatePostgresFields() error {
	if c.SSLMode == "" {
		return errors.New("PostgreSQL SSL mode is required")
	}

	return nil
}

// validateMariaDBFields validates MariaDB-specific fields
func (c *DatabaseConfig) validateMariaDBFields() error {
	if c.RootPassword == "" {
		return errors.New("MariaDB root password is required")
	}

	return nil
}
