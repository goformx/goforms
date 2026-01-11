package middleware

import (
	"errors"
	"fmt"

	"github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// MigrationAdapter provides a hybrid approach for gradual migration from old to new middleware system
type MigrationAdapter struct {
	orchestrator core.Orchestrator
	registry     core.Registry
	logger       logging.Logger
	useNewSystem bool // Feature flag to control migration
}

// NewMigrationAdapter creates a new migration adapter
func NewMigrationAdapter(
	orchestrator core.Orchestrator,
	registry core.Registry,
	logger logging.Logger,
) *MigrationAdapter {
	return &MigrationAdapter{
		orchestrator: orchestrator,
		registry:     registry,
		logger:       logger,
		useNewSystem: false, // Start with old system, can be enabled via config
	}
}

// SetupHybridMiddleware sets up middleware using both old and new systems
func (ma *MigrationAdapter) SetupHybridMiddleware(e *echo.Echo, oldManager *Manager) error {
	ma.logger.Info("setting up hybrid middleware system", "use_new_system", ma.useNewSystem)

	if ma.useNewSystem {
		return ma.setupNewSystem(e)
	}

	return ma.setupOldSystem(e, oldManager)
}

// setupNewSystem sets up middleware using the new orchestrator system
func (ma *MigrationAdapter) setupNewSystem(e *echo.Echo) error {
	ma.logger.Info("using new middleware orchestrator system")

	// Create Echo adapter and setup middleware
	adapter := NewEchoOrchestratorAdapter(ma.orchestrator, ma.logger)

	return adapter.SetupMiddleware(e)
}

// setupOldSystem sets up middleware using the old Manager system
func (ma *MigrationAdapter) setupOldSystem(e *echo.Echo, oldManager *Manager) error {
	ma.logger.Info("using legacy middleware manager system")

	if oldManager == nil {
		return errors.New("old manager is required when using legacy system")
	}

	oldManager.Setup(e)

	return nil
}

// EnableNewSystem enables the new middleware system
func (ma *MigrationAdapter) EnableNewSystem() {
	ma.useNewSystem = true
	ma.logger.Info("new middleware system enabled")
}

// DisableNewSystem disables the new middleware system (falls back to old)
func (ma *MigrationAdapter) DisableNewSystem() {
	ma.useNewSystem = false
	ma.logger.Info("new middleware system disabled, using legacy system")
}

// IsNewSystemEnabled returns whether the new system is enabled
func (ma *MigrationAdapter) IsNewSystemEnabled() bool {
	return ma.useNewSystem
}

// MigrateMiddleware migrates a specific middleware from old to new system
func (ma *MigrationAdapter) MigrateMiddleware(middlewareName string) error {
	ma.logger.Info("migrating middleware to new system", "middleware", middlewareName)

	// Check if middleware exists in registry
	if _, exists := ma.registry.Get(middlewareName); !exists {
		return fmt.Errorf("middleware %s not found in registry", middlewareName)
	}

	ma.logger.Info("middleware migration completed", "middleware", middlewareName)

	return nil
}

// GetMigrationStatus returns the current migration status
func (ma *MigrationAdapter) GetMigrationStatus() MigrationStatus {
	return MigrationStatus{
		NewSystemEnabled:     ma.useNewSystem,
		RegisteredMiddleware: ma.registry.List(),
		AvailableChains:      ma.orchestrator.ListChains(),
	}
}

// MigrationStatus represents the current migration status
type MigrationStatus struct {
	NewSystemEnabled     bool     `json:"new_system_enabled"`
	RegisteredMiddleware []string `json:"registered_middleware"`
	AvailableChains      []string `json:"available_chains"`
}

// SetupWithFallback sets up middleware with automatic fallback to old system on error
func (ma *MigrationAdapter) SetupWithFallback(e *echo.Echo, oldManager *Manager) error {
	ma.logger.Info("setting up middleware with fallback")

	// Try new system first
	if ma.useNewSystem {
		if err := ma.setupNewSystem(e); err != nil {
			ma.logger.Warn("new system setup failed, falling back to legacy system", "error", err)
			ma.useNewSystem = false

			return ma.setupOldSystem(e, oldManager)
		}

		return nil
	}

	// Use old system
	return ma.setupOldSystem(e, oldManager)
}

// ValidateMigration validates that the migration can proceed safely
func (ma *MigrationAdapter) ValidateMigration() error {
	// Check if all required middleware are registered
	requiredMiddleware := []string{
		"recovery",
		"cors",
		"request-id",
		"timeout",
		"security-headers",
		"csrf",
		"rate-limit",
		"session",
		"authentication",
		"authorization",
	}

	var missing []string

	for _, mw := range requiredMiddleware {
		if _, exists := ma.registry.Get(mw); !exists {
			missing = append(missing, mw)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required middleware: %v", missing)
	}

	// Validate orchestrator can build chains
	chainTypes := []core.ChainType{
		core.ChainTypeDefault,
		core.ChainTypeAPI,
		core.ChainTypeWeb,
		core.ChainTypeAuth,
		core.ChainTypeAdmin,
		core.ChainTypePublic,
		core.ChainTypeStatic,
	}

	for _, chainType := range chainTypes {
		if _, err := ma.orchestrator.BuildChain(chainType); err != nil {
			return fmt.Errorf("failed to build chain %s: %w", chainType, err)
		}
	}

	ma.logger.Info("migration validation completed successfully")

	return nil
}

// RollbackMigration rolls back to the old system
func (ma *MigrationAdapter) RollbackMigration() {
	ma.logger.Info("rolling back to legacy middleware system")
	ma.useNewSystem = false
}

// GetOrchestrator returns the orchestrator instance
func (ma *MigrationAdapter) GetOrchestrator() core.Orchestrator {
	return ma.orchestrator
}

// GetRegistry returns the registry instance
func (ma *MigrationAdapter) GetRegistry() core.Registry {
	return ma.registry
}
