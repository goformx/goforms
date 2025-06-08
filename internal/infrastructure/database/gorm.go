package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

const (
	// DefaultPingTimeout is the default timeout for database ping operations
	DefaultPingTimeout = 5 * time.Second
)

// GormDB wraps the GORM database connection
type GormDB struct {
	*gorm.DB
	logger logging.Logger
}

// NewGormDB creates a new GORM database connection
func NewGormDB(cfg *config.Config, appLogger logging.Logger) (*GormDB, error) {
	// Configure GORM logger
	gormLogger := logger.New(
		&GormLogWriter{logger: appLogger},
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: false,
			Colorful:                  cfg.App.IsDevelopment(),
			ParameterizedQueries:      true,
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
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true, // Enable prepared statements for better performance
	})
	if err != nil {
		appLogger.Error("failed to connect to database",
			logging.Error(err),
			logging.String("driver", "postgres"),
			logging.String("dsn", fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
				cfg.Database.Postgres.Host,
				cfg.Database.Postgres.Port,
				cfg.Database.Postgres.User,
				cfg.Database.Postgres.Name,
				cfg.Database.Postgres.SSLMode,
			)),
		)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.Postgres.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.Postgres.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.Postgres.ConnMaxLifetime)

	// Verify connection
	if pingErr := sqlDB.Ping(); pingErr != nil {
		appLogger.Error("failed to ping database",
			logging.Error(pingErr),
		)
		return nil, fmt.Errorf("failed to ping database: %w", pingErr)
	}

	appLogger.Info("successfully connected to database",
		logging.String("driver", "postgres"),
	)

	return &GormDB{
		DB:     db,
		logger: appLogger,
	}, nil
}

// Close closes the database connection
func (db *GormDB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if closeErr := sqlDB.Close(); closeErr != nil {
		return fmt.Errorf("failed to close database connection: %w", closeErr)
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
	w.logger.Info("gorm query",
		logging.String("query", string(p)),
		logging.String("type", "raw_query"),
		logging.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)
	return len(p), nil
}

// Printf implements logger.Writer interface
func (w *GormLogWriter) Printf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	w.logger.Info("gorm query",
		logging.String("query", msg),
		logging.String("type", "formatted_query"),
		logging.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
		logging.String("args", fmt.Sprintf("%+v", args)),
	)
}

// Error implements logger.Writer interface
func (w *GormLogWriter) Error(msg string, err error) {
	errorType := "database_error"
	if errors.Is(err, gorm.ErrRecordNotFound) {
		errorType = "record_not_found"
	} else if errors.Is(err, gorm.ErrInvalidDB) {
		errorType = "invalid_db"
	} else if errors.Is(err, gorm.ErrInvalidTransaction) {
		errorType = "invalid_transaction"
	}

	w.logger.Error("gorm error",
		logging.String("message", msg),
		logging.Error(err),
		logging.String("type", errorType),
		logging.String("error_type", fmt.Sprintf("%T", err)),
		logging.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
		logging.String("stack_trace", fmt.Sprintf("%+v", err)),
		logging.String("sql_error", err.Error()),
		logging.String("sql_state", getSQLState(err)),
	)
}

// getSQLState extracts the SQL state from a database error
func getSQLState(err error) string {
	if err == nil {
		return ""
	}

	// Try to extract SQL state from error message
	if sqlErr, ok := err.(interface{ SQLState() string }); ok {
		return sqlErr.SQLState()
	}

	// Try to extract SQL state from error message
	if sqlErr, ok := err.(interface{ GetSQLState() string }); ok {
		return sqlErr.GetSQLState()
	}

	return ""
}

func (db *GormDB) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, DefaultPingTimeout)
	defer cancel()

	return db.DB.WithContext(pingCtx).Raw("SELECT 1").Error
}
