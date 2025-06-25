package health

import (
	"context"
	"time"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Status represents the possible health check statuses
type Status string

const (
	// StatusHealthy represents a healthy component status
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// ComponentStatus represents the status of a system component
type ComponentStatus struct {
	Status    Status    `json:"status"`
	Message   string    `json:"message,omitempty"`
	LastCheck time.Time `json:"last_check"`
}

// HealthStatus represents the health status of the system
type HealthStatus struct {
	Status     Status                     `json:"status"`
	Components map[string]ComponentStatus `json:"components"`
	Timestamp  time.Time                  `json:"timestamp"`
}

// Checker defines the interface for component health checks
type Checker interface {
	// Check performs a health check of the component
	Check(ctx context.Context) error
}

// Service defines the health check service interface
type Service interface {
	// CheckHealth performs a health check of the system
	CheckHealth(ctx context.Context) (*HealthStatus, error)
}

// service implements the health check service
type service struct {
	logger   logging.Logger
	checkers map[string]Checker
}

// NewService creates a new health check service
func NewService(logger logging.Logger, checkers map[string]Checker) Service {
	return &service{
		logger:   logger,
		checkers: checkers,
	}
}

// CheckHealth performs a health check of the system
func (s *service) CheckHealth(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		Status:     StatusHealthy,
		Components: make(map[string]ComponentStatus),
		Timestamp:  time.Now(),
	}

	// Check each component
	for name, checker := range s.checkers {
		componentStatus := ComponentStatus{
			Status:    StatusHealthy,
			LastCheck: time.Now(),
		}

		if err := checker.Check(ctx); err != nil {
			s.logger.Error("health check failed",
				"component", name,
				"error", err,
			)
			status.Status = StatusUnhealthy
			componentStatus.Status = StatusUnhealthy
			componentStatus.Message = err.Error()
		}

		status.Components[name] = componentStatus
	}

	return status, nil
}
