// Package config provides configuration helpers.
package config

// Individual config providers for fine-grained dependency injection

// NewAppConfig provides app configuration
func NewAppConfig(cfg *Config) AppConfig {
	return cfg.App
}

// NewDatabaseConfig provides database configuration
func NewDatabaseConfig(cfg *Config) DatabaseConfig {
	return cfg.Database
}

// NewSecurityConfig provides security configuration
func NewSecurityConfig(cfg *Config) SecurityConfig {
	return cfg.Security
}

// NewEmailConfig provides email configuration
func NewEmailConfig(cfg *Config) EmailConfig {
	return cfg.Email
}

// NewStorageConfig provides storage configuration
func NewStorageConfig(cfg *Config) StorageConfig {
	return cfg.Storage
}

// NewCacheConfig provides cache configuration
func NewCacheConfig(cfg *Config) CacheConfig {
	return cfg.Cache
}

// NewLoggingConfig provides logging configuration
func NewLoggingConfig(cfg *Config) LoggingConfig {
	return cfg.Logging
}

// NewSessionConfig provides session configuration
func NewSessionConfig(cfg *Config) SessionConfig {
	return cfg.Session
}

// NewAuthConfig provides authentication configuration
func NewAuthConfig(cfg *Config) AuthConfig {
	return cfg.Auth
}

// NewFormConfig provides form configuration
func NewFormConfig(cfg *Config) FormConfig {
	return cfg.Form
}

// NewAPIConfig provides API configuration
func NewAPIConfig(cfg *Config) APIConfig {
	return cfg.API
}

// NewWebConfig provides web configuration
func NewWebConfig(cfg *Config) WebConfig {
	return cfg.Web
}

// NewUserConfig provides user configuration
func NewUserConfig(cfg *Config) UserConfig {
	return cfg.User
}
