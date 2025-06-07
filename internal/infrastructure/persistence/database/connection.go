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
