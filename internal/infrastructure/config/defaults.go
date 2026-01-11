// Package config provides default configuration values.
package config

import "time"

// Default ports
const (
	DefaultAppPort   = 8080
	DefaultDBPort    = 5432
	DefaultRedisPort = 6379
	DefaultSMTPPort  = 587
)

// Default timeouts
const (
	DefaultReadTimeout    = 15 * time.Second
	DefaultWriteTimeout   = 15 * time.Second
	DefaultIdleTimeout    = 60 * time.Second
	DefaultRequestTimeout = 30 * time.Second
	DefaultConnLifetime   = 5 * time.Minute
	DefaultConnIdleTime   = 5 * time.Minute
	DefaultSessionMaxAge  = 24 * time.Hour
	DefaultAuthTimeout    = 30 * time.Minute
	DefaultLockoutTime    = 15 * time.Minute
)

// Default connection pool settings
const (
	DefaultMaxOpenConns = 25
	DefaultMaxIdleConns = 25
)

// Default security settings
const (
	DefaultCSRFTokenLength = 32
	DefaultCookieMaxAge    = 86400 // 24 hours in seconds
	DefaultRateLimitRPS    = 100
	DefaultRateLimitBurst  = 200
	DefaultAPIRateLimitRPS = 1000
	DefaultAPIRateBurst    = 2000
)

// Default size limits
const (
	DefaultMaxFileSize     = 10 * 1024 * 1024 // 10MB
	DefaultMaxFormMemory   = 32 * 1024 * 1024 // 32MB
	DefaultMaxFields       = 100
	DefaultMaxErrors       = 10
	DefaultMemoryCacheSize = 1000
)

// Default logging settings
const (
	DefaultLogMaxSize    = 100 // MB
	DefaultLogMaxBackups = 3
	DefaultLogMaxAge     = 28 // days
)

// Default auth settings
const (
	DefaultPasswordMinLength = 8
	DefaultMaxLoginAttempts  = 5
	DefaultMaxRetries        = 3
)

// Validation thresholds
const (
	MinPasswordLengthThreshold = 6
	MinSecretLength            = 32
)
