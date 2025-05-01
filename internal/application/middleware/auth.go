package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// JWTMiddleware handles JWT authentication
type JWTMiddleware struct {
	userService user.Service
	secret      string
	logger      logging.Logger
}

// NewJWTMiddleware creates a new JWT middleware
func NewJWTMiddleware(userService user.Service, secret string) (echo.MiddlewareFunc, error) {
	logger, err := logging.NewLogger(false, "auth")
	if err != nil {
		return nil, fmt.Errorf("failed to create auth logger: %w", err)
	}

	m := &JWTMiddleware{
		userService: userService,
		secret:      secret,
		logger:      logger,
	}
	return m.Handle, nil
}

// Handle processes JWT authentication
func (m *JWTMiddleware) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Request().URL.Path
		method := c.Request().Method

		// Skip authentication for static files and special files
		if strings.HasPrefix(path, "/static/") || 
		   path == "/favicon.ico" ||
		   path == "/robots.txt" {
			m.logger.Debug("JWT skipped: static content", 
				logging.String("path", path),
				logging.String("reason", "static content path"))
			return next(c)
		}

		// Skip authentication for validation API endpoints
		if strings.HasPrefix(path, "/api/validation/") {
			m.logger.Debug("JWT skipped: validation API", 
				logging.String("path", path),
				logging.String("reason", "validation API endpoint"))
			return next(c)
		}

		// Skip authentication for GET requests to public pages
		if method == http.MethodGet && (
			strings.HasPrefix(path, "/login") ||
			strings.HasPrefix(path, "/signup") ||
			strings.HasPrefix(path, "/forgot-password") ||
			strings.HasPrefix(path, "/contact") ||
			strings.HasPrefix(path, "/demo")) {
			m.logger.Debug("JWT skipped: public page", 
				logging.String("path", path),
				logging.String("method", method),
				logging.String("reason", "public page"))
			return next(c)
		}

		// Get token from header
		authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
		if authHeader == "" {
			return m.handleAuthError(c, fmt.Errorf("missing authorization header"))
		}

		// Parse token
		token, err := m.parseToken(authHeader)
		if err != nil {
			return m.handleAuthError(c, err)
		}

		// Validate token
		if !token.Valid {
			return m.handleAuthError(c, fmt.Errorf("invalid token"))
		}

		// Get user from token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return m.handleAuthError(c, fmt.Errorf("invalid token claims"))
		}

		// Get user ID from claims
		userID, ok := claims["sub"].(string)
		if !ok {
			return m.handleAuthError(c, fmt.Errorf("invalid user ID in token"))
		}

		// Get user from service
		user, err := m.userService.GetByID(c.Request().Context(), userID)
		if err != nil {
			return m.handleAuthError(c, fmt.Errorf("user not found: %w", err))
		}

		// Set user in context
		c.Set("user", user)

		return next(c)
	}
}

// handleAuthError handles authentication errors
func (m *JWTMiddleware) handleAuthError(c echo.Context, err error) error {
	m.logger.Error("authentication failed",
		logging.String("path", c.Path()),
		logging.String("method", c.Request().Method),
		logging.Error(err),
	)
	return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
}

// parseToken parses and validates a JWT token
func (m *JWTMiddleware) parseToken(authHeader string) (*jwt.Token, error) {
	// Extract token from header
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	// Parse token
	token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
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
