package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// JWTMiddleware handles JWT authentication
type JWTMiddleware struct {
	userService user.Service
	secret      string
	logger      logging.Logger
	config      *config.Config
}

// NewJWTMiddleware creates a new JWT middleware
func NewJWTMiddleware(
	userService user.Service,
	secret string,
	logger logging.Logger,
	cfg *config.Config,
) echo.MiddlewareFunc {
	if userService == nil {
		panic("JWTMiddleware initialization failed: userService is required")
	}
	if logger == nil {
		panic("JWTMiddleware initialization failed: logger is required")
	}
	if cfg == nil {
		panic("JWTMiddleware initialization failed: config is required")
	}
	if secret == "" {
		panic("JWTMiddleware initialization failed: secret is required")
	}

	return (&JWTMiddleware{
		userService: userService,
		secret:      secret,
		logger:      logger,
		config:      cfg,
	}).Handle
}

// isAuthExempt checks if the path is exempt from authentication
func (m *JWTMiddleware) isAuthExempt(path string) bool {
	return strings.HasPrefix(path, "/"+m.config.Static.DistDir+"/") ||
		path == "/favicon.ico" || path == "/robots.txt" ||
		strings.HasPrefix(path, "/api/validation/") ||
		strings.HasPrefix(path, "/login") || strings.HasPrefix(path, "/signup") ||
		strings.HasPrefix(path, "/forgot-password") || strings.HasPrefix(path, "/contact") ||
		strings.HasPrefix(path, "/demo") || strings.HasPrefix(path, "/api/v1/auth/login")
}

// Handle processes JWT authentication
func (m *JWTMiddleware) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Request().URL.Path

		// Skip authentication for exempt paths
		if m.isAuthExempt(path) {
			m.logger.Debug("skipping auth check",
				logging.StringField("path", path))
			return next(c)
		}

		// Get token from header
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Warn("authorization header missing",
				logging.StringField("path", c.Path()),
				logging.StringField("method", c.Request().Method),
				logging.StringField("ip", c.RealIP()),
				logging.StringField("user_agent", c.Request().UserAgent()))
			return m.handleAuthError(c, echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header"))
		}

		// Parse token
		token, err := m.parseToken(authHeader)
		if err != nil {
			return m.handleAuthError(c, echo.NewHTTPError(http.StatusUnauthorized, err.Error()))
		}

		// Validate token claims
		if !token.Valid {
			return m.handleAuthError(c, echo.NewHTTPError(http.StatusForbidden, "invalid token claims"))
		}

		// Get user ID from claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return m.handleAuthError(c, echo.NewHTTPError(http.StatusForbidden, "invalid token claims format"))
		}

		userID, err := extractUserID(claims)
		if err != nil {
			return m.handleAuthError(c, echo.NewHTTPError(http.StatusForbidden, err.Error()))
		}

		// Get user from service
		userData, err := m.userService.GetByID(c.Request().Context(), userID)
		if err != nil {
			return m.handleAuthError(c, echo.NewHTTPError(http.StatusForbidden, "user not found or inactive"))
		}

		// Set user in context
		c.Set("user", userData)

		return next(c)
	}
}

// handleAuthError handles authentication errors
func (m *JWTMiddleware) handleAuthError(c echo.Context, err error) error {
	// Extract status code from HTTPError if available
	status := http.StatusUnauthorized
	var he *echo.HTTPError
	if errors.As(err, &he) {
		status = he.Code
	}

	m.logger.Error("auth check failed",
		logging.StringField("path", c.Path()),
		logging.StringField("method", c.Request().Method),
		logging.IntField("status", status),
		logging.ErrorField("error", err))
	return err
}

// parseToken parses and validates a JWT token
func (m *JWTMiddleware) parseToken(authHeader string) (*jwt.Token, error) {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errors.New("invalid authorization header format, expected 'Bearer <token>'")
	}

	token, err := jwt.Parse(parts[1], func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return token, nil
}

// extractUserID extracts and validates the user ID from JWT claims
func extractUserID(claims jwt.MapClaims) (string, error) {
	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return "", errors.New("invalid or missing user_id claim")
	}
	return userID, nil
}
