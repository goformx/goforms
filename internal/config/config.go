package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	RateLimit RateLimitConfig
}

type ServerConfig struct {
	Host string
	Port int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RateLimitConfig struct {
	Rate  int
	Burst int
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host: getEnvString("SERVER_HOST", "localhost"),
			Port: getEnvInt("SERVER_PORT", 8080),
		},
		Database: DatabaseConfig{
			Host:     getEnvString("POSTGRES_HOSTNAME", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnvString("POSTGRES_USER", "postgres"),
			Password: getEnvString("POSTGRES_PASSWORD", "postgres"),
			DBName:   getEnvString("POSTGRES_DB", "goforms"),
			SSLMode:  getEnvString("DB_SSL_MODE", "disable"),
		},
		RateLimit: RateLimitConfig{
			Rate:  getEnvInt("RATE_LIMIT", 100),
			Burst: getEnvInt("RATE_BURST", 5),
		},
	}

	return cfg, nil
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
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
