package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

const (
	// minArgs is the minimum number of arguments required
	minArgs = 2
	// defaultSourceURL is the default location of migration files
	defaultSourceURL = "file://migrations"
)

func main() {
	sourceURL, command := parseFlags()
	if err := performMigration(sourceURL, command); err != nil {
		log.Fatal(err)
	}
}

func parseFlags() (sourceURL, command string) {
	flag.StringVar(&sourceURL, "source", defaultSourceURL, "Migration source URL")
	flag.StringVar(&command, "command", "up", "Migration command (up/down)")
	flag.Parse()

	if len(os.Args) < minArgs {
		log.Fatal("Please provide a migration command (up/down)")
	}
	return sourceURL, command
}

func setupLogger() (logging.Logger, error) {
	logFactory := logging.NewFactory()
	return logFactory.CreateLogger()
}

func setupDatabase(logger logging.Logger) (*database.Database, error) {
	cfg, err := config.New(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return database.NewDB(cfg, logger)
}

func createMigrator(db *database.Database, cfg *config.Config, sourceURL string) (*migrate.Migrate, error) {
	driver, err := mysql.WithInstance(db.DB.DB, &mysql.Config{
		MigrationsTable: "schema_migrations",
		DatabaseName:    cfg.Database.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to configure MySQL driver: %w", err)
	}

	return migrate.NewWithDatabaseInstance(sourceURL, "mysql", driver)
}

func handleDirtyState(m *migrate.Migrate, version uint) error {
	if version > uint(math.MaxInt) {
		return fmt.Errorf("version value %d overflows int", version)
	}
	if err := m.Force(int(version)); err != nil {
		return fmt.Errorf("failed to force version: %w", err)
	}
	log.Printf("Successfully forced version %d", version)
	return nil
}

func performMigration(sourceURL, command string) error {
	logger, err := setupLogger()
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	cfg, err := config.New(logger)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	db, err := setupDatabase(logger)
	if err != nil {
		return err
	}

	m, err := createMigrator(db, cfg, sourceURL)
	if err != nil {
		db.Close()
		return err
	}

	defer func() {
		sourceErr, closeErr := m.Close()
		dbErr := db.Close()
		if sourceErr != nil || dbErr != nil || closeErr != nil {
			log.Printf("Error closing resources: migrate: %v, close: %v, db: %v", sourceErr, closeErr, dbErr)
		}
	}()

	version, dirty, versionErr := m.Version()
	if versionErr != nil && !errors.Is(versionErr, migrate.ErrNilVersion) {
		return fmt.Errorf("failed to get migration version: %w", versionErr)
	}
	log.Printf("Current migration version: %d, dirty: %v", version, dirty)

	if dirty {
		log.Printf("Database is dirty at version %d, attempting to fix...", version)
		if dirtyErr := handleDirtyState(m, version); dirtyErr != nil {
			return dirtyErr
		}
	}

	if migrationErr := runMigration(m, command); migrationErr != nil {
		return migrationErr
	}

	return checkFinalVersion(m)
}

func checkFinalVersion(m *migrate.Migrate) error {
	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			log.Printf("New migration version: none (database is at base state), dirty: %v", dirty)
			return nil
		}
		return fmt.Errorf("failed to get final migration version: %w", err)
	}
	log.Printf("New migration version: %d, dirty: %v", version, dirty)
	return nil
}

func runMigration(m *migrate.Migrate, cmd string) error {
	switch cmd {
	case "up":
		return handleUpMigration(m)
	case "down":
		return handleDownMigration(m)
	case "down all":
		return handleDownAllMigration(m)
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func handleUpMigration(m *migrate.Migrate) error {
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Printf("No migration needed")
			return nil
		}
		return fmt.Errorf("migration error: %w", err)
	}
	return nil
}

func handleDownMigration(m *migrate.Migrate) error {
	if err := m.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Printf("No migration needed")
			return nil
		}
		return fmt.Errorf("migration error: %w", err)
	}
	return nil
}

func handleDownAllMigration(m *migrate.Migrate) error {
	for {
		if err := m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Printf("All migrations reverted")
				return nil
			}
			return fmt.Errorf("migration error: %w", err)
		}
	}
}
