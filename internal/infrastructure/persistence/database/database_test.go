package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
)

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.DatabaseConfig
		expected string
	}{
		{
			name: "valid configuration",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     3306,
				User:     "test_user",
				Password: "test_pass",
				Name:     "test_db",
			},
			expected: "test_user:test_pass@tcp(localhost:3306)/test_db?parseTime=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := buildDSN(tt.config)
			assert.Equal(t, tt.expected, dsn)
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "invalid configuration - empty host",
			cfg: &config.Config{
				Database: config.DatabaseConfig{
					Port:           3306,
					User:           "test_user",
					Password:       "test_pass",
					Name:           "test_db",
					MaxOpenConns:   10,
					MaxIdleConns:   5,
					ConnMaxLifetme: time.Hour,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid configuration - empty user",
			cfg: &config.Config{
				Database: config.DatabaseConfig{
					Host:           "localhost",
					Port:           3306,
					Password:       "test_pass",
					Name:           "test_db",
					MaxOpenConns:   10,
					MaxIdleConns:   5,
					ConnMaxLifetme: time.Hour,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We only test invalid configurations since valid ones would try to connect
			if tt.wantErr {
				_, err := New(tt.cfg)
				assert.Error(t, err)
			}
		})
	}
}
