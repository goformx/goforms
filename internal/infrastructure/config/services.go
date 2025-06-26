package config

import (
	"time"
)

// EmailConfig holds email-related configuration
type EmailConfig struct {
	Host     string `envconfig:"GOFORMS_EMAIL_HOST"`
	Port     int    `envconfig:"GOFORMS_EMAIL_PORT" default:"587"`
	Username string `envconfig:"GOFORMS_EMAIL_USERNAME"`
	Password string `envconfig:"GOFORMS_EMAIL_PASSWORD"`
	From     string `envconfig:"GOFORMS_EMAIL_FROM"`
}

// StorageConfig holds storage-related configuration
type StorageConfig struct {
	Type     string `envconfig:"GOFORMS_STORAGE_TYPE" default:"local"`
	LocalDir string `envconfig:"GOFORMS_STORAGE_LOCAL_DIR" default:"./storage"`
}

// CacheConfig holds cache-related configuration
type CacheConfig struct {
	Type    string        `envconfig:"GOFORMS_CACHE_TYPE" default:"memory"`
	TTL     time.Duration `envconfig:"GOFORMS_CACHE_TTL" default:"1h"`
	MaxSize int           `envconfig:"GOFORMS_CACHE_MAX_SIZE" default:"1000"`
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level      string `envconfig:"GOFORMS_LOG_LEVEL" default:"info"`
	Format     string `envconfig:"GOFORMS_LOG_FORMAT" default:"json"`
	Output     string `envconfig:"GOFORMS_LOG_OUTPUT" default:"stdout"`
	MaxSize    int    `envconfig:"GOFORMS_LOG_MAX_SIZE" default:"100"`
	MaxBackups int    `envconfig:"GOFORMS_LOG_MAX_BACKUPS" default:"3"`
	MaxAge     int    `envconfig:"GOFORMS_LOG_MAX_AGE" default:"28"`
	Compress   bool   `envconfig:"GOFORMS_LOG_COMPRESS" default:"true"`
}

// SessionConfig holds session-related configuration
type SessionConfig struct {
	Type       string        `envconfig:"GOFORMS_SESSION_TYPE" default:"none"`
	Secret     string        `envconfig:"GOFORMS_SESSION_SECRET"`
	TTL        time.Duration `envconfig:"GOFORMS_SESSION_TTL" default:"24h"`
	Secure     bool          `envconfig:"GOFORMS_SESSION_SECURE" default:"true"`
	HTTPOnly   bool          `envconfig:"GOFORMS_SESSION_HTTP_ONLY" default:"true"`
	CookieName string        `envconfig:"GOFORMS_SESSION_COOKIE_NAME" default:"session"`
	StoreFile  string        `envconfig:"GOFORMS_SESSION_STORE_FILE" default:"storage/sessions/sessions.json"`
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	PasswordCost int `envconfig:"GOFORMS_PASSWORD_COST" default:"12"`
}
