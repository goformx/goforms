package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
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
	SSLMode  string `validate:"required,oneof=disable enable verify-ca verify-full"`
}

type RateLimitConfig struct {
	Rate  int `validate:"required,min=1"`
	Burst int `validate:"required,min=1"`
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host: getEnvString("SERVER_HOST", "localhost"),
			Port: getEnvInt("SERVER_PORT", 8080),
		},
		Database: DatabaseConfig{
			Host:     getEnvString("MYSQL_HOSTNAME", "db"),
			Port:     getEnvInt("MYSQL_PORT", 3306),
			User:     getEnvString("MYSQL_USER", "goforms"),
			Password: getEnvString("MYSQL_PASSWORD", "goforms"),
			DBName:   getEnvString("MYSQL_DATABASE", "goforms"),
			SSLMode:  getEnvString("DB_SSL_MODE", "disable"),
		},
		RateLimit: RateLimitConfig{
			Rate:  getEnvInt("RATE_LIMIT", 100),
			Burst: getEnvInt("RATE_BURST", 5),
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

func getEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
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

func NewConfig() *Config {
	cfg := &Config{}

	// Set default values
	cfg.Server.Host = os.Getenv("SERVER_HOST")
	if cfg.Server.Host == "" {
		cfg.Server.Host = "localhost"
	}

	// Default port 8080
	cfg.Server.Port = 8080

	// Default rate limit of 100 requests per second
	cfg.RateLimit.Rate = 100

	return cfg
}
