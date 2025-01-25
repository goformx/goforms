// Package database provides database connection and management
package database

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Database wraps the SQL connection pool
type Database struct {
	*sqlx.DB
	logger logging.Logger
}

// NewDB creates a new database connection
func NewDB(cfg *config.Config, logger logging.Logger) (*Database, error) {
	logger.Debug("building database connection string",
		logging.String("host", cfg.Database.Host),
		logging.Int("port", cfg.Database.Port),
		logging.String("name", cfg.Database.Name),
		logging.String("user", cfg.Database.User),
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
			logging.Error(err),
			logging.String("host", cfg.Database.Host),
			logging.Int("port", cfg.Database.Port),
		)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Debug("setting database connection parameters",
		logging.Int("max_open_conns", cfg.Database.MaxOpenConns),
		logging.Int("max_idle_conns", cfg.Database.MaxIdleConns),
	)

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetme)

	// Verify connection
	logger.Debug("pinging database to verify connection")
	if err := db.Ping(); err != nil {
		logger.Error("failed to ping database", logging.Error(err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("successfully connected to database",
		logging.String("host", cfg.Database.Host),
		logging.Int("port", cfg.Database.Port),
		logging.String("name", cfg.Database.Name),
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
		db.logger.Error("failed to close database connection", logging.Error(err))
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	db.logger.Debug("database connection closed successfully")
	return nil
}

// Begin starts a new transaction with detailed logging
func (db *Database) Begin() (*sqlx.Tx, error) {
	db.logger.Debug("beginning database transaction")
	tx, err := db.DB.Beginx()
	if err != nil {
		db.logger.Error("failed to begin transaction", logging.Error(err))
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	db.logger.Debug("transaction started successfully")
	return tx, nil
}
