package database

import (
	"fmt"
	"os"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

// NewConfig creates database configuration from environment variables
func NewConfig(logger logging.Logger) *Config {
	logger.Debug("loading database configuration")

	config := &Config{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "3306"),
		Name:     getEnvOrDefault("DB_NAME", "goforms"),
		User:     getEnvOrDefault("DB_USER", "root"),
		Password: os.Getenv("DB_PASSWORD"),
		SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
	}

	logger.Debug("database configuration loaded",
		logging.String("host", config.Host),
		logging.String("port", config.Port),
		logging.String("name", config.Name),
		logging.String("user", config.User),
		logging.String("ssl_mode", config.SSLMode),
	)

	return config
}

// DSN returns the database connection string
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true&tls=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
