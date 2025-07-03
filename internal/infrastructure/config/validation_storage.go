// Package config provides validation utilities for Viper-based configuration
package config

import (
	"strings"
)

// validateStorageConfig validates storage configuration
func validateStorageConfig(cfg StorageConfig, result *ValidationResult) {
	validateStorageType(cfg, result)
	validateStorageLocal(cfg, result)
	validateStorageS3(cfg, result)
	validateStorageLimits(cfg, result)
}

func validateStorageType(cfg StorageConfig, result *ValidationResult) {
	if cfg.Type == "" {
		result.AddError("storage.type", "storage type is required", cfg.Type)

		return
	}

	supportedTypes := []string{"local", "s3"}
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

func validateStorageS3(cfg StorageConfig, result *ValidationResult) {
	if !strings.EqualFold(cfg.Type, "s3") {
		return
	}

	if cfg.S3.Bucket == "" {
		result.AddError("storage.s3.bucket",
			"S3 bucket is required for S3 storage", cfg.S3.Bucket)
	}

	if cfg.S3.Region == "" {
		result.AddError("storage.s3.region",
			"S3 region is required for S3 storage", cfg.S3.Region)
	}

	if cfg.S3.AccessKey == "" {
		result.AddError("storage.s3.access_key",
			"S3 access key is required for S3 storage", "***")
	}

	if cfg.S3.SecretKey == "" {
		result.AddError("storage.s3.secret_key",
			"S3 secret key is required for S3 storage", "***")
	}
}

func validateStorageLimits(cfg StorageConfig, result *ValidationResult) {
	if cfg.MaxSize <= 0 {
		result.AddError("storage.max_size",
			"storage max size must be positive", cfg.MaxSize)
	}
}
