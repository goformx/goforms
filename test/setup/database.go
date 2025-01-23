package setup

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"

	// Import the file source driver
	"testing"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// TestDB manages test database setup and teardown
type TestDB struct {
	DB *sqlx.DB
}

// NewTestDB creates a new test database connection
func NewTestDB() (*TestDB, error) {
	// Use environment variables with defaults
	dbUser := os.Getenv("TEST_DB_USER")
	if dbUser == "" {
		dbUser = "goforms_test"
	}
	dbPass := os.Getenv("TEST_DB_PASSWORD")
	if dbPass == "" {
		dbPass = "goforms_test"
	}
	dbName := os.Getenv("TEST_DB_NAME")
	if dbName == "" {
		dbName = "goforms_test"
	}
	dbHost := os.Getenv("TEST_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost" // Default to Docker service name
	}
	dbPort := os.Getenv("TEST_DB_PORT")
	if dbPort == "" {
		dbPort = "3307" // Use test Docker port
	}

	// Build connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&allowNativePasswords=true",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// Debug connection info
	log.Printf("Attempting to connect to database with DSN: %s:%s@tcp(%s:%s)/%s",
		dbUser, "[REDACTED]", dbHost, dbPort, dbName)

	// Retry connection up to 5 times with exponential backoff
	var db *sqlx.DB
	var err error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = sqlx.Connect("mysql", dsn)
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
			log.Printf("Failed to connect, retrying in %v... (attempt %d/%d)", waitTime, i+1, maxRetries)
			time.Sleep(waitTime)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database after %d attempts: %w", maxRetries, err)
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
	// Add all your tables here
	tables := []string{
		"subscriptions",
		"contacts",
		// Add other tables as needed
	}

	for _, table := range tables {
		_, err := tdb.DB.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			return fmt.Errorf("failed to clear data from %s: %w", table, err)
		}
	}
	return nil
}

// Cleanup closes the database connection without dropping the database
func (tdb *TestDB) Cleanup() error {
	if tdb.DB != nil {
		if err := tdb.ClearData(); err != nil {
			return fmt.Errorf("failed to clear test data: %w", err)
		}
		return tdb.DB.Close()
	}
	return nil
}

// TestNewTestDBRetries tests the retry mechanism for database connections
func TestNewTestDBRetries(t *testing.T) {
	// Save original environment variables
	origUser := os.Getenv("TEST_DB_USER")
	origPass := os.Getenv("TEST_DB_PASSWORD")
	origName := os.Getenv("TEST_DB_NAME")
	origHost := os.Getenv("TEST_DB_HOST")
	origPort := os.Getenv("TEST_DB_PORT")

	// Set invalid connection details to force retries
	os.Setenv("TEST_DB_HOST", "nonexistent-host")
	os.Setenv("TEST_DB_PORT", "1234")

	start := time.Now()
	db, err := NewTestDB()
	duration := time.Since(start)

	// Verify that it took some time due to retries
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.True(t, duration >= 30*time.Second) // 5 retries with exponential backoff

	// Restore original environment variables
	os.Setenv("TEST_DB_USER", origUser)
	os.Setenv("TEST_DB_PASSWORD", origPass)
	os.Setenv("TEST_DB_NAME", origName)
	os.Setenv("TEST_DB_HOST", origHost)
	os.Setenv("TEST_DB_PORT", origPort)
}
