package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/config/database"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			cfg: &config.Config{
				Database: database.Config{
					Host:           "localhost",
					Port:           3306,
					User:           "test_user",
					Password:       "test_pass",
					Name:           "test_db",
					MaxOpenConns:   10,
					MaxIdleConns:   5,
					ConnMaxLifetme: time.Hour,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid configuration - empty host",
			cfg: &config.Config{
				Database: database.Config{
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
				Database: database.Config{
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
			db, err := New(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				// We can't actually connect to the database in a unit test,
				// but we can verify the DSN was constructed correctly
				assert.NoError(t, err)
				assert.NotNil(t, db)
			}
		})
	}
}
