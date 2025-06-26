package config_test

import (
	"os"
	"testing"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Clear existing environment variables that might interfere
	envVarsToClear := []string{
		"GOFORMS_DB_CONNECTION",
		"GOFORMS_DB_PORT",
	}
	for _, key := range envVarsToClear {
		os.Unsetenv(key)
	}

	// Set minimal required environment variables
	envVars := map[string]string{
		"GOFORMS_DB_HOST":              "localhost",
		"GOFORMS_DB_DATABASE":          "test_db",
		"GOFORMS_DB_USERNAME":          "test_user",
		"GOFORMS_DB_PASSWORD":          "test_pass",
		"GOFORMS_DB_ROOT_PASSWORD":     "root_pass",
		"GOFORMS_SECURITY_CSRF_SECRET": "test_csrf_secret_32_characters_long",
		"GOFORMS_ADMIN_EMAIL":          "admin@test.com",
		"GOFORMS_ADMIN_PASSWORD":       "admin_pass",
		"GOFORMS_ADMIN_FIRST_NAME":     "Admin",
		"GOFORMS_ADMIN_LAST_NAME":      "User",
		"GOFORMS_USER_EMAIL":           "user@test.com",
		"GOFORMS_USER_PASSWORD":        "user_pass",
		"GOFORMS_USER_FIRST_NAME":      "Test",
		"GOFORMS_USER_LAST_NAME":       "User",
	}

	// Set environment variables
	for key, value := range envVars {
		t.Setenv(key, value)
	}

	cfg, cfgErr := config.New()
	require.NoError(t, cfgErr)
	assert.NotNil(t, cfg)

	// Test default values
	assert.Equal(t, "GoFormX", cfg.App.Name)
	assert.Equal(t, "production", cfg.App.Env)
	assert.Equal(t, 8090, cfg.App.Port)
	assert.Equal(t, "mariadb", cfg.Database.Connection)
	assert.Equal(t, 3306, cfg.Database.Port)
}

func TestAppConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"development", "development", true},
		{"Development", "Development", true},
		{"DEVELOPMENT", "DEVELOPMENT", true},
		{"production", "production", false},
		{"staging", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &config.AppConfig{Env: tt.env}
			assert.Equal(t, tt.want, app.IsDevelopment())
		})
	}
}

func TestAppConfig_GetServerURL(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.AppConfig
		want string
	}{
		{
			name: "URL field set",
			cfg: config.AppConfig{
				URL:    "https://example.com:8080",
				Scheme: "http",
				Host:   "localhost",
				Port:   3000,
			},
			want: "https://example.com:8080",
		},
		{
			name: "URL field not set",
			cfg: config.AppConfig{
				URL:    "",
				Scheme: "https",
				Host:   "example.com",
				Port:   8443,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cfg.GetServerURL())
		})
	}
}

func TestAppConfig_GetServerPort(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.AppConfig
		want int
	}{
		{
			name: "use port field",
			cfg: config.AppConfig{
				URL:  "https://example.com:8080",
				Port: 3000,
			},
			want: 3000,
		},
		{
			name: "port field only",
			cfg: config.AppConfig{
				Port: 3000,
			},
			want: 3000,
		},
		{
			name: "URL without port, use port field",
			cfg: config.AppConfig{
				URL:  "https://example.com",
				Port: 3000,
			},
			want: 3000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cfg.GetServerPort())
		})
	}
}

func TestSecurityConfig_GetCSPDirectives(t *testing.T) {
	tests := []struct {
		name      string
		security  config.SecurityConfig
		appConfig config.AppConfig
		want      string
	}{
		{
			name: "development environment",
			security: config.SecurityConfig{
				CSP: config.CSPConfig{Enabled: true},
			},
			appConfig: config.AppConfig{Env: "development"},
			want: "default-src 'self'; " +
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' http://localhost:5173 https://cdn.form.io blob:; " +
				"worker-src 'self' blob:; " +
				"style-src 'self' 'unsafe-inline' http://localhost:5173 https://cdn.form.io; " +
				"img-src 'self' data:; " +
				"font-src 'self' http://localhost:5173; " +
				"connect-src 'self' http://localhost:5173 ws://localhost:5173; " +
				"frame-ancestors 'none'; " +
				"base-uri 'self'; " +
				"form-action 'self'",
		},
		{
			name: "production environment",
			security: config.SecurityConfig{
				CSP: config.CSPConfig{Enabled: true},
			},
			appConfig: config.AppConfig{Env: "production"},
			want: "default-src 'self'; " +
				"script-src 'self' 'unsafe-inline'; " +
				"style-src 'self' 'unsafe-inline'; " +
				"img-src 'self' data:; " +
				"font-src 'self'; " +
				"connect-src 'self'; " +
				"frame-ancestors 'none'; " +
				"base-uri 'self'; " +
				"form-action 'self'",
		},
		{
			name: "custom CSP directives",
			security: config.SecurityConfig{
				CSP: config.CSPConfig{
					Enabled:    true,
					Directives: "default-src 'none'; script-src 'self'",
				},
			},
			appConfig: config.AppConfig{Env: "production"},
			want:      "default-src 'none'; script-src 'self'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.security.GetCSPDirectives(&tt.appConfig)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestLoader(t *testing.T) {
	// Create a temporary .env file
	envContent := `GOFORMS_APP_NAME=TestApp
GOFORMS_APP_ENV=test
GOFORMS_DB_HOST=localhost
GOFORMS_DB_DATABASE=test_db
GOFORMS_DB_USERNAME=test_user
GOFORMS_DB_PASSWORD=test_pass
GOFORMS_DB_ROOT_PASSWORD=root_pass
GOFORMS_SECURITY_CSRF_SECRET=test_csrf_secret_32_characters_long
GOFORMS_ADMIN_EMAIL=admin@test.com
GOFORMS_ADMIN_PASSWORD=admin_pass
GOFORMS_ADMIN_FIRST_NAME=Admin
GOFORMS_ADMIN_LAST_NAME=User
GOFORMS_USER_EMAIL=user@test.com
GOFORMS_USER_PASSWORD=user_pass
GOFORMS_USER_FIRST_NAME=Test
GOFORMS_USER_LAST_NAME=User`

	tmpfile, err := os.CreateTemp("", "test.env")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(envContent)
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	cfg, err := config.LoadFromFile(tmpfile.Name())
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "TestApp", cfg.App.Name)
	assert.Equal(t, "test", cfg.App.Env)
}
