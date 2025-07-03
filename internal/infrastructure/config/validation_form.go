// Package config provides validation utilities for Viper-based configuration
package config

// validateFormConfig validates form configuration
func validateFormConfig(cfg FormConfig, result *ValidationResult) {
	if cfg.MaxFileSize <= 0 {
		result.AddError("form.max_file_size",
			"max file size must be positive", cfg.MaxFileSize)
	}

	if cfg.MaxFields <= 0 {
		result.AddError("form.max_fields",
			"max fields must be positive", cfg.MaxFields)
	}

	if cfg.MaxMemory <= 0 {
		result.AddError("form.max_memory",
			"max memory must be positive", cfg.MaxMemory)
	}

	if cfg.Validation.MaxErrors <= 0 {
		result.AddError("form.validation.max_errors",
			"max errors must be positive", cfg.Validation.MaxErrors)
	}
}
