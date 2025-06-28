// Package config provides configuration management for the GoForms application.
// It uses environment variables to configure various aspects of the application
// including database connections, security settings, logging, and more.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/fx"
)

// Config represents the complete application configuration
type Config struct {
	App      AppConfig      `json:"app"`
	Database DatabaseConfig `json:"database"`
	Security SecurityConfig `json:"security"`
	Email    EmailConfig    `json:"email"`
	Storage  StorageConfig  `json:"storage"`
	Cache    CacheConfig    `json:"cache"`
	Logging  LoggingConfig  `json:"logging"`
	Session  SessionConfig  `json:"session"`
	Auth     AuthConfig     `json:"auth"`
	Form     FormConfig     `json:"form"`
	API      APIConfig      `json:"api"`
	Web      WebConfig      `json:"web"`
	User     UserConfig     `json:"user"`
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	config := &Config{}

	// Load App config
	config.App = AppConfig{
		Name:           getEnvOrDefault("GOFORMS_APP_NAME", "GoForms"),
		Version:        getEnvOrDefault("GOFORMS_APP_VERSION", "1.0.0"),
		Environment:    getEnvOrDefault("GOFORMS_APP_ENV", "development"),
		Debug:          getEnvBool("GOFORMS_APP_DEBUG", true),
		LogLevel:       getEnvOrDefault("GOFORMS_APP_LOGLEVEL", "info"),
		URL:            getEnvOrDefault("GOFORMS_APP_URL", "http://localhost:8080"),
		Scheme:         getEnvOrDefault("GOFORMS_APP_SCHEME", "http"),
		Port:           getEnvInt("GOFORMS_APP_PORT", 8080),
		Host:           getEnvOrDefault("GOFORMS_APP_HOST", "localhost"),
		ReadTimeout:    getEnvDuration("GOFORMS_APP_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:   getEnvDuration("GOFORMS_APP_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:    getEnvDuration("GOFORMS_APP_IDLE_TIMEOUT", 60*time.Second),
		RequestTimeout: getEnvDuration("GOFORMS_APP_REQUEST_TIMEOUT", 30*time.Second),
		ViteDevHost:    getEnvOrDefault("GOFORMS_VITE_DEV_HOST", "localhost"),
		ViteDevPort:    getEnvOrDefault("GOFORMS_VITE_DEV_PORT", "5173"),
	}

	// Load Database config
	config.Database = DatabaseConfig{
		Driver:          getEnvOrDefault("GOFORMS_DB_CONNECTION", "postgres"),
		Host:            getEnvOrDefault("GOFORMS_DB_HOST", "localhost"),
		Port:            getEnvInt("GOFORMS_DB_PORT", 5432),
		Name:            getEnvOrDefault("GOFORMS_DB_DATABASE", "goforms"),
		Username:        getEnvOrDefault("GOFORMS_DB_USERNAME", "goforms"),
		Password:        getEnvOrDefault("GOFORMS_DB_PASSWORD", "goforms"),
		SSLMode:         getEnvOrDefault("GOFORMS_DB_SSLMODE", "disable"),
		MaxOpenConns:    getEnvInt("GOFORMS_DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvInt("GOFORMS_DB_MAX_IDLE_CONNS", 25),
		ConnMaxLifetime: getEnvDuration("GOFORMS_DB_CONN_MAX_LIFETIME", 5*time.Minute),
		ConnMaxIdleTime: getEnvDuration("GOFORMS_DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
	}

	// Load Security config
	config.Security = SecurityConfig{
		CSRF: CSRFConfig{
			Enabled:    getEnvBool("GOFORMS_SECURITY_CSRF_ENABLED", true),
			Secret:     getEnvOrDefault("GOFORMS_SECURITY_CSRF_SECRET", "csrf-secret"),
			TokenName:  getEnvOrDefault("GOFORMS_SECURITY_CSRF_TOKEN_NAME", "_token"),
			HeaderName: getEnvOrDefault("GOFORMS_SECURITY_CSRF_HEADER_NAME", "X-CSRF-Token"),
		},
		CORS: CORSConfig{
			Enabled:          getEnvBool("GOFORMS_SECURITY_CORS_ENABLED", true),
			AllowedOrigins:   getEnvStringSlice("GOFORMS_SECURITY_CORS_ALLOWED_ORIGINS", []string{"*"}),
			AllowedMethods:   getEnvStringSlice("GOFORMS_SECURITY_CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders:   getEnvStringSlice("GOFORMS_SECURITY_CORS_ALLOWED_HEADERS", []string{"*"}),
			ExposedHeaders:   getEnvStringSlice("GOFORMS_SECURITY_CORS_EXPOSED_HEADERS", []string{}),
			AllowCredentials: getEnvBool("GOFORMS_SECURITY_CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           getEnvInt("GOFORMS_SECURITY_CORS_MAX_AGE", 86400),
		},
		RateLimit: RateLimitConfig{
			Enabled: getEnvBool("GOFORMS_SECURITY_RATE_LIMIT_ENABLED", true),
			RPS:     getEnvInt("GOFORMS_SECURITY_RATE_LIMIT_RPS", 100),
			Burst:   getEnvInt("GOFORMS_SECURITY_RATE_LIMIT_BURST", 200),
		},
		CSP: CSPConfig{
			Enabled:    getEnvBool("GOFORMS_SECURITY_CSP_ENABLED", true),
			DefaultSrc: getEnvOrDefault("GOFORMS_SECURITY_CSP_DEFAULT_SRC", "'self'"),
			ScriptSrc:  getEnvOrDefault("GOFORMS_SECURITY_CSP_SCRIPT_SRC", "'self' 'unsafe-inline'"),
			StyleSrc:   getEnvOrDefault("GOFORMS_SECURITY_CSP_STYLE_SRC", "'self' 'unsafe-inline'"),
			ImgSrc:     getEnvOrDefault("GOFORMS_SECURITY_CSP_IMG_SRC", "'self' data: https:"),
			ConnectSrc: getEnvOrDefault("GOFORMS_SECURITY_CSP_CONNECT_SRC", "'self'"),
			FontSrc:    getEnvOrDefault("GOFORMS_SECURITY_CSP_FONT_SRC", "'self'"),
			ObjectSrc:  getEnvOrDefault("GOFORMS_SECURITY_CSP_OBJECT_SRC", "'none'"),
			MediaSrc:   getEnvOrDefault("GOFORMS_SECURITY_CSP_MEDIA_SRC", "'self'"),
			FrameSrc:   getEnvOrDefault("GOFORMS_SECURITY_CSP_FRAME_SRC", "'none'"),
			ReportURI:  getEnvOrDefault("GOFORMS_SECURITY_CSP_REPORT_URI", ""),
		},
		TLS: TLSConfig{
			Enabled:  getEnvBool("GOFORMS_SECURITY_TLS_ENABLED", false),
			CertFile: getEnvOrDefault("GOFORMS_SECURITY_TLS_CERT_FILE", ""),
			KeyFile:  getEnvOrDefault("GOFORMS_SECURITY_TLS_KEY_FILE", ""),
		},
		Encryption: EncryptionConfig{
			Key: getEnvOrDefault("GOFORMS_SECURITY_ENCRYPTION_KEY", ""),
		},
	}

	// Load Email config
	config.Email = EmailConfig{
		Host:     getEnvOrDefault("GOFORMS_EMAIL_HOST", ""),
		Port:     getEnvInt("GOFORMS_EMAIL_PORT", 587),
		Username: getEnvOrDefault("GOFORMS_EMAIL_USERNAME", ""),
		Password: getEnvOrDefault("GOFORMS_EMAIL_PASSWORD", ""),
		From:     getEnvOrDefault("GOFORMS_EMAIL_FROM", ""),
		UseTLS:   getEnvBool("GOFORMS_EMAIL_USE_TLS", true),
		UseSSL:   getEnvBool("GOFORMS_EMAIL_USE_SSL", false),
		Template: getEnvOrDefault("GOFORMS_EMAIL_TEMPLATE", "default"),
	}

	// Load Storage config
	config.Storage = StorageConfig{
		Type: getEnvOrDefault("GOFORMS_STORAGE_TYPE", "local"),
		Local: LocalStorageConfig{
			Path: getEnvOrDefault("GOFORMS_STORAGE_LOCAL_PATH", "./uploads"),
		},
		S3: S3StorageConfig{
			Bucket:    getEnvOrDefault("GOFORMS_STORAGE_S3_BUCKET", ""),
			Region:    getEnvOrDefault("GOFORMS_STORAGE_S3_REGION", "us-east-1"),
			AccessKey: getEnvOrDefault("GOFORMS_STORAGE_S3_ACCESS_KEY", ""),
			SecretKey: getEnvOrDefault("GOFORMS_STORAGE_S3_SECRET_KEY", ""),
			Endpoint:  getEnvOrDefault("GOFORMS_STORAGE_S3_ENDPOINT", ""),
		},
		MaxSize:     getEnvInt64("GOFORMS_STORAGE_MAX_SIZE", 10*1024*1024), // 10MB
		AllowedExts: getEnvStringSlice("GOFORMS_STORAGE_ALLOWED_EXTENSIONS", []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"}),
	}

	// Load Cache config
	config.Cache = CacheConfig{
		Type: getEnvOrDefault("GOFORMS_CACHE_TYPE", "memory"),
		Redis: RedisConfig{
			Host:     getEnvOrDefault("GOFORMS_CACHE_REDIS_HOST", "localhost"),
			Port:     getEnvInt("GOFORMS_CACHE_REDIS_PORT", 6379),
			Password: getEnvOrDefault("GOFORMS_CACHE_REDIS_PASSWORD", ""),
			DB:       getEnvInt("GOFORMS_CACHE_REDIS_DB", 0),
		},
		Memory: MemoryConfig{
			MaxSize: getEnvInt("GOFORMS_CACHE_MEMORY_MAX_SIZE", 1000),
		},
		TTL: getEnvDuration("GOFORMS_CACHE_TTL", 1*time.Hour),
	}

	// Load Logging config
	config.Logging = LoggingConfig{
		Level:      getEnvOrDefault("GOFORMS_LOGGING_LEVEL", "info"),
		Format:     getEnvOrDefault("GOFORMS_LOGGING_FORMAT", "json"),
		Output:     getEnvOrDefault("GOFORMS_LOGGING_OUTPUT", "stdout"),
		File:       getEnvOrDefault("GOFORMS_LOGGING_FILE", "logs/app.log"),
		MaxSize:    getEnvInt("GOFORMS_LOGGING_MAX_SIZE", 100),
		MaxBackups: getEnvInt("GOFORMS_LOGGING_MAX_BACKUPS", 3),
		MaxAge:     getEnvInt("GOFORMS_LOGGING_MAX_AGE", 28),
		Compress:   getEnvBool("GOFORMS_LOGGING_COMPRESS", true),
	}

	// Load Session config
	config.Session = SessionConfig{
		Type:       getEnvOrDefault("GOFORMS_SESSION_TYPE", "cookie"),
		Secret:     getEnvOrDefault("GOFORMS_SESSION_SECRET", "session-secret"),
		MaxAge:     getEnvDuration("GOFORMS_SESSION_MAX_AGE", 24*time.Hour),
		Domain:     getEnvOrDefault("GOFORMS_SESSION_DOMAIN", ""),
		Path:       getEnvOrDefault("GOFORMS_SESSION_PATH", "/"),
		Secure:     getEnvBool("GOFORMS_SESSION_SECURE", false),
		HTTPOnly:   getEnvBool("GOFORMS_SESSION_HTTP_ONLY", true),
		SameSite:   getEnvOrDefault("GOFORMS_SESSION_SAME_SITE", "lax"),
		Store:      getEnvOrDefault("GOFORMS_SESSION_STORE", "memory"),
		StoreFile:  getEnvOrDefault("GOFORMS_SESSION_STORE_FILE", "storage/sessions/sessions.json"),
		CookieName: getEnvOrDefault("GOFORMS_SESSION_COOKIE_NAME", "session"),
	}

	// Load Auth config
	config.Auth = AuthConfig{
		RequireEmailVerification: getEnvBool("GOFORMS_AUTH_REQUIRE_EMAIL_VERIFICATION", false),
		PasswordMinLength:        getEnvInt("GOFORMS_AUTH_PASSWORD_MIN_LENGTH", 8),
		PasswordRequireSpecial:   getEnvBool("GOFORMS_AUTH_PASSWORD_REQUIRE_SPECIAL", true),
		SessionTimeout:           getEnvDuration("GOFORMS_AUTH_SESSION_TIMEOUT", 30*time.Minute),
		MaxLoginAttempts:         getEnvInt("GOFORMS_AUTH_MAX_LOGIN_ATTEMPTS", 5),
		LockoutDuration:          getEnvDuration("GOFORMS_AUTH_LOCKOUT_DURATION", 15*time.Minute),
	}

	// Load Form config
	config.Form = FormConfig{
		MaxFileSize:      getEnvInt64("GOFORMS_FORM_MAX_FILE_SIZE", 10*1024*1024), // 10MB
		AllowedFileTypes: getEnvStringSlice("GOFORMS_FORM_ALLOWED_FILE_TYPES", []string{"image/jpeg", "image/png", "image/gif", "application/pdf"}),
		MaxFields:        getEnvInt("GOFORMS_FORM_MAX_FIELDS", 100),
		MaxMemory:        getEnvInt64("GOFORMS_FORM_MAX_MEMORY", 32*1024*1024), // 32MB
		Validation: ValidationConfig{
			StrictMode: getEnvBool("GOFORMS_FORM_VALIDATION_STRICT_MODE", false),
			MaxErrors:  getEnvInt("GOFORMS_FORM_VALIDATION_MAX_ERRORS", 10),
		},
	}

	// Load API config
	config.API = APIConfig{
		Version:    getEnvOrDefault("GOFORMS_API_VERSION", "v1"),
		Prefix:     getEnvOrDefault("GOFORMS_API_PREFIX", "/api"),
		Timeout:    getEnvDuration("GOFORMS_API_TIMEOUT", 30*time.Second),
		MaxRetries: getEnvInt("GOFORMS_API_MAX_RETRIES", 3),
		RateLimit: RateLimitConfig{
			Enabled: getEnvBool("GOFORMS_API_RATE_LIMIT_ENABLED", true),
			RPS:     getEnvInt("GOFORMS_API_RATE_LIMIT_RPS", 1000),
			Burst:   getEnvInt("GOFORMS_API_RATE_LIMIT_BURST", 2000),
		},
	}

	// Load Web config
	config.Web = WebConfig{
		TemplateDir:  getEnvOrDefault("GOFORMS_WEB_TEMPLATE_DIR", "templates"),
		StaticDir:    getEnvOrDefault("GOFORMS_WEB_STATIC_DIR", "static"),
		AssetsDir:    getEnvOrDefault("GOFORMS_WEB_ASSETS_DIR", "assets"),
		ReadTimeout:  getEnvDuration("GOFORMS_WEB_READ_TIMEOUT", 15*time.Second),
		WriteTimeout: getEnvDuration("GOFORMS_WEB_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:  getEnvDuration("GOFORMS_WEB_IDLE_TIMEOUT", 60*time.Second),
		Gzip:         getEnvBool("GOFORMS_WEB_GZIP", true),
	}

	// Load User config
	config.User = UserConfig{
		Admin: AdminUserConfig{
			Email:    getEnvOrDefault("GOFORMS_USER_ADMIN_EMAIL", "admin@example.com"),
			Password: getEnvOrDefault("GOFORMS_USER_ADMIN_PASSWORD", "admin123"),
			Name:     getEnvOrDefault("GOFORMS_USER_ADMIN_NAME", "Administrator"),
		},
		Default: DefaultUserConfig{
			Role:        getEnvOrDefault("GOFORMS_USER_DEFAULT_ROLE", "user"),
			Permissions: getEnvStringSlice("GOFORMS_USER_DEFAULT_PERMISSIONS", []string{"read"}),
		},
	}

	// Validate configuration
	if err := config.validateConfig(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// NewConfigProvider creates an Fx provider for the configuration
func NewConfigProvider() fx.Option {
	return fx.Provide(LoadFromEnv)
}

// validateConfig validates the configuration
func (c *Config) validateConfig() error {
	var errs []string

	// Validate App config
	if err := c.App.Validate(); err != nil {
		errs = append(errs, err.Error())
	}

	// Validate Database config
	if err := c.Database.Validate(); err != nil {
		errs = append(errs, err.Error())
	}

	// Validate Security config
	if err := c.Security.Validate(); err != nil {
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

// Helper functions for environment variable parsing

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
