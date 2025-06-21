package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents the complete application configuration
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Security SecurityConfig
	Email    EmailConfig
	Storage  StorageConfig
	Cache    CacheConfig
	Logging  LoggingConfig
	Session  SessionConfig
	Auth     AuthConfig
	Form     FormConfig
	API      APIConfig
	Web      WebConfig
	User     UserConfig
}

// AppConfig holds application-level configuration
type AppConfig struct {
	// Application Info
	Name     string `envconfig:"GOFORMS_APP_NAME" default:"GoFormX"`
	Env      string `envconfig:"GOFORMS_APP_ENV" default:"production"`
	Debug    bool   `envconfig:"GOFORMS_APP_DEBUG" default:"false"`
	LogLevel string `envconfig:"GOFORMS_APP_LOGLEVEL" default:"info"`

	// Server Settings
	Scheme         string        `envconfig:"GOFORMS_APP_SCHEME" default:"http"`
	Port           int           `envconfig:"GOFORMS_APP_PORT" default:"8090"`
	Host           string        `envconfig:"GOFORMS_APP_HOST" default:"0.0.0.0"`
	ReadTimeout    time.Duration `envconfig:"GOFORMS_APP_READ_TIMEOUT" default:"5s"`
	WriteTimeout   time.Duration `envconfig:"GOFORMS_APP_WRITE_TIMEOUT" default:"10s"`
	IdleTimeout    time.Duration `envconfig:"GOFORMS_APP_IDLE_TIMEOUT" default:"120s"`
	RequestTimeout time.Duration `envconfig:"GOFORMS_APP_REQUEST_TIMEOUT" default:"30s"`

	// Development Settings
	ViteDevHost string `envconfig:"GOFORMS_VITE_DEV_HOST" default:"localhost"`
	ViteDevPort string `envconfig:"GOFORMS_VITE_DEV_PORT" default:"3000"`
}

// IsDevelopment returns true if the application is running in development mode
func (c *AppConfig) IsDevelopment() bool {
	return strings.EqualFold(c.Env, "development")
}

// DatabaseConfig holds all database-related configuration
type DatabaseConfig struct {
	// Common database settings
	Connection      string        `envconfig:"GOFORMS_DB_CONNECTION" default:"mariadb"`
	Host            string        `envconfig:"GOFORMS_DB_HOST" validate:"required"`
	Port            int           `envconfig:"GOFORMS_DB_PORT" default:"3306"`
	Database        string        `envconfig:"GOFORMS_DB_DATABASE" validate:"required"`
	Username        string        `envconfig:"GOFORMS_DB_USERNAME" validate:"required"`
	Password        string        `envconfig:"GOFORMS_DB_PASSWORD" validate:"required"`
	MaxOpenConns    int           `envconfig:"GOFORMS_DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int           `envconfig:"GOFORMS_DB_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime time.Duration `envconfig:"GOFORMS_DB_CONN_MAX_LIFETIME" default:"5m"`

	// PostgreSQL specific settings
	SSLMode string `envconfig:"GOFORMS_DB_SSLMODE" default:"disable"`

	// MariaDB specific settings
	RootPassword string `envconfig:"GOFORMS_DB_ROOT_PASSWORD"`

	// Logging configuration
	Logging struct {
		// SlowThreshold is the threshold for logging slow queries
		SlowThreshold time.Duration `envconfig:"GOFORMS_DB_SLOW_THRESHOLD" default:"1s"`
		// Parameterized enables logging of query parameters
		Parameterized bool `envconfig:"GOFORMS_DB_LOG_PARAMETERS" default:"false"`
		// IgnoreNotFound determines whether to ignore record not found errors
		IgnoreNotFound bool `envconfig:"GOFORMS_DB_IGNORE_NOT_FOUND" default:"false"`
		// LogLevel determines the verbosity of database logging
		// Valid values: "silent", "error", "warn", "info"
		LogLevel string `envconfig:"GOFORMS_DB_LOG_LEVEL" default:"warn"`
	} `envconfig:"GOFORMS_DB_LOGGING"`
}

// CSRFConfig holds CSRF-related configuration
type CSRFConfig struct {
	Enabled        bool   `envconfig:"GOFORMS_SECURITY_CSRF_ENABLED" default:"true"`
	Secret         string `envconfig:"GOFORMS_SECURITY_CSRF_SECRET" validate:"required"`
	TokenLength    int    `envconfig:"GOFORMS_SECURITY_CSRF_TOKEN_LENGTH" default:"32"`
	TokenLookup    string `envconfig:"GOFORMS_SECURITY_CSRF_TOKEN_LOOKUP" default:"header:X-Csrf-Token"`
	ContextKey     string `envconfig:"GOFORMS_SECURITY_CSRF_CONTEXT_KEY" default:"csrf"`
	CookieName     string `envconfig:"GOFORMS_SECURITY_CSRF_COOKIE_NAME" default:"_csrf"`
	CookiePath     string `envconfig:"GOFORMS_SECURITY_CSRF_COOKIE_PATH" default:"/"`
	CookieDomain   string `envconfig:"GOFORMS_SECURITY_CSRF_COOKIE_DOMAIN" default:""`
	CookieHTTPOnly bool   `envconfig:"GOFORMS_SECURITY_CSRF_COOKIE_HTTP_ONLY" default:"true"`
	CookieSameSite string `envconfig:"GOFORMS_SECURITY_CSRF_COOKIE_SAME_SITE" default:"Lax"`
	CookieMaxAge   int    `envconfig:"GOFORMS_SECURITY_CSRF_COOKIE_MAX_AGE" default:"86400"`
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	// CSRF protection
	CSRF CSRFConfig `envconfig:"CSRF"`

	// CORS configuration
	CORS CORSConfig `envconfig:"CORS"`

	// Rate limiting configuration
	RateLimit RateLimitConfig `envconfig:"RATE_LIMIT"`

	// Security headers configuration
	Headers SecurityHeadersConfig `envconfig:"HEADERS"`

	// Content Security Policy configuration
	CSP CSPConfig `envconfig:"CSP"`

	// Cookie security
	SecureCookie bool `envconfig:"GOFORMS_SECURITY_SECURE_COOKIE" default:"true"`

	// Debug mode
	Debug bool `envconfig:"GOFORMS_SECURITY_DEBUG" default:"false"`
}

// CORSConfig holds CORS-related configuration
type CORSConfig struct {
	Enabled          bool     `envconfig:"GOFORMS_SECURITY_CORS_ENABLED" default:"true"`
	AllowedOrigins   []string `envconfig:"GOFORMS_SECURITY_CORS_ORIGINS" default:"http://localhost:3000"`
	AllowedMethods   []string `envconfig:"GOFORMS_SECURITY_CORS_METHODS" default:"GET,POST,PUT,DELETE,OPTIONS"`
	AllowedHeaders   []string `envconfig:"GOFORMS_SECURITY_CORS_HEADERS" default:"Content-Type,Authorization,X-Csrf-Token,X-Requested-With"`
	AllowCredentials bool     `envconfig:"GOFORMS_SECURITY_CORS_CREDENTIALS" default:"true"`
	MaxAge           int      `envconfig:"GOFORMS_SECURITY_CORS_MAX_AGE" default:"3600"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool          `envconfig:"GOFORMS_SECURITY_RATE_LIMIT_ENABLED" default:"true"`
	Requests    int           `envconfig:"GOFORMS_SECURITY_RATE_LIMIT_REQUESTS" default:"100"`
	Window      time.Duration `envconfig:"GOFORMS_SECURITY_RATE_LIMIT_WINDOW" default:"1m"`
	Burst       int           `envconfig:"GOFORMS_SECURITY_RATE_LIMIT_BURST" default:"20"`
	PerIP       bool          `envconfig:"GOFORMS_SECURITY_RATE_LIMIT_PER_IP" default:"true"`
	SkipPaths   []string      `envconfig:"GOFORMS_SECURITY_RATE_LIMIT_SKIP_PATHS"`
	SkipMethods []string      `envconfig:"GOFORMS_SECURITY_RATE_LIMIT_SKIP_METHODS" default:"GET,HEAD,OPTIONS"`
}

// SecurityHeadersConfig holds security headers configuration
type SecurityHeadersConfig struct {
	XFrameOptions           string `envconfig:"GOFORMS_SECURITY_X_FRAME_OPTIONS" default:"DENY"`
	XContentTypeOptions     string `envconfig:"GOFORMS_SECURITY_X_CONTENT_TYPE_OPTIONS" default:"nosniff"`
	XXSSProtection          string `envconfig:"GOFORMS_SECURITY_X_XSS_PROTECTION" default:"1; mode=block"`
	ReferrerPolicy          string `envconfig:"GOFORMS_SECURITY_REFERRER_POLICY" default:"strict-origin-when-cross-origin"`
	StrictTransportSecurity string `envconfig:"GOFORMS_SECURITY_HSTS" default:"max-age=31536000; includeSubDomains"`
}

// CSPConfig holds Content Security Policy configuration
type CSPConfig struct {
	Enabled    bool   `envconfig:"GOFORMS_SECURITY_CSP_ENABLED" default:"true"`
	Directives string `envconfig:"GOFORMS_SECURITY_CSP_DIRECTIVES"`
}

// GetCSPDirectives returns the Content Security Policy directives based on environment
func (s *SecurityConfig) GetCSPDirectives(appConfig *AppConfig) string {
	if appConfig.IsDevelopment() {
		return "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' http://localhost:3000; " +
			"style-src 'self' 'unsafe-inline' http://localhost:3000; " +
			"img-src 'self' data:; " +
			"font-src 'self' http://localhost:3000; " +
			"connect-src 'self' http://localhost:3000 ws://localhost:3000; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"
	}

	// If custom CSP directives are provided via environment, use them
	if s.CSP.Directives != "" {
		return s.CSP.Directives
	}

	// Generate CSP directives based on environment
	return "default-src 'self'; " +
		"script-src 'self' 'unsafe-inline'; " +
		"style-src 'self' 'unsafe-inline'; " +
		"img-src 'self' data:; " +
		"font-src 'self'; " +
		"connect-src 'self'; " +
		"frame-ancestors 'none'; " +
		"base-uri 'self'; " +
		"form-action 'self'"
}

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
	StoreFile  string        `envconfig:"GOFORMS_SESSION_STORE_FILE" default:"tmp/sessions.json"`
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	PasswordCost int `envconfig:"GOFORMS_PASSWORD_COST" default:"12"`
}

// FormConfig holds form-related configuration
type FormConfig struct {
	MaxFileSize    int64    `envconfig:"GOFORMS_MAX_FILE_SIZE" default:"10485760"` // 10MB
	AllowedTypes   []string `envconfig:"GOFORMS_ALLOWED_FILE_TYPES" default:"image/jpeg,image/png,application/pdf"`
	MaxSubmissions int      `envconfig:"GOFORMS_MAX_SUBMISSIONS" default:"1000"`
	RetentionDays  int      `envconfig:"GOFORMS_RETENTION_DAYS" default:"90"`
}

// APIConfig holds API-related configuration
type APIConfig struct {
	Version   string `envconfig:"GOFORMS_API_VERSION" default:"v1"`
	Prefix    string `envconfig:"GOFORMS_API_PREFIX" default:"/api"`
	RateLimit int    `envconfig:"GOFORMS_API_RATE_LIMIT" default:"100"`
	Timeout   int    `envconfig:"GOFORMS_API_TIMEOUT" default:"30"`
}

// WebConfig holds web-related configuration
type WebConfig struct {
	BaseURL      string `envconfig:"GOFORMS_WEB_BASE_URL" default:"http://localhost:8090"`
	AssetsDir    string `envconfig:"GOFORMS_WEB_ASSETS_DIR" default:"./assets"`
	TemplatesDir string `envconfig:"GOFORMS_WEB_TEMPLATES_DIR" default:"./templates"`
}

// UserConfig holds user-related configuration
type UserConfig struct {
	Admin struct {
		Email     string `envconfig:"GOFORMS_ADMIN_EMAIL" validate:"required,email"`
		Password  string `envconfig:"GOFORMS_ADMIN_PASSWORD" validate:"required"`
		FirstName string `envconfig:"GOFORMS_ADMIN_FIRST_NAME" validate:"required"`
		LastName  string `envconfig:"GOFORMS_ADMIN_LAST_NAME" validate:"required"`
	} `envconfig:"GOFORMS_ADMIN"`

	Default struct {
		Email     string `envconfig:"GOFORMS_USER_EMAIL" validate:"required,email"`
		Password  string `envconfig:"GOFORMS_USER_PASSWORD" validate:"required"`
		FirstName string `envconfig:"GOFORMS_USER_FIRST_NAME" validate:"required"`
		LastName  string `envconfig:"GOFORMS_USER_LAST_NAME" validate:"required"`
	} `envconfig:"GOFORMS_USER"`
}

// Validation errors
var (
	ErrMissingAppName    = errors.New("application name is required")
	ErrInvalidPort       = errors.New("port must be between 1 and 65535")
	ErrMissingDBDriver   = errors.New("database driver is required")
	ErrMissingDBHost     = errors.New("database host is required")
	ErrMissingDBUser     = errors.New("database user is required")
	ErrMissingDBPassword = errors.New("database password is required")
	ErrMissingDBName     = errors.New("database name is required")
	ErrMissingCSRFSecret = errors.New("CSRF secret is required when CSRF is enabled")
	ErrInvalidTimeout    = errors.New("timeout duration must be positive")
	ErrInvalidRateLimit  = errors.New("rate limit must be positive")
	ErrInvalidMaxConns   = errors.New("max connections must be positive")
	ErrInvalidDBLogLevel = errors.New("invalid database log level")
)

// validateAppConfig validates the application configuration
func (c *Config) validateAppConfig() error {
	var errs []string

	if c.App.Name == "" {
		errs = append(errs, "app name is required")
	}
	if c.App.Port <= 0 || c.App.Port > 65535 {
		errs = append(errs, "app port must be between 1 and 65535")
	}
	if c.App.ReadTimeout <= 0 {
		errs = append(errs, "read timeout must be positive")
	}
	if c.App.WriteTimeout <= 0 {
		errs = append(errs, "write timeout must be positive")
	}
	if c.App.IdleTimeout <= 0 {
		errs = append(errs, "idle timeout must be positive")
	}

	if len(errs) > 0 {
		return fmt.Errorf("app config validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// validateDatabaseConfig validates database configuration
func (c *Config) validateDatabaseConfig() error {
	var errs []string

	if c.Database.Host == "" {
		errs = append(errs, "database host is required")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		errs = append(errs, "database port must be between 1 and 65535")
	}
	if c.Database.Username == "" {
		errs = append(errs, "database username is required")
	}
	if c.Database.Password == "" {
		errs = append(errs, "database password is required")
	}
	if c.Database.Database == "" {
		errs = append(errs, "database name is required")
	}

	// Validate database-specific settings
	switch c.Database.Connection {
	case "postgres":
		if c.Database.SSLMode == "" {
			errs = append(errs, "PostgreSQL SSL mode is required")
		}
	case "mariadb":
		if c.Database.RootPassword == "" {
			errs = append(errs, "MariaDB root password is required")
		}
	default:
		errs = append(errs, "unsupported database connection type")
	}

	if len(errs) > 0 {
		return fmt.Errorf("database config validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// validateSecurityConfig validates security configuration
func (c *Config) validateSecurityConfig() error {
	var errs []string

	if c.Security.CSRF.Enabled && c.Security.CSRF.Secret == "" {
		errs = append(errs, "CSRF secret is required when CSRF is enabled")
	}

	if len(errs) > 0 {
		return fmt.Errorf("security config validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// validateConfig validates the configuration
func (c *Config) validateConfig() error {
	var errs []string

	// Validate App config
	if err := c.validateAppConfig(); err != nil {
		errs = append(errs, err.Error())
	}

	// Validate Database config
	if err := c.validateDatabaseConfig(); err != nil {
		errs = append(errs, err.Error())
	}

	// Validate Security config
	if err := c.validateSecurityConfig(); err != nil {
		errs = append(errs, err.Error())
	}

	// Validate Session config only if session type is not "none"
	if c.Session.Type != "none" && c.Session.Secret == "" {
		errs = append(errs, "session secret is required when session type is not 'none'")
	}

	// Validate Email config only if email host is set
	if c.Email.Host != "" {
		if c.Email.Username == "" {
			errs = append(errs, "Email username is required when email host is set")
		}
		if c.Email.Password == "" {
			errs = append(errs, "Email password is required when email host is set")
		}
		if c.Email.From == "" {
			errs = append(errs, "Email from address is required when email host is set")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// New creates a new Config instance
func New() (*Config, error) {
	var config Config

	// Load environment variables
	if err := envconfig.Process("GOFORMS", &config); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Validate required fields
	if err := config.validateConfig(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}
