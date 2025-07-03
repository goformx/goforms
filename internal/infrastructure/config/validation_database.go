// Package config provides validation utilities for Viper-based configuration
package config

import (
	"strings"
)

// validateDatabaseConfig validates database configuration
func validateDatabaseConfig(cfg DatabaseConfig, result *ValidationResult) {
	validateDatabaseConfigDriverPresence(cfg, result)
	validateDatabaseConfigDriver(cfg, result)
	validateDatabaseConfigHost(cfg, result)
	validateDatabaseConfigPort(cfg, result)
	validateDatabaseConfigName(cfg, result)
	validateDatabaseConfigUsername(cfg, result)
	validateDatabaseConfigPool(cfg, result)
}

func validateDatabaseConfigDriverPresence(cfg DatabaseConfig, result *ValidationResult) {
	if cfg.Driver == "" {
		result.AddError("database.driver", "database driver is required", cfg.Driver)
	}
}

func validateDatabaseConfigDriver(cfg DatabaseConfig, result *ValidationResult) {
	supportedDrivers := []string{"postgres", "mysql", "mariadb"}
	driverValid := false

	for _, driver := range supportedDrivers {
		if strings.EqualFold(cfg.Driver, driver) {
			driverValid = true

			break
		}
	}

	if !driverValid {
		result.AddError("database.driver", "unsupported database driver", cfg.Driver)
	}
}

func validateDatabaseConfigHost(cfg DatabaseConfig, result *ValidationResult) {
	if cfg.Host == "" {
		result.AddError("database.host", "database host is required", cfg.Host)
	}
}

func validateDatabaseConfigPort(cfg DatabaseConfig, result *ValidationResult) {
	if cfg.Port <= 0 || cfg.Port > 65535 {
		result.AddError("database.port", "database port must be between 1 and 65535", cfg.Port)
	}
}

func validateDatabaseConfigName(cfg DatabaseConfig, result *ValidationResult) {
	if cfg.Name == "" {
		result.AddError("database.name", "database name is required", cfg.Name)
	}
}

func validateDatabaseConfigUsername(cfg DatabaseConfig, result *ValidationResult) {
	if cfg.Username == "" {
		result.AddError("database.username", "database username is required", cfg.Username)
	}
}

func validateDatabaseConfigPool(cfg DatabaseConfig, result *ValidationResult) {
	if cfg.MaxOpenConns <= 0 {
		result.AddError("database.max_open_conns", "max open connections must be positive", cfg.MaxOpenConns)
	}

	if cfg.MaxIdleConns <= 0 {
		result.AddError("database.max_idle_conns", "max idle connections must be positive", cfg.MaxIdleConns)
	}

	if cfg.ConnMaxLifetime <= 0 {
		result.AddError("database.conn_max_lifetime", "connection max lifetime must be positive", cfg.ConnMaxLifetime)
	}

	if cfg.ConnMaxIdleTime <= 0 {
		result.AddError("database.conn_max_idle_time", "connection max idle time must be positive", cfg.ConnMaxIdleTime)
	}
}
