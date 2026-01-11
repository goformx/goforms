// Package config provides validation utilities for Viper-based configuration
package config

import (
	"path/filepath"
	"strings"
)

// validateSessionConfig validates session configuration
func validateSessionConfig(cfg SessionConfig, result *ValidationResult) {
	validateSessionType(cfg, result)
	validateSessionSecret(cfg, result)
	validateSessionDuration(cfg, result)
	validateSessionFile(cfg, result)
}

func validateSessionType(cfg SessionConfig, result *ValidationResult) {
	if cfg.Type == "" {
		result.AddError("session.type", "session type is required", cfg.Type)

		return
	}

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
}

func validateSessionSecret(cfg SessionConfig, result *ValidationResult) {
	if cfg.Secret == "" {
		result.AddError("session.secret", "session secret is required", "***")
	} else if len(cfg.Secret) < MinSecretLength {
		result.AddError("session.secret",
			"session secret must be at least 32 characters long", "***")
	}
}

func validateSessionDuration(cfg SessionConfig, result *ValidationResult) {
	if cfg.MaxAge <= 0 {
		result.AddError("session.max_age",
			"session max age must be positive", cfg.MaxAge)
	}
}

func validateSessionFile(cfg SessionConfig, result *ValidationResult) {
	if !strings.EqualFold(cfg.Type, "file") {
		return
	}

	if cfg.StoreFile == "" {
		result.AddError("session.store_file",
			"session store file is required for file sessions", cfg.StoreFile)

		return
	}
	// Ensure session directory is writable
	sessionDir := filepath.Dir(cfg.StoreFile)
	if !isWritableDirectory(sessionDir) {
		result.AddError("session.store_file",
			"session directory must be writable", sessionDir)
	}
}
