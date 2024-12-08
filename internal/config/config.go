package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig    `validate:"required"`
	Database  DatabaseConfig  `validate:"required"`
	RateLimit RateLimitConfig `validate:"required"`
}

type ServerConfig struct {
	Host string `validate:"required"`
	Port int    `validate:"required,min=1,max=65535"`
}

type DatabaseConfig struct {
	Host     string `validate:"required"`
	Port     int    `validate:"required,min=1,max=65535"`
	User     string `validate:"required"`
	Password string `validate:"required"`
	DBName   string `validate:"required"`
}

type RateLimitConfig struct {
	Rate  int `validate:"required,min=1"`
	Burst int `validate:"required,min=1"`
}

// Load initializes configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Only log warning as .env is optional in production
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Host: getEnvStringOrPanic("SERVER_HOST", "localhost"),
			Port: getEnvIntOrPanic("SERVER_PORT", 8090),
		},
		Database: DatabaseConfig{
			Host:     getEnvStringOrPanic("MYSQL_HOSTNAME", "db"),
			Port:     getEnvIntOrPanic("MYSQL_PORT", 3306),
			User:     getEnvStringOrPanic("MYSQL_USER", "goforms"),
			Password: getEnvStringOrPanic("MYSQL_PASSWORD", "goforms"),
			DBName:   getEnvStringOrPanic("MYSQL_DATABASE", "goforms"),
		},
		RateLimit: RateLimitConfig{
			Rate:  getEnvIntOrPanic("RATE_LIMIT", 100),
			Burst: getEnvIntOrPanic("RATE_BURST", 5),
		},
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		d.User, d.Password, d.Host, d.Port, d.DBName)
}

// getEnvStringOrPanic gets an environment variable or returns the default value
// It panics if the environment variable is empty and no default is provided
func getEnvStringOrPanic(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	if defaultValue == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return defaultValue
}

// getEnvIntOrPanic gets an environment variable as integer or returns the default value
// It panics if the environment variable is not a valid integer
func getEnvIntOrPanic(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			panic(fmt.Sprintf("environment variable %s must be a valid integer: %v", key, err))
		}
		return intValue
	}
	return defaultValue
}

// Provide a constructor for fx
func NewConfig() (*Config, error) {
	return Load()
}
