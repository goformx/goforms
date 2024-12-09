package server

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		want        *Config
		wantErr     bool
		errContains string
	}{
		{
			name: "default values",
			want: &Config{
				Host: "localhost",
				Port: 8090,
				Timeouts: TimeoutConfig{
					Read:  15 * time.Second,
					Write: 15 * time.Second,
					Idle:  60 * time.Second,
				},
			},
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"SERVER_HOST":          "0.0.0.0",
				"SERVER_PORT":          "9000",
				"SERVER_READ_TIMEOUT":  "30s",
				"SERVER_WRITE_TIMEOUT": "30s",
				"SERVER_IDLE_TIMEOUT":  "120s",
			},
			want: &Config{
				Host: "0.0.0.0",
				Port: 9000,
				Timeouts: TimeoutConfig{
					Read:  30 * time.Second,
					Write: 30 * time.Second,
					Idle:  120 * time.Second,
				},
			},
		},
		{
			name: "invalid port",
			envVars: map[string]string{
				"SERVER_PORT": "invalid",
			},
			wantErr:     true,
			errContains: "failed to process server config",
		},
		{
			name: "invalid timeout",
			envVars: map[string]string{
				"SERVER_READ_TIMEOUT": "invalid",
			},
			wantErr:     true,
			errContains: "failed to process server config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment before each test
			os.Clearenv()

			// Set environment variables for test
			for k, v := range tt.envVars {
				err := os.Setenv(k, v)
				require.NoError(t, err)
			}

			// Run test
			got, err := NewConfig()

			// Check error
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			// Check result
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
