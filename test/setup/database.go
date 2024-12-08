package setup

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

// TestDB manages test database setup and teardown
type TestDB struct {
	DB *sqlx.DB
}

// NewTestDB creates a new test database connection
func NewTestDB() (*TestDB, error) {
	// Only load .env.test for testing
	if err := godotenv.Load(".env.test"); err != nil {
		return nil, fmt.Errorf("failed to load test environment (.env.test): %w", err)
	}

	dbUser := getEnvOrDefault("MYSQL_USER", "goforms_test")
	dbPass := getEnvOrDefault("MYSQL_PASSWORD", "goforms_test")
	dbHost := getEnvOrDefault("MYSQL_HOSTNAME", "localhost")
	dbPort := getEnvOrDefault("MYSQL_PORT", "3306")
	dbName := getEnvOrDefault("MYSQL_DATABASE", "goforms_test")

	// Create DSN for test database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// Connect to database
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	return &TestDB{DB: db}, nil
}

// RunMigrations runs all database migrations
func (tdb *TestDB) RunMigrations() error {
	migrationPath := "file://../../migrations"

	driver, err := mysql.WithInstance(tdb.DB.DB, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationPath,
		"mysql",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// ClearData removes all data from tables while preserving structure
func (tdb *TestDB) ClearData() error {
	_, err := tdb.DB.Exec("DELETE FROM subscriptions")
	if err != nil {
		return fmt.Errorf("failed to clear test data: %w", err)
	}
	return nil
}

// Cleanup closes the database connection and optionally drops the database
func (tdb *TestDB) Cleanup(dropDB bool) error {
	if tdb.DB != nil {
		if dropDB {
			if _, err := tdb.DB.Exec("DROP DATABASE IF EXISTS goforms_test"); err != nil {
				return fmt.Errorf("failed to drop test database: %w", err)
			}
		}
		return tdb.DB.Close()
	}
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
