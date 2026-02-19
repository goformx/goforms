// Package config provides Viper-based configuration management for the GoForms application.
// It supports multiple configuration formats (JSON, YAML, TOML, ENV) and sources.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// ViperConfig represents the Viper-based configuration loader
type ViperConfig struct {
	viper          *viper.Viper
	configFilePath string // Path to loaded config file, available after Load()
}

// GetConfigFilePath returns the path to the loaded config file
func (vc *ViperConfig) GetConfigFilePath() string {
	return vc.configFilePath
}

// NewViperConfig creates a new Viper configuration instance
func NewViperConfig() *ViperConfig {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Configure Viper with best practices
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Bind DB_* environment variables to database.* config keys
	// This allows users to use the common DB_ prefix convention
	_ = v.BindEnv("database.host", "DB_HOST")
	_ = v.BindEnv("database.port", "DB_PORT")
	_ = v.BindEnv("database.name", "DB_NAME", "DB_DATABASE")
	_ = v.BindEnv("database.username", "DB_USERNAME", "DB_USER")
	_ = v.BindEnv("database.password", "DB_PASSWORD")
	_ = v.BindEnv("database.driver", "DB_CONNECTION", "DB_DRIVER")
	_ = v.BindEnv("database.ssl_mode", "DB_SSL_MODE")

	// Bind CORS_* environment variables for convenience
	_ = v.BindEnv("security.cors.allowed_origins", "CORS_ALLOWED_ORIGINS", "CORS_ORIGINS")
	_ = v.BindEnv("security.cors.allowed_methods", "CORS_ALLOWED_METHODS")
	_ = v.BindEnv("security.cors.allowed_headers", "CORS_ALLOWED_HEADERS")
	_ = v.BindEnv("security.cors.allow_credentials", "CORS_ALLOW_CREDENTIALS")
	_ = v.BindEnv("security.cors.max_age", "CORS_MAX_AGE")

	// Bind GOFORMS_SHARED_SECRET for Laravel-Go assertion verification
	_ = v.BindEnv("security.assertion.secret", "GOFORMS_SHARED_SECRET")

	// Set config file search paths (order matters - first found wins)
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/goforms")
	v.AddConfigPath("$HOME/.goforms")

	// Set config file names (without extension) - try multiple formats
	v.SetConfigName("config")
	v.SetConfigType("yaml") // Default to YAML

	return &ViperConfig{viper: v}
}

// Load loads configuration using Viper with improved error handling
func (vc *ViperConfig) Load() (*Config, error) {
	if err := vc.loadConfigFiles(); err != nil {
		return nil, fmt.Errorf("failed to load configuration files: %w", err)
	}

	config := &Config{}

	if err := vc.loadAllConfigSections(config); err != nil {
		return nil, fmt.Errorf("failed to load configuration sections: %w", err)
	}

	// Validate configuration with detailed error reporting
	if err := config.validateConfig(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// loadConfigFiles loads configuration files with better error handling
func (vc *ViperConfig) loadConfigFiles() error {
	// Try to read config file
	if err := vc.viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// Config file not found is not an error - we can use environment variables
			// Silent in production, no logging needed here
			return nil
		}

		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Store config file path for later logging (once logger is available)
	vc.configFilePath = vc.viper.ConfigFileUsed()

	// Try to merge additional config files (like .env)
	if err := vc.viper.MergeInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// Additional config file not found is not an error
			return nil
		}

		return fmt.Errorf("failed to merge additional config: %w", err)
	}

	return nil
}

// loadAllConfigSections loads all configuration sections
func (vc *ViperConfig) loadAllConfigSections(config *Config) error {
	loaders := []func(*Config) error{
		vc.loadAppConfig,
		vc.loadDatabaseConfig,
		vc.loadSecurityConfig,
		vc.loadEmailConfig,
		vc.loadStorageConfig,
		vc.loadCacheConfig,
		vc.loadLoggingConfig,
		vc.loadSessionConfig,
		vc.loadAuthConfig,
		vc.loadFormConfig,
		vc.loadAPIConfig,
		vc.loadWebConfig,
		vc.loadUserConfig,
	}

	for _, loader := range loaders {
		if err := loader(config); err != nil {
			return err
		}
	}

	return nil
}

// loadAppConfig loads application configuration
func (vc *ViperConfig) loadAppConfig(config *Config) error {
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

	return nil
}

// loadDatabaseConfig loads database configuration
func (vc *ViperConfig) loadDatabaseConfig(config *Config) error {
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

	return nil
}

// loadCSRFConfig loads CSRF configuration from viper
func (vc *ViperConfig) loadCSRFConfig() CSRFConfig {
	return CSRFConfig{
		Enabled:        vc.viper.GetBool("security.csrf.enabled"),
		Secret:         vc.viper.GetString("security.csrf.secret"),
		TokenName:      vc.viper.GetString("security.csrf.token_name"),
		HeaderName:     vc.viper.GetString("security.csrf.header_name"),
		TokenLength:    vc.viper.GetInt("security.csrf.token_length"),
		TokenLookup:    vc.viper.GetString("security.csrf.token_lookup"),
		ContextKey:     vc.viper.GetString("security.csrf.context_key"),
		CookieName:     vc.viper.GetString("security.csrf.cookie_name"),
		CookiePath:     vc.viper.GetString("security.csrf.cookie_path"),
		CookieDomain:   vc.viper.GetString("security.csrf.cookie_domain"),
		CookieHTTPOnly: vc.viper.GetBool("security.csrf.cookie_http_only"),
		CookieSameSite: vc.viper.GetString("security.csrf.cookie_same_site"),
		CookieMaxAge:   vc.viper.GetInt("security.csrf.cookie_max_age"),
	}
}

// loadCORSConfig loads CORS configuration from viper
func (vc *ViperConfig) loadCORSConfig() CORSConfig {
	return CORSConfig{
		Enabled:          vc.viper.GetBool("security.cors.enabled"),
		AllowedOrigins:   vc.viper.GetStringSlice("security.cors.allowed_origins"),
		AllowedMethods:   vc.viper.GetStringSlice("security.cors.allowed_methods"),
		AllowedHeaders:   vc.viper.GetStringSlice("security.cors.allowed_headers"),
		ExposedHeaders:   vc.viper.GetStringSlice("security.cors.exposed_headers"),
		AllowCredentials: vc.viper.GetBool("security.cors.allow_credentials"),
		MaxAge:           vc.viper.GetInt("security.cors.max_age"),
	}
}

// loadAssertionConfig loads assertion verification configuration from viper
func (vc *ViperConfig) loadAssertionConfig() AssertionConfig {
	return AssertionConfig{
		Secret:               vc.viper.GetString("security.assertion.secret"),
		TimestampSkewSeconds: vc.viper.GetInt("security.assertion.timestamp_skew_seconds"),
	}
}

// loadAPIKeyConfig loads API key configuration from viper
func (vc *ViperConfig) loadAPIKeyConfig() APIKeyConfig {
	// Support environment variable with comma-separated keys
	keysEnv := os.Getenv("API_KEYS")
	var keys []string
	if keysEnv != "" {
		keys = strings.Split(keysEnv, ",")
		// Trim whitespace from each key
		for i, key := range keys {
			keys[i] = strings.TrimSpace(key)
		}
	} else {
		keys = vc.viper.GetStringSlice("security.api_key.keys")
	}

	headerName := vc.viper.GetString("security.api_key.header_name")
	if headerName == "" {
		headerName = "X-API-Key" // Default header name
	}

	return APIKeyConfig{
		Enabled:     vc.viper.GetBool("security.api_key.enabled"),
		Keys:        keys,
		HeaderName:  headerName,
		QueryParam:  vc.viper.GetString("security.api_key.query_param"),
		SkipPaths:   vc.viper.GetStringSlice("security.api_key.skip_paths"),
		SkipMethods: vc.viper.GetStringSlice("security.api_key.skip_methods"),
	}
}

// loadRateLimitConfig loads rate limit configuration from viper
func (vc *ViperConfig) loadRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:  vc.viper.GetBool("security.rate_limit.enabled"),
		RPS:      vc.viper.GetInt("security.rate_limit.rps"),
		Requests: vc.viper.GetInt("security.rate_limit.rps"),
		Burst:    vc.viper.GetInt("security.rate_limit.burst"),
		Window:   vc.viper.GetDuration("security.rate_limit.window"),
		PerIP:    vc.viper.GetBool("security.rate_limit.per_ip"),
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
			"/robots.txt",
			"/static/",
			"/assets/",
		},
		SkipMethods: []string{"OPTIONS"},
	}
}

// loadCSPConfig loads CSP configuration from viper
func (vc *ViperConfig) loadCSPConfig() CSPConfig {
	return CSPConfig{
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
	}
}

// loadSecurityHeadersConfig loads security headers configuration from viper
func (vc *ViperConfig) loadSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		Enabled:                 vc.viper.GetBool("security.security_headers.enabled"),
		XFrameOptions:           vc.viper.GetString("security.security_headers.x_frame_options"),
		XContentTypeOptions:     vc.viper.GetString("security.security_headers.x_content_type_options"),
		XXSSProtection:          vc.viper.GetString("security.security_headers.x_xss_protection"),
		ReferrerPolicy:          vc.viper.GetString("security.security_headers.referrer_policy"),
		PermissionsPolicy:       vc.viper.GetString("security.security_headers.permissions_policy"),
		StrictTransportSecurity: vc.viper.GetString("security.security_headers.strict_transport_security"),
	}
}

// loadSecurityConfig loads security configuration
func (vc *ViperConfig) loadSecurityConfig(config *Config) error {
	config.Security = SecurityConfig{
		CSRF:      vc.loadCSRFConfig(),
		CORS:      vc.loadCORSConfig(),
		RateLimit: vc.loadRateLimitConfig(),
		CSP:       vc.loadCSPConfig(),
		TLS: TLSConfig{
			Enabled:  vc.viper.GetBool("security.tls.enabled"),
			CertFile: vc.viper.GetString("security.tls.cert_file"),
			KeyFile:  vc.viper.GetString("security.tls.key_file"),
		},
		Encryption: EncryptionConfig{
			Key: vc.viper.GetString("security.encryption.key"),
		},
		SecurityHeaders: vc.loadSecurityHeadersConfig(),
		CookieSecurity: CookieSecurityConfig{
			Secure:   vc.viper.GetBool("security.cookie_security.secure"),
			HTTPOnly: vc.viper.GetBool("security.cookie_security.http_only"),
			SameSite: vc.viper.GetString("security.cookie_security.same_site"),
			Path:     vc.viper.GetString("security.cookie_security.path"),
			Domain:   vc.viper.GetString("security.cookie_security.domain"),
			MaxAge:   vc.viper.GetInt("security.cookie_security.max_age"),
		},
		TrustProxy: TrustProxyConfig{
			Enabled:        vc.viper.GetBool("security.trust_proxy.enabled"),
			TrustedProxies: vc.viper.GetStringSlice("security.trust_proxy.trusted_proxies"),
		},
		Assertion:    vc.loadAssertionConfig(),
		APIKey:       vc.loadAPIKeyConfig(),
		SecureCookie: vc.viper.GetBool("security.secure_cookie"),
		Debug:        vc.viper.GetBool("security.debug"),
	}

	return nil
}

// loadEmailConfig loads email configuration
func (vc *ViperConfig) loadEmailConfig(config *Config) error {
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

	return nil
}

// loadStorageConfig loads storage configuration
func (vc *ViperConfig) loadStorageConfig(config *Config) error {
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

	return nil
}

// loadCacheConfig loads cache configuration
func (vc *ViperConfig) loadCacheConfig(config *Config) error {
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

	return nil
}

// loadLoggingConfig loads logging configuration
func (vc *ViperConfig) loadLoggingConfig(config *Config) error {
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

	return nil
}

// loadSessionConfig loads session configuration
func (vc *ViperConfig) loadSessionConfig(config *Config) error {
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

	return nil
}

// loadAuthConfig loads authentication configuration
func (vc *ViperConfig) loadAuthConfig(config *Config) error {
	config.Auth = AuthConfig{
		RequireEmailVerification: vc.viper.GetBool("auth.require_email_verification"),
		PasswordMinLength:        vc.viper.GetInt("auth.password_min_length"),
		PasswordRequireSpecial:   vc.viper.GetBool("auth.password_require_special"),
		SessionTimeout:           vc.viper.GetDuration("auth.session_timeout"),
		MaxLoginAttempts:         vc.viper.GetInt("auth.max_login_attempts"),
		LockoutDuration:          vc.viper.GetDuration("auth.lockout_duration"),
	}

	return nil
}

// loadFormConfig loads form configuration
func (vc *ViperConfig) loadFormConfig(config *Config) error {
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

	return nil
}

// loadAPIConfig loads API configuration
func (vc *ViperConfig) loadAPIConfig(config *Config) error {
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

	return nil
}

// loadWebConfig loads web configuration
func (vc *ViperConfig) loadWebConfig(config *Config) error {
	config.Web = WebConfig{
		TemplateDir:  vc.viper.GetString("web.template_dir"),
		StaticDir:    vc.viper.GetString("web.static_dir"),
		AssetsDir:    vc.viper.GetString("web.assets_dir"),
		ReadTimeout:  vc.viper.GetDuration("web.read_timeout"),
		WriteTimeout: vc.viper.GetDuration("web.write_timeout"),
		IdleTimeout:  vc.viper.GetDuration("web.idle_timeout"),
		Gzip:         vc.viper.GetBool("web.gzip"),
	}

	return nil
}

// loadUserConfig loads user configuration
func (vc *ViperConfig) loadUserConfig(config *Config) error {
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

	return nil
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

		if mergeErr := vc.viper.MergeInConfig(); mergeErr != nil {
			return nil, fmt.Errorf("failed to merge env config: %w", mergeErr)
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

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	setAppDefaults(v)
	setDatabaseDefaults(v)
	setSecurityDefaults(v)
	setEmailDefaults(v)
	setStorageDefaults(v)
	setCacheDefaults(v)
	setLoggingDefaults(v)
	setSessionDefaults(v)
	setAuthDefaults(v)
	setFormDefaults(v)
	setAPIDefaults(v)
	setWebDefaults(v)
	setUserDefaults(v)
}

// setAppDefaults sets application default values
func setAppDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "GoForms")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", true)
	v.SetDefault("app.log_level", "info")
	v.SetDefault("app.url", "http://localhost:8080")
	v.SetDefault("app.scheme", "http")
	v.SetDefault("app.port", DefaultAppPort)
	v.SetDefault("app.host", "localhost")
	v.SetDefault("app.read_timeout", DefaultReadTimeout)
	v.SetDefault("app.write_timeout", DefaultWriteTimeout)
	v.SetDefault("app.idle_timeout", DefaultIdleTimeout)
	v.SetDefault("app.request_timeout", DefaultRequestTimeout)
	v.SetDefault("app.vite_dev_host", "localhost")
	v.SetDefault("app.vite_dev_port", "5173")
}

// setDatabaseDefaults sets database default values
func setDatabaseDefaults(v *viper.Viper) {
	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", DefaultDBPort)
	v.SetDefault("database.name", "goforms")
	v.SetDefault("database.username", "goforms")
	v.SetDefault("database.password", "goforms")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", DefaultMaxOpenConns)
	v.SetDefault("database.max_idle_conns", DefaultMaxIdleConns)
	v.SetDefault("database.conn_max_lifetime", DefaultConnLifetime)
	v.SetDefault("database.conn_max_idle_time", DefaultConnIdleTime)
}

// setCSRFDefaults sets CSRF default values
func setCSRFDefaults(v *viper.Viper) {
	v.SetDefault("security.csrf.enabled", true)
	v.SetDefault("security.csrf.secret", "csrf-secret")
	v.SetDefault("security.csrf.token_name", "_token")
	v.SetDefault("security.csrf.header_name", "X-Csrf-Token")
	v.SetDefault("security.csrf.token_length", DefaultCSRFTokenLength)
	v.SetDefault("security.csrf.token_lookup", "header:X-Csrf-Token")
	v.SetDefault("security.csrf.context_key", "csrf")
	v.SetDefault("security.csrf.cookie_name", "_csrf")
	v.SetDefault("security.csrf.cookie_path", "/")
	v.SetDefault("security.csrf.cookie_domain", "")
	v.SetDefault("security.csrf.cookie_http_only", true)
	v.SetDefault("security.csrf.cookie_same_site", "Lax")
	v.SetDefault("security.csrf.cookie_max_age", DefaultCookieMaxAge)
}

// setCORSDefaults sets CORS default values
func setCORSDefaults(v *viper.Viper) {
	v.SetDefault("security.cors.enabled", true)
	v.SetDefault("security.cors.allowed_origins", []string{"*"})
	v.SetDefault("security.cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	allowedHeaders := []string{"Content-Type", "Authorization", "X-Csrf-Token", "X-Requested-With", "X-API-Key"}
	v.SetDefault("security.cors.allowed_headers", allowedHeaders)
	v.SetDefault("security.cors.exposed_headers", []string{})
	v.SetDefault("security.cors.allow_credentials", true)
	v.SetDefault("security.cors.max_age", DefaultCookieMaxAge)
}

const defaultAssertionTimestampSkewSeconds = 60

// setAssertionDefaults sets assertion verification default values
func setAssertionDefaults(v *viper.Viper) {
	v.SetDefault("security.assertion.secret", "")
	v.SetDefault("security.assertion.timestamp_skew_seconds", defaultAssertionTimestampSkewSeconds)
}

// setAPIKeyDefaults sets API key default values
func setAPIKeyDefaults(v *viper.Viper) {
	v.SetDefault("security.api_key.enabled", false)
	v.SetDefault("security.api_key.keys", []string{})
	v.SetDefault("security.api_key.header_name", "X-API-Key")
	v.SetDefault("security.api_key.query_param", "")
	v.SetDefault("security.api_key.skip_paths", []string{})
	v.SetDefault("security.api_key.skip_methods", []string{"OPTIONS"})
}

// setCSPDefaults sets CSP default values
func setCSPDefaults(v *viper.Viper) {
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
}

// setSecurityHeadersDefaults sets security headers default values
func setSecurityHeadersDefaults(v *viper.Viper) {
	v.SetDefault("security.security_headers.enabled", true)
	v.SetDefault("security.security_headers.x_frame_options", "DENY")
	v.SetDefault("security.security_headers.x_content_type_options", "nosniff")
	v.SetDefault("security.security_headers.x_xss_protection", "1; mode=block")
	v.SetDefault("security.security_headers.referrer_policy", "strict-origin-when-cross-origin")
	v.SetDefault("security.security_headers.permissions_policy", "camera=(), microphone=(), geolocation=()")
	v.SetDefault("security.security_headers.strict_transport_security", "")
}

// setSecurityDefaults sets security default values
func setSecurityDefaults(v *viper.Viper) {
	setCSRFDefaults(v)
	setCORSDefaults(v)
	setAssertionDefaults(v)
	setAPIKeyDefaults(v)
	v.SetDefault("security.rate_limit.enabled", false)
	v.SetDefault("security.rate_limit.rps", DefaultRateLimitRPS)
	v.SetDefault("security.rate_limit.burst", DefaultRateLimitBurst)
	v.SetDefault("security.rate_limit.window", "1m")
	v.SetDefault("security.rate_limit.per_ip", false)
	setCSPDefaults(v)
	v.SetDefault("security.tls.enabled", false)
	v.SetDefault("security.encryption.key", "")
	v.SetDefault("security.secure_cookie", false)
	v.SetDefault("security.debug", false)
	setSecurityHeadersDefaults(v)
	v.SetDefault("security.cookie_security.secure", false)
	v.SetDefault("security.cookie_security.http_only", true)
	v.SetDefault("security.cookie_security.same_site", "Lax")
	v.SetDefault("security.cookie_security.path", "/")
	v.SetDefault("security.cookie_security.domain", "")
	v.SetDefault("security.cookie_security.max_age", DefaultCookieMaxAge)
	v.SetDefault("security.trust_proxy.enabled", true)
	v.SetDefault("security.trust_proxy.trusted_proxies", []string{"127.0.0.1", "::1"})
}

// setEmailDefaults sets email default values
func setEmailDefaults(v *viper.Viper) {
	v.SetDefault("email.port", DefaultSMTPPort)
	v.SetDefault("email.use_tls", true)
	v.SetDefault("email.use_ssl", false)
	v.SetDefault("email.template", "default")
}

// setStorageDefaults sets storage default values
func setStorageDefaults(v *viper.Viper) {
	v.SetDefault("storage.type", "local")
	v.SetDefault("storage.local.path", "./uploads")
	v.SetDefault("storage.s3.region", "us-east-1")
	v.SetDefault("storage.max_size", DefaultMaxFileSize)
	v.SetDefault("storage.allowed_extensions", []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"})
}

// setCacheDefaults sets cache default values
func setCacheDefaults(v *viper.Viper) {
	v.SetDefault("cache.type", "memory")
	v.SetDefault("cache.redis.host", "localhost")
	v.SetDefault("cache.redis.port", DefaultRedisPort)
	v.SetDefault("cache.redis.db", 0)
	v.SetDefault("cache.memory.max_size", DefaultMemoryCacheSize)
	v.SetDefault("cache.ttl", 1*time.Hour)
}

// setLoggingDefaults sets logging default values
func setLoggingDefaults(v *viper.Viper) {
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")
	v.SetDefault("logging.file", "logs/app.log")
	v.SetDefault("logging.max_size", DefaultLogMaxSize)
	v.SetDefault("logging.max_backups", DefaultLogMaxBackups)
	v.SetDefault("logging.max_age", DefaultLogMaxAge)
	v.SetDefault("logging.compress", true)
}

// setSessionDefaults sets session default values
func setSessionDefaults(v *viper.Viper) {
	v.SetDefault("session.type", "cookie")
	v.SetDefault("session.secret", "session-secret")
	v.SetDefault("session.max_age", DefaultSessionMaxAge)
	v.SetDefault("session.path", "/")
	v.SetDefault("session.secure", false)
	v.SetDefault("session.http_only", true)
	v.SetDefault("session.same_site", "lax")
	v.SetDefault("session.store", "memory")
	v.SetDefault("session.store_file", "storage/sessions/sessions.json")
	v.SetDefault("session.cookie_name", "session")
}

// setAuthDefaults sets authentication default values
func setAuthDefaults(v *viper.Viper) {
	v.SetDefault("auth.require_email_verification", false)
	v.SetDefault("auth.password_min_length", DefaultPasswordMinLength)
	v.SetDefault("auth.password_require_special", true)
	v.SetDefault("auth.session_timeout", DefaultAuthTimeout)
	v.SetDefault("auth.max_login_attempts", DefaultMaxLoginAttempts)
	v.SetDefault("auth.lockout_duration", DefaultLockoutTime)
}

// setFormDefaults sets form default values
func setFormDefaults(v *viper.Viper) {
	v.SetDefault("form.max_file_size", DefaultMaxFileSize)
	v.SetDefault("form.allowed_file_types", []string{"image/jpeg", "image/png", "image/gif", "application/pdf"})
	v.SetDefault("form.max_fields", DefaultMaxFields)
	v.SetDefault("form.max_memory", DefaultMaxFormMemory)
	v.SetDefault("form.validation.strict_mode", false)
	v.SetDefault("form.validation.max_errors", DefaultMaxErrors)
}

// setAPIDefaults sets API default values
func setAPIDefaults(v *viper.Viper) {
	v.SetDefault("api.version", "v1")
	v.SetDefault("api.prefix", "/api")
	v.SetDefault("api.timeout", DefaultRequestTimeout)
	v.SetDefault("api.max_retries", DefaultMaxRetries)
	v.SetDefault("api.rate_limit.enabled", true)
	v.SetDefault("api.rate_limit.rps", DefaultAPIRateLimitRPS)
	v.SetDefault("api.rate_limit.burst", DefaultAPIRateBurst)
}

// setWebDefaults sets web default values
func setWebDefaults(v *viper.Viper) {
	v.SetDefault("web.template_dir", "templates")
	v.SetDefault("web.static_dir", "static")
	v.SetDefault("web.assets_dir", "assets")
	v.SetDefault("web.read_timeout", DefaultReadTimeout)
	v.SetDefault("web.write_timeout", DefaultWriteTimeout)
	v.SetDefault("web.idle_timeout", DefaultIdleTimeout)
	v.SetDefault("web.gzip", true)
}

// setUserDefaults sets user default values
func setUserDefaults(v *viper.Viper) {
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
