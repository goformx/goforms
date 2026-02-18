package middleware

import (
	"fmt"

	"github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// EchoOrchestratorAdapter adapts the new middleware orchestrator to work with Echo
type EchoOrchestratorAdapter struct {
	orchestrator core.Orchestrator
	logger       logging.Logger
}

// NewEchoOrchestratorAdapter creates a new Echo orchestrator adapter
func NewEchoOrchestratorAdapter(orchestrator core.Orchestrator, logger logging.Logger) *EchoOrchestratorAdapter {
	return &EchoOrchestratorAdapter{
		orchestrator: orchestrator,
		logger:       logger,
	}
}

// SetupMiddleware sets up middleware chains on the Echo instance
func (ea *EchoOrchestratorAdapter) SetupMiddleware(e *echo.Echo) error {
	// Build and apply different middleware chains based on path patterns

	// Default chain for all routes
	if err := ea.setupDefaultChain(e); err != nil {
		return fmt.Errorf("failed to setup default chain: %w", err)
	}

	// API chain for API routes
	if err := ea.setupAPIChain(e); err != nil {
		return fmt.Errorf("failed to setup API chain: %w", err)
	}

	// Web chain for web routes
	if err := ea.setupWebChain(e); err != nil {
		return fmt.Errorf("failed to setup web chain: %w", err)
	}

	// Auth chain for authentication routes
	if err := ea.setupAuthChain(e); err != nil {
		return fmt.Errorf("failed to setup auth chain: %w", err)
	}

	// Admin chain for admin routes
	if err := ea.setupAdminChain(e); err != nil {
		return fmt.Errorf("failed to setup admin chain: %w", err)
	}

	// Public chain for public routes
	if err := ea.setupPublicChain(e); err != nil {
		return fmt.Errorf("failed to setup public chain: %w", err)
	}

	// Static chain for static assets
	if err := ea.setupStaticChain(e); err != nil {
		return fmt.Errorf("failed to setup static chain: %w", err)
	}

	ea.logger.Info("middleware chains setup completed")

	return nil
}

// setupDefaultChain sets up the default middleware chain
func (ea *EchoOrchestratorAdapter) setupDefaultChain(e *echo.Echo) error {
	chain, err := ea.orchestrator.BuildChain(core.ChainTypeDefault)
	if err != nil {
		return fmt.Errorf("failed to build default chain: %w", err)
	}

	echoMiddleware := ea.convertChainToEcho(chain)
	for _, mw := range echoMiddleware {
		e.Use(mw)
	}

	ea.logger.Info("default middleware chain applied", "middleware_count", len(echoMiddleware))

	return nil
}

// setupAPIChain sets up the API middleware chain
func (ea *EchoOrchestratorAdapter) setupAPIChain(e *echo.Echo) error {
	chain, err := ea.orchestrator.BuildChain(core.ChainTypeAPI)
	if err != nil {
		return fmt.Errorf("failed to build API chain: %w", err)
	}

	echoMiddleware := ea.convertChainToEcho(chain)
	for _, mw := range echoMiddleware {
		e.Use(mw)
	}

	ea.logger.Info("API middleware chain applied", "middleware_count", len(echoMiddleware))

	return nil
}

// setupWebChain sets up the web middleware chain
func (ea *EchoOrchestratorAdapter) setupWebChain(e *echo.Echo) error {
	chain, err := ea.orchestrator.BuildChain(core.ChainTypeWeb)
	if err != nil {
		return fmt.Errorf("failed to build web chain: %w", err)
	}

	echoMiddleware := ea.convertChainToEcho(chain)
	for _, mw := range echoMiddleware {
		e.Use(mw)
	}

	ea.logger.Info("web middleware chain applied", "middleware_count", len(echoMiddleware))

	return nil
}

// setupAuthChain sets up the auth middleware chain
func (ea *EchoOrchestratorAdapter) setupAuthChain(e *echo.Echo) error {
	chain, err := ea.orchestrator.BuildChain(core.ChainTypeAuth)
	if err != nil {
		return fmt.Errorf("failed to build auth chain: %w", err)
	}

	echoMiddleware := ea.convertChainToEcho(chain)
	for _, mw := range echoMiddleware {
		e.Use(mw)
	}

	ea.logger.Info("auth middleware chain applied", "middleware_count", len(echoMiddleware))

	return nil
}

// setupAdminChain sets up the admin middleware chain
func (ea *EchoOrchestratorAdapter) setupAdminChain(e *echo.Echo) error {
	chain, err := ea.orchestrator.BuildChain(core.ChainTypeAdmin)
	if err != nil {
		return fmt.Errorf("failed to build admin chain: %w", err)
	}

	echoMiddleware := ea.convertChainToEcho(chain)
	for _, mw := range echoMiddleware {
		e.Use(mw)
	}

	ea.logger.Info("admin middleware chain applied", "middleware_count", len(echoMiddleware))

	return nil
}

// setupPublicChain sets up the public middleware chain
func (ea *EchoOrchestratorAdapter) setupPublicChain(e *echo.Echo) error {
	chain, err := ea.orchestrator.BuildChain(core.ChainTypePublic)
	if err != nil {
		return fmt.Errorf("failed to build public chain: %w", err)
	}

	echoMiddleware := ea.convertChainToEcho(chain)
	for _, mw := range echoMiddleware {
		e.Use(mw)
	}

	ea.logger.Info("public middleware chain applied", "middleware_count", len(echoMiddleware))

	return nil
}

// setupStaticChain sets up the static middleware chain
func (ea *EchoOrchestratorAdapter) setupStaticChain(e *echo.Echo) error {
	chain, err := ea.orchestrator.BuildChain(core.ChainTypeStatic)
	if err != nil {
		return fmt.Errorf("failed to build static chain: %w", err)
	}

	echoMiddleware := ea.convertChainToEcho(chain)
	for _, mw := range echoMiddleware {
		e.Use(mw)
	}

	ea.logger.Info("static middleware chain applied", "middleware_count", len(echoMiddleware))

	return nil
}

// convertChainToEcho converts a middleware chain to Echo middleware functions
func (ea *EchoOrchestratorAdapter) convertChainToEcho(chain core.Chain) []echo.MiddlewareFunc {
	echoMiddleware := make([]echo.MiddlewareFunc, 0, 1)

	// For now, return a simple no-op middleware
	// This will be expanded to convert actual middleware
	echoMiddleware = append(echoMiddleware, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ea.logger.Debug("middleware chain processing", "chain_length", chain.Length())

			return next(c)
		}
	})

	return echoMiddleware
}

// BuildChainForPath builds a middleware chain for a specific path
func (ea *EchoOrchestratorAdapter) BuildChainForPath(path string) (core.Chain, error) {
	// Determine chain type based on path
	chainType := ea.determineChainType(path)

	chain, err := ea.orchestrator.BuildChain(chainType)
	if err != nil {
		return nil, fmt.Errorf("failed to build chain for path %s: %w", path, err)
	}

	return chain, nil
}

// determineChainType determines the appropriate chain type for a given path
func (ea *EchoOrchestratorAdapter) determineChainType(path string) core.ChainType {
	switch {
	case ea.isAPIPath(path):
		return core.ChainTypeAPI
	case ea.isWebPath(path):
		return core.ChainTypeWeb
	case ea.isAuthPath(path):
		return core.ChainTypeAuth
	case ea.isAdminPath(path):
		return core.ChainTypeAdmin
	case ea.isPublicPath(path):
		return core.ChainTypePublic
	case ea.isStaticPath(path):
		return core.ChainTypeStatic
	default:
		return core.ChainTypeDefault
	}
}

// isAPIPath checks if the path is an API path
func (ea *EchoOrchestratorAdapter) isAPIPath(path string) bool {
	return len(path) >= 4 && path[:4] == "/api"
}

// isWebPath checks if the path is a web path
func (ea *EchoOrchestratorAdapter) isWebPath(path string) bool {
	return len(path) >= 10 && path[:10] == "/dashboard" ||
		len(path) >= 6 && path[:6] == "/forms"
}

// isAuthPath checks if the path is an auth path
func (ea *EchoOrchestratorAdapter) isAuthPath(path string) bool {
	return path == "/login" || path == "/signup" || path == "/logout" ||
		path == "/forgot-password" || path == "/reset-password"
}

// isAdminPath checks if the path is an admin path
func (ea *EchoOrchestratorAdapter) isAdminPath(path string) bool {
	return len(path) >= 7 && path[:7] == "/admin"
}

// isPublicPath checks if the path is a public path
func (ea *EchoOrchestratorAdapter) isPublicPath(path string) bool {
	return len(path) >= 8 && path[:8] == "/public"
}

// isStaticPath checks if the path is a static path
func (ea *EchoOrchestratorAdapter) isStaticPath(path string) bool {
	return len(path) >= 8 && path[:8] == "/static" ||
		len(path) >= 8 && path[:8] == "/assets"
}
