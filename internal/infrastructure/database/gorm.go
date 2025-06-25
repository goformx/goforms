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

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true, // Enable prepared statements for better performance
	}

	var db *gorm.DB
	var err error

	// Create database connection based on the selected driver
	switch cfg.Database.Connection {
	case "postgres":
		// Build DSN for PostgreSQL
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Database,
			cfg.Database.SSLMode,
		)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)

	case "mariadb":
		// Build DSN for MariaDB
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Database,
		)
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)

	default:
		return nil, fmt.Errorf("unsupported database connection type: %s", cfg.Database.Connection)
	}

	if err != nil {
		appLogger.Error("failed to connect to database", "error", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Verify connection
	if pingErr := sqlDB.Ping(); pingErr != nil {
		appLogger.Error("failed to ping database", "error", pingErr)
		return nil, fmt.Errorf("failed to ping database: %w", pingErr)
	}

	appLogger.Info("database connection established",
		"driver", cfg.Database.Connection,
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"max_open_conns", cfg.Database.MaxOpenConns)

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
		if ra, ok := args[rowsAffectedArgPos].(int64); ok {
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
	switch db.DB.Dialector.Name() {
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
