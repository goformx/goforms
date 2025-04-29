package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
)

const (
	// minArgs is the minimum number of arguments required
	minArgs = 2
	// defaultSourceURL is the default location of migration files
	defaultSourceURL = "file://migrations"
)

func main() {
	var (
		sourceURL string
		command   string
	)

	flag.StringVar(&sourceURL, "source", defaultSourceURL, "Migration source URL")
	flag.StringVar(&command, "command", "up", "Migration command (up/down)")
	flag.Parse()

	if len(os.Args) < minArgs {
		log.Fatal("Please provide a migration command (up/down)")
	}

	if err := performMigration(sourceURL, command); err != nil {
		log.Fatal(err)
	}
}

func performMigration(sourceURL, command string) error {
	// Load configuration
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create database connection
	db, err := database.NewDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure MariaDB driver
	driver, err := mysql.WithInstance(db.DB, &mysql.Config{
		MigrationsTable: "schema_migrations",
		DatabaseName:    cfg.Database.Name,
	})
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to configure MySQL driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(sourceURL, "mysql", driver)
	if err != nil {
		db.Close()
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer func() {
		sourceErr, closeErr := m.Close()
		dbErr := db.Close()
		if sourceErr != nil || dbErr != nil || closeErr != nil {
			log.Printf("Error closing resources: migrate: %v, close: %v, db: %v", sourceErr, closeErr, dbErr)
		}
	}()

	// Check current version
	version, dirty, versionErr := m.Version()
	if versionErr != nil && !errors.Is(versionErr, migrate.ErrNilVersion) {
		return fmt.Errorf("failed to get migration version: %w", versionErr)
	}
	log.Printf("Current migration version: %d, dirty: %v", version, dirty)

	migrationErr := runMigration(m, command)
	if migrationErr != nil && !errors.Is(migrationErr, migrate.ErrNilVersion) {
		return fmt.Errorf("migration failed: %w", migrationErr)
	}

	// Check new version
	version, dirty, versionErr = m.Version()
	if versionErr != nil {
		return fmt.Errorf("failed to get final migration version: %w", versionErr)
	}
	log.Printf("New migration version: %d, dirty: %v", version, dirty)

	return nil
}

func runMigration(m *migrate.Migrate, cmd string) error {
	var migrationErr error

	switch cmd {
	case "up":
		migrationErr = m.Up()
	case "down":
		migrationErr = m.Down()
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}

	if migrationErr != nil {
		if errors.Is(migrationErr, migrate.ErrNoChange) {
			log.Printf("No migration needed")
			return nil
		}
		return fmt.Errorf("migration error: %w", migrationErr)
	}

	return nil
}
