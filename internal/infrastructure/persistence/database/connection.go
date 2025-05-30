package database

import (
	"context"
	"fmt"
	"strconv"

	_ "github.com/go-sql-driver/mysql" // MySQL driver for database/sql
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq" // PostgreSQL driver for database/sql
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// DB wraps sqlx.DB with lifecycle management
type DB struct {
	*sqlx.DB
	logger logging.Logger
	config *config.DatabaseConfig
}

// NewDB creates a new database connection with proper configuration
func NewDB(lc fx.Lifecycle, cfg *config.Config, logger logging.Logger) (*DB, error) {
	logger.Debug("initializing database connection",
		logging.StringField("host", cfg.Database.Host),
		logging.StringField("port", strconv.Itoa(cfg.Database.Port)),
		logging.StringField("name", cfg.Database.Name),
		logging.StringField("user", cfg.Database.User),
	)

	// Construct DSN
	dsn := buildDSN(&cfg.Database)

	logger.Debug("connecting to database")

	// Open connection
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		logger.Error("failed to connect to database",
			logging.ErrorField("error", err),
			logging.StringField("host", cfg.Database.Host),
			logging.StringField("port", strconv.Itoa(cfg.Database.Port)),
			logging.StringField("user", cfg.Database.User),
			logging.StringField("database", cfg.Database.Name),
		)
		return nil, fmt.Errorf("failed to connect to database %s@%s:%d/%s: %w",
			cfg.Database.User,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Name,
			err,
		)
	}

	logger.Debug("setting database connection parameters",
		logging.IntField("max_open_conns", cfg.Database.MaxOpenConns),
		logging.IntField("max_idle_conns", cfg.Database.MaxIdleConns),
	)

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxLifetime) // Using same value for idle time

	// Verify connection
	logger.Debug("pinging database to verify connection")
	pingErr := db.Ping()
	if pingErr != nil {
		logger.Error("failed to ping database",
			logging.ErrorField("error", pingErr),
		)
		return nil, fmt.Errorf("failed to ping database: %w", pingErr)
	}

	wrappedDB := &DB{
		DB:     db,
		logger: logger,
		config: &cfg.Database,
	}

	// Register lifecycle hooks
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logger.Debug("verifying database connection on startup")
			return db.Ping()
		},
		OnStop: func(context.Context) error {
			logger.Debug("closing database connection")
			return db.Close()
		},
	})

	logger.Info("successfully connected to database",
		logging.StringField("host", cfg.Database.Host),
		logging.StringField("port", strconv.Itoa(cfg.Database.Port)),
		logging.StringField("name", cfg.Database.Name),
	)

	return wrappedDB, nil
}

// buildDSN constructs the database connection string
func buildDSN(dbConfig *config.DatabaseConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name)
}

// PingContext implements the PingContexter interface for Echo
func (db *DB) PingContext(c echo.Context) error {
	return db.Ping()
}

// WithTx executes a function within a transaction
func (db *DB) WithTx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	db.logger.Debug("beginning database transaction")

	tx, err := db.beginTransaction(ctx)
	if err != nil {
		return err
	}

	// Handle panics and errors
	return db.executeInTransaction(tx, fn)
}

// beginTransaction starts a new database transaction
func (db *DB) beginTransaction(ctx context.Context) (*sqlx.Tx, error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		db.logger.Error("failed to begin transaction", logging.ErrorField("error", err))
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

// executeInTransaction executes the function within the transaction
func (db *DB) executeInTransaction(tx *sqlx.Tx, fn func(*sqlx.Tx) error) error {
	// Handle panics
	defer func() {
		if p := recover(); p != nil {
			db.handlePanic(tx, p)
		}
	}()

	// Execute the function
	if err := fn(tx); err != nil {
		return db.handleTransactionError(tx, err)
	}

	// Commit the transaction
	return db.commitTransaction(tx)
}

// handlePanic handles transaction panics
func (db *DB) handlePanic(tx *sqlx.Tx, p any) {
	db.logger.Error("rolling back transaction due to panic",
		logging.AnyField("panic", p),
	)
	if rbErr := tx.Rollback(); rbErr != nil {
		db.logger.Error("failed to rollback transaction after panic",
			logging.ErrorField("error", rbErr),
		)
	}
	panic(p) // re-throw panic after rollback
}

// handleTransactionError handles transaction errors
func (db *DB) handleTransactionError(tx *sqlx.Tx, err error) error {
	db.logger.Error("rolling back transaction due to error",
		logging.ErrorField("error", err),
	)
	if rbErr := tx.Rollback(); rbErr != nil {
		db.logger.Error("failed to rollback transaction",
			logging.ErrorField("error", rbErr),
		)
		return fmt.Errorf("rollback failed: %w (original error: %w)", rbErr, err)
	}
	return err
}

// commitTransaction commits the transaction
func (db *DB) commitTransaction(tx *sqlx.Tx) error {
	db.logger.Debug("committing transaction")
	if err := tx.Commit(); err != nil {
		db.logger.Error("failed to commit transaction",
			logging.ErrorField("error", err),
		)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	db.logger.Debug("transaction completed successfully")
	return nil
}
