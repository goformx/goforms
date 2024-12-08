package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/jonesrussell/goforms/internal/config/database"
	"github.com/jonesrussell/goforms/internal/config/server"
)

// Config represents the complete application configuration
type Config struct {
	App       AppConfig       `validate:"required"`
	Server    server.Config   `validate:"required"`
	Database  database.Config `validate:"required"`
	Security  SecurityConfig  `validate:"required"`
	RateLimit RateLimitConfig `validate:"required"`
	Logging   LoggingConfig   `validate:"required"`
}

// AppConfig contains basic application settings
type AppConfig struct {
	Name string `validate:"required"`
	Env  string `validate:"required,oneof=development staging production"`
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	CorsAllowedOrigins []string `validate:"required"`
	CorsAllowedMethods []string `validate:"required"`
	CorsAllowedHeaders []string `validate:"required"`
	CorsMaxAge         int      `validate:"required,min=0"`
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	Rate  int `validate:"required,min=1"`
	Burst int `validate:"required,min=1"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `validate:"required,oneof=debug info warn error"`
	Format string `validate:"required,oneof=json console"`
}

// New provides the application configuration for fx
func New() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	cfg := &Config{
		App: AppConfig{
			Name: getEnvString("APP_NAME", "goforms"),
			Env:  getEnvString("APP_ENV", "development"),
		},
		Server: server.Config{
			Host: getEnvString("SERVER_HOST", "localhost"),
			Port: getEnvInt("SERVER_PORT", 8090),
			Timeouts: server.TimeoutConfig{
				Read:  getEnvDuration("READ_TIMEOUT", 5*time.Second),
				Write: getEnvDuration("WRITE_TIMEOUT", 10*time.Second),
				Idle:  getEnvDuration("IDLE_TIMEOUT", 120*time.Second),
			},
		},
		Database: database.Config{
			Host: getEnvString("DB_HOSTNAME", "db"),
			Port: getEnvInt("DB_PORT", 3306),
			Credentials: database.Credentials{
				User:     getEnvString("DB_USER", "goforms"),
				Password: getEnvString("DB_PASSWORD", "goforms"),
				DBName:   getEnvString("DB_DATABASE", "goforms"),
			},
			ConnectionPool: database.PoolConfig{
				MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
				MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
				ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			},
		},
		Security: SecurityConfig{
			CorsAllowedOrigins: getEnvStringSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
			CorsAllowedMethods: getEnvStringSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE"}),
			CorsAllowedHeaders: getEnvStringSlice("CORS_ALLOWED_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization"}),
			CorsMaxAge:         getEnvInt("CORS_MAX_AGE", 300),
		},
		RateLimit: RateLimitConfig{
			Rate:  getEnvInt("RATE_LIMIT", 100),
			Burst: getEnvInt("RATE_BURST", 5),
		},
		Logging: LoggingConfig{
			Level:  getEnvString("LOG_LEVEL", "debug"),
			Format: getEnvString("LOG_FORMAT", "json"),
		},
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Helper functions for environment variables
func getEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
