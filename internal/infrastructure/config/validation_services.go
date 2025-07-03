// Package config provides validation utilities for Viper-based configuration
package config

import (
	"strings"
)

// =============================================================================
// EMAIL CONFIGURATION VALIDATION
// =============================================================================

// validateEmailConfig validates email configuration
func validateEmailConfig(cfg EmailConfig, result *ValidationResult) {
	validateEmailConfigHost(cfg, result)
	validateEmailConfigPort(cfg, result)
	validateEmailConfigCredentials(cfg, result)
	validateEmailConfigTLS(cfg, result)
}

func validateEmailConfigHost(cfg EmailConfig, result *ValidationResult) {
	if cfg.Host == "" {
		result.AddError("email.host", "email host is required", cfg.Host)
	}
}

func validateEmailConfigPort(cfg EmailConfig, result *ValidationResult) {
	if cfg.Port <= 0 || cfg.Port > 65535 {
		result.AddError("email.port", "email port must be between 1 and 65535", cfg.Port)
	}
}

func validateEmailConfigCredentials(cfg EmailConfig, result *ValidationResult) {
	if cfg.Username == "" {
		result.AddError("email.username", "email username is required", cfg.Username)
	}

	if cfg.Password == "" {
		result.AddError("email.password", "email password is required", cfg.Password)
	}
}

func validateEmailConfigTLS(cfg EmailConfig, result *ValidationResult) {
	if cfg.UseTLS {
		// TLS validation logic if needed
	}
}

// =============================================================================
// CACHE CONFIGURATION VALIDATION
// =============================================================================

// validateCacheConfig validates cache configuration
func validateCacheConfig(cfg CacheConfig, result *ValidationResult) {
	validateCacheConfigType(cfg, result)
	validateCacheConfigTTL(cfg, result)
	validateCacheConfigRedis(cfg, result)
	validateCacheConfigMemory(cfg, result)
}

func validateCacheConfigType(cfg CacheConfig, result *ValidationResult) {
	validTypes := []string{"memory", "redis"}
	typeValid := false

	for _, cacheType := range validTypes {
		if strings.EqualFold(cfg.Type, cacheType) {
			typeValid = true

			break
		}
	}

	if !typeValid {
		result.AddError("cache.type", "cache type must be one of: memory, redis", cfg.Type)
	}
}

func validateCacheConfigTTL(cfg CacheConfig, result *ValidationResult) {
	if cfg.TTL <= 0 {
		result.AddError("cache.ttl", "cache TTL must be positive", cfg.TTL)
	}
}

func validateCacheConfigRedis(cfg CacheConfig, result *ValidationResult) {
	if strings.EqualFold(cfg.Type, "redis") {
		if cfg.Redis.Host == "" {
			result.AddError("cache.redis.host", "Redis host is required when using Redis cache", cfg.Redis.Host)
		}

		if cfg.Redis.Port <= 0 || cfg.Redis.Port > 65535 {
			result.AddError("cache.redis.port", "Redis port must be between 1 and 65535", cfg.Redis.Port)
		}

		if cfg.Redis.DB < 0 {
			result.AddError("cache.redis.db", "Redis database number must be non-negative", cfg.Redis.DB)
		}
	}
}

func validateCacheConfigMemory(cfg CacheConfig, result *ValidationResult) {
	if strings.EqualFold(cfg.Type, "memory") {
		if cfg.Memory.MaxSize <= 0 {
			result.AddError("cache.memory.max_size", "memory cache max size must be positive", cfg.Memory.MaxSize)
		}
	}
}

// =============================================================================
// STORAGE CONFIGURATION VALIDATION
// =============================================================================

// validateStorageConfig validates storage configuration
func validateStorageConfig(cfg StorageConfig, result *ValidationResult) {
	validateStorageConfigType(cfg, result)
	validateStorageConfigLocal(cfg, result)
	validateStorageConfigS3(cfg, result)
}

func validateStorageConfigType(cfg StorageConfig, result *ValidationResult) {
	validTypes := []string{"local", "s3"}
	typeValid := false

	for _, storageType := range validTypes {
		if strings.EqualFold(cfg.Type, storageType) {
			typeValid = true

			break
		}
	}

	if !typeValid {
		result.AddError("storage.type", "storage type must be one of: local, s3", cfg.Type)
	}
}

func validateStorageConfigLocal(cfg StorageConfig, result *ValidationResult) {
	if strings.EqualFold(cfg.Type, "local") {
		if cfg.Local.Path == "" {
			result.AddError("storage.local.path", "local storage path is required", cfg.Local.Path)
		}

		if !isWritableDirectory(cfg.Local.Path) {
			result.AddError("storage.local.path", "local storage path must be a writable directory", cfg.Local.Path)
		}
	}
}

func validateStorageConfigS3(cfg StorageConfig, result *ValidationResult) {
	if strings.EqualFold(cfg.Type, "s3") {
		if cfg.S3.Bucket == "" {
			result.AddError("storage.s3.bucket", "S3 bucket is required", cfg.S3.Bucket)
		}

		if cfg.S3.Region == "" {
			result.AddError("storage.s3.region", "S3 region is required", cfg.S3.Region)
		}

		if cfg.S3.AccessKey == "" {
			result.AddError("storage.s3.access_key", "S3 access key is required", cfg.S3.AccessKey)
		}

		if cfg.S3.SecretKey == "" {
			result.AddError("storage.s3.secret_key", "S3 secret key is required", cfg.S3.SecretKey)
		}
	}
}

// =============================================================================
// LOGGING CONFIGURATION VALIDATION
// =============================================================================

// validateLoggingConfig validates logging configuration
func validateLoggingConfig(cfg LoggingConfig, result *ValidationResult) {
	validateLoggingConfigLevel(cfg, result)
	validateLoggingConfigOutput(cfg, result)
	validateLoggingConfigFile(cfg, result)
}

func validateLoggingConfigLevel(cfg LoggingConfig, result *ValidationResult) {
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	levelValid := false

	for _, level := range validLevels {
		if strings.EqualFold(cfg.Level, level) {
			levelValid = true

			break
		}
	}

	if !levelValid {
		result.AddError("logging.level", "log level must be one of: debug, info, warn, error, fatal", cfg.Level)
	}
}

func validateLoggingConfigOutput(cfg LoggingConfig, result *ValidationResult) {
	validOutputs := []string{"stdout", "stderr", "file"}
	outputValid := false

	for _, output := range validOutputs {
		if strings.EqualFold(cfg.Output, output) {
			outputValid = true

			break
		}
	}

	if !outputValid {
		result.AddError("logging.output", "log output must be one of: stdout, stderr, file", cfg.Output)
	}
}

func validateLoggingConfigFile(cfg LoggingConfig, result *ValidationResult) {
	if strings.EqualFold(cfg.Output, "file") {
		if cfg.File == "" {
			result.AddError("logging.file", "log file path is required when using file output", cfg.File)
		}

		if cfg.MaxSize <= 0 {
			result.AddError("logging.max_size", "log file max size must be positive", cfg.MaxSize)
		}

		if cfg.MaxBackups <= 0 {
			result.AddError("logging.max_backups", "log file max backups must be positive", cfg.MaxBackups)
		}
	}
}

// =============================================================================
// SESSION CONFIGURATION VALIDATION
// =============================================================================

// validateSessionConfig validates session configuration
func validateSessionConfig(cfg SessionConfig, result *ValidationResult) {
	validateSessionConfigStore(cfg, result)
	validateSessionConfigTTL(cfg, result)
	validateSessionConfigRedis(cfg, result)
}

func validateSessionConfigStore(cfg SessionConfig, result *ValidationResult) {
	validStores := []string{"memory", "redis", "database"}
	storeValid := false

	for _, store := range validStores {
		if strings.EqualFold(cfg.Store, store) {
			storeValid = true

			break
		}
	}

	if !storeValid {
		result.AddError("session.store", "session store must be one of: memory, redis, database", cfg.Store)
	}
}

func validateSessionConfigTTL(cfg SessionConfig, result *ValidationResult) {
	if cfg.MaxAge <= 0 {
		result.AddError("session.max_age", "session max age must be positive", cfg.MaxAge)
	}
}

func validateSessionConfigRedis(cfg SessionConfig, result *ValidationResult) {
	// Redis session validation is handled by the session store implementation
	// SessionConfig doesn't have direct Redis fields
}
