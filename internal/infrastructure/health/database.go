// Package health provides health check utilities for infrastructure components such as the database.
package health

import (
	"context"
	"database/sql"
	"fmt"
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
	if err := c.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}
