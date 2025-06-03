package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Rate limit configuration
const (
	authRateLimit  = 5  // requests per second
	authBurstLimit = 10 // maximum burst size
	authWindowSize = 1 * time.Minute
)

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	userService user.Service
	logger      logging.Logger
	config      *config.Config
	limiter     *rate.Limiter
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(
	userService user.Service,
	logger logging.Logger,
	cfg *config.Config,
) *AuthMiddleware {
	if userService == nil {
		panic("AuthMiddleware initialization failed: userService is required")
	}
	if logger == nil {
		panic("AuthMiddleware initialization failed: logger is required")
	}
	if cfg == nil {
		panic("AuthMiddleware initialization failed: config is required")
	}

	return &AuthMiddleware{
		userService: userService,
		logger:      logger,
		config:      cfg,
		limiter:     rate.NewLimiter(rate.Limit(authRateLimit), authBurstLimit),
	}
}

// Middleware creates a new authentication middleware function
func (m *AuthMiddleware) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip auth check for certain paths
			if m.isAuthExempt(c.Request().URL.Path) {
				return next(c)
			}

			// Get session from context
			session, ok := c.Get("session").(*Session)
			if !ok {
				return m.handleAuthError(c, "no session found")
			}

			// Get user from service
			userData, err := m.userService.GetUserByID(c.Request().Context(), session.UserID)
			if err != nil {
				return m.handleAuthError(c, "user not found or inactive")
			}

			// Set user in context
			c.Set("user", userData)
			c.Set("user_id", session.UserID)
			c.Set("user_email", session.Email)
			c.Set("user_role", session.Role)

			return next(c)
		}
	}
}

// isAuthExempt checks if the path is exempt from authentication
func (m *AuthMiddleware) isAuthExempt(path string) bool {
	return isStaticFile(path) ||
		strings.HasPrefix(path, "/api/validation/") ||
		strings.HasPrefix(path, "/login") || strings.HasPrefix(path, "/signup") ||
		strings.HasPrefix(path, "/forgot-password") || strings.HasPrefix(path, "/contact") ||
		strings.HasPrefix(path, "/demo")
}

// handleAuthError handles authentication errors
func (m *AuthMiddleware) handleAuthError(c echo.Context, message string) error {
	m.logger.Error("auth check failed",
		logging.StringField("path", c.Path()),
		logging.StringField("method", c.Request().Method),
		logging.StringField("error", message))

	// For API requests, return 401
	if c.Request().Header.Get("Accept") == "application/json" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": message,
		})
	}

	// For web requests, redirect to login
	return c.Redirect(http.StatusSeeOther, "/login")
}

// SecurityHeaders adds security-related headers to responses
func (m *AuthMiddleware) SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Add security headers
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			return next(c)
		}
	}
}

// Note: isPublicRoute is defined in middleware.go
