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
}

// NewDB creates a new database connection
func NewDB(cfg *config.Config, logger logging.Logger) (*Database, error) {
	logger.Debug("building database connection string",
		logging.StringField("host", cfg.Database.Host),
		logging.IntField("port", cfg.Database.Port),
		logging.StringField("name", cfg.Database.Name),
		logging.StringField("user", cfg.Database.User),
	)

	// Construct DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	logger.Debug("connecting to database")

	// Open connection
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		logger.Error("failed to connect to database",
			logging.ErrorField("error", err),
			logging.StringField("host", cfg.Database.Host),
			logging.IntField("port", cfg.Database.Port),
		)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Debug("setting database connection parameters",
		logging.IntField("max_open_conns", cfg.Database.MaxOpenConns),
		logging.IntField("max_idle_conns", cfg.Database.MaxIdleConns),
	)

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

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
		logging.StringField("host", cfg.Database.Host),
		logging.IntField("port", cfg.Database.Port),
		logging.StringField("name", cfg.Database.Name),
	)

	return &Database{
		DB:     db,
		logger: logger,
	}, nil
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
