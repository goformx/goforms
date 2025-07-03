package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goformx/goforms/internal/infrastructure/config"
)

// createValidConfig creates a valid test configuration
func createValidConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:           "Test App",
			Environment:    "development",
			Port:           8080,
			ReadTimeout:    5,
			WriteTimeout:   5,
			IdleTimeout:    5,
			RequestTimeout: 5,
		},
		Database: config.DatabaseConfig{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     5432,
			Name:     "testdb",
			Username: "testuser",
			Password: "testpass",
			SSLMode:  "disable",
		},
		Security: config.SecurityConfig{
			CSRF:            config.CSRFConfig{Enabled: false},
			CORS:            config.CORSConfig{Enabled: false},
			TLS:             config.TLSConfig{Enabled: false},
			CSP:             config.CSPConfig{Enabled: false},
			RateLimit:       config.RateLimitConfig{Enabled: false, RPS: 1, Burst: 1, Window: 1},
			SecurityHeaders: config.SecurityHeadersConfig{Enabled: false},
			CookieSecurity:  config.CookieSecurityConfig{Secure: true, HTTPOnly: true, SameSite: "Lax"},
		},
		Session: config.SessionConfig{
			Type:   "cookie",
			Secret: "this-is-a-very-long-session-secret-1234567890",
		},
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
	}{
		{
			name:        "valid config",
			config:      createValidConfig(),
			expectError: false,
		},
		{
			name: "invalid app config",
			config: &config.Config{
				App: config.AppConfig{
					Name: "", // Invalid: empty name
				},
				Database: config.DatabaseConfig{
					Driver:   "postgres",
					Host:     "localhost",
					Port:     5432,
					Name:     "testdb",
					Username: "testuser",
					Password: "testpass",
				},
				Security: config.SecurityConfig{
					CSRF: config.CSRFConfig{Enabled: false},
					CORS: config.CORSConfig{Enabled: false},
				},
			},
			expectError: true,
		},
		{
			name: "invalid database config",
			config: &config.Config{
				App: config.AppConfig{
					Name:        "Test App",
					Environment: "development",
					Port:        8080,
				},
				Database: config.DatabaseConfig{
					Driver: "", // Invalid: empty driver
				},
				Security: config.SecurityConfig{
					CSRF: config.CSRFConfig{Enabled: false},
					CORS: config.CORSConfig{Enabled: false},
				},
			},
			expectError: true,
		},
		{
			name: "invalid security config",
			config: &config.Config{
				App: config.AppConfig{
					Name:        "Test App",
					Environment: "development",
					Port:        8080,
				},
				Database: config.DatabaseConfig{
					Driver:   "postgres",
					Host:     "localhost",
					Port:     5432,
					Name:     "testdb",
					Username: "testuser",
					Password: "testpass",
				},
				Security: config.SecurityConfig{
					CSRF: config.CSRFConfig{
						Enabled: true,
						Secret:  "short", // Invalid: too short
					},
					CORS: config.CORSConfig{Enabled: false},
				},
			},
			expectError: true,
		},
		{
			name: "session config without secret",
			config: &config.Config{
				App: config.AppConfig{
					Name:        "Test App",
					Environment: "development",
					Port:        8080,
				},
				Database: config.DatabaseConfig{
					Driver:   "postgres",
					Host:     "localhost",
					Port:     5432,
					Name:     "testdb",
					Username: "testuser",
					Password: "testpass",
				},
				Security: config.SecurityConfig{
					CSRF: config.CSRFConfig{Enabled: false},
					CORS: config.CORSConfig{Enabled: false},
				},
				Session: config.SessionConfig{
					Type:   "cookie",
					Secret: "", // Invalid: empty secret for cookie session
				},
			},
			expectError: true,
		},
		{
			name: "email config with host but missing credentials",
			config: &config.Config{
				App: config.AppConfig{
					Name:        "Test App",
					Environment: "development",
					Port:        8080,
				},
				Database: config.DatabaseConfig{
					Driver:   "postgres",
					Host:     "localhost",
					Port:     5432,
					Name:     "testdb",
					Username: "testuser",
					Password: "testpass",
				},
				Security: config.SecurityConfig{
					CSRF: config.CSRFConfig{Enabled: false},
					CORS: config.CORSConfig{Enabled: false},
				},
				Email: config.EmailConfig{
					Host:     "smtp.example.com",
					Username: "", // Invalid: empty username when host is set
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.config.IsValid()

			if tt.expectError {
				assert.False(t, isValid)

				return
			}

			assert.True(t, isValid)
		})
	}
}

func TestConfig_IsValid(t *testing.T) {
	validConfig := createValidConfig()

	assert.True(t, validConfig.IsValid())

	invalidConfig := &config.Config{
		App: config.AppConfig{
			Name: "", // Invalid: empty name
		},
		Database: config.DatabaseConfig{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     5432,
			Name:     "testdb",
			Username: "testuser",
			Password: "testpass",
		},
		Security: config.SecurityConfig{
			CSRF: config.CSRFConfig{Enabled: false},
			CORS: config.CORSConfig{Enabled: false},
		},
	}

	assert.False(t, invalidConfig.IsValid())
}

func TestConfig_EnvironmentMethods(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		isProd      bool
		isDev       bool
		isStaging   bool
		isTest      bool
	}{
		{
			name:        "production",
			environment: "production",
			isProd:      true,
			isDev:       false,
			isStaging:   false,
			isTest:      false,
		},
		{
			name:        "development",
			environment: "development",
			isProd:      false,
			isDev:       true,
			isStaging:   false,
			isTest:      false,
		},
		{
			name:        "staging",
			environment: "staging",
			isProd:      false,
			isDev:       false,
			isStaging:   true,
			isTest:      false,
		},
		{
			name:        "test",
			environment: "test",
			isProd:      false,
			isDev:       false,
			isStaging:   false,
			isTest:      true,
		},
		{
			name:        "case insensitive",
			environment: "PRODUCTION",
			isProd:      true,
			isDev:       false,
			isStaging:   false,
			isTest:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				App: config.AppConfig{
					Environment: tt.environment,
				},
			}

			assert.Equal(t, tt.isProd, cfg.IsProduction())
			assert.Equal(t, tt.isDev, cfg.IsDevelopment())
			assert.Equal(t, tt.isStaging, cfg.IsStaging())
			assert.Equal(t, tt.isTest, cfg.IsTest())
		})
	}
}

func TestConfig_GetEnvironment(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Environment: "PRODUCTION",
		},
	}

	assert.Equal(t, "production", cfg.GetEnvironment())
}

func TestConfig_GetConfigSummary(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "Test App",
			Environment: "development",
			Debug:       true,
			URL:         "http://localhost:8080",
		},
		Database: config.DatabaseConfig{
			Driver: "postgres",
			Host:   "localhost",
			Port:   5432,
			Name:   "testdb",
		},
		Security: config.SecurityConfig{
			CSRF:      config.CSRFConfig{Enabled: true},
			CORS:      config.CORSConfig{Enabled: false},
			RateLimit: config.RateLimitConfig{Enabled: true},
			CSP:       config.CSPConfig{Enabled: true},
		},
		Email: config.EmailConfig{
			Host: "smtp.example.com",
		},
		Cache: config.CacheConfig{
			Type: "memory",
		},
		Storage: config.StorageConfig{
			Type: "local",
		},
		Session: config.SessionConfig{
			Type: "cookie",
		},
	}

	summary := cfg.GetConfigSummary()

	// Test app section
	app, ok := summary["app"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Test App", app["name"])
	assert.Equal(t, "development", app["environment"])
	assert.Equal(t, true, app["debug"])
	assert.Equal(t, "http://localhost:8080", app["url"])

	// Test database section
	db, ok := summary["database"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "postgres", db["driver"])
	assert.Equal(t, "localhost", db["host"])
	assert.Equal(t, 5432, db["port"])
	assert.Equal(t, "testdb", db["name"])

	// Test security section
	security, ok := summary["security"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, security["csrf_enabled"])
	assert.Equal(t, false, security["cors_enabled"])
	assert.Equal(t, true, security["rate_limit_enabled"])
	assert.Equal(t, true, security["csp_enabled"])

	// Test services section
	services, ok := summary["services"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, services["email_configured"])
	assert.Equal(t, "memory", services["cache_type"])
	assert.Equal(t, "local", services["storage_type"])
	assert.Equal(t, "cookie", services["session_type"])
}
