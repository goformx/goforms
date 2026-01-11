// Package config provides validation utilities for Viper-based configuration
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// File permission constants
const (
	permUserWrite = 0o200
	permUserRead  = 0o400
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
	Value   any
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// ValidationResult contains the results of configuration validation
type ValidationResult struct {
	IsValid bool
	Errors  []ValidationError
}

// AddError adds a validation error to the result
func (r *ValidationResult) AddError(field, message string, value any) {
	r.IsValid = false
	r.Errors = append(r.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// ValidateConfig validates the complete configuration with detailed error reporting
func ValidateConfig(cfg *Config) ValidationResult {
	result := ValidationResult{IsValid: true}

	// Validate each configuration section
	validateAppConfig(cfg.App, &result)
	validateDatabaseConfig(cfg.Database, &result)
	validateSecurityConfig(cfg.Security, &result)
	validateEmailConfig(cfg.Email, &result)
	validateStorageConfig(cfg.Storage, &result)
	validateCacheConfig(cfg.Cache, &result)
	validateLoggingConfig(cfg.Logging, &result)
	validateSessionConfig(cfg.Session, &result)
	validateAuthConfig(cfg.Auth, &result)
	validateFormConfig(cfg.Form, &result)
	validateAPIConfig(cfg.API, &result)
	validateWebConfig(cfg.Web, &result)
	validateUserConfig(cfg.User, &result)

	// Validate cross-section dependencies
	validateCrossSectionDependencies(cfg, &result)

	return result
}

// validateCrossSectionDependencies validates dependencies between configuration sections
func validateCrossSectionDependencies(cfg *Config, result *ValidationResult) {
	// Check if TLS is enabled but secure cookies are disabled
	if cfg.Security.TLS.Enabled && !cfg.Security.SecureCookie {
		result.AddError("security.secure_cookie",
			"secure cookies should be enabled when TLS is enabled",
			cfg.Security.SecureCookie)
	}

	// Check if session store is Redis but Redis cache is not configured
	if strings.EqualFold(cfg.Session.Store, "redis") && !strings.EqualFold(cfg.Cache.Type, "redis") {
		result.AddError("session.store",
			"Redis session store requires Redis cache configuration",
			cfg.Session.Store)
	}

	// Check if email verification is required but email is not configured
	if cfg.Auth.RequireEmailVerification && cfg.Email.Host == "" {
		result.AddError(
			"auth.require_email_verification",
			"email verification requires email configuration",
			cfg.Auth.RequireEmailVerification,
		)
	}
}

// Helper functions

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

// isWritableDirectory checks if a directory is writable
func isWritableDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir() && (info.Mode()&permUserWrite) != 0
}

// isReadableDirectory checks if a directory is readable
func isReadableDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir() && (info.Mode()&permUserRead) != 0
}

// isValidEmail checks if an email address is valid
func isValidEmail(email string) bool {
	// Simple email validation - in production, consider using a more robust library
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// ValidateEnvironmentVariables validates that required environment variables are set
func ValidateEnvironmentVariables() ValidationResult {
	result := ValidationResult{IsValid: true}

	// Check for required environment variables
	requiredVars := []string{
		"APP_NAME",
		"DB_HOST",
		"DB_NAME",
		"DB_USERNAME",
		"DB_PASSWORD",
	}

	for _, envVar := range requiredVars {
		if os.Getenv(envVar) == "" {
			result.AddError(envVar, "required environment variable is not set", "")
		}
	}

	// Validate environment variable formats
	if port := os.Getenv("APP_PORT"); port != "" {
		if _, err := strconv.Atoi(port); err != nil {
			result.AddError("APP_PORT", "must be a valid integer", port)
		}
	}

	if timeout := os.Getenv("APP_READ_TIMEOUT"); timeout != "" {
		if _, err := time.ParseDuration(timeout); err != nil {
			result.AddError("APP_READ_TIMEOUT", "must be a valid duration", timeout)
		}
	}

	return result
}
