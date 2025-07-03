// Package config provides validation utilities for Viper-based configuration
package config

import (
	"strings"
)

// =============================================================================
// AUTH CONFIGURATION VALIDATION
// =============================================================================

// validateAuthConfig validates authentication configuration
func validateAuthConfig(cfg AuthConfig, result *ValidationResult) {
	validateAuthConfigPasswordPolicy(cfg, result)
	validateAuthConfigSession(cfg, result)
	validateAuthConfigEmailVerification(cfg, result)
}

func validateAuthConfigPasswordPolicy(cfg AuthConfig, result *ValidationResult) {
	if cfg.PasswordMinLength <= 0 {
		result.AddError("auth.password_min_length", "password min length must be positive", cfg.PasswordMinLength)
	}

	if cfg.PasswordMinLength < 6 {
		result.AddError("auth.password_min_length", "password min length must be at least 6", cfg.PasswordMinLength)
	}
}

func validateAuthConfigSession(cfg AuthConfig, result *ValidationResult) {
	if cfg.SessionTimeout <= 0 {
		result.AddError("auth.session_timeout", "session timeout must be positive", cfg.SessionTimeout)
	}

	if cfg.MaxLoginAttempts <= 0 {
		result.AddError("auth.max_login_attempts", "max login attempts must be positive", cfg.MaxLoginAttempts)
	}

	if cfg.LockoutDuration <= 0 {
		result.AddError("auth.lockout_duration", "lockout duration must be positive", cfg.LockoutDuration)
	}
}

func validateAuthConfigEmailVerification(cfg AuthConfig, result *ValidationResult) {
	if cfg.RequireEmailVerification {
		// Email verification requires email configuration
		// This is validated in cross-section dependencies
	}
}

// =============================================================================
// FORM CONFIGURATION VALIDATION
// =============================================================================

// validateFormConfig validates form configuration
func validateFormConfig(cfg FormConfig, result *ValidationResult) {
	validateFormConfigMaxFields(cfg, result)
	validateFormConfigFileUpload(cfg, result)
	validateFormConfigValidation(cfg, result)
}

func validateFormConfigMaxFields(cfg FormConfig, result *ValidationResult) {
	if cfg.MaxFields <= 0 {
		result.AddError("form.max_fields", "max fields must be positive", cfg.MaxFields)
	}

	if cfg.MaxFields > 100 {
		result.AddError("form.max_fields", "max fields cannot exceed 100", cfg.MaxFields)
	}
}

func validateFormConfigFileUpload(cfg FormConfig, result *ValidationResult) {
	if cfg.MaxFileSize <= 0 {
		result.AddError("form.max_file_size", "max file size must be positive", cfg.MaxFileSize)
	}

	if cfg.MaxMemory <= 0 {
		result.AddError("form.max_memory", "max memory must be positive", cfg.MaxMemory)
	}

	if len(cfg.AllowedFileTypes) == 0 {
		result.AddError("form.allowed_file_types", "at least one allowed file type is required", cfg.AllowedFileTypes)
	}
}

func validateFormConfigValidation(cfg FormConfig, result *ValidationResult) {
	if cfg.Validation.MaxErrors <= 0 {
		result.AddError("form.validation.max_errors", "max errors must be positive", cfg.Validation.MaxErrors)
	}
}

// =============================================================================
// USER CONFIGURATION VALIDATION
// =============================================================================

// validateUserConfig validates user configuration
func validateUserConfig(cfg UserConfig, result *ValidationResult) {
	validateUserConfigAdmin(cfg, result)
	validateUserConfigDefault(cfg, result)
}

func validateUserConfigAdmin(cfg UserConfig, result *ValidationResult) {
	if cfg.Admin.Email == "" {
		result.AddError("user.admin.email", "admin email is required", cfg.Admin.Email)
	}

	if !isValidEmail(cfg.Admin.Email) {
		result.AddError("user.admin.email", "admin email must be valid", cfg.Admin.Email)
	}

	if cfg.Admin.Password == "" {
		result.AddError("user.admin.password", "admin password is required", cfg.Admin.Password)
	}

	if len(cfg.Admin.Password) < 8 {
		result.AddError("user.admin.password", "admin password must be at least 8 characters", len(cfg.Admin.Password))
	}
}

func validateUserConfigDefault(cfg UserConfig, result *ValidationResult) {
	if cfg.Default.Role == "" {
		result.AddError("user.default.role", "default role is required", cfg.Default.Role)
	}

	if len(cfg.Default.Permissions) == 0 {
		result.AddError("user.default.permissions", "at least one default permission is required", cfg.Default.Permissions)
	}
}

// =============================================================================
// API CONFIGURATION VALIDATION
// =============================================================================

// validateAPIConfig validates API configuration
func validateAPIConfig(cfg APIConfig, result *ValidationResult) {
	validateAPIConfigVersion(cfg, result)
	validateAPIConfigTimeout(cfg, result)
	validateAPIConfigRateLimit(cfg, result)
}

func validateAPIConfigVersion(cfg APIConfig, result *ValidationResult) {
	if cfg.Version == "" {
		result.AddError("api.version", "API version is required", cfg.Version)
	}

	validVersions := []string{"v1", "v2"}
	versionValid := false

	for _, version := range validVersions {
		if strings.EqualFold(cfg.Version, version) {
			versionValid = true

			break
		}
	}

	if !versionValid {
		result.AddError("api.version", "API version must be one of: v1, v2", cfg.Version)
	}
}

func validateAPIConfigTimeout(cfg APIConfig, result *ValidationResult) {
	if cfg.Timeout <= 0 {
		result.AddError("api.timeout", "API timeout must be positive", cfg.Timeout)
	}

	if cfg.MaxRetries < 0 {
		result.AddError("api.max_retries", "max retries must be non-negative", cfg.MaxRetries)
	}
}

func validateAPIConfigRateLimit(cfg APIConfig, result *ValidationResult) {
	if cfg.RateLimit.Enabled {
		if cfg.RateLimit.RPS <= 0 {
			result.AddError("api.rate_limit.rps", "API rate limit RPS must be positive", cfg.RateLimit.RPS)
		}
	}
}

// =============================================================================
// WEB CONFIGURATION VALIDATION
// =============================================================================

// validateWebConfig validates web configuration
func validateWebConfig(cfg WebConfig, result *ValidationResult) {
	validateWebConfigTemplate(cfg, result)
	validateWebConfigStatic(cfg, result)
	validateWebConfigTimeouts(cfg, result)
}

func validateWebConfigTemplate(cfg WebConfig, result *ValidationResult) {
	if cfg.TemplateDir == "" {
		result.AddError("web.template_dir", "template directory is required", cfg.TemplateDir)
	}

	if !isReadableDirectory(cfg.TemplateDir) {
		result.AddError("web.template_dir", "template directory must be a readable directory", cfg.TemplateDir)
	}
}

func validateWebConfigStatic(cfg WebConfig, result *ValidationResult) {
	if cfg.StaticDir == "" {
		result.AddError("web.static_dir", "static directory is required", cfg.StaticDir)
	}

	if !isReadableDirectory(cfg.StaticDir) {
		result.AddError("web.static_dir", "static directory must be a readable directory", cfg.StaticDir)
	}
}

func validateWebConfigTimeouts(cfg WebConfig, result *ValidationResult) {
	if cfg.ReadTimeout <= 0 {
		result.AddError("web.read_timeout", "read timeout must be positive", cfg.ReadTimeout)
	}

	if cfg.WriteTimeout <= 0 {
		result.AddError("web.write_timeout", "write timeout must be positive", cfg.WriteTimeout)
	}

	if cfg.IdleTimeout <= 0 {
		result.AddError("web.idle_timeout", "idle timeout must be positive", cfg.IdleTimeout)
	}
}
