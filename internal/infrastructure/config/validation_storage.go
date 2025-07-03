// Package config provides validation utilities for Viper-based configuration
package config

import (
	"strings"
)

// validateStorageConfig validates storage configuration
func validateStorageConfig(cfg StorageConfig, result *ValidationResult) {
	validateStorageType(cfg, result)
	validateStorageLocal(cfg, result)
	validateStorageLimits(cfg, result)
}

func validateStorageType(cfg StorageConfig, result *ValidationResult) {
	if cfg.Type == "" {
		result.AddError("storage.type", "storage type is required", cfg.Type)

		return
	}

	supportedTypes := []string{"local"}
	for _, storageType := range supportedTypes {
		if strings.EqualFold(cfg.Type, storageType) {
			return
		}
	}

	result.AddError("storage.type", "unsupported storage type", cfg.Type)
}

func validateStorageLocal(cfg StorageConfig, result *ValidationResult) {
	if !strings.EqualFold(cfg.Type, "local") {
		return
	}

	if cfg.Local.Path == "" {
		result.AddError("storage.local.path",
			"local storage path is required", cfg.Local.Path)

		return
	}

	if !isWritableDirectory(cfg.Local.Path) {
		result.AddError("storage.local.path",
			"local storage path must be a writable directory", cfg.Local.Path)
	}
}

func validateStorageLimits(cfg StorageConfig, result *ValidationResult) {
	if cfg.MaxSize <= 0 {
		result.AddError("storage.max_size",
			"storage max size must be positive", cfg.MaxSize)
	}
}
