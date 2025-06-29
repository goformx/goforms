package config

import (
	"time"
)

// EmailConfig holds email-related configuration
type EmailConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	From       string `json:"from"`
	UseTLS     bool   `json:"use_tls"`
	UseSSL     bool   `json:"use_ssl"`
	Template   string `json:"template"`
	Timeout    int    `json:"timeout"`
	MaxRetries int    `json:"max_retries"`
}

// StorageConfig holds storage-related configuration
type StorageConfig struct {
	Type        string             `json:"type"`
	Local       LocalStorageConfig `json:"local"`
	S3          S3StorageConfig    `json:"s3"`
	MaxSize     int64              `json:"max_size"`
	AllowedExts []string           `json:"allowed_exts"`
}

// LocalStorageConfig holds local storage configuration
type LocalStorageConfig struct {
	Path string `json:"path"`
}

// S3StorageConfig holds S3 storage configuration
type S3StorageConfig struct {
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Endpoint  string `json:"endpoint"`
}

// CacheConfig holds cache-related configuration
type CacheConfig struct {
	Type   string        `json:"type"`
	Redis  RedisConfig   `json:"redis"`
	Memory MemoryConfig  `json:"memory"`
	TTL    time.Duration `json:"ttl"`
}

// RedisConfig holds Redis cache configuration
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// MemoryConfig holds memory cache configuration
type MemoryConfig struct {
	MaxSize int `json:"max_size"`
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	File       string `json:"file"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
	Compress   bool   `json:"compress"`
}

// SessionConfig holds session-related configuration
type SessionConfig struct {
	Type       string        `json:"type"`
	Secret     string        `json:"secret"`
	MaxAge     time.Duration `json:"max_age"`
	Domain     string        `json:"domain"`
	Path       string        `json:"path"`
	Secure     bool          `json:"secure"`
	HTTPOnly   bool          `json:"http_only"`
	SameSite   string        `json:"same_site"`
	Store      string        `json:"store"`
	StoreFile  string        `json:"store_file"`
	CookieName string        `json:"cookie_name"`
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	RequireEmailVerification bool          `json:"require_email_verification"`
	PasswordMinLength        int           `json:"password_min_length"`
	PasswordRequireSpecial   bool          `json:"password_require_special"`
	SessionTimeout           time.Duration `json:"session_timeout"`
	MaxLoginAttempts         int           `json:"max_login_attempts"`
	LockoutDuration          time.Duration `json:"lockout_duration"`
}
