package middleware

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/goformx/goforms/internal/infrastructure/config"
)

// NewRecoveryMiddleware creates a new recovery middleware
func NewRecoveryMiddleware() core.Middleware {
	return &recoveryMiddleware{
		name:     "recovery",
		priority: 10,
	}
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware() core.Middleware {
	return &corsMiddleware{
		name:     "cors",
		priority: 20,
	}
}

// NewSecurityHeadersMiddleware creates a new security headers middleware
func NewSecurityHeadersMiddleware() core.Middleware {
	return &securityHeadersMiddleware{
		name:     "security-headers",
		priority: 50,
	}
}

// NewRequestIDMiddleware creates a new request ID middleware
func NewRequestIDMiddleware() core.Middleware {
	return &requestIDMiddleware{
		name:     "request-id",
		priority: 30,
	}
}

// NewTimeoutMiddleware creates a new timeout middleware
func NewTimeoutMiddleware() core.Middleware {
	return &timeoutMiddleware{
		name:     "timeout",
		priority: 40,
	}
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware() core.Middleware {
	return &loggingMiddleware{
		name:     "logging",
		priority: 90,
	}
}

// NewCSRFMiddleware creates a new CSRF middleware
func NewCSRFMiddleware() core.Middleware {
	return &csrfMiddleware{
		name:     "csrf",
		priority: 60,
	}
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware() core.Middleware {
	return &rateLimitMiddleware{
		name:     "rate-limit",
		priority: 70,
	}
}

// NewInputValidationMiddleware creates a new input validation middleware
func NewInputValidationMiddleware() core.Middleware {
	return &inputValidationMiddleware{
		name:     "input-validation",
		priority: 80,
	}
}

// NewSessionMiddleware creates a new session middleware
func NewSessionMiddleware() core.Middleware {
	return &sessionMiddleware{
		name:     "session",
		priority: 100,
	}
}

// NewAuthenticationMiddleware creates a new authentication middleware
func NewAuthenticationMiddleware() core.Middleware {
	return &authenticationMiddleware{
		name:     "authentication",
		priority: 110,
	}
}

// NewAuthorizationMiddleware creates a new authorization middleware
func NewAuthorizationMiddleware() core.Middleware {
	return &authorizationMiddleware{
		name:     "authorization",
		priority: 120,
	}
}

// Base middleware implementations

type recoveryMiddleware struct {
	name     string
	priority int
}

func (m *recoveryMiddleware) Process(ctx context.Context, req core.Request, next core.Handler) core.Response {
	defer func() {
		if r := recover(); r != nil {
			// Log the panic and return error response
			// In a real implementation, this would use the logger from context
			_ = r // Suppress unused variable warning
		}
	}()

	return next(ctx, req)
}

func (m *recoveryMiddleware) Name() string {
	return m.name
}

func (m *recoveryMiddleware) Priority() int {
	return m.priority
}

type corsMiddleware struct {
	name     string
	priority int
}

func (m *corsMiddleware) Process(ctx context.Context, req core.Request, next core.Handler) core.Response {
	// CORS logic would be implemented here
	return next(ctx, req)
}

func (m *corsMiddleware) Name() string {
	return m.name
}

func (m *corsMiddleware) Priority() int {
	return m.priority
}

type securityHeadersMiddleware struct {
	name     string
	priority int
}

func (m *securityHeadersMiddleware) Process(ctx context.Context, req core.Request, next core.Handler) core.Response {
	// Security headers logic would be implemented here
	return next(ctx, req)
}

func (m *securityHeadersMiddleware) Name() string {
	return m.name
}

func (m *securityHeadersMiddleware) Priority() int {
	return m.priority
}

type requestIDMiddleware struct {
	name     string
	priority int
}

func (m *requestIDMiddleware) Process(ctx context.Context, req core.Request, next core.Handler) core.Response {
	// Request ID logic would be implemented here
	return next(ctx, req)
}

func (m *requestIDMiddleware) Name() string {
	return m.name
}

func (m *requestIDMiddleware) Priority() int {
	return m.priority
}

type timeoutMiddleware struct {
	name     string
	priority int
}

func (m *timeoutMiddleware) Process(ctx context.Context, req core.Request, next core.Handler) core.Response {
	// Timeout logic would be implemented here
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return next(timeoutCtx, req)
}

func (m *timeoutMiddleware) Name() string {
	return m.name
}

func (m *timeoutMiddleware) Priority() int {
	return m.priority
}

type loggingMiddleware struct {
	name     string
	priority int
}

func (m *loggingMiddleware) Process(ctx context.Context, req core.Request, next core.Handler) core.Response {
	// Logging logic would be implemented here
	return next(ctx, req)
}

func (m *loggingMiddleware) Name() string {
	return m.name
}

func (m *loggingMiddleware) Priority() int {
	return m.priority
}

type csrfMiddleware struct {
	name     string
	priority int
}

func (m *csrfMiddleware) Process(ctx context.Context, req core.Request, next core.Handler) core.Response {
	// Get CSRF configuration from context
	cfg := m.getCSRFConfig(ctx)
	if cfg == nil {
		// If no CSRF config, skip CSRF processing
		return next(ctx, req)
	}

	// Check if CSRF should be skipped for this path
	if m.shouldSkipCSRF(req.Path(), cfg) {
		return next(ctx, req)
	}

	method := req.Method()

	// For GET requests, generate and set CSRF token
	if method == "GET" {
		token := m.generateCSRFToken(cfg)
		req.Set("csrf", token)

		// Set CSRF token in response headers for frontend access
		resp := next(ctx, req)
		resp.AddHeader("X-Csrf-Token", token)

		return resp
	}

	// For non-GET requests, validate CSRF token
	if !m.validateCSRFToken(req, cfg) {
		return core.NewErrorResponse(http.StatusForbidden, fmt.Errorf("CSRF token validation failed"))
	}

	return next(ctx, req)
}

func (m *csrfMiddleware) Name() string {
	return m.name
}

func (m *csrfMiddleware) Priority() int {
	return m.priority
}

// getCSRFConfig retrieves CSRF configuration from context
func (m *csrfMiddleware) getCSRFConfig(ctx context.Context) *config.CSRFConfig {
	// Try to get config from context
	if cfg, ok := ctx.Value("csrf_config").(*config.CSRFConfig); ok {
		return cfg
	}

	// Return default config for development
	return &config.CSRFConfig{
		Enabled:        true,
		Secret:         "default-csrf-secret-key-for-development",
		TokenName:      "_token",
		HeaderName:     "X-Csrf-Token",
		TokenLength:    32,
		ContextKey:     "csrf",
		CookieName:     "_csrf",
		CookiePath:     "/",
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
		CookieMaxAge:   86400,
	}
}

// shouldSkipCSRF checks if CSRF should be skipped for the given path
func (m *csrfMiddleware) shouldSkipCSRF(path string, cfg *config.CSRFConfig) bool {
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

	// Check configured skip paths
	for _, skipPath := range cfg.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

// generateCSRFToken generates a new CSRF token
func (m *csrfMiddleware) generateCSRFToken(cfg *config.CSRFConfig) string {
	// Generate random bytes
	randomBytes := make([]byte, cfg.TokenLength)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to timestamp-based token if crypto/rand fails
		timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
		randomBytes = []byte(timestamp)
	}

	// Create token by combining secret with random bytes
	tokenData := cfg.Secret + string(randomBytes)
	hash := sha256.Sum256([]byte(tokenData))

	// Return base64 encoded token
	return base64.StdEncoding.EncodeToString(hash[:])
}

// validateCSRFToken validates the CSRF token in the request
func (m *csrfMiddleware) validateCSRFToken(req core.Request, cfg *config.CSRFConfig) bool {
	// Get token from various sources
	token := m.extractCSRFToken(req, cfg)
	if token == "" {
		return false
	}

	// For now, we'll accept any non-empty token in development
	// In production, you'd want to validate against stored tokens
	return len(token) > 0
}

// extractCSRFToken extracts CSRF token from request headers, cookies, or form data
func (m *csrfMiddleware) extractCSRFToken(req core.Request, cfg *config.CSRFConfig) string {
	// Try header first
	if token := req.Headers().Get(cfg.HeaderName); token != "" {
		return token
	}

	// Try form data
	if form, err := req.Form(); err == nil {
		if token := form.Get(cfg.TokenName); token != "" {
			return token
		}
	}

	// Try cookies
	if cookies := req.Cookies(); len(cookies) > 0 {
		for _, cookie := range cookies {
			if cookie.Name == cfg.CookieName {
				return cookie.Value
			}
		}
	}

	return ""
}

type rateLimitMiddleware struct {
	name     string
	priority int
}

func (m *rateLimitMiddleware) Process(ctx context.Context, req core.Request, next core.Handler) core.Response {
	// Rate limiting logic would be implemented here
	return next(ctx, req)
}

func (m *rateLimitMiddleware) Name() string {
	return m.name
}

func (m *rateLimitMiddleware) Priority() int {
	return m.priority
}

type inputValidationMiddleware struct {
	name     string
	priority int
}

func (m *inputValidationMiddleware) Process(ctx context.Context, req core.Request, next core.Handler) core.Response {
	// Input validation logic would be implemented here
	return next(ctx, req)
}

func (m *inputValidationMiddleware) Name() string {
	return m.name
}

func (m *inputValidationMiddleware) Priority() int {
	return m.priority
}

type sessionMiddleware struct {
	name     string
	priority int
}

func (m *sessionMiddleware) Process(ctx context.Context, req core.Request, next core.Handler) core.Response {
	// Session logic would be implemented here
	return next(ctx, req)
}

func (m *sessionMiddleware) Name() string {
	return m.name
}

func (m *sessionMiddleware) Priority() int {
	return m.priority
}

type authenticationMiddleware struct {
	name     string
	priority int
}

func (m *authenticationMiddleware) Process(ctx context.Context, req core.Request, next core.Handler) core.Response {
	// Authentication logic would be implemented here
	return next(ctx, req)
}

func (m *authenticationMiddleware) Name() string {
	return m.name
}

func (m *authenticationMiddleware) Priority() int {
	return m.priority
}

type authorizationMiddleware struct {
	name     string
	priority int
}

func (m *authorizationMiddleware) Process(ctx context.Context, req core.Request, next core.Handler) core.Response {
	// Authorization logic would be implemented here
	return next(ctx, req)
}

func (m *authorizationMiddleware) Name() string {
	return m.name
}

func (m *authorizationMiddleware) Priority() int {
	return m.priority
}
