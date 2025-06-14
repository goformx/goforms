package health

import (
	"context"
	"database/sql"
)

// DatabaseChecker implements the Checker interface for database health checks
type DatabaseChecker struct {
	db *sql.DB
}

// NewDatabaseChecker creates a new database health checker
func NewDatabaseChecker(db *sql.DB) *DatabaseChecker {
	return &DatabaseChecker{db: db}
}

// Check performs a database health check
func (c *DatabaseChecker) Check(ctx context.Context) error {
	return c.db.PingContext(ctx)
}
