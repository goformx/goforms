package database_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/persistence/database"
)

func TestNewDB(t *testing.T) {
	ctx := t.Context()

	// Start MariaDB container
	req := testcontainers.ContainerRequest{
		Image:        "mariadb:10.6",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "test",
			"MYSQL_DATABASE":      "test_db",
		},
		WaitingFor: wait.ForLog("ready for connections"),
	}

	mariadbC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer mariadbC.Terminate(ctx)

	// Get container host and port
	host, err := mariadbC.Host(ctx)
	require.NoError(t, err)
	port, err := mariadbC.MappedPort(ctx, "3306")
	require.NoError(t, err)

	// Create test app
	app := fxtest.New(t,
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.NewLogger(true, "test")
			},
			func() *config.Config {
				return &config.Config{
					Database: config.DatabaseConfig{
						Host:           host,
						Port:           port.Int(),
						Name:           "test_db",
						User:           "root",
						Password:       "test",
						MaxOpenConns:   10,
						MaxIdleConns:   5,
						ConnMaxLifetme: time.Hour,
					},
				}
			},
			database.NewDB,
		),
		fx.Invoke(func(db *database.DB) {
			require.NotNil(t, db)
			require.NoError(t, db.Ping())
		}),
	)

	// Start the app
	require.NoError(t, app.Start(ctx))
	stopErr := app.Stop(ctx)
	if stopErr != nil {
		t.Fatalf("Failed to stop app: %v", stopErr)
	}
}

func TestNewConnection(t *testing.T) {
	ctx := t.Context()
	app := fx.New(
		fx.Provide(
			database.NewDB,
			func() *config.Config {
				return &config.Config{
					Database: config.DatabaseConfig{
						Host:     "localhost",
						Port:     5432,
						User:     "postgres",
						Password: "postgres",
						Name:     "testdb",
					},
				}
			},
		),
		fx.Invoke(func(db *database.DB) {
			if db == nil {
				t.Error("Expected database connection to be initialized")
			}
		}),
	)

	if err := app.Start(ctx); err != nil {
		t.Fatalf("Failed to start app: %v", err)
	}

	if err := app.Stop(ctx); err != nil {
		t.Fatalf("Failed to stop app: %v", err)
	}
}
