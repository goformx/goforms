package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
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
	app := fxtest.New(t,
		fx.Provide(
			func() logging.Logger {
				return logging.NewTestLogger()
			},
			func() *config.Config {
				return &config.Config{
					Database: config.DatabaseConfig{
						Host:           "localhost",
						Port:           3306,
						Name:           "test_db",
						User:           "test_user",
						Password:       "test_pass",
						MaxOpenConns:   10,
						MaxIdleConns:   5,
						ConnMaxLifetme: time.Hour,
					},
				}
			},
			NewDB,
		),
	)

	defer func() {
		if err := app.Stop(context.Background()); err != nil {
			t.Errorf("failed to stop app: %v", err)
		}
	}()

	if err := app.Start(context.Background()); err != nil {
		t.Errorf("failed to start app: %v", err)
	}
}
