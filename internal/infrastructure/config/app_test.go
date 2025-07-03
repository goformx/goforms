package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goformx/goforms/internal/infrastructure/config"
)

func TestAppConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		appConfig   config.AppConfig
		expectError bool
	}{
		{
			name: "valid app config",
			appConfig: config.AppConfig{
				Name:           "Test App",
				Version:        "1.0.0",
				Environment:    "development",
				Port:           8080,
				Host:           "localhost",
				ReadTimeout:    5,
				WriteTimeout:   5,
				IdleTimeout:    5,
				RequestTimeout: 5,
			},
			expectError: false,
		},
		{
			name: "empty name",
			appConfig: config.AppConfig{
				Name:        "",
				Version:     "1.0.0",
				Environment: "development",
				Port:        8080,
			},
			expectError: true,
		},
		{
			name: "invalid port",
			appConfig: config.AppConfig{
				Name:        "Test App",
				Version:     "1.0.0",
				Environment: "development",
				Port:        0,
			},
			expectError: true,
		},
		{
			name: "invalid port too high",
			appConfig: config.AppConfig{
				Name:        "Test App",
				Version:     "1.0.0",
				Environment: "development",
				Port:        70000,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.appConfig.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAppConfig_GetServerURL(t *testing.T) {
	appConfig := config.AppConfig{
		URL: "http://localhost:8080",
	}

	result := appConfig.GetServerURL()
	assert.Equal(t, "http://localhost:8080", result)
}

func TestAppConfig_GetServerPort(t *testing.T) {
	appConfig := config.AppConfig{
		Port: 8080,
	}

	result := appConfig.GetServerPort()
	assert.Equal(t, 8080, result)
}

func TestAppConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name      string
		appConfig config.AppConfig
		expected  bool
	}{
		{
			name: "development environment",
			appConfig: config.AppConfig{
				Environment: "development",
			},
			expected: true,
		},
		{
			name: "production environment",
			appConfig: config.AppConfig{
				Environment: "production",
			},
			expected: false,
		},
		{
			name: "staging environment",
			appConfig: config.AppConfig{
				Environment: "staging",
			},
			expected: false,
		},
		{
			name: "test environment",
			appConfig: config.AppConfig{
				Environment: "test",
			},
			expected: false,
		},
		{
			name: "case insensitive development",
			appConfig: config.AppConfig{
				Environment: "DEVELOPMENT",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appConfig.IsDevelopment()
			assert.Equal(t, tt.expected, result)
		})
	}
}
