// Package config provides configuration module for Fx dependency injection
package config

import (
	"go.uber.org/fx"
)

// Module provides the configuration module for Fx
var Module = fx.Module("config",
	// Use Viper configuration provider instead of LoadFromEnv
	NewViperConfigProvider(),
	fx.Provide(NewAppConfig),
	fx.Provide(NewDatabaseConfig),
	fx.Provide(NewSecurityConfig),
	fx.Provide(NewEmailConfig),
	fx.Provide(NewStorageConfig),
	fx.Provide(NewCacheConfig),
	fx.Provide(NewLoggingConfig),
	fx.Provide(NewSessionConfig),
	fx.Provide(NewAuthConfig),
	fx.Provide(NewFormConfig),
	fx.Provide(NewAPIConfig),
	fx.Provide(NewWebConfig),
	fx.Provide(NewUserConfig),
)

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
