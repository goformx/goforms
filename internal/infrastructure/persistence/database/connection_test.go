package database_test

import (
	"testing"
	"time"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/persistence/database"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestNewDB(t *testing.T) {
	app := fxtest.New(t,
		fx.Provide(
			func() logging.Logger {
				return logging.NewLogger(true, "test")
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
			database.NewDB,
		),
	)

	require.NoError(t, app.Start(t.Context()))
	defer app.Stop(t.Context())

	// Get the database instance
	var db *database.DB
	require.NoError(t, app.Start(t.Context()))
	require.NoError(t, app.Stop(t.Context()))
	require.NotNil(t, db)
}
