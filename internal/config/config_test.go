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
	tests := []struct {
		name    string
		envVars map[string]string
		check   func(*testing.T, *Config)
	}{
		{
			name: "default security settings",
			envVars: map[string]string{
				"APP_NAME":    "goforms",
				"APP_ENV":     "development",
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
			},
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, []string{"http://localhost:3000"}, cfg.Security.CorsAllowedOrigins)
				assert.Equal(t, []string{"GET,POST,PUT,DELETE,OPTIONS"}, cfg.Security.CorsAllowedMethods)
				assert.Equal(t, 3600, cfg.Security.CorsMaxAge)
				assert.Equal(t, []string{"127.0.0.1,::1"}, cfg.Security.TrustedProxies)
				assert.Equal(t, 30*time.Second, cfg.Security.RequestTimeout)
			},
		},
		{
			name: "custom security settings",
			envVars: map[string]string{
				"APP_NAME":             "goforms",
				"APP_ENV":              "development",
				"DB_USER":              "testuser",
				"DB_PASSWORD":          "testpass",
				"DB_NAME":              "testdb",
				"CORS_ALLOWED_ORIGINS": "https://example.com",
				"CORS_ALLOWED_METHODS": "GET,POST",
				"CORS_MAX_AGE":         "7200",
				"TRUSTED_PROXIES":      "10.0.0.1",
				"REQUEST_TIMEOUT":      "60s",
			},
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, []string{"https://example.com"}, cfg.Security.CorsAllowedOrigins)
				assert.Equal(t, []string{"GET,POST"}, cfg.Security.CorsAllowedMethods)
				assert.Equal(t, 7200, cfg.Security.CorsMaxAge)
				assert.Equal(t, []string{"10.0.0.1"}, cfg.Security.TrustedProxies)
				assert.Equal(t, 60*time.Second, cfg.Security.RequestTimeout)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := New()
			require.NoError(t, err)
			require.NotNil(t, cfg)

			tt.check(t, cfg)
		})
	}
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
				"APP_NAME":    "goforms",
				"APP_ENV":     "development",
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
				assert.Equal(t, []string{"/health,/metrics"}, cfg.RateLimit.ExemptPaths)
			},
		},
		{
			name: "custom rate limit settings",
			envVars: map[string]string{
				"APP_NAME":                "goforms",
				"APP_ENV":                 "development",
				"DB_USER":                 "testuser",
				"DB_PASSWORD":             "testpass",
				"DB_NAME":                 "testdb",
				"RATE_LIMIT_ENABLED":      "false",
				"RATE_LIMIT_RATE":         "50",
				"RATE_LIMIT_BURST":        "10",
				"RATE_LIMIT_WINDOW":       "2m",
				"RATE_LIMIT_PER_IP":       "false",
				"RATE_LIMIT_EXEMPT_PATHS": "/status,/ping",
			},
			check: func(t *testing.T, cfg *Config) {
				assert.False(t, cfg.RateLimit.Enabled)
				assert.Equal(t, 50, cfg.RateLimit.Rate)
				assert.Equal(t, 10, cfg.RateLimit.Burst)
				assert.Equal(t, 2*time.Minute, cfg.RateLimit.TimeWindow)
				assert.False(t, cfg.RateLimit.PerIP)
				assert.Equal(t, []string{"/status,/ping"}, cfg.RateLimit.ExemptPaths)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := New()
			require.NoError(t, err)
			require.NotNil(t, cfg)

			tt.check(t, cfg)
		})
	}
}
