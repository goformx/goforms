package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Save original env vars
	originalEnv := map[string]string{
		"APP_NAME":    os.Getenv("APP_NAME"),
		"APP_ENV":     os.Getenv("APP_ENV"),
		"DB_USER":     os.Getenv("DB_USER"),
		"DB_PASSWORD": os.Getenv("DB_PASSWORD"),
		"DB_NAME":     os.Getenv("DB_NAME"),
		"DB_HOST":     os.Getenv("DB_HOST"),
		"DB_PORT":     os.Getenv("DB_PORT"),
		"LOG_LEVEL":   os.Getenv("LOG_LEVEL"),
		"LOG_FORMAT":  os.Getenv("LOG_FORMAT"),
		"SERVER_PORT": os.Getenv("SERVER_PORT"),
		"SERVER_HOST": os.Getenv("SERVER_HOST"),
	}

	// Cleanup function to restore original env vars
	defer func() {
		for k, v := range originalEnv {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	}()

	tests := []struct {
		name      string
		envVars   map[string]string
		wantError bool
	}{
		{
			name: "valid configuration",
			envVars: map[string]string{
				"APP_NAME":             "goforms",
				"APP_ENV":              "development",
				"DB_USER":              "testuser",
				"DB_PASSWORD":          "testpass",
				"DB_NAME":              "testdb",
				"DB_HOST":              "localhost",
				"DB_PORT":              "3306",
				"LOG_LEVEL":            "debug",
				"LOG_FORMAT":           "json",
				"SERVER_PORT":          "8080",
				"SERVER_HOST":          "localhost",
				"CORS_ALLOWED_ORIGINS": "http://localhost:3000",
				"CORS_ALLOWED_METHODS": "GET,POST,PUT,DELETE,OPTIONS",
				"TRUSTED_PROXIES":      "127.0.0.1,::1",
				"RATE_LIMIT_ENABLED":   "true",
				"RATE_LIMIT_RATE":      "100",
				"RATE_LIMIT_BURST":     "5",
				"RATE_LIMIT_WINDOW":    "1m",
			},
			wantError: false,
		},
		{
			name: "invalid environment",
			envVars: map[string]string{
				"APP_NAME": "goforms",
				"APP_ENV":  "invalid",
			},
			wantError: true,
		},
		{
			name: "missing required database config",
			envVars: map[string]string{
				"APP_NAME": "goforms",
				"APP_ENV":  "development",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear existing env vars first
			for k := range originalEnv {
				os.Unsetenv(k)
			}

			// Set environment variables for test
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := New()
			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)

			// Verify configuration values
			if tt.name == "valid configuration" {
				assert.Equal(t, "goforms", cfg.App.Name)
				assert.Equal(t, "development", cfg.App.Env)
				assert.Equal(t, "testuser", cfg.Database.User)
				assert.Equal(t, "testpass", cfg.Database.Password)
				assert.Equal(t, "testdb", cfg.Database.Name)
				assert.Equal(t, "localhost", cfg.Database.Host)
				assert.Equal(t, 3306, cfg.Database.Port)
			}
		})
	}
}

func TestSecurityConfig(t *testing.T) {
	t.Run("default_security_settings", func(t *testing.T) {
		// Set required database config
		os.Setenv("DB_USER", "testuser")
		os.Setenv("DB_PASSWORD", "testpass")
		os.Setenv("DB_NAME", "testdb")

		// Clean up after test
		defer func() {
			os.Unsetenv("DB_USER")
			os.Unsetenv("DB_PASSWORD")
			os.Unsetenv("DB_NAME")
			os.Unsetenv("SECURITY_CORS_ALLOWED_METHODS")
			os.Unsetenv("SECURITY_TRUSTED_PROXIES")
		}()

		config, err := New()
		assert.NoError(t, err)

		assert.Equal(t, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, config.Security.CorsAllowedMethods)
		assert.Equal(t, []string{"127.0.0.1", "::1"}, config.Security.TrustedProxies)
	})

	t.Run("custom_security_settings", func(t *testing.T) {
		// Set required database config
		os.Setenv("DB_USER", "testuser")
		os.Setenv("DB_PASSWORD", "testpass")
		os.Setenv("DB_NAME", "testdb")

		// Set custom security values
		os.Setenv("SECURITY_CORS_ALLOWED_METHODS", "GET,POST")
		os.Setenv("SECURITY_TRUSTED_PROXIES", "10.0.0.1")

		// Clean up after test
		defer func() {
			os.Unsetenv("DB_USER")
			os.Unsetenv("DB_PASSWORD")
			os.Unsetenv("DB_NAME")
			os.Unsetenv("SECURITY_CORS_ALLOWED_METHODS")
			os.Unsetenv("SECURITY_TRUSTED_PROXIES")
		}()

		config, err := New()
		assert.NoError(t, err)

		assert.Equal(t, []string{"GET", "POST"}, config.Security.CorsAllowedMethods)
		assert.Equal(t, []string{"10.0.0.1"}, config.Security.TrustedProxies)
	})
}

func TestRateLimitConfig(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		check   func(*testing.T, *Config)
	}{
		{
			name: "default rate limit settings",
			envVars: map[string]string{
				// Required database config
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
			},
			check: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.RateLimit.Enabled)
				assert.Equal(t, 100, cfg.RateLimit.Rate)
				assert.Equal(t, 5, cfg.RateLimit.Burst)
				assert.Equal(t, time.Minute, cfg.RateLimit.TimeWindow)
				assert.True(t, cfg.RateLimit.PerIP)
				assert.Equal(t, []string{"/health", "/metrics"}, cfg.RateLimit.ExemptPaths)
			},
		},
		{
			name: "custom rate limit settings",
			envVars: map[string]string{
				// Required database config
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
				// Rate limit config
				"RATE_LIMIT_ENABLED":      "true",
				"RATE_LIMIT_PER_IP":       "true",
				"RATE_LIMIT_RATE":         "200",
				"RATE_LIMIT_BURST":        "10",
				"RATE_LIMIT_TIME_WINDOW":  "2m",
				"RATE_LIMIT_EXEMPT_PATHS": "/health,/status",
			},
			check: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.RateLimit.Enabled)
				assert.True(t, cfg.RateLimit.PerIP)
				assert.Equal(t, 200, cfg.RateLimit.Rate)
				assert.Equal(t, 10, cfg.RateLimit.Burst)
				assert.Equal(t, 2*time.Minute, cfg.RateLimit.TimeWindow)
				assert.Equal(t, []string{"/health", "/status"}, cfg.RateLimit.ExemptPaths)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Clean up after test
			defer os.Clearenv()

			cfg, err := New()
			require.NoError(t, err)
			require.NotNil(t, cfg)

			tt.check(t, cfg)
		})
	}
}
