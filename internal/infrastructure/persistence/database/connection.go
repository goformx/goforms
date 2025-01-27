package database

import (
	"context"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
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
		logging.String("host", cfg.Database.Host),
		logging.String("port", fmt.Sprintf("%d", cfg.Database.Port)),
		logging.String("name", cfg.Database.Name),
		logging.String("user", cfg.Database.User),
	)

	// Construct DSN
	dsn := buildDSN(&cfg.Database)

	logger.Debug("connecting to database")

	// Open connection
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		logger.Error("failed to connect to database",
			logging.Error(err),
			logging.String("host", cfg.Database.Host),
			logging.String("port", fmt.Sprintf("%d", cfg.Database.Port)),
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
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxLifetme) // Using same value for idle time

	// Verify connection
	logger.Debug("pinging database to verify connection")
	if err := db.Ping(); err != nil {
		logger.Error("failed to ping database", logging.Error(err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
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
		logging.String("host", cfg.Database.Host),
		logging.String("port", fmt.Sprintf("%d", cfg.Database.Port)),
		logging.String("name", cfg.Database.Name),
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

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		db.logger.Error("failed to begin transaction", logging.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			db.logger.Error("rolling back transaction due to panic",
				logging.Any("panic", p),
			)
			_ = tx.Rollback()
			panic(p) // re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		db.logger.Error("rolling back transaction due to error",
			logging.Error(err),
		)
		if rbErr := tx.Rollback(); rbErr != nil {
			db.logger.Error("failed to rollback transaction",
				logging.Error(rbErr),
			)
			return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	db.logger.Debug("committing transaction")
	if err := tx.Commit(); err != nil {
		db.logger.Error("failed to commit transaction",
			logging.Error(err),
		)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	db.logger.Debug("transaction completed successfully")
	return nil
}
