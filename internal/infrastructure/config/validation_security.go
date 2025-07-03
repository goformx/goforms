// Package config provides validation utilities for Viper-based configuration
package config

// validateSecurityConfig validates security configuration
func validateSecurityConfig(cfg SecurityConfig, result *ValidationResult) {
	validateSecurityCSRF(cfg, result)
	validateSecurityCORS(cfg, result)
	validateSecurityRateLimit(cfg, result)
	validateSecurityTLS(cfg, result)
}

func validateSecurityCSRF(cfg SecurityConfig, result *ValidationResult) {
	if cfg.CSRF.Enabled && len(cfg.CSRF.Secret) < 32 {
		result.AddError("security.csrf.secret",
			"CSRF secret must be at least 32 characters long", "***")
	}
}

func validateSecurityCORS(cfg SecurityConfig, result *ValidationResult) {
	if !cfg.CORS.Enabled {
		return
	}

	if len(cfg.CORS.AllowedOrigins) == 0 {
		result.AddError("security.cors.allowed_origins",
			"at least one allowed origin is required when CORS is enabled",
			cfg.CORS.AllowedOrigins)
	}

	// Check for wildcard with credentials
	for _, origin := range cfg.CORS.AllowedOrigins {
		if origin == "*" && cfg.CORS.AllowCredentials {
			result.AddError("security.cors.allowed_origins",
				"wildcard origin (*) cannot be used with allow_credentials=true", origin)
		}
	}
}

func validateSecurityRateLimit(cfg SecurityConfig, result *ValidationResult) {
	if !cfg.RateLimit.Enabled {
		return
	}

	if cfg.RateLimit.RPS <= 0 {
		result.AddError("security.rate_limit.rps",
			"rate limit RPS must be positive", cfg.RateLimit.RPS)
	}

	if cfg.RateLimit.Burst <= 0 {
		result.AddError("security.rate_limit.burst",
			"rate limit burst must be positive", cfg.RateLimit.Burst)
	}

	if cfg.RateLimit.Window <= 0 {
		result.AddError("security.rate_limit.window",
			"rate limit window must be positive", cfg.RateLimit.Window)
	}
}

func validateSecurityTLS(cfg SecurityConfig, result *ValidationResult) {
	if !cfg.TLS.Enabled {
		return
	}

	if cfg.TLS.CertFile == "" {
		result.AddError("security.tls.cert_file",
			"TLS certificate file is required when TLS is enabled", cfg.TLS.CertFile)
	}

	if cfg.TLS.KeyFile == "" {
		result.AddError("security.tls.key_file",
			"TLS key file is required when TLS is enabled", cfg.TLS.KeyFile)
	}

	// Validate file existence
	if cfg.TLS.CertFile != "" && !fileExists(cfg.TLS.CertFile) {
		result.AddError("security.tls.cert_file",
			"TLS certificate file does not exist", cfg.TLS.CertFile)
	}

	if cfg.TLS.KeyFile != "" && !fileExists(cfg.TLS.KeyFile) {
		result.AddError("security.tls.key_file",
			"TLS key file does not exist", cfg.TLS.KeyFile)
	}
}
