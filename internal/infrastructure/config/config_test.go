package config_test

import (
	"testing"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/stretchr/testify/require"
)

func TestNew_ValidConfig(t *testing.T) {
	envVars := map[string]string{
		"APP_NAME":             "testapp",
		"APP_ENV":              "development",
		"APP_DEBUG":            "true",
		"DB_USER":              "testuser",
		"DB_PASSWORD":          "testpass",
		"DB_NAME":              "testdb",
		"DB_HOST":              "localhost",
		"DB_PORT":              "3306",
		"APP_PORT":             "8080",
		"APP_HOST":             "localhost",
		"CORS_ALLOWED_ORIGINS": "http://localhost:3000",
		"CORS_ALLOWED_METHODS": "GET,POST,PUT,DELETE,OPTIONS",
	}

	// Set environment variables for test
	for k, v := range envVars {
		t.Setenv(k, v)
	}

	cfg, err := config.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config but got nil")
	}

	// Verify App configuration
	if cfg.App.Name != "testapp" {
		t.Errorf("expected App.Name to be %q, got %q", "testapp", cfg.App.Name)
	}
	if cfg.App.Env != "development" {
		t.Errorf("expected App.Env to be %q, got %q", "development", cfg.App.Env)
	}
	if !cfg.App.Debug {
		t.Error("expected App.Debug to be true")
	}
	if cfg.App.Port != 8080 {
		t.Errorf("expected App.Port to be %d, got %d", 8080, cfg.App.Port)
	}
	if cfg.App.Host != "localhost" {
		t.Errorf("expected App.Host to be %q, got %q", "localhost", cfg.App.Host)
	}

	// Verify Database configuration
	if cfg.Database.User != "testuser" {
		t.Errorf("expected Database.User to be %q, got %q", "testuser", cfg.Database.User)
	}
	if cfg.Database.Password != "testpass" {
		t.Errorf("expected Database.Password to be %q, got %q", "testpass", cfg.Database.Password)
	}
	if cfg.Database.Name != "testdb" {
		t.Errorf("expected Database.Name to be %q, got %q", "testdb", cfg.Database.Name)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("expected Database.Host to be %q, got %q", "localhost", cfg.Database.Host)
	}
	if cfg.Database.Port != 3306 {
		t.Errorf("expected Database.Port to be %d, got %d", 3306, cfg.Database.Port)
	}
}

func TestNew_InvalidConfig(t *testing.T) {
	envVars := map[string]string{
		"APP_NAME":    "testapp",
		"APP_ENV":     "development",
		"APP_DEBUG":   "true",
		"APP_PORT":    "8080",
		"APP_HOST":    "localhost",
		"DB_USER":     "testuser",
		"DB_PASSWORD": "testpass",
		"DB_NAME":     "testdb",
		"DB_HOST":     "localhost",
		"DB_PORT":     "invalid", // Invalid port number
	}

	// Set environment variables for test
	for k, v := range envVars {
		t.Setenv(k, v)
	}

	_, err := config.New()
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestSecurityConfig(t *testing.T) {
	t.Run("default_security_settings", func(t *testing.T) {
		// Set required database config
		t.Setenv("DB_USER", "testuser")
		t.Setenv("DB_PASSWORD", "testpass")
		t.Setenv("DB_NAME", "testdb")

		config, err := config.New()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		if len(config.Security.CorsAllowedMethods) != len(expectedMethods) {
			t.Errorf("expected %d methods, got %d", len(expectedMethods), len(config.Security.CorsAllowedMethods))
		}
		for i, method := range expectedMethods {
			if config.Security.CorsAllowedMethods[i] != method {
				t.Errorf("expected method %q at index %d, got %q", method, i, config.Security.CorsAllowedMethods[i])
			}
		}
	})

	t.Run("custom_security_settings", func(t *testing.T) {
		// Set required database config
		t.Setenv("DB_USER", "testuser")
		t.Setenv("DB_PASSWORD", "testpass")
		t.Setenv("DB_NAME", "testdb")
		t.Setenv("CORS_ALLOWED_METHODS", "GET,POST")

		config, err := config.New()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedMethods := []string{"GET", "POST"}
		if len(config.Security.CorsAllowedMethods) != len(expectedMethods) {
			t.Errorf("expected %d methods, got %d", len(expectedMethods), len(config.Security.CorsAllowedMethods))
		}
		for i, method := range expectedMethods {
			if config.Security.CorsAllowedMethods[i] != method {
				t.Errorf("expected method %q at index %d, got %q", method, i, config.Security.CorsAllowedMethods[i])
			}
		}
	})
}

func TestRateLimitConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected config.RateLimitConfig
	}{
		{
			name: "default values",
			envVars: map[string]string{},
			expected: config.RateLimitConfig{
				Enabled: true,
				Rate:    100,
				Burst:   50,
			},
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"RATE_LIMIT_ENABLED": "false",
				"RATE_LIMIT_RATE":    "200",
				"RATE_LIMIT_BURST":   "100",
			},
			expected: config.RateLimitConfig{
				Enabled: false,
				Rate:    200,
				Burst:   100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			// Create config instance
			cfg, err := config.New()
			require.NoError(t, err)

			// Verify rate limit config
			require.Equal(t, tt.expected.Enabled, cfg.RateLimit.Enabled)
			require.Equal(t, tt.expected.Rate, cfg.RateLimit.Rate)
			require.Equal(t, tt.expected.Burst, cfg.RateLimit.Burst)
		})
	}
}
