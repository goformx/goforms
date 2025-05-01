package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
)

// CORSOriginsDecoder handles parsing of CORS allowed origins
type CORSOriginsDecoder []string

func (c *CORSOriginsDecoder) Decode(value string) error {
	if value == "" {
		*c = []string{"http://localhost:3000"}
		return nil
	}
	*c = strings.Split(value, ",")
	return nil
}

// CORSMethodsDecoder handles parsing of CORS allowed methods
type CORSMethodsDecoder []string

func (c *CORSMethodsDecoder) Decode(value string) error {
	if value == "" {
		*c = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		return nil
	}
	*c = strings.Split(value, ",")
	return nil
}

// CORSHeadersDecoder handles parsing of CORS allowed headers
type CORSHeadersDecoder []string

func (c *CORSHeadersDecoder) Decode(value string) error {
	if value == "" {
		*c = []string{"Origin", "Content-Type", "Accept"}
		return nil
	}
	*c = strings.Split(value, ",")
	return nil
}

// Config represents the complete application configuration
type Config struct {
	App       AppConfig
	Server    ServerConfig
	Database  DatabaseConfig
	Security  SecurityConfig
	RateLimit RateLimitConfig
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name  string `envconfig:"GOFORMS_APP_NAME" default:"GoForms"`
	Env   string `envconfig:"GOFORMS_APP_ENV" default:"production"`
	Debug bool   `envconfig:"GOFORMS_APP_DEBUG" default:"false"`
	Port  int    `envconfig:"GOFORMS_APP_PORT" default:"8090"`
	Host  string `envconfig:"GOFORMS_APP_HOST" default:"localhost"`
}

// DatabaseConfig holds all database-related configuration
type DatabaseConfig struct {
	Host           string        `envconfig:"GOFORMS_DB_HOST" validate:"required"`
	Port           int           `envconfig:"GOFORMS_DB_PORT" validate:"required"`
	User           string        `envconfig:"GOFORMS_DB_USER" validate:"required"`
	Password       string        `envconfig:"GOFORMS_DB_PASSWORD" validate:"required"`
	Name           string        `envconfig:"GOFORMS_DB_NAME" validate:"required"`
	MaxOpenConns   int           `envconfig:"GOFORMS_DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns   int           `envconfig:"GOFORMS_DB_MAX_IDLE_CONNS" default:"25"`
	ConnMaxLifetme time.Duration `envconfig:"GOFORMS_DB_CONN_MAX_LIFETIME" default:"5m"`
}

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Host            string        `envconfig:"GOFORMS_APP_HOST" default:"localhost"`
	Port            int           `envconfig:"GOFORMS_APP_PORT" default:"8099"`
	ReadTimeout     time.Duration `envconfig:"GOFORMS_READ_TIMEOUT" default:"5s"`
	WriteTimeout    time.Duration `envconfig:"GOFORMS_WRITE_TIMEOUT" default:"10s"`
	IdleTimeout     time.Duration `envconfig:"GOFORMS_IDLE_TIMEOUT" default:"120s"`
	ShutdownTimeout time.Duration `envconfig:"GOFORMS_SHUTDOWN_TIMEOUT" default:"30s"`
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	JWTSecret            string `envconfig:"GOFORMS_JWT_SECRET" validate:"required"`
	CSRF                 CSRFConfig
	CorsAllowedOrigins   CORSOriginsDecoder `envconfig:"GOFORMS_CORS_ALLOWED_ORIGINS"`
	CorsAllowedMethods   CORSMethodsDecoder `envconfig:"GOFORMS_CORS_ALLOWED_METHODS"`
	CorsAllowedHeaders   CORSHeadersDecoder `envconfig:"GOFORMS_CORS_ALLOWED_HEADERS"`
	CorsMaxAge           int                `envconfig:"GOFORMS_CORS_MAX_AGE" default:"3600"`
	CorsAllowCredentials bool               `envconfig:"GOFORMS_CORS_ALLOW_CREDENTIALS" default:"true"`
	RequestTimeout       time.Duration      `envconfig:"GOFORMS_REQUEST_TIMEOUT" default:"30s"`

	// Form-specific CORS settings
	FormCorsAllowedOrigins CORSOriginsDecoder `envconfig:"GOFORMS_FORM_CORS_ALLOWED_ORIGINS"`
	FormCorsAllowedMethods CORSMethodsDecoder `envconfig:"GOFORMS_FORM_CORS_ALLOWED_METHODS"`
	FormCorsAllowedHeaders CORSHeadersDecoder `envconfig:"GOFORMS_FORM_CORS_ALLOWED_HEADERS"`
	FormRateLimit          int                `envconfig:"GOFORMS_FORM_RATE_LIMIT" default:"20"`
	FormRateLimitWindow    time.Duration      `envconfig:"GOFORMS_FORM_RATE_LIMIT_WINDOW" default:"1m"`
}

// CSRFConfig holds CSRF-related configuration
type CSRFConfig struct {
	Enabled bool   `envconfig:"GOFORMS_CSRF_ENABLED" default:"true"`
	Secret  string `envconfig:"GOFORMS_CSRF_SECRET" validate:"required"`
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	Enabled    bool          `envconfig:"GOFORMS_RATE_LIMIT_ENABLED" default:"true"`
	Rate       int           `envconfig:"GOFORMS_RATE_LIMIT" default:"100"`
	Burst      int           `envconfig:"GOFORMS_RATE_BURST" default:"5"`
	TimeWindow time.Duration `envconfig:"GOFORMS_RATE_LIMIT_TIME_WINDOW" default:"1m"`
	PerIP      bool          `envconfig:"GOFORMS_RATE_LIMIT_PER_IP" default:"true"`
}

// New creates a new Config with default values
func New() (*Config, error) {
	var cfg Config

	// Debug environment variables
	if os.Getenv("GOFORMS_APP_DEBUG") == "true" {
		fmt.Fprintln(os.Stdout, "\n=== Environment Variables ===")
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, "GOFORMS_") {
				fmt.Fprintln(os.Stdout, env)
			}
		}
		fmt.Fprintln(os.Stdout, "===========================\n")
	}

	if err := envconfig.Process("GOFORMS", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}

	// Debug output after configuration is loaded
	if cfg.App.Debug {
		fmt.Fprintln(os.Stdout, "\n=== Loaded Configuration ===")
		fmt.Fprintf(os.Stdout, "App Configuration:\n")
		fmt.Fprintf(os.Stdout, "  Name: %s\n", cfg.App.Name)
		fmt.Fprintf(os.Stdout, "  Env: %s\n", cfg.App.Env)
		fmt.Fprintf(os.Stdout, "  Debug: %v\n", cfg.App.Debug)
		fmt.Fprintf(os.Stdout, "  Port: %d\n", cfg.App.Port)
		fmt.Fprintf(os.Stdout, "  Host: %s\n", cfg.App.Host)
		
		fmt.Fprintf(os.Stdout, "\nServer Configuration:\n")
		fmt.Fprintf(os.Stdout, "  Host: %s\n", cfg.Server.Host)
		fmt.Fprintf(os.Stdout, "  Port: %d\n", cfg.Server.Port)
		fmt.Fprintf(os.Stdout, "  Read Timeout: %v\n", cfg.Server.ReadTimeout)
		fmt.Fprintf(os.Stdout, "  Write Timeout: %v\n", cfg.Server.WriteTimeout)
		fmt.Fprintf(os.Stdout, "  Idle Timeout: %v\n", cfg.Server.IdleTimeout)
		
		fmt.Fprintf(os.Stdout, "\nDatabase Configuration:\n")
		fmt.Fprintf(os.Stdout, "  Host: %s\n", cfg.Database.Host)
		fmt.Fprintf(os.Stdout, "  Port: %d\n", cfg.Database.Port)
		fmt.Fprintf(os.Stdout, "  Name: %s\n", cfg.Database.Name)
		fmt.Fprintf(os.Stdout, "  User: %s\n", cfg.Database.User)
		fmt.Fprintf(os.Stdout, "  Max Open Connections: %d\n", cfg.Database.MaxOpenConns)
		fmt.Fprintf(os.Stdout, "  Max Idle Connections: %d\n", cfg.Database.MaxIdleConns)
		fmt.Fprintf(os.Stdout, "  Connection Max Lifetime: %v\n", cfg.Database.ConnMaxLifetme)
		
		fmt.Fprintf(os.Stdout, "\nSecurity Configuration:\n")
		fmt.Fprintf(os.Stdout, "  CSRF Enabled: %v\n", cfg.Security.CSRF.Enabled)
		fmt.Fprintf(os.Stdout, "  CORS Allowed Origins: %v\n", cfg.Security.CorsAllowedOrigins)
		fmt.Fprintf(os.Stdout, "  CORS Allowed Methods: %v\n", cfg.Security.CorsAllowedMethods)
		fmt.Fprintf(os.Stdout, "  CORS Allowed Headers: %v\n", cfg.Security.CorsAllowedHeaders)
		fmt.Fprintf(os.Stdout, "  CORS Max Age: %d\n", cfg.Security.CorsMaxAge)
		fmt.Fprintf(os.Stdout, "  CORS Allow Credentials: %v\n", cfg.Security.CorsAllowCredentials)
		
		fmt.Fprintf(os.Stdout, "\nRate Limit Configuration:\n")
		fmt.Fprintf(os.Stdout, "  Enabled: %v\n", cfg.RateLimit.Enabled)
		fmt.Fprintf(os.Stdout, "  Rate: %d\n", cfg.RateLimit.Rate)
		fmt.Fprintf(os.Stdout, "  Burst: %d\n", cfg.RateLimit.Burst)
		fmt.Fprintf(os.Stdout, "  Time Window: %v\n", cfg.RateLimit.TimeWindow)
		fmt.Fprintf(os.Stdout, "  Per IP: %v\n", cfg.RateLimit.PerIP)
		fmt.Fprintln(os.Stdout, "===========================\n")
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
