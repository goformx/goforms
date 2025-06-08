package health

import (
	"context"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Service defines the health check service interface
type Service interface {
	// CheckHealth performs a health check of the system
	CheckHealth(ctx context.Context) (*HealthStatus, error)
}

// HealthStatus represents the health status of the system
type HealthStatus struct {
	Status     string            `json:"status"`
	Components map[string]string `json:"components"`
}

// Repository defines the interface for health check operations
type Repository interface {
	// PingContext checks if the database is accessible
	PingContext(ctx context.Context) error
}

// service implements the health check service
type service struct {
	logger     logging.Logger
	repository Repository
}

// NewService creates a new health check service
func NewService(logger logging.Logger, repository Repository) Service {
	return &service{
		logger:     logger,
		repository: repository,
	}
}

// CheckHealth performs a health check of the system
func (s *service) CheckHealth(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		Status:     "healthy",
		Components: make(map[string]string),
	}

	// Check database connectivity
	if err := s.repository.PingContext(ctx); err != nil {
		s.logger.Error("health check failed",
			logging.Error(err),
			logging.String("component", "database"),
		)
		status.Status = "unhealthy"
		status.Components["database"] = "down"
		return status, err
	}

	status.Components["database"] = "up"
	return status, nil
}
