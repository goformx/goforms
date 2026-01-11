// Package config provides validation utilities for Viper-based configuration
package config

// validateAuthConfig validates authentication configuration
func validateAuthConfig(cfg AuthConfig, result *ValidationResult) {
	if cfg.PasswordMinLength < MinPasswordLengthThreshold {
		result.AddError("auth.password_min_length",
			"password minimum length must be at least 6", cfg.PasswordMinLength)
	}

	if cfg.SessionTimeout <= 0 {
		result.AddError("auth.session_timeout",
			"session timeout must be positive", cfg.SessionTimeout)
	}

	if cfg.MaxLoginAttempts <= 0 {
		result.AddError("auth.max_login_attempts",
			"max login attempts must be positive", cfg.MaxLoginAttempts)
	}

	if cfg.LockoutDuration <= 0 {
		result.AddError("auth.lockout_duration",
			"lockout duration must be positive", cfg.LockoutDuration)
	}
}
