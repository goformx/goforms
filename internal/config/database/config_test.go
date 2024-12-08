package database

import (
	"os"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validation(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Host:           "localhost",
				Port:           3306,
				User:           "testuser",
				Password:       "testpass",
				Name:           "testdb",
				MaxOpenConns:   25,
				MaxIdleConns:   5,
				ConnMaxLifetme: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: Config{
				Port:           3306,
				User:           "testuser",
				Password:       "testpass",
				Name:           "testdb",
				MaxOpenConns:   25,
				MaxIdleConns:   5,
				ConnMaxLifetme: 5 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "missing user",
			config: Config{
				Host:           "localhost",
				Port:           3306,
				Password:       "testpass",
				Name:           "testdb",
				MaxOpenConns:   25,
				MaxIdleConns:   5,
				ConnMaxLifetme: 5 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: Config{
				Host:           "localhost",
				Port:           0,
				User:           "testuser",
				Password:       "testpass",
				Name:           "testdb",
				MaxOpenConns:   25,
				MaxIdleConns:   5,
				ConnMaxLifetme: 5 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid max open conns",
			config: Config{
				Host:           "localhost",
				Port:           3306,
				User:           "testuser",
				Password:       "testpass",
				Name:           "testdb",
				MaxOpenConns:   0,
				MaxIdleConns:   5,
				ConnMaxLifetme: 5 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid connection lifetime",
			config: Config{
				Host:           "localhost",
				Port:           3306,
				User:           "testuser",
				Password:       "testpass",
				Name:           "testdb",
				MaxOpenConns:   25,
				MaxIdleConns:   5,
				ConnMaxLifetme: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_DefaultValues(t *testing.T) {
	// Clear any existing env vars first
	os.Clearenv()

	// Set required environment variables with correct prefix
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_MAX_OPEN_CONNS", "25")
	os.Setenv("DB_MAX_IDLE_CONNS", "5")
	os.Setenv("DB_CONN_MAX_LIFETIME", "5m")

	// Clean up after test
	defer func() {
		os.Clearenv()
	}()

	// Use NewConfig to get config with environment variables
	config, err := NewConfig()
	require.NoError(t, err, "Failed to create config")
	require.NotNil(t, config, "Config should not be nil")

	// Test all fields
	assert.Equal(t, "testuser", config.User)
	assert.Equal(t, "testpass", config.Password)
	assert.Equal(t, "testdb", config.Name)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 3306, config.Port)
	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, config.ConnMaxLifetme)
}
