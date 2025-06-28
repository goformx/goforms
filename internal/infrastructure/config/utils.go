// Package config provides utility functions for configuration management.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigUtils provides utility functions for configuration management
type ConfigUtils struct{}

// NewConfigUtils creates a new configuration utilities instance
func NewConfigUtils() *ConfigUtils {
	return &ConfigUtils{}
}

// GetConfigFilePaths returns the search paths for configuration files
func (cu *ConfigUtils) GetConfigFilePaths() []string {
	return []string{
		".",
		"./config",
		"/etc/goforms",
		"$HOME/.goforms",
	}
}

// GetConfigFileNames returns the possible configuration file names
func (cu *ConfigUtils) GetConfigFileNames() []string {
	return []string{
		"config",
		"config.yaml",
		"config.yml",
		"config.json",
		"config.toml",
		"config.env",
	}
}

// FindConfigFile searches for a configuration file in the standard paths
func (cu *ConfigUtils) FindConfigFile() (string, error) {
	paths := cu.GetConfigFilePaths()
	names := cu.GetConfigFileNames()

	for _, path := range paths {
		// Expand environment variables in path
		expandedPath := os.ExpandEnv(path)

		for _, name := range names {
			fullPath := filepath.Join(expandedPath, name)
			if _, err := os.Stat(fullPath); err == nil {
				return fullPath, nil
			}
		}
	}

	return "", fmt.Errorf("no configuration file found in paths: %v", paths)
}

// ValidateConfigFile validates that a configuration file exists and is readable
func (cu *ConfigUtils) ValidateConfigFile(filepath string) error {
	if filepath == "" {
		return fmt.Errorf("config file path is empty")
	}

	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", filepath)
	}

	// Check if file is readable
	if _, err := os.ReadFile(filepath); err != nil {
		return fmt.Errorf("config file is not readable: %s", err)
	}

	return nil
}

// GetEnvironmentFromFile attempts to determine the environment from a config file
func (cu *ConfigUtils) GetEnvironmentFromFile(filepath string) (string, error) {
	if filepath == "" {
		return "development", nil // Default environment
	}

	// Extract environment from filename
	base := filepath.Base(filepath)
	if strings.Contains(base, "production") {
		return "production", nil
	}
	if strings.Contains(base, "staging") {
		return "staging", nil
	}
	if strings.Contains(base, "test") {
		return "test", nil
	}
	if strings.Contains(base, "development") {
		return "development", nil
	}

	return "development", nil // Default environment
}

// FormatConfigSummary formats a configuration summary for display
func (cu *ConfigUtils) FormatConfigSummary(summary map[string]interface{}) string {
	var result strings.Builder

	result.WriteString("Configuration Summary:\n")
	result.WriteString("=====================\n\n")

	for section, data := range summary {
		result.WriteString(fmt.Sprintf("%s:\n", strings.Title(section)))
		if sectionData, ok := data.(map[string]interface{}); ok {
			for key, value := range sectionData {
				result.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}

// SanitizeConfigForLogging removes sensitive information from config for logging
func (cu *ConfigUtils) SanitizeConfigForLogging(config *Config) *Config {
	if config == nil {
		return nil
	}

	// Create a copy to avoid modifying the original
	sanitized := *config

	// Sanitize database password
	if sanitized.Database.Password != "" {
		sanitized.Database.Password = "[REDACTED]"
	}

	// Sanitize security secrets
	if sanitized.Security.CSRF.Secret != "" {
		sanitized.Security.CSRF.Secret = "[REDACTED]"
	}
	if sanitized.Security.Encryption.Key != "" {
		sanitized.Security.Encryption.Key = "[REDACTED]"
	}

	// Sanitize session secret
	if sanitized.Session.Secret != "" {
		sanitized.Session.Secret = "[REDACTED]"
	}

	// Sanitize email password
	if sanitized.Email.Password != "" {
		sanitized.Email.Password = "[REDACTED]"
	}

	// Sanitize S3 credentials
	if sanitized.Storage.S3.AccessKey != "" {
		sanitized.Storage.S3.AccessKey = "[REDACTED]"
	}
	if sanitized.Storage.S3.SecretKey != "" {
		sanitized.Storage.S3.SecretKey = "[REDACTED]"
	}

	// Sanitize Redis password
	if sanitized.Cache.Redis.Password != "" {
		sanitized.Cache.Redis.Password = "[REDACTED]"
	}

	// Sanitize user passwords
	if sanitized.User.Admin.Password != "" {
		sanitized.User.Admin.Password = "[REDACTED]"
	}

	return &sanitized
}
