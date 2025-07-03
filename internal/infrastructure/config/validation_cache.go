// Package config provides validation utilities for Viper-based configuration
package config

import (
	"strings"
)

// validateCacheConfig validates cache configuration
func validateCacheConfig(cfg CacheConfig, result *ValidationResult) {
	validateCacheType(cfg, result)
	validateCacheRedis(cfg, result)
	validateCacheTTL(cfg, result)
}

func validateCacheType(cfg CacheConfig, result *ValidationResult) {
	if cfg.Type == "" {
		result.AddError("cache.type", "cache type is required", cfg.Type)
	}

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
}

func validateCacheRedis(cfg CacheConfig, result *ValidationResult) {
	if !strings.EqualFold(cfg.Type, "redis") {
		return
	}

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

func validateCacheTTL(cfg CacheConfig, result *ValidationResult) {
	if cfg.TTL <= 0 {
		result.AddError("cache.ttl", "cache TTL must be positive", cfg.TTL)
	}
}
