// Package database provides database connection and management
package database

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql" // MySQL driver for database/sql
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver for database/sql

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Database wraps the SQL connection pool
type Database struct {
	*sqlx.DB
	logger logging.Logger
	driver string
}

// NewDB creates a new database connection
func NewDB(cfg *config.Config, logger logging.Logger) (*Database, error) {
	var dsn string

	// Select configuration based on driver
	switch cfg.Database.Driver {
	case "postgres":
		logger.Debug("building PostgreSQL connection string",
			logging.StringField("host", cfg.Database.Postgres.Host),
			logging.IntField("port", cfg.Database.Postgres.Port),
			logging.StringField("name", cfg.Database.Postgres.Name),
			logging.StringField("user", cfg.Database.Postgres.User),
		)
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Database.Postgres.Host,
			cfg.Database.Postgres.Port,
			cfg.Database.Postgres.User,
			cfg.Database.Postgres.Password,
			cfg.Database.Postgres.Name,
			cfg.Database.Postgres.SSLMode,
		)
	case "mariadb", "mysql":
		logger.Debug("building MariaDB connection string",
			logging.StringField("host", cfg.Database.MariaDB.Host),
			logging.IntField("port", cfg.Database.MariaDB.Port),
			logging.StringField("name", cfg.Database.MariaDB.Name),
			logging.StringField("user", cfg.Database.MariaDB.User),
		)
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			cfg.Database.MariaDB.User,
			cfg.Database.MariaDB.Password,
			cfg.Database.MariaDB.Host,
			cfg.Database.MariaDB.Port,
			cfg.Database.MariaDB.Name,
		)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	logger.Debug("connecting to database")

	// Open connection
	driver := cfg.Database.Driver
	if driver == "mariadb" {
		driver = "mysql"
	}
	db, err := sqlx.Connect(driver, dsn)
	if err != nil {
		logger.Error("failed to connect to database",
			logging.ErrorField("error", err),
			logging.StringField("driver", driver),
		)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool based on selected database
	switch cfg.Database.Driver {
	case "postgres":
		db.SetMaxOpenConns(cfg.Database.Postgres.MaxOpenConns)
		db.SetMaxIdleConns(cfg.Database.Postgres.MaxIdleConns)
		db.SetConnMaxLifetime(cfg.Database.Postgres.ConnMaxLifetime)
	case "mariadb", "mysql":
		db.SetMaxOpenConns(cfg.Database.MariaDB.MaxOpenConns)
		db.SetMaxIdleConns(cfg.Database.MariaDB.MaxIdleConns)
		db.SetConnMaxLifetime(cfg.Database.MariaDB.ConnMaxLifetime)
	}

	// Verify connection
	logger.Debug("pinging database to verify connection")
	pingErr := db.Ping()
	if pingErr != nil {
		logger.Error("failed to ping database",
			logging.ErrorField("error", pingErr),
		)
		return nil, fmt.Errorf("failed to ping database: %w", pingErr)
	}

	logger.Info("successfully connected to database",
		logging.StringField("driver", cfg.Database.Driver),
	)

	return &Database{
		DB:     db,
		logger: logger,
		driver: cfg.Database.Driver,
	}, nil
}

// GetPlaceholder returns the appropriate placeholder for the current database driver
func (db *Database) GetPlaceholder(index int) string {
	if db.driver == "postgres" {
		return fmt.Sprintf("$%d", index)
	}
	return "?"
}

// Close closes the database connection
func (db *Database) Close() error {
	db.logger.Debug("closing database connection")
	if err := db.DB.Close(); err != nil {
		db.logger.Error("failed to close database connection", logging.ErrorField("error", err))
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	db.logger.Debug("database connection closed successfully")
	return nil
}

// Begin starts a new transaction with detailed logging
func (db *Database) Begin() (*sqlx.Tx, error) {
	db.logger.Debug("beginning database transaction")
	tx, err := db.Beginx()
	if err != nil {
		db.logger.Error("failed to begin transaction", logging.ErrorField("error", err))
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	db.logger.Debug("transaction started successfully")
	return tx, nil
}
