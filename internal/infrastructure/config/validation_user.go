// Package config provides validation utilities for Viper-based configuration
package config

// validateUserConfig validates user configuration
func validateUserConfig(cfg UserConfig, result *ValidationResult) {
	// Validate admin user configuration
	if cfg.Admin.Email != "" && !isValidEmail(cfg.Admin.Email) {
		result.AddError("user.admin.email",
			"invalid admin email format", cfg.Admin.Email)
	}

	if cfg.Admin.Password != "" && len(cfg.Admin.Password) < 8 {
		result.AddError("user.admin.password",
			"admin password must be at least 8 characters long", "***")
	}

	// Validate default user configuration
	if cfg.Default.Role == "" {
		result.AddError("user.default.role",
			"default user role is required", cfg.Default.Role)
	}
}
