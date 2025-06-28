// Package config provides Viper-based configuration management for the GoForms application.
// It supports multiple configuration formats (JSON, YAML, TOML, ENV) and sources.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// ViperConfig represents the Viper-based configuration loader
type ViperConfig struct {
	viper *viper.Viper
}

// NewViperConfig creates a new Viper configuration instance
func NewViperConfig() *ViperConfig {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Configure Viper
	v.SetEnvPrefix("GOFORMS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set config file search paths
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/goforms")
	v.AddConfigPath("$HOME/.goforms")

	// Set config file names (without extension)
	v.SetConfigName("config")
	v.SetConfigType("yaml") // Default to YAML

	return &ViperConfig{viper: v}
}

// Load loads configuration using Viper
func (vc *ViperConfig) Load() (*Config, error) {
	// Try to read config file
	if err := vc.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, continue with environment variables only
	}

	// Load .env file if it exists
	if err := vc.viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to merge config: %w", err)
		}
	}

	// Create config struct
	config := &Config{}

	// Load App config
	config.App = AppConfig{
		Name:           vc.viper.GetString("app.name"),
		Version:        vc.viper.GetString("app.version"),
		Environment:    vc.viper.GetString("app.environment"),
		Debug:          vc.viper.GetBool("app.debug"),
		LogLevel:       vc.viper.GetString("app.log_level"),
		URL:            vc.viper.GetString("app.url"),
		Scheme:         vc.viper.GetString("app.scheme"),
		Port:           vc.viper.GetInt("app.port"),
		Host:           vc.viper.GetString("app.host"),
		ReadTimeout:    vc.viper.GetDuration("app.read_timeout"),
		WriteTimeout:   vc.viper.GetDuration("app.write_timeout"),
		IdleTimeout:    vc.viper.GetDuration("app.idle_timeout"),
		RequestTimeout: vc.viper.GetDuration("app.request_timeout"),
		ViteDevHost:    vc.viper.GetString("app.vite_dev_host"),
		ViteDevPort:    vc.viper.GetString("app.vite_dev_port"),
	}

	// Load Database config
	config.Database = DatabaseConfig{
		Driver:          vc.viper.GetString("database.driver"),
		Host:            vc.viper.GetString("database.host"),
		Port:            vc.viper.GetInt("database.port"),
		Name:            vc.viper.GetString("database.name"),
		Username:        vc.viper.GetString("database.username"),
		Password:        vc.viper.GetString("database.password"),
		SSLMode:         vc.viper.GetString("database.ssl_mode"),
		MaxOpenConns:    vc.viper.GetInt("database.max_open_conns"),
		MaxIdleConns:    vc.viper.GetInt("database.max_idle_conns"),
		ConnMaxLifetime: vc.viper.GetDuration("database.conn_max_lifetime"),
		ConnMaxIdleTime: vc.viper.GetDuration("database.conn_max_idle_time"),
	}

	// Load Security config
	config.Security = SecurityConfig{
		CSRF: CSRFConfig{
			Enabled:    vc.viper.GetBool("security.csrf.enabled"),
			Secret:     vc.viper.GetString("security.csrf.secret"),
			TokenName:  vc.viper.GetString("security.csrf.token_name"),
			HeaderName: vc.viper.GetString("security.csrf.header_name"),
		},
		CORS: CORSConfig{
			Enabled:          vc.viper.GetBool("security.cors.enabled"),
			AllowedOrigins:   vc.viper.GetStringSlice("security.cors.allowed_origins"),
			AllowedMethods:   vc.viper.GetStringSlice("security.cors.allowed_methods"),
			AllowedHeaders:   vc.viper.GetStringSlice("security.cors.allowed_headers"),
			ExposedHeaders:   vc.viper.GetStringSlice("security.cors.exposed_headers"),
			AllowCredentials: vc.viper.GetBool("security.cors.allow_credentials"),
			MaxAge:           vc.viper.GetInt("security.cors.max_age"),
		},
		RateLimit: RateLimitConfig{
			Enabled: vc.viper.GetBool("security.rate_limit.enabled"),
			RPS:     vc.viper.GetInt("security.rate_limit.rps"),
			Burst:   vc.viper.GetInt("security.rate_limit.burst"),
		},
		CSP: CSPConfig{
			Enabled:    vc.viper.GetBool("security.csp.enabled"),
			DefaultSrc: vc.viper.GetString("security.csp.default_src"),
			ScriptSrc:  vc.viper.GetString("security.csp.script_src"),
			StyleSrc:   vc.viper.GetString("security.csp.style_src"),
			ImgSrc:     vc.viper.GetString("security.csp.img_src"),
			ConnectSrc: vc.viper.GetString("security.csp.connect_src"),
			FontSrc:    vc.viper.GetString("security.csp.font_src"),
			ObjectSrc:  vc.viper.GetString("security.csp.object_src"),
			MediaSrc:   vc.viper.GetString("security.csp.media_src"),
			FrameSrc:   vc.viper.GetString("security.csp.frame_src"),
			ReportURI:  vc.viper.GetString("security.csp.report_uri"),
		},
		TLS: TLSConfig{
			Enabled:  vc.viper.GetBool("security.tls.enabled"),
			CertFile: vc.viper.GetString("security.tls.cert_file"),
			KeyFile:  vc.viper.GetString("security.tls.key_file"),
		},
		Encryption: EncryptionConfig{
			Key: vc.viper.GetString("security.encryption.key"),
		},
		SecureCookie: vc.viper.GetBool("security.secure_cookie"),
		Debug:        vc.viper.GetBool("security.debug"),
	}

	// Load Email config
	config.Email = EmailConfig{
		Host:     vc.viper.GetString("email.host"),
		Port:     vc.viper.GetInt("email.port"),
		Username: vc.viper.GetString("email.username"),
		Password: vc.viper.GetString("email.password"),
		From:     vc.viper.GetString("email.from"),
		UseTLS:   vc.viper.GetBool("email.use_tls"),
		UseSSL:   vc.viper.GetBool("email.use_ssl"),
		Template: vc.viper.GetString("email.template"),
	}

	// Load Storage config
	config.Storage = StorageConfig{
		Type: vc.viper.GetString("storage.type"),
		Local: LocalStorageConfig{
			Path: vc.viper.GetString("storage.local.path"),
		},
		S3: S3StorageConfig{
			Bucket:    vc.viper.GetString("storage.s3.bucket"),
			Region:    vc.viper.GetString("storage.s3.region"),
			AccessKey: vc.viper.GetString("storage.s3.access_key"),
			SecretKey: vc.viper.GetString("storage.s3.secret_key"),
			Endpoint:  vc.viper.GetString("storage.s3.endpoint"),
		},
		MaxSize:     vc.viper.GetInt64("storage.max_size"),
		AllowedExts: vc.viper.GetStringSlice("storage.allowed_extensions"),
	}

	// Load Cache config
	config.Cache = CacheConfig{
		Type: vc.viper.GetString("cache.type"),
		Redis: RedisConfig{
			Host:     vc.viper.GetString("cache.redis.host"),
			Port:     vc.viper.GetInt("cache.redis.port"),
			Password: vc.viper.GetString("cache.redis.password"),
			DB:       vc.viper.GetInt("cache.redis.db"),
		},
		Memory: MemoryConfig{
			MaxSize: vc.viper.GetInt("cache.memory.max_size"),
		},
		TTL: vc.viper.GetDuration("cache.ttl"),
	}

	// Load Logging config
	config.Logging = LoggingConfig{
		Level:      vc.viper.GetString("logging.level"),
		Format:     vc.viper.GetString("logging.format"),
		Output:     vc.viper.GetString("logging.output"),
		File:       vc.viper.GetString("logging.file"),
		MaxSize:    vc.viper.GetInt("logging.max_size"),
		MaxBackups: vc.viper.GetInt("logging.max_backups"),
		MaxAge:     vc.viper.GetInt("logging.max_age"),
		Compress:   vc.viper.GetBool("logging.compress"),
	}

	// Load Session config
	config.Session = SessionConfig{
		Type:       vc.viper.GetString("session.type"),
		Secret:     vc.viper.GetString("session.secret"),
		MaxAge:     vc.viper.GetDuration("session.max_age"),
		Domain:     vc.viper.GetString("session.domain"),
		Path:       vc.viper.GetString("session.path"),
		Secure:     vc.viper.GetBool("session.secure"),
		HTTPOnly:   vc.viper.GetBool("session.http_only"),
		SameSite:   vc.viper.GetString("session.same_site"),
		Store:      vc.viper.GetString("session.store"),
		StoreFile:  vc.viper.GetString("session.store_file"),
		CookieName: vc.viper.GetString("session.cookie_name"),
	}

	// Load Auth config
	config.Auth = AuthConfig{
		RequireEmailVerification: vc.viper.GetBool("auth.require_email_verification"),
		PasswordMinLength:        vc.viper.GetInt("auth.password_min_length"),
		PasswordRequireSpecial:   vc.viper.GetBool("auth.password_require_special"),
		SessionTimeout:           vc.viper.GetDuration("auth.session_timeout"),
		MaxLoginAttempts:         vc.viper.GetInt("auth.max_login_attempts"),
		LockoutDuration:          vc.viper.GetDuration("auth.lockout_duration"),
	}

	// Load Form config
	config.Form = FormConfig{
		MaxFileSize:      vc.viper.GetInt64("form.max_file_size"),
		AllowedFileTypes: vc.viper.GetStringSlice("form.allowed_file_types"),
		MaxFields:        vc.viper.GetInt("form.max_fields"),
		MaxMemory:        vc.viper.GetInt64("form.max_memory"),
		Validation: ValidationConfig{
			StrictMode: vc.viper.GetBool("form.validation.strict_mode"),
			MaxErrors:  vc.viper.GetInt("form.validation.max_errors"),
		},
	}

	// Load API config
	config.API = APIConfig{
		Version:    vc.viper.GetString("api.version"),
		Prefix:     vc.viper.GetString("api.prefix"),
		Timeout:    vc.viper.GetDuration("api.timeout"),
		MaxRetries: vc.viper.GetInt("api.max_retries"),
		RateLimit: RateLimitConfig{
			Enabled: vc.viper.GetBool("api.rate_limit.enabled"),
			RPS:     vc.viper.GetInt("api.rate_limit.rps"),
			Burst:   vc.viper.GetInt("api.rate_limit.burst"),
		},
	}

	// Load Web config
	config.Web = WebConfig{
		TemplateDir:  vc.viper.GetString("web.template_dir"),
		StaticDir:    vc.viper.GetString("web.static_dir"),
		AssetsDir:    vc.viper.GetString("web.assets_dir"),
		ReadTimeout:  vc.viper.GetDuration("web.read_timeout"),
		WriteTimeout: vc.viper.GetDuration("web.write_timeout"),
		IdleTimeout:  vc.viper.GetDuration("web.idle_timeout"),
		Gzip:         vc.viper.GetBool("web.gzip"),
	}

	// Load User config
	config.User = UserConfig{
		Admin: AdminUserConfig{
			Email:    vc.viper.GetString("user.admin.email"),
			Password: vc.viper.GetString("user.admin.password"),
			Name:     vc.viper.GetString("user.admin.name"),
		},
		Default: DefaultUserConfig{
			Role:        vc.viper.GetString("user.default.role"),
			Permissions: vc.viper.GetStringSlice("user.default.permissions"),
		},
	}

	// Validate configuration
	if err := config.validateConfig(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// LoadForEnvironment loads configuration for a specific environment
func (vc *ViperConfig) LoadForEnvironment(env string) (*Config, error) {
	// Set environment-specific config file
	vc.viper.SetConfigName(fmt.Sprintf("config.%s", env))

	// Also try to load .env file for the environment
	envFile := fmt.Sprintf(".env.%s", env)
	if _, err := os.Stat(envFile); err == nil {
		vc.viper.SetConfigFile(envFile)
		vc.viper.SetConfigType("env")
		if err := vc.viper.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("failed to merge env config: %w", err)
		}
	}

	config, err := vc.Load()
	if err != nil {
		return nil, err
	}

	// Override the environment setting
	config.App.Environment = env

	return config, nil
}

// WatchConfig watches for configuration changes and reloads when files change
func (vc *ViperConfig) WatchConfig(callback func()) {
	vc.viper.WatchConfig()
	vc.viper.OnConfigChange(func(e fsnotify.Event) {
		callback()
	})
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.name", "GoForms")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", true)
	v.SetDefault("app.log_level", "info")
	v.SetDefault("app.url", "http://localhost:8080")
	v.SetDefault("app.scheme", "http")
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.host", "localhost")
	v.SetDefault("app.read_timeout", 15*time.Second)
	v.SetDefault("app.write_timeout", 15*time.Second)
	v.SetDefault("app.idle_timeout", 60*time.Second)
	v.SetDefault("app.request_timeout", 30*time.Second)
	v.SetDefault("app.vite_dev_host", "localhost")
	v.SetDefault("app.vite_dev_port", "5173")

	// Database defaults
	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.name", "goforms")
	v.SetDefault("database.username", "goforms")
	v.SetDefault("database.password", "goforms")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 25)
	v.SetDefault("database.conn_max_lifetime", 5*time.Minute)
	v.SetDefault("database.conn_max_idle_time", 5*time.Minute)

	// Security defaults
	v.SetDefault("security.csrf.enabled", true)
	v.SetDefault("security.csrf.secret", "csrf-secret")
	v.SetDefault("security.csrf.token_name", "_token")
	v.SetDefault("security.csrf.header_name", "X-CSRF-Token")
	v.SetDefault("security.cors.enabled", true)
	v.SetDefault("security.cors.allowed_origins", []string{"*"})
	v.SetDefault("security.cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("security.cors.allowed_headers", []string{"*"})
	v.SetDefault("security.cors.exposed_headers", []string{})
	v.SetDefault("security.cors.allow_credentials", true)
	v.SetDefault("security.cors.max_age", 86400)
	v.SetDefault("security.rate_limit.enabled", true)
	v.SetDefault("security.rate_limit.rps", 100)
	v.SetDefault("security.rate_limit.burst", 200)
	v.SetDefault("security.csp.enabled", true)
	v.SetDefault("security.csp.default_src", "'self'")
	v.SetDefault("security.csp.script_src", "'self' 'unsafe-inline'")
	v.SetDefault("security.csp.style_src", "'self' 'unsafe-inline'")
	v.SetDefault("security.csp.img_src", "'self' data: https:")
	v.SetDefault("security.csp.connect_src", "'self'")
	v.SetDefault("security.csp.font_src", "'self'")
	v.SetDefault("security.csp.object_src", "'none'")
	v.SetDefault("security.csp.media_src", "'self'")
	v.SetDefault("security.csp.frame_src", "'none'")
	v.SetDefault("security.tls.enabled", false)
	v.SetDefault("security.encryption.key", "")
	v.SetDefault("security.secure_cookie", false)
	v.SetDefault("security.debug", false)

	// Email defaults
	v.SetDefault("email.port", 587)
	v.SetDefault("email.use_tls", true)
	v.SetDefault("email.use_ssl", false)
	v.SetDefault("email.template", "default")

	// Storage defaults
	v.SetDefault("storage.type", "local")
	v.SetDefault("storage.local.path", "./uploads")
	v.SetDefault("storage.s3.region", "us-east-1")
	v.SetDefault("storage.max_size", 10*1024*1024) // 10MB
	v.SetDefault("storage.allowed_extensions", []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"})

	// Cache defaults
	v.SetDefault("cache.type", "memory")
	v.SetDefault("cache.redis.host", "localhost")
	v.SetDefault("cache.redis.port", 6379)
	v.SetDefault("cache.redis.db", 0)
	v.SetDefault("cache.memory.max_size", 1000)
	v.SetDefault("cache.ttl", 1*time.Hour)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")
	v.SetDefault("logging.file", "logs/app.log")
	v.SetDefault("logging.max_size", 100)
	v.SetDefault("logging.max_backups", 3)
	v.SetDefault("logging.max_age", 28)
	v.SetDefault("logging.compress", true)

	// Session defaults
	v.SetDefault("session.type", "cookie")
	v.SetDefault("session.secret", "session-secret")
	v.SetDefault("session.max_age", 24*time.Hour)
	v.SetDefault("session.path", "/")
	v.SetDefault("session.secure", false)
	v.SetDefault("session.http_only", true)
	v.SetDefault("session.same_site", "lax")
	v.SetDefault("session.store", "memory")
	v.SetDefault("session.store_file", "storage/sessions/sessions.json")
	v.SetDefault("session.cookie_name", "session")

	// Auth defaults
	v.SetDefault("auth.require_email_verification", false)
	v.SetDefault("auth.password_min_length", 8)
	v.SetDefault("auth.password_require_special", true)
	v.SetDefault("auth.session_timeout", 30*time.Minute)
	v.SetDefault("auth.max_login_attempts", 5)
	v.SetDefault("auth.lockout_duration", 15*time.Minute)

	// Form defaults
	v.SetDefault("form.max_file_size", 10*1024*1024) // 10MB
	v.SetDefault("form.allowed_file_types", []string{"image/jpeg", "image/png", "image/gif", "application/pdf"})
	v.SetDefault("form.max_fields", 100)
	v.SetDefault("form.max_memory", 32*1024*1024) // 32MB
	v.SetDefault("form.validation.strict_mode", false)
	v.SetDefault("form.validation.max_errors", 10)

	// API defaults
	v.SetDefault("api.version", "v1")
	v.SetDefault("api.prefix", "/api")
	v.SetDefault("api.timeout", 30*time.Second)
	v.SetDefault("api.max_retries", 3)
	v.SetDefault("api.rate_limit.enabled", true)
	v.SetDefault("api.rate_limit.rps", 1000)
	v.SetDefault("api.rate_limit.burst", 2000)

	// Web defaults
	v.SetDefault("web.template_dir", "templates")
	v.SetDefault("web.static_dir", "static")
	v.SetDefault("web.assets_dir", "assets")
	v.SetDefault("web.read_timeout", 15*time.Second)
	v.SetDefault("web.write_timeout", 15*time.Second)
	v.SetDefault("web.idle_timeout", 60*time.Second)
	v.SetDefault("web.gzip", true)

	// User defaults
	v.SetDefault("user.admin.email", "admin@example.com")
	v.SetDefault("user.admin.password", "admin123")
	v.SetDefault("user.admin.name", "Administrator")
	v.SetDefault("user.default.role", "user")
	v.SetDefault("user.default.permissions", []string{"read"})
}

// NewViperConfigProvider creates an Fx provider for Viper configuration
func NewViperConfigProvider() fx.Option {
	return fx.Provide(func() (*Config, error) {
		vc := NewViperConfig()
		return vc.Load()
	})
}
