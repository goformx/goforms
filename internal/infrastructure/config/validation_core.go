// Package config provides validation utilities for Viper-based configuration
package config

import (
	"net/url"
	"strings"
)

// =============================================================================
// APP CONFIGURATION VALIDATION
// =============================================================================

// validateAppConfig validates application configuration
func validateAppConfig(cfg AppConfig, result *ValidationResult) {
	validateAppConfigName(cfg, result)
	validateAppConfigPort(cfg, result)
	validateAppConfigTimeouts(cfg, result)
	validateAppConfigURL(cfg, result)
	validateAppConfigEnvironment(cfg, result)
}

func validateAppConfigName(cfg AppConfig, result *ValidationResult) {
	if cfg.Name == "" {
		result.AddError("app.name", "application name is required", cfg.Name)
	}
}

func validateAppConfigPort(cfg AppConfig, result *ValidationResult) {
	if cfg.Port <= 0 || cfg.Port > 65535 {
		result.AddError("app.port", "port must be between 1 and 65535", cfg.Port)
	}
}

func validateAppConfigTimeouts(cfg AppConfig, result *ValidationResult) {
	if cfg.ReadTimeout <= 0 {
		result.AddError("app.read_timeout", "read timeout must be positive", cfg.ReadTimeout)
	}

	if cfg.WriteTimeout <= 0 {
		result.AddError("app.write_timeout", "write timeout must be positive", cfg.WriteTimeout)
	}

	if cfg.IdleTimeout <= 0 {
		result.AddError("app.idle_timeout", "idle timeout must be positive", cfg.IdleTimeout)
	}
}

func validateAppConfigURL(cfg AppConfig, result *ValidationResult) {
	if cfg.URL != "" {
		if _, err := url.Parse(cfg.URL); err != nil {
			result.AddError("app.url", "invalid URL format", cfg.URL)
		}
	}
}

func validateAppConfigEnvironment(cfg AppConfig, result *ValidationResult) {
	validEnvironments := []string{"development", "staging", "production", "test"}
	envValid := false

	for _, env := range validEnvironments {
		if strings.EqualFold(cfg.Environment, env) {
			envValid = true

			break
		}
	}

	if !envValid {
		result.AddError(
			"app.environment",
			"environment must be one of: development, staging, production, test",
			cfg.Environment,
		)
	}
}

// =============================================================================
// DATABASE CONFIGURATION VALIDATION
// =============================================================================

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

// =============================================================================
// SECURITY CONFIGURATION VALIDATION
// =============================================================================

// validateSecurityConfig validates security configuration
func validateSecurityConfig(cfg SecurityConfig, result *ValidationResult) {
	validateSecurityConfigSecret(cfg, result)
	validateSecurityConfigTLS(cfg, result)
	validateSecurityConfigCORS(cfg, result)
	validateSecurityConfigRateLimit(cfg, result)
	validateSecurityConfigSession(cfg, result)
}

func validateSecurityConfigSecret(cfg SecurityConfig, result *ValidationResult) {
	// Security secret validation is handled by individual components (CSRF, etc.)
	// No global secret field in SecurityConfig
}

func validateSecurityConfigTLS(cfg SecurityConfig, result *ValidationResult) {
	if cfg.TLS.Enabled {
		if cfg.TLS.CertFile == "" {
			result.AddError("security.tls.cert_file", "TLS certificate file is required when TLS is enabled", cfg.TLS.CertFile)
		}

		if cfg.TLS.KeyFile == "" {
			result.AddError("security.tls.key_file", "TLS key file is required when TLS is enabled", cfg.TLS.KeyFile)
		}

		if !fileExists(cfg.TLS.CertFile) {
			result.AddError("security.tls.cert_file", "TLS certificate file does not exist", cfg.TLS.CertFile)
		}

		if !fileExists(cfg.TLS.KeyFile) {
			result.AddError("security.tls.key_file", "TLS key file does not exist", cfg.TLS.KeyFile)
		}
	}
}

func validateSecurityConfigCORS(cfg SecurityConfig, result *ValidationResult) {
	if cfg.CORS.Enabled {
		if len(cfg.CORS.AllowedOrigins) == 0 {
			result.AddError("security.cors.allowed_origins", "at least one allowed origin is required when CORS is enabled", cfg.CORS.AllowedOrigins)
		}

		if len(cfg.CORS.AllowedMethods) == 0 {
			result.AddError("security.cors.allowed_methods", "at least one allowed method is required when CORS is enabled", cfg.CORS.AllowedMethods)
		}
	}
}

func validateSecurityConfigRateLimit(cfg SecurityConfig, result *ValidationResult) {
	if cfg.RateLimit.Enabled {
		if cfg.RateLimit.RPS <= 0 {
			result.AddError("security.rate_limit.rps", "RPS must be positive", cfg.RateLimit.RPS)
		}

		if cfg.RateLimit.Burst <= 0 {
			result.AddError("security.rate_limit.burst", "burst must be positive", cfg.RateLimit.Burst)
		}
	}
}

func validateSecurityConfigSession(cfg SecurityConfig, result *ValidationResult) {
	// Session timeout is handled in AuthConfig, not SecurityConfig
	if cfg.SecureCookie && !cfg.TLS.Enabled {
		result.AddError("security.secure_cookie", "secure cookies require TLS to be enabled", cfg.SecureCookie)
	}
}
