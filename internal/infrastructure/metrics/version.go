// Package metrics provides application metrics collection and reporting
// functionality for monitoring application performance and health.
package metrics

import (
	"sync"
	"time"

	"github.com/goformx/goforms/internal/infrastructure/version"
)

// VersionMetrics tracks version-related metrics
type VersionMetrics struct {
	mu sync.RWMutex

	// Version information
	Version   string
	BuildTime time.Time
	GitCommit string
	GoVersion string

	// Runtime metrics
	StartTime time.Time
	Uptime    time.Duration
}

// NewVersionMetrics creates new version metrics
func NewVersionMetrics() *VersionMetrics {
	info := version.GetInfo()

	buildTime, err := time.Parse(time.RFC3339, info.BuildTime)
	if err != nil {
		// Log error or use zero time
		buildTime = time.Time{}
	}

	return &VersionMetrics{
		Version:   info.Version,
		BuildTime: buildTime,
		GitCommit: info.GitCommit,
		GoVersion: info.GoVersion,
		StartTime: time.Now(),
	}
}

// Update updates the metrics
func (m *VersionMetrics) Update() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Uptime = time.Since(m.StartTime)
}

// GetMetrics returns the current metrics
func (m *VersionMetrics) GetMetrics() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]any{
		"version":    m.Version,
		"build_time": m.BuildTime.Format(time.RFC3339),
		"git_commit": m.GitCommit,
		"go_version": m.GoVersion,
		"uptime":     m.Uptime.String(),
		"start_time": m.StartTime.Format(time.RFC3339),
	}
}

// GetUptime returns the current uptime
func (m *VersionMetrics) GetUptime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.Uptime
}

// GetStartTime returns the start time
func (m *VersionMetrics) GetStartTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.StartTime
}
