package middleware

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/session"
	"github.com/labstack/echo/v4"
)

// EchoOrchestratorAdapter adapts the new middleware orchestrator to work with Echo
type EchoOrchestratorAdapter struct {
	orchestrator   core.Orchestrator
	logger         logging.Logger
	sessionManager *session.Manager
	accessManager  *access.Manager
}

// NewEchoOrchestratorAdapter creates a new Echo orchestrator adapter
func NewEchoOrchestratorAdapter(
	orchestrator core.Orchestrator,
	logger logging.Logger,
	sessionManager *session.Manager,
	accessManager *access.Manager,
) *EchoOrchestratorAdapter {
	return &EchoOrchestratorAdapter{
		orchestrator:   orchestrator,
		logger:         logger,
		sessionManager: sessionManager,
		accessManager:  accessManager,
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

	echoMiddleware := ea.ConvertChainToEcho(chain)
	for _, mw := range echoMiddleware {
		e.Use(mw)
	}

	return nil
}

// setupAPIChain sets up the API middleware chain
func (ea *EchoOrchestratorAdapter) setupAPIChain(e *echo.Echo) error {
	chain, err := ea.orchestrator.BuildChain(core.ChainTypeAPI)
	if err != nil {
		return fmt.Errorf("failed to build API chain: %w", err)
	}

	echoMiddleware := ea.ConvertChainToEcho(chain)
	for _, mw := range echoMiddleware {
		e.Use(mw)
	}

	return nil
}

// setupWebChain sets up the web middleware chain
func (ea *EchoOrchestratorAdapter) setupWebChain(e *echo.Echo) error {
	chain, err := ea.orchestrator.BuildChain(core.ChainTypeWeb)
	if err != nil {
		return fmt.Errorf("failed to build web chain: %w", err)
	}

	echoMiddleware := ea.ConvertChainToEcho(chain)
	for _, mw := range echoMiddleware {
		e.Use(mw)
	}

	return nil
}

// setupAuthChain sets up the auth middleware chain
func (ea *EchoOrchestratorAdapter) setupAuthChain(e *echo.Echo) error {
	chain, err := ea.orchestrator.BuildChain(core.ChainTypeAuth)
	if err != nil {
		return fmt.Errorf("failed to build auth chain: %w", err)
	}

	echoMiddleware := ea.ConvertChainToEcho(chain)

	for _, mw := range echoMiddleware {
		e.Use(mw)
	}

	return nil
}

// setupAdminChain sets up the admin middleware chain
func (ea *EchoOrchestratorAdapter) setupAdminChain(e *echo.Echo) error {
	chain, err := ea.orchestrator.BuildChain(core.ChainTypeAdmin)
	if err != nil {
		return fmt.Errorf("failed to build admin chain: %w", err)
	}

	echoMiddleware := ea.ConvertChainToEcho(chain)
	for _, mw := range echoMiddleware {
		e.Use(mw)
	}

	return nil
}

// setupPublicChain sets up the public middleware chain
func (ea *EchoOrchestratorAdapter) setupPublicChain(e *echo.Echo) error {
	chain, err := ea.orchestrator.BuildChain(core.ChainTypePublic)
	if err != nil {
		return fmt.Errorf("failed to build public chain: %w", err)
	}

	echoMiddleware := ea.ConvertChainToEcho(chain)
	for _, mw := range echoMiddleware {
		e.Use(mw)
	}

	return nil
}

// setupStaticChain sets up the static middleware chain
func (ea *EchoOrchestratorAdapter) setupStaticChain(e *echo.Echo) error {
	chain, err := ea.orchestrator.BuildChain(core.ChainTypeStatic)
	if err != nil {
		return fmt.Errorf("failed to build static chain: %w", err)
	}

	echoMiddleware := ea.ConvertChainToEcho(chain)
	for _, mw := range echoMiddleware {
		e.Use(mw)
	}

	return nil
}

// ConvertChainToEcho converts a middleware chain to Echo middleware functions
func (ea *EchoOrchestratorAdapter) ConvertChainToEcho(chain core.Chain) []echo.MiddlewareFunc {
	var echoMiddleware []echo.MiddlewareFunc

	// Log the middleware in the chain
	middlewares := chain.List()
	ea.logger.Debug("Converting chain to Echo middleware", "chain_length", len(middlewares))

	for _, mw := range middlewares {
		ea.logger.Debug("Chain middleware", "name", mw.Name())
	}

	// Add CSRF middleware for auth routes
	if ea.shouldApplyCSRF(chain) {
		ea.logger.Debug("Adding CSRF middleware to chain")
		echoMiddleware = append(echoMiddleware, ea.createCSRFMiddleware())
	} else {
		ea.logger.Debug("CSRF middleware not added to chain")
	}

	// Convert all middleware in the chain to Echo middleware
	for _, mw := range middlewares {
		echoMw := ea.convertMiddlewareToEcho(mw)
		if echoMw != nil {
			echoMiddleware = append(echoMiddleware, echoMw)

			ea.logger.Debug("Added Echo middleware", "name", mw.Name())
		} else {
			ea.logger.Debug("No Echo middleware for", "name", mw.Name())
		}
	}

	ea.logger.Debug("Final Echo middleware count", "count", len(echoMiddleware))

	return echoMiddleware
}

// convertMiddlewareToEcho converts a core middleware to an Echo middleware function
func (ea *EchoOrchestratorAdapter) convertMiddlewareToEcho(mw core.Middleware) echo.MiddlewareFunc {
	switch mw.Name() {
	case "session":
		return ea.createSessionMiddleware()
	case "authentication":
		return ea.createAuthenticationMiddleware()
	case "authorization":
		return ea.createAuthorizationMiddleware()
	case "csrf":
		return ea.createCSRFMiddleware()
	case "security-headers":
		return ea.createSecurityHeadersMiddleware()
	default:
		return nil
	}
}

// createSessionMiddleware creates an Echo session middleware
func (ea *EchoOrchestratorAdapter) createSessionMiddleware() echo.MiddlewareFunc {
	return ea.sessionManager.SessionMiddleware()
}

// createAuthenticationMiddleware creates an Echo authentication middleware
func (ea *EchoOrchestratorAdapter) createAuthenticationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if user_id is set in context (by session middleware)
			userID, exists := context.GetUserID(c)
			if !exists || userID == "" {
				path := c.Request().URL.Path

				// Check if this is an API request
				isAPIRequest := strings.HasPrefix(path, "/api/")
				acceptsJSON := strings.Contains(c.Request().Header.Get("Accept"), "application/json")

				if isAPIRequest || acceptsJSON {
					// Return JSON error response for API requests
					return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
				}

				// For web requests, redirect to login
				return c.Redirect(http.StatusSeeOther, "/login")
			}

			return next(c)
		}
	}
}

// createAuthorizationMiddleware creates an Echo authorization middleware
func (ea *EchoOrchestratorAdapter) createAuthorizationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// For now, just allow access if user is authenticated
			// In the future, this would check roles and permissions
			userID, exists := context.GetUserID(c)
			if !exists || userID == "" {
				return echo.NewHTTPError(http.StatusForbidden, "authorization required")
			}

			return next(c)
		}
	}
}

// shouldApplyCSRF checks if CSRF middleware should be applied to this chain
func (ea *EchoOrchestratorAdapter) shouldApplyCSRF(chain core.Chain) bool {
	// Check if the chain contains CSRF middleware
	middlewares := chain.List()

	for _, mw := range middlewares {
		if mw.Name() == "csrf" {
			ea.logger.Debug("CSRF middleware found in chain", "name", mw.Name())

			return true
		}
	}

	ea.logger.Debug("CSRF middleware not found in chain")

	return false
}

const httpMethodGET = "GET"

// createCSRFMiddleware creates an Echo CSRF middleware
func (ea *EchoOrchestratorAdapter) createCSRFMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			method := c.Request().Method

			// Skip CSRF for static files and health endpoints
			if ea.shouldSkipCSRF(path) {
				ea.logger.Debug("CSRF middleware skipped", "path", path, "method", method)

				return next(c)
			}

			// For GET requests, generate and set CSRF token
			if method == httpMethodGET {
				token := ea.generateCSRFToken()
				c.Set("csrf", token)
				ea.logger.Debug("CSRF token generated and set in context", "path", path, "token", token)
			}

			// For non-GET requests, validate CSRF token
			if method != httpMethodGET {
				valid := ea.validateCSRFToken(c)
				ea.logger.Debug("CSRF token validation", "path", path, "method", method, "valid", valid)

				if !valid {
					ea.logger.Warn("CSRF token validation failed", "path", path, "method", method)

					return echo.NewHTTPError(http.StatusForbidden, "CSRF token validation failed")
				}
			}

			return next(c)
		}
	}
}

// shouldSkipCSRF checks if CSRF should be skipped for the given path
func (ea *EchoOrchestratorAdapter) shouldSkipCSRF(path string) bool {
	// Skip for static files
	if strings.HasPrefix(path, "/static/") || strings.HasPrefix(path, "/assets/") {
		return true
	}

	// Skip for health and monitoring endpoints
	if strings.HasPrefix(path, "/health") || strings.HasPrefix(path, "/metrics") {
		return true
	}

	// Skip for public API endpoints
	if strings.HasPrefix(path, "/api/public/") {
		return true
	}

	return false
}

// generateCSRFToken generates a new CSRF token
func (ea *EchoOrchestratorAdapter) generateCSRFToken() string {
	// Generate random bytes
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to timestamp-based token if crypto/rand fails
		timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
		randomBytes = []byte(timestamp)
	}

	// Create token by combining secret with random bytes
	secret := "default-csrf-secret-key-for-development"
	tokenData := secret + string(randomBytes)
	hash := sha256.Sum256([]byte(tokenData))

	// Return base64 encoded token
	return base64.StdEncoding.EncodeToString(hash[:])
}

// validateCSRFToken validates the CSRF token in the request
func (ea *EchoOrchestratorAdapter) validateCSRFToken(c echo.Context) bool {
	token := c.Request().Header.Get("X-Csrf-Token")
	if token != "" {
		return true // For now, accept any non-empty token
	}

	formToken := c.FormValue("_token")
	if formToken != "" {
		return true
	}

	cookie, err := c.Cookie("_csrf")
	if err == nil && cookie.Value != "" {
		return true
	}

	return false
}

// BuildChainForPath builds a middleware chain for a specific path
func (ea *EchoOrchestratorAdapter) BuildChainForPath(path string) (core.Chain, error) {
	// Determine chain type based on path
	chainType := ea.determineChainType(path)
	ea.logger.Debug("Building chain for path", "path", path, "chain_type", chainType)

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

// isAuthPath checks if the path is an auth path (requires authentication)
func (ea *EchoOrchestratorAdapter) isAuthPath(path string) bool {
	// Auth paths are for authenticated users (like logout)
	return path == "/logout"
}

// isAdminPath checks if the path is an admin path
func (ea *EchoOrchestratorAdapter) isAdminPath(path string) bool {
	return len(path) >= 7 && path[:7] == "/admin"
}

// isPublicPath checks if the path is a public path
func (ea *EchoOrchestratorAdapter) isPublicPath(path string) bool {
	// Public paths that don't require authentication
	isPublic := path == "/" ||
		path == "/login" ||
		path == "/signup" ||
		path == "/forgot-password" ||
		path == "/reset-password" ||
		len(path) >= 8 && path[:8] == "/public"

	return isPublic
}

// isStaticPath checks if the path is a static path
func (ea *EchoOrchestratorAdapter) isStaticPath(path string) bool {
	return len(path) >= 8 && path[:8] == "/static" ||
		len(path) >= 8 && path[:8] == "/assets"
}

// createSecurityHeadersMiddleware creates an Echo security headers middleware
func (ea *EchoOrchestratorAdapter) createSecurityHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Set security headers
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			c.Response().Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

			// Set Strict-Transport-Security only for HTTPS
			if c.IsTLS() {
				c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			return next(c)
		}
	}
}
