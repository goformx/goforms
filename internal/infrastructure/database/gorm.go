// Package database provides database connection and ORM utilities for the application.
package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/mysql"
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
	// ConnectionPoolWarningThreshold is the percentage of max connections that triggers a warning
	ConnectionPoolWarningThreshold = 0.8
	// ConnectionPoolPercentageMultiplier is used to convert ratio to percentage
	ConnectionPoolPercentageMultiplier = 100
)

// GormDB wraps the GORM database connection
type GormDB struct {
	*gorm.DB
	logger logging.Logger
}

// TickerDuration controls how often the connection pool is monitored
var TickerDuration = 1 * time.Minute

// New creates a new GORM database connection
func New(cfg *config.Config, appLogger logging.Logger) (*GormDB, error) {
	// Configure GORM logger
	gormLogger := configureGormLogger(cfg, appLogger)

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true, // Enable prepared statements for better performance
	}

	// Create database connection
	db, err := createDatabaseConnection(cfg, gormConfig)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	if poolErr := configureConnectionPool(db, cfg); poolErr != nil {
		return nil, poolErr
	}

	// Verify connection
	if verifyErr := verifyConnection(db, appLogger); verifyErr != nil {
		return nil, verifyErr
	}

	appLogger.Info("database connection established",
		"driver", cfg.Database.Driver,
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"max_open_conns", cfg.Database.MaxOpenConns)

	return &GormDB{
		DB:     db,
		logger: appLogger,
	}, nil
}

// configureGormLogger configures the GORM logger with the specified settings
func configureGormLogger(cfg *config.Config, appLogger logging.Logger) logger.Interface {
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
	return logger.New(
		&GormLogWriter{logger: appLogger},
		logger.Config{
			SlowThreshold:             cfg.Database.Logging.SlowThreshold,
			LogLevel:                  gormLogLevel,
			IgnoreRecordNotFoundError: cfg.Database.Logging.IgnoreNotFound,
			ParameterizedQueries:      cfg.Database.Logging.Parameterized,
			Colorful:                  cfg.App.IsDevelopment(),
		},
	)
}

// createDatabaseConnection creates a database connection based on the configuration
func createDatabaseConnection(cfg *config.Config, gormConfig *gorm.Config) (*gorm.DB, error) {
	var db *gorm.DB

	var err error

	// Create database connection based on the selected driver
	switch cfg.Database.Driver {
	case "postgres":
		dsn := buildPostgresDSN(cfg)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	case "mariadb":
		dsn := buildMariaDBDSN(cfg)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
	default:
		return nil, fmt.Errorf("unsupported database connection type: %s", cfg.Database.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// buildPostgresDSN builds the PostgreSQL connection string
func buildPostgresDSN(cfg *config.Config) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
}

// buildMariaDBDSN builds the MariaDB connection string
func buildMariaDBDSN(cfg *config.Config) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)
}

// configureConnectionPool configures the database connection pool
func configureConnectionPool(db *gorm.DB, cfg *config.Config) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	return nil
}

// verifyConnection verifies the database connection by pinging it
func verifyConnection(db *gorm.DB, appLogger logging.Logger) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if pingErr := sqlDB.Ping(); pingErr != nil {
		appLogger.Error("failed to ping database", "error", pingErr)

		return fmt.Errorf("failed to ping database: %w", pingErr)
	}

	return nil
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
	// Use format properly as required by GORM logger interface
	message := fmt.Sprintf(format, args...)
	w.logger.Debug("GORM", "sql", message)

	if len(args) < durationArgPos+1 {
		return
	}

	query, ok := args[queryArgPos].(string)
	if !ok {
		query = "unknown query"
	}

	duration, ok := args[durationArgPos].(time.Duration)
	if !ok {
		duration = 0
	}

	rowsAffected := int64(0)

	if len(args) > rowsAffectedArgPos {
		if ra, raOk := args[rowsAffectedArgPos].(int64); raOk {
			rowsAffected = ra
		}
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

	errorType := w.getErrorType(err)

	w.logger.Error("database error",
		"message", msg,
		"type", errorType,
		"error", err)
}

// getErrorType determines the error type based on the GORM error
func (w *GormLogWriter) getErrorType(err error) string {
	for gormErr, errorType := range gormErrorTypes {
		if errors.Is(err, gormErr) {
			return errorType
		}
	}

	return "database_error"
}

// gormErrorTypes maps GORM errors to their corresponding error types
var gormErrorTypes = map[error]string{
	gorm.ErrInvalidDB:             "invalid_db",
	gorm.ErrInvalidTransaction:    "invalid_transaction",
	gorm.ErrNotImplemented:        "not_implemented",
	gorm.ErrMissingWhereClause:    "missing_where_clause",
	gorm.ErrUnsupportedDriver:     "unsupported_driver",
	gorm.ErrRegistered:            "already_registered",
	gorm.ErrInvalidField:          "invalid_field",
	gorm.ErrEmptySlice:            "empty_slice",
	gorm.ErrDryRunModeUnsupported: "dry_run_unsupported",
	gorm.ErrInvalidData:           "invalid_data",
	gorm.ErrUnsupportedRelation:   "unsupported_relation",
	gorm.ErrPrimaryKeyRequired:    "primary_key_required",
}

// MonitorConnectionPool monitors the database connection pool and logs metrics
func (db *GormDB) MonitorConnectionPool(ctx context.Context) {
	db.logger.Debug("starting MonitorConnectionPool")

	ticker := time.NewTicker(TickerDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			db.logger.Debug("MonitorConnectionPool context done")

			return
		case <-ticker.C:
			db.logger.Debug("MonitorConnectionPool tick")
			db.collectAndLogMetrics()
		}
	}
}

// collectAndLogMetrics collects and logs database connection pool metrics
func (db *GormDB) collectAndLogMetrics() {
	db.logger.Debug("collectAndLogMetrics called")

	sqlDB, err := db.DB.DB()
	if err != nil {
		db.logger.Error("failed to get database instance", map[string]any{"error": err})

		return
	}

	stats := sqlDB.Stats()
	metrics := map[string]any{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration,
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}

	// Add database-specific metrics
	db.addDatabaseSpecificMetrics(metrics)

	// Log the metrics
	db.logger.Info("database connection pool status", map[string]any{"metrics": metrics})

	// Check for high usage
	if float64(stats.InUse)/float64(stats.MaxOpenConnections) > ConnectionPoolWarningThreshold {
		db.logger.Warn("database connection pool usage is high",
			map[string]any{
				"in_use":   stats.InUse,
				"max_open": stats.MaxOpenConnections,
			})
	}

	// Check for long wait times
	if stats.WaitDuration > time.Second*5 {
		db.logger.Warn("database connection wait time is high",
			map[string]any{
				"wait_duration": stats.WaitDuration,
				"wait_count":    stats.WaitCount,
			})
	}
}

// addDatabaseSpecificMetrics adds database-specific metrics to the metrics map
func (db *GormDB) addDatabaseSpecificMetrics(metrics map[string]any) {
	switch db.Name() {
	case "postgres":
		db.addPostgresMetrics(metrics)
	case "mysql":
		db.addMySQLMetrics(metrics)
	}
}

// addPostgresMetrics adds PostgreSQL-specific metrics
func (db *GormDB) addPostgresMetrics(metrics map[string]any) {
	var pgStats struct {
		ActiveConnections  int64
		IdleConnections    int64
		WaitingConnections int64
	}

	// Get active connections
	if err := db.DB.Raw(
		"SELECT count(*) as active_connections FROM pg_stat_activity WHERE state = 'active'",
	).Scan(&pgStats.ActiveConnections).Error; err == nil {
		metrics["postgres_active_connections"] = pgStats.ActiveConnections
	}

	// Get idle connections
	if err := db.DB.Raw(
		"SELECT count(*) as idle_connections FROM pg_stat_activity WHERE state = 'idle'",
	).Scan(&pgStats.IdleConnections).Error; err == nil {
		metrics["postgres_idle_connections"] = pgStats.IdleConnections
	}

	// Get waiting connections
	if err := db.DB.Raw(
		"SELECT count(*) as waiting_connections FROM pg_stat_activity WHERE wait_event_type IS NOT NULL",
	).Scan(&pgStats.WaitingConnections).Error; err == nil {
		metrics["postgres_waiting_connections"] = pgStats.WaitingConnections
	}
}

// addMySQLMetrics adds MySQL-specific metrics
func (db *GormDB) addMySQLMetrics(metrics map[string]any) {
	var mysqlStats []struct {
		VariableName string
		Value        string
	}

	if err := db.DB.Raw(
		"SHOW STATUS WHERE Variable_name IN ('Threads_connected', 'Threads_running', 'Threads_waiting')",
	).Scan(&mysqlStats).Error; err == nil {
		for _, stat := range mysqlStats {
			metrics["mysql_"+strings.ToLower(stat.VariableName)] = stat.Value
		}
	}
}

// Ping checks the database connection by executing a simple query
func (db *GormDB) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, DefaultPingTimeout)
	defer cancel()

	return db.DB.WithContext(pingCtx).Raw("SELECT 1").Error
}

// NewWithDB creates a new GormDB instance with an existing DB connection
func NewWithDB(db *gorm.DB, appLogger logging.Logger) *GormDB {
	return &GormDB{
		DB:     db,
		logger: appLogger,
	}
}

// GetDB returns the underlying GORM DB instance
func (db *GormDB) GetDB() *gorm.DB {
	return db.DB
}
