package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validation(t *testing.T) {
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
				DBName:         "testdb",
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
				DBName:         "testdb",
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
				DBName:         "testdb",
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
				DBName:         "testdb",
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
				DBName:         "testdb",
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
				DBName:         "testdb",
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
	expectedUser := "testuser"
	expectedPass := "testpass"
	expectedDB := "testdb"

	// Use NewConfig to get defaults
	config := NewConfig()

	// Override with test values
	config.User = expectedUser
	config.Password = expectedPass
	config.DBName = expectedDB

	// Test all fields, including explicitly set ones
	assert.Equal(t, expectedUser, config.User)
	assert.Equal(t, expectedPass, config.Password)
	assert.Equal(t, expectedDB, config.DBName)

	// Test default values
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 3306, config.Port)
	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, config.ConnMaxLifetme)
}
