// Package config provides validation utilities for Viper-based configuration
package config

import (
	"net/url"
	"strings"
)

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
