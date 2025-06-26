package config

import (
	"os"
	"testing"
	"time"

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

	config, err := New()
	require.NoError(t, err)
	assert.NotNil(t, config)

	// Test default values
	assert.Equal(t, "GoFormX", config.App.Name)
	assert.Equal(t, "production", config.App.Env)
	assert.Equal(t, 8090, config.App.Port)
	assert.Equal(t, "mariadb", config.Database.Connection)
	assert.Equal(t, 3306, config.Database.Port)
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
			app := &AppConfig{Env: tt.env}
			assert.Equal(t, tt.want, app.IsDevelopment())
		})
	}
}

func TestAppConfig_GetServerURL(t *testing.T) {
	tests := []struct {
		name   string
		config AppConfig
		want   string
	}{
		{
			name: "URL field set",
			config: AppConfig{
				URL:    "https://example.com:8080",
				Scheme: "http",
				Host:   "localhost",
				Port:   3000,
			},
			want: "https://example.com:8080",
		},
		{
			name: "URL field not set",
			config: AppConfig{
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
			assert.Equal(t, tt.want, tt.config.GetServerURL())
		})
	}
}

func TestAppConfig_GetServerPort(t *testing.T) {
	tests := []struct {
		name   string
		config AppConfig
		want   int
	}{
		{
			name: "use port field",
			config: AppConfig{
				URL:  "https://example.com:8080",
				Port: 3000,
			},
			want: 3000,
		},
		{
			name: "port field only",
			config: AppConfig{
				Port: 3000,
			},
			want: 3000,
		},
		{
			name: "URL without port, use port field",
			config: AppConfig{
				URL:  "https://example.com",
				Port: 3000,
			},
			want: 3000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.config.GetServerPort())
		})
	}
}

func TestSecurityConfig_GetCSPDirectives(t *testing.T) {
	tests := []struct {
		name      string
		security  SecurityConfig
		appConfig AppConfig
		want      string
	}{
		{
			name: "development environment",
			security: SecurityConfig{
				CSP: CSPConfig{Enabled: true},
			},
			appConfig: AppConfig{Env: "development"},
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
			security: SecurityConfig{
				CSP: CSPConfig{Enabled: true},
			},
			appConfig: AppConfig{Env: "production"},
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
			security: SecurityConfig{
				CSP: CSPConfig{
					Enabled:    true,
					Directives: "default-src 'none'; script-src 'self'",
				},
			},
			appConfig: AppConfig{Env: "production"},
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

func TestValidateAppConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				App: AppConfig{
					Name:         "TestApp",
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 10 * time.Second,
					IdleTimeout:  120 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "empty app name",
			config: Config{
				App: AppConfig{
					Name:         "",
					Port:         8080,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 10 * time.Second,
					IdleTimeout:  120 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "app name is required",
		},
		{
			name: "invalid port",
			config: Config{
				App: AppConfig{
					Name:         "TestApp",
					Port:         -1,
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 10 * time.Second,
					IdleTimeout:  120 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "app port must be between 1 and 65535",
		},
		{
			name: "invalid timeout",
			config: Config{
				App: AppConfig{
					Name:         "TestApp",
					Port:         8080,
					ReadTimeout:  0,
					WriteTimeout: 10 * time.Second,
					IdleTimeout:  120 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "read timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateAppConfig()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDatabaseConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid mariadb config",
			config: Config{
				Database: DatabaseConfig{
					Connection:   "mariadb",
					Host:         "localhost",
					Port:         3306,
					Database:     "testdb",
					Username:     "user",
					Password:     "pass",
					RootPassword: "rootpass",
				},
			},
			wantErr: false,
		},
		{
			name: "valid postgres config",
			config: Config{
				Database: DatabaseConfig{
					Connection: "postgres",
					Host:       "localhost",
					Port:       5432,
					Database:   "testdb",
					Username:   "user",
					Password:   "pass",
					SSLMode:    "disable",
				},
			},
			wantErr: false,
		},
		{
			name: "missing required fields",
			config: Config{
				Database: DatabaseConfig{
					Connection: "mariadb",
				},
			},
			wantErr: true,
			errMsg:  "database host is required",
		},
		{
			name: "unsupported database type",
			config: Config{
				Database: DatabaseConfig{
					Connection:   "oracle",
					Host:         "localhost",
					Port:         1521,
					Database:     "testdb",
					Username:     "user",
					Password:     "pass",
					RootPassword: "rootpass",
				},
			},
			wantErr: true,
			errMsg:  "unsupported database connection type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateDatabaseConfig()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSecurityConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid security config",
			config: Config{
				Security: SecurityConfig{
					CSRF: CSRFConfig{
						Enabled: true,
						Secret:  "test_secret_32_characters_long!",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "CSRF enabled without secret",
			config: Config{
				Security: SecurityConfig{
					CSRF: CSRFConfig{
						Enabled: true,
						Secret:  "",
					},
				},
			},
			wantErr: true,
			errMsg:  "CSRF secret is required when CSRF is enabled",
		},
		{
			name: "CSRF disabled",
			config: Config{
				Security: SecurityConfig{
					CSRF: CSRFConfig{
						Enabled: false,
						Secret:  "",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateSecurityConfig()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
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

	_, err = tmpfile.Write([]byte(envContent))
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	config, err := LoadFromFile(tmpfile.Name())
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "TestApp", config.App.Name)
	assert.Equal(t, "test", config.App.Env)
}
