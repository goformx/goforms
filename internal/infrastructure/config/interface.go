// Package config provides configuration interfaces and implementations.
package config

// ConfigInterface defines the interface for configuration access
//
//go:generate mockgen -typed -source=interface.go -destination=../../../test/mocks/config/mock_config.go -package=config
type ConfigInterface interface {
	// GetApp returns the application configuration
	GetApp() AppConfig

	// GetDatabase returns the database configuration
	GetDatabase() DatabaseConfig

	// GetSecurity returns the security configuration
	GetSecurity() SecurityConfig

	// GetEmail returns the email configuration
	GetEmail() EmailConfig

	// GetStorage returns the storage configuration
	GetStorage() StorageConfig

	// GetCache returns the cache configuration
	GetCache() CacheConfig

	// GetLogging returns the logging configuration
	GetLogging() LoggingConfig

	// GetSession returns the session configuration
	GetSession() SessionConfig

	// GetAuth returns the auth configuration
	GetAuth() AuthConfig

	// GetForm returns the form configuration
	GetForm() FormConfig

	// GetAPI returns the API configuration
	GetAPI() APIConfig

	// GetWeb returns the web configuration
	GetWeb() WebConfig

	// GetUser returns the user configuration
	GetUser() UserConfig

	// GetMiddleware returns the middleware configuration
	GetMiddleware() MiddlewareConfig

	// GetEnvironment returns the current environment
	GetEnvironment() string

	// IsProduction returns true if running in production
	IsProduction() bool

	// IsDevelopment returns true if running in development
	IsDevelopment() bool

	// IsStaging returns true if running in staging
	IsStaging() bool

	// IsTest returns true if running in test
	IsTest() bool

	// IsValid checks if the configuration is valid
	IsValid() bool

	// GetConfigSummary returns a summary of the current configuration
	GetConfigSummary() map[string]any
}
