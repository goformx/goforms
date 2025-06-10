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
	// MinArgsLength represents the minimum number of arguments needed for a query
	MinArgsLength = 2
	// GORM query argument positions
	queryArgPos        = 0
	durationArgPos     = 1
	rowsAffectedArgPos = 2
)

// GormDB wraps the GORM database connection
type GormDB struct {
	*gorm.DB
	logger logging.Logger
}

// NewGormDB creates a new GORM database connection
func NewGormDB(cfg *config.Config, appLogger logging.Logger) (*GormDB, error) {
	// Map our log levels to GORM log levels
	var gormLogLevel logger.LogLevel
	switch cfg.Database.Logging.LogLevel {
	case "silent":
		gormLogLevel = logger.Silent
	case "error":
		gormLogLevel = logger.Error
	case "warn":
		gormLogLevel = logger.Warn
	case "info":
		gormLogLevel = logger.Info
	default:
		gormLogLevel = logger.Warn // Default to warn level
	}

	// Configure GORM logger with enhanced settings
	gormLogger := logger.New(
		&GormLogWriter{logger: appLogger},
		logger.Config{
			SlowThreshold:             cfg.Database.Logging.SlowThreshold,
			LogLevel:                  gormLogLevel,
			IgnoreRecordNotFoundError: cfg.Database.Logging.IgnoreNotFound,
			ParameterizedQueries:      cfg.Database.Logging.Parameterized,
			Colorful:                  cfg.App.IsDevelopment(),
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
		appLogger.Error("failed to connect to database", "error", err)
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
		appLogger.Error("failed to ping database", "error", pingErr)
		return nil, fmt.Errorf("failed to ping database: %w", pingErr)
	}

	appLogger.Info("database connection established",
		"host", cfg.Database.Postgres.Host,
		"port", cfg.Database.Postgres.Port,
		"max_open_conns", cfg.Database.Postgres.MaxOpenConns)

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
		db.logger.Error("failed to close database connection", "error", closeErr)
		return fmt.Errorf("failed to close database connection: %w", closeErr)
	}

	return nil
}

// GormLogWriter implements io.Writer for GORM logger
type GormLogWriter struct {
	logger logging.Logger
}

// Write implements io.Writer interface
func (w *GormLogWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// Printf implements logger.Writer interface
func (w *GormLogWriter) Printf(format string, args ...any) {
	if len(args) < durationArgPos+1 {
		return
	}

	query, _ := args[queryArgPos].(string)
	duration, _ := args[durationArgPos].(time.Duration)
	rowsAffected := int64(0)
	if len(args) > rowsAffectedArgPos {
		rowsAffected, _ = args[rowsAffectedArgPos].(int64)
	}

	// Log all queries in debug mode
	w.logger.Debug("database query",
		"query", query,
		"duration", duration,
		"rows_affected", rowsAffected)

	// Warn on slow queries
	if duration > time.Millisecond*100 {
		w.logger.Warn("slow query detected",
			"query", query,
			"duration", duration,
			"rows_affected", rowsAffected,
			"threshold", "100ms")
	}
}

// Error implements logger.Writer interface
func (w *GormLogWriter) Error(msg string, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		w.logger.Debug("record not found",
			"message", msg,
			"error", err)
		return
	}

	errorType := "database_error"
	switch {
	case errors.Is(err, gorm.ErrInvalidDB):
		errorType = "invalid_db"
	case errors.Is(err, gorm.ErrInvalidTransaction):
		errorType = "invalid_transaction"
	case errors.Is(err, gorm.ErrNotImplemented):
		errorType = "not_implemented"
	case errors.Is(err, gorm.ErrMissingWhereClause):
		errorType = "missing_where_clause"
	case errors.Is(err, gorm.ErrUnsupportedDriver):
		errorType = "unsupported_driver"
	case errors.Is(err, gorm.ErrRegistered):
		errorType = "already_registered"
	case errors.Is(err, gorm.ErrInvalidField):
		errorType = "invalid_field"
	case errors.Is(err, gorm.ErrEmptySlice):
		errorType = "empty_slice"
	case errors.Is(err, gorm.ErrDryRunModeUnsupported):
		errorType = "dry_run_unsupported"
	case errors.Is(err, gorm.ErrInvalidData):
		errorType = "invalid_data"
	case errors.Is(err, gorm.ErrUnsupportedRelation):
		errorType = "unsupported_relation"
	case errors.Is(err, gorm.ErrPrimaryKeyRequired):
		errorType = "primary_key_required"
	}

	w.logger.Error("database error",
		"message", msg,
		"type", errorType,
		"error", err)
}

// MonitorConnectionPool periodically checks the database connection pool status
func (db *GormDB) MonitorConnectionPool(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sqlDB, err := db.DB.DB()
			if err != nil {
				db.logger.Error("failed to get database instance for monitoring",
					"error", err)
				continue
			}

			stats := sqlDB.Stats()
			db.logger.Info("database connection pool status",
				"max_open_connections", stats.MaxOpenConnections,
				"open_connections", stats.OpenConnections,
				"in_use", stats.InUse,
				"idle", stats.Idle,
				"wait_count", stats.WaitCount,
				"wait_duration", stats.WaitDuration,
				"max_idle_closed", stats.MaxIdleClosed,
				"max_lifetime_closed", stats.MaxLifetimeClosed)

			// Alert on high connection usage
			if float64(stats.InUse)/float64(stats.MaxOpenConnections) > 0.8 {
				db.logger.Warn("high database connection usage",
					"in_use", stats.InUse,
					"max_open", stats.MaxOpenConnections,
					"usage_percentage", float64(stats.InUse)/float64(stats.MaxOpenConnections)*100)
			}
		}
	}
}

func (db *GormDB) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, DefaultPingTimeout)
	defer cancel()

	return db.DB.WithContext(pingCtx).Raw("SELECT 1").Error
}
