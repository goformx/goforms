package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"go.uber.org/fx"
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

func TestNewDB(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: &config.Config{
				Database: config.DatabaseConfig{
					Host:     "localhost",
					Port:     3306,
					User:     "test_user",
					Password: "test_pass",
					Name:     "test_db",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logging.NewTestLogger()
			var db *DB
			var err error

			app := fx.New(
				fx.NopLogger,
				fx.Supply(tt.config, logger),
				fx.Provide(
					func(lc fx.Lifecycle) (*DB, error) {
						return NewDB(lc, tt.config, logger)
					},
				),
				fx.Populate(&db),
			)
			err = app.Start(context.Background())
			defer app.Stop(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
			}
		})
	}
}
