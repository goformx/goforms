// Package config provides validation utilities for Viper-based configuration
package config

import (
	"path/filepath"
	"strings"
)

// validateLoggingConfig validates logging configuration
func validateLoggingConfig(cfg LoggingConfig, result *ValidationResult) {
	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	levelValid := false

	for _, level := range validLevels {
		if strings.EqualFold(cfg.Level, level) {
			levelValid = true

			break
		}
	}

	if !levelValid {
		result.AddError("logging.level", "invalid log level", cfg.Level)
	}

	// Validate log format
	validFormats := []string{"json", "console"}
	formatValid := false

	for _, format := range validFormats {
		if strings.EqualFold(cfg.Format, format) {
			formatValid = true

			break
		}
	}

	if !formatValid {
		result.AddError("logging.format", "invalid log format", cfg.Format)
	}

	// Validate file logging configuration
	if cfg.Output == "file" {
		if cfg.File == "" {
			result.AddError("logging.file",
				"log file path is required when output is file", cfg.File)
		} else {
			// Ensure log directory is writable
			logDir := filepath.Dir(cfg.File)
			if !isWritableDirectory(logDir) {
				result.AddError("logging.file",
					"log directory must be writable", logDir)
			}
		}
	}

	// Validate rotation settings
	if cfg.MaxSize <= 0 {
		result.AddError("logging.max_size",
			"log max size must be positive", cfg.MaxSize)
	}

	if cfg.MaxBackups < 0 {
		result.AddError("logging.max_backups",
			"log max backups must be non-negative", cfg.MaxBackups)
	}

	if cfg.MaxAge < 0 {
		result.AddError("logging.max_age",
			"log max age must be non-negative", cfg.MaxAge)
	}
}
