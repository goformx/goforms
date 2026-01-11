package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/goformx/goforms/internal/infrastructure/config"
)

func TestDatabaseConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		dbConfig    config.DatabaseConfig
		expectError bool
	}{
		{
			name: "valid database config",
			dbConfig: config.DatabaseConfig{
				Driver:          "postgres",
				Host:            "localhost",
				Port:            5432,
				Name:            "testdb",
				Username:        "testuser",
				Password:        "testpass",
				SSLMode:         "disable",
				ConnMaxLifetime: 1,
				ConnMaxIdleTime: 1,
			},
			expectError: false,
		},
		{
			name: "empty driver",
			dbConfig: config.DatabaseConfig{
				Driver:   "",
				Host:     "localhost",
				Port:     5432,
				Name:     "testdb",
				Username: "testuser",
				Password: "testpass",
			},
			expectError: true,
		},
		{
			name: "empty host",
			dbConfig: config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "",
				Port:     5432,
				Name:     "testdb",
				Username: "testuser",
				Password: "testpass",
			},
			expectError: true,
		},
		{
			name: "invalid port",
			dbConfig: config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     0,
				Name:     "testdb",
				Username: "testuser",
				Password: "testpass",
			},
			expectError: true,
		},
		{
			name: "port too high",
			dbConfig: config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     70000,
				Name:     "testdb",
				Username: "testuser",
				Password: "testpass",
			},
			expectError: true,
		},
		{
			name: "empty database name",
			dbConfig: config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Name:     "",
				Username: "testuser",
				Password: "testpass",
			},
			expectError: true,
		},
		{
			name: "empty username",
			dbConfig: config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Name:     "testdb",
				Username: "",
				Password: "testpass",
			},
			expectError: true,
		},
		{
			name: "empty password",
			dbConfig: config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Name:     "testdb",
				Username: "testuser",
				Password: "",
			},
			expectError: true,
		},
		{
			name: "invalid max open connections",
			dbConfig: config.DatabaseConfig{
				Driver:       "postgres",
				Host:         "localhost",
				Port:         5432,
				Name:         "testdb",
				Username:     "testuser",
				Password:     "testpass",
				MaxOpenConns: -1,
			},
			expectError: true,
		},
		{
			name: "invalid max idle connections",
			dbConfig: config.DatabaseConfig{
				Driver:       "postgres",
				Host:         "localhost",
				Port:         5432,
				Name:         "testdb",
				Username:     "testuser",
				Password:     "testpass",
				MaxIdleConns: -1,
			},
			expectError: true,
		},
		{
			name: "invalid connection max lifetime",
			dbConfig: config.DatabaseConfig{
				Driver:          "postgres",
				Host:            "localhost",
				Port:            5432,
				Name:            "testdb",
				Username:        "testuser",
				Password:        "testpass",
				ConnMaxLifetime: -1 * time.Second,
			},
			expectError: true,
		},
		{
			name: "invalid connection max idle time",
			dbConfig: config.DatabaseConfig{
				Driver:          "postgres",
				Host:            "localhost",
				Port:            5432,
				Name:            "testdb",
				Username:        "testuser",
				Password:        "testpass",
				ConnMaxIdleTime: -1 * time.Second,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dbConfig.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDatabaseConfig_ValidateDriverSpecific(t *testing.T) {
	tests := []struct {
		name        string
		dbConfig    config.DatabaseConfig
		expectError bool
	}{
		{
			name: "valid postgres config",
			dbConfig: config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Name:     "testdb",
				Username: "testuser",
				Password: "testpass",
				SSLMode:  "disable",
			},
			expectError: false,
		},
		{
			name: "postgres without SSL mode",
			dbConfig: config.DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Name:     "testdb",
				Username: "testuser",
				Password: "testpass",
				SSLMode:  "",
			},
			expectError: true,
		},
		{
			name: "valid mariadb config",
			dbConfig: config.DatabaseConfig{
				Driver:       "mariadb",
				Host:         "localhost",
				Port:         5432,
				Name:         "testdb",
				Username:     "testuser",
				Password:     "testpass",
				RootPassword: "rootpass",
			},
			expectError: false,
		},
		{
			name: "mariadb without root password",
			dbConfig: config.DatabaseConfig{
				Driver:       "mariadb",
				Host:         "localhost",
				Port:         5432,
				Name:         "testdb",
				Username:     "testuser",
				Password:     "testpass",
				RootPassword: "",
			},
			expectError: true,
		},
		{
			name: "unsupported driver",
			dbConfig: config.DatabaseConfig{
				Driver:   "sqlite",
				Host:     "localhost",
				Port:     5432,
				Name:     "testdb",
				Username: "testuser",
				Password: "testpass",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dbConfig.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
