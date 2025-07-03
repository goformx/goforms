// Package config provides validation utilities for Viper-based configuration
package config

// validateAPIConfig validates API configuration
func validateAPIConfig(cfg APIConfig, result *ValidationResult) {
	if cfg.Version == "" {
		result.AddError("api.version", "API version is required", cfg.Version)
	}

	if cfg.Timeout <= 0 {
		result.AddError("api.timeout", "API timeout must be positive", cfg.Timeout)
	}

	if cfg.MaxRetries < 0 {
		result.AddError("api.max_retries",
			"API max retries must be non-negative", cfg.MaxRetries)
	}

	// Validate API rate limiting
	if cfg.RateLimit.Enabled {
		if cfg.RateLimit.RPS <= 0 {
			result.AddError("api.rate_limit.rps",
				"API rate limit RPS must be positive", cfg.RateLimit.RPS)
		}

		if cfg.RateLimit.Burst <= 0 {
			result.AddError("api.rate_limit.burst",
				"API rate limit burst must be positive", cfg.RateLimit.Burst)
		}
	}
}
