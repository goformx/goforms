package setup

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/database"
)

// TestDB manages test database setup and teardown
type TestDB struct {
	DB *sqlx.DB
}

// NewTestDB creates a new test database connection
func NewTestDB() (*TestDB, error) {
	// Set test environment variables if not already set
	setTestEnvDefaults()

	// Connect to MySQL as root to set up the test database and permissions
	rootDB, err := sqlx.Connect("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%s)/",
			"root", // Use root user for initial setup
			os.Getenv("MYSQL_ROOT_PASSWORD"),
			os.Getenv("MYSQL_HOSTNAME"),
			os.Getenv("MYSQL_PORT"),
		))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL as root: %w", err)
	}
	defer rootDB.Close()

	// Create test database and set up permissions
	if err := setupTestDatabase(rootDB); err != nil {
		return nil, fmt.Errorf("failed to setup test database: %w", err)
	}

	// Now create the real connection using the application user
	cfg, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := database.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
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

	// Drop existing tables
	if err := tdb.dropTables(); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	// Create migrations table
	if err := tdb.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Cleanup removes test data and optionally drops the database
func (tdb *TestDB) Cleanup(dropDB bool) error {
	if tdb.DB != nil {
		if dropDB {
			// Connect as root to drop the database
			rootDB, err := sqlx.Connect("mysql",
				fmt.Sprintf("%s:%s@tcp(%s:%s)/",
					"root",
					os.Getenv("MYSQL_ROOT_PASSWORD"),
					os.Getenv("MYSQL_HOSTNAME"),
					os.Getenv("MYSQL_PORT"),
				))
			if err != nil {
				return fmt.Errorf("failed to connect to MySQL as root for cleanup: %w", err)
			}
			defer rootDB.Close()

			if _, err := rootDB.Exec("DROP DATABASE IF EXISTS goforms_test"); err != nil {
				return fmt.Errorf("failed to drop test database: %w", err)
			}
		}
		return tdb.DB.Close()
	}
	return nil
}

// ClearData removes all data from tables while preserving structure
func (tdb *TestDB) ClearData() error {
	_, err := tdb.DB.Exec("DELETE FROM subscriptions")
	if err != nil {
		return fmt.Errorf("failed to clear subscription data: %w", err)
	}
	return nil
}

// Helper functions
func setTestEnvDefaults() {
	envDefaults := map[string]string{
		"MYSQL_HOSTNAME":          "localhost",
		"MYSQL_PORT":              "3306",
		"MYSQL_ROOT_PASSWORD":     "rootpassword", // Added root password
		"MYSQL_USER":              "goforms",
		"MYSQL_PASSWORD":          "goforms",
		"MYSQL_DATABASE":          "goforms_test",
		"MYSQL_MAX_OPEN_CONNS":    "25",
		"MYSQL_MAX_IDLE_CONNS":    "5",
		"MYSQL_CONN_MAX_LIFETIME": "5m",
	}

	for key, value := range envDefaults {
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

func setupTestDatabase(db *sqlx.DB) error {
	// Drop database if it exists
	if _, err := db.Exec("DROP DATABASE IF EXISTS goforms_test"); err != nil {
		return fmt.Errorf("failed to drop existing database: %w", err)
	}

	// Create test database
	if _, err := db.Exec("CREATE DATABASE goforms_test"); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// Create user if not exists and grant privileges
	if _, err := db.Exec(fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED BY '%s'",
		os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"))); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Grant privileges
	if _, err := db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON goforms_test.* TO '%s'@'%%'",
		os.Getenv("MYSQL_USER"))); err != nil {
		return fmt.Errorf("failed to grant privileges: %w", err)
	}

	if _, err := db.Exec("FLUSH PRIVILEGES"); err != nil {
		return fmt.Errorf("failed to flush privileges: %w", err)
	}

	return nil
}

func (tdb *TestDB) dropTables() error {
	_, err := tdb.DB.Exec("DROP TABLE IF EXISTS schema_migrations, subscriptions")
	if err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}
	return nil
}

func (tdb *TestDB) createMigrationsTable() error {
	_, err := tdb.DB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version bigint NOT NULL,
			dirty boolean NOT NULL,
			PRIMARY KEY (version)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	return nil
}
