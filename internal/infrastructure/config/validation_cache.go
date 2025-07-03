// Package config provides validation utilities for Viper-based configuration
package config

import (
	"strings"
)

// validateCacheConfig validates cache configuration
func validateCacheConfig(cfg CacheConfig, result *ValidationResult) {
	if cfg.Type == "" {
		result.AddError("cache.type", "cache type is required", cfg.Type)
	}

	// Validate cache type
	supportedTypes := []string{"memory", "redis"}
	typeValid := false

	for _, cacheType := range supportedTypes {
		if strings.EqualFold(cfg.Type, cacheType) {
			typeValid = true

			break
		}
	}

	if !typeValid {
		result.AddError("cache.type", "unsupported cache type", cfg.Type)
	}

	// Validate Redis configuration
	if strings.EqualFold(cfg.Type, "redis") {
		if cfg.Redis.Host == "" {
			result.AddError("cache.redis.host",
				"Redis host is required for Redis cache", cfg.Redis.Host)
		}

		if cfg.Redis.Port <= 0 || cfg.Redis.Port > 65535 {
			result.AddError("cache.redis.port",
				"Redis port must be between 1 and 65535", cfg.Redis.Port)
		}

		if cfg.Redis.DB < 0 {
			result.AddError("cache.redis.db",
				"Redis database number must be non-negative", cfg.Redis.DB)
		}
	}

	// Validate TTL
	if cfg.TTL <= 0 {
		result.AddError("cache.ttl", "cache TTL must be positive", cfg.TTL)
	}
}
