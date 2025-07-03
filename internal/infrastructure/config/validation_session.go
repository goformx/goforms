// Package config provides validation utilities for Viper-based configuration
package config

import (
	"path/filepath"
	"strings"
)

// validateSessionConfig validates session configuration
func validateSessionConfig(cfg SessionConfig, result *ValidationResult) {
	if cfg.Type == "" {
		result.AddError("session.type", "session type is required", cfg.Type)
	}

	// Validate session type
	supportedTypes := []string{"cookie", "redis", "file"}
	typeValid := false

	for _, sessionType := range supportedTypes {
		if strings.EqualFold(cfg.Type, sessionType) {
			typeValid = true

			break
		}
	}

	if !typeValid {
		result.AddError("session.type", "unsupported session type", cfg.Type)
	}

	// Validate session secret
	if cfg.Secret == "" {
		result.AddError("session.secret", "session secret is required", "***")
	} else if len(cfg.Secret) < 32 {
		result.AddError("session.secret",
			"session secret must be at least 32 characters long", "***")
	}

	// Validate session duration
	if cfg.MaxAge <= 0 {
		result.AddError("session.max_age",
			"session max age must be positive", cfg.MaxAge)
	}

	// Validate file session configuration
	if strings.EqualFold(cfg.Type, "file") {
		if cfg.StoreFile == "" {
			result.AddError("session.store_file",
				"session store file is required for file sessions", cfg.StoreFile)
		} else {
			// Ensure session directory is writable
			sessionDir := filepath.Dir(cfg.StoreFile)
			if !isWritableDirectory(sessionDir) {
				result.AddError("session.store_file",
					"session directory must be writable", sessionDir)
			}
		}
	}
}
