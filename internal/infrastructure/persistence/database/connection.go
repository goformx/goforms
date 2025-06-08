package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// DB wraps GORM DB with lifecycle management
type DB struct {
	*gorm.DB
	logger logging.Logger
}

var db *gorm.DB

// GetDB returns the singleton database instance
func GetDB() *gorm.DB {
	return db
}

// NewDB creates a new database connection
func NewDB(cfg *config.Config, appLogger logging.Logger) (*DB, error) {
	// Configure GORM logger
	gormLogger := logger.New(
		&GormLogWriter{logger: appLogger},
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Build DSN for PostgreSQL
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Postgres.Host,
		cfg.Database.Postgres.Port,
		cfg.Database.Postgres.User,
		cfg.Database.Postgres.Password,
		cfg.Database.Postgres.Name,
		cfg.Database.Postgres.SSLMode,
	)

	// Open connection
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		appLogger.Error("failed to connect to database",
			logging.ErrorField("error", err),
			logging.StringField("driver", "postgres"),
		)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.Postgres.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.Postgres.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.Postgres.ConnMaxLifetime)

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		appLogger.Error("failed to ping database",
			logging.ErrorField("error", err),
		)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	appLogger.Info("successfully connected to database",
		logging.StringField("driver", "postgres"),
	)

	db = gormDB

	return &DB{
		DB:     gormDB,
		logger: appLogger,
	}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	db.logger.Debug("closing database connection")
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		db.logger.Error("failed to close database connection", logging.ErrorField("error", err))
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	db.logger.Debug("database connection closed successfully")
	return nil
}

// GormLogWriter implements io.Writer for GORM logger
type GormLogWriter struct {
	logger logging.Logger
}

// Write implements io.Writer interface
func (w *GormLogWriter) Write(p []byte) (n int, err error) {
	w.logger.Debug("gorm query", logging.StringField("query", string(p)))
	return len(p), nil
}

// Printf implements logger.Writer interface
func (w *GormLogWriter) Printf(format string, args ...any) {
	w.logger.Debug("gorm query", logging.StringField("query", fmt.Sprintf(format, args...)))
}
