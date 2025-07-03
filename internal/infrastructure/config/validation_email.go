// Package config provides validation utilities for Viper-based configuration
package config

// validateEmailConfig validates email configuration
func validateEmailConfig(cfg EmailConfig, result *ValidationResult) {
	if cfg.Host != "" {
		if cfg.Username == "" {
			result.AddError("email.username",
				"email username is required when host is specified", cfg.Username)
		}

		if cfg.Port <= 0 || cfg.Port > 65535 {
			result.AddError("email.port",
				"email port must be between 1 and 65535", cfg.Port)
		}

		// Validate email format
		if cfg.From != "" && !isValidEmail(cfg.From) {
			result.AddError("email.from", "invalid email format", cfg.From)
		}
	}
}
