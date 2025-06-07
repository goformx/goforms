package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

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
	var db *sqlx.DB
	var err error

	// Select configuration based on driver
	switch cfg.Database.Driver {
	case "mariadb":
		logger.Debug("initializing MariaDB connection",
			logging.StringField("host", cfg.Database.MariaDB.Host),
			logging.StringField("port", strconv.Itoa(cfg.Database.MariaDB.Port)),
			logging.StringField("name", cfg.Database.MariaDB.Name),
			logging.StringField("user", cfg.Database.MariaDB.User),
		)
		db, err = connectMariaDB(&cfg.Database.MariaDB)
	case "postgres":
		logger.Debug("initializing PostgreSQL connection",
			logging.StringField("host", cfg.Database.Postgres.Host),
			logging.StringField("port", strconv.Itoa(cfg.Database.Postgres.Port)),
			logging.StringField("name", cfg.Database.Postgres.Name),
			logging.StringField("user", cfg.Database.Postgres.User),
		)
		db, err = connectPostgres(&cfg.Database.Postgres)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	if err != nil {
		logger.Error("failed to connect to database",
			logging.ErrorField("error", err),
			logging.StringField("driver", cfg.Database.Driver),
		)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool based on driver
	if cfg.Database.Driver == "mariadb" {
		db.SetMaxOpenConns(cfg.Database.MariaDB.MaxOpenConns)
		db.SetMaxIdleConns(cfg.Database.MariaDB.MaxIdleConns)
		db.SetConnMaxLifetime(cfg.Database.MariaDB.ConnMaxLifetime)
		db.SetConnMaxIdleTime(cfg.Database.MariaDB.ConnMaxLifetime)
	} else {
		db.SetMaxOpenConns(cfg.Database.Postgres.MaxOpenConns)
		db.SetMaxIdleConns(cfg.Database.Postgres.MaxIdleConns)
		db.SetConnMaxLifetime(cfg.Database.Postgres.ConnMaxLifetime)
		db.SetConnMaxIdleTime(cfg.Database.Postgres.ConnMaxLifetime)
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
		logging.StringField("driver", cfg.Database.Driver),
	)

	return wrappedDB, nil
}

// connectMariaDB establishes a connection to MariaDB
func connectMariaDB(dbConfig *struct {
	Host            string        `envconfig:"GOFORMS_MARIADB_HOST" default:"mariadb"`
	Port            int           `envconfig:"GOFORMS_MARIADB_PORT" default:"3306"`
	User            string        `envconfig:"GOFORMS_MARIADB_USER" default:"goforms"`
	Password        string        `envconfig:"GOFORMS_MARIADB_PASSWORD" default:"goforms"`
	Name            string        `envconfig:"GOFORMS_MARIADB_NAME" default:"goforms"`
	MaxOpenConns    int           `envconfig:"GOFORMS_MARIADB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int           `envconfig:"GOFORMS_MARIADB_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime time.Duration `envconfig:"GOFORMS_MARIADB_CONN_MAX_LIFETIME" default:"5m"`
}) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
	)

	return sqlx.Connect("mysql", dsn)
}

// connectPostgres establishes a connection to PostgreSQL
func connectPostgres(dbConfig *struct {
	Host            string        `envconfig:"GOFORMS_POSTGRES_HOST" default:"postgres"`
	Port            int           `envconfig:"GOFORMS_POSTGRES_PORT" default:"5432"`
	User            string        `envconfig:"GOFORMS_POSTGRES_USER" default:"goforms"`
	Password        string        `envconfig:"GOFORMS_POSTGRES_PASSWORD" default:"goforms"`
	Name            string        `envconfig:"GOFORMS_POSTGRES_DB" default:"goforms"`
	SSLMode         string        `envconfig:"GOFORMS_POSTGRES_SSLMODE" default:"disable"`
	MaxOpenConns    int           `envconfig:"GOFORMS_POSTGRES_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int           `envconfig:"GOFORMS_POSTGRES_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime time.Duration `envconfig:"GOFORMS_POSTGRES_CONN_MAX_LIFETIME" default:"5m"`
}) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Name,
		dbConfig.SSLMode,
	)

	return sqlx.Connect("postgres", dsn)
}

// PingContext implements the PingContexter interface for Echo
func (db *DB) PingContext(c echo.Context) error {
	return db.Ping()
}
