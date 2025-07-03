// Package config provides validation utilities for Viper-based configuration
package config

// validateWebConfig validates web configuration
func validateWebConfig(cfg WebConfig, result *ValidationResult) {
	if cfg.ReadTimeout <= 0 {
		result.AddError("web.read_timeout",
			"web read timeout must be positive", cfg.ReadTimeout)
	}

	if cfg.WriteTimeout <= 0 {
		result.AddError("web.write_timeout",
			"web write timeout must be positive", cfg.WriteTimeout)
	}

	if cfg.IdleTimeout <= 0 {
		result.AddError("web.idle_timeout",
			"web idle timeout must be positive", cfg.IdleTimeout)
	}

	// Validate template directory
	if cfg.TemplateDir != "" && !isReadableDirectory(cfg.TemplateDir) {
		result.AddError("web.template_dir",
			"template directory must be readable", cfg.TemplateDir)
	}

	// Validate static directory
	if cfg.StaticDir != "" && !isReadableDirectory(cfg.StaticDir) {
		result.AddError("web.static_dir",
			"static directory must be readable", cfg.StaticDir)
	}
}
