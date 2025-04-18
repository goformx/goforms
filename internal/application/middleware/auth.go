package middleware

import (
	"errors"
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
		// Skip auth for public routes
		if m.isPublicRoute(c.Path()) {
			return next(c)
		}

		// Get token from header
		token, err := m.getTokenFromHeader(c)
		if err != nil {
			return m.handleAuthError(c, err)
		}

		// Validate token
		claims, err := m.validateToken(token)
		if err != nil {
			return m.handleAuthError(c, err)
		}

		// Set user context
		if setErr := m.setUserContext(c, claims); setErr != nil {
			return setErr
		}

		// Check role-based access
		if roleErr := m.checkRoleAccess(c); roleErr != nil {
			return roleErr
		}

		return next(c)
	}
}

// isPublicRoute checks if the route is public
func (m *JWTMiddleware) isPublicRoute(path string) bool {
	publicRoutes := []string{
		"/health",
		"/login",
		"/signup",
		"/api/v1/contact",
		"/api/v1/subscription",
	}
	for _, route := range publicRoutes {
		if strings.HasPrefix(path, route) {
			return true
		}
	}
	return false
}

// getTokenFromHeader extracts token from Authorization header
func (m *JWTMiddleware) getTokenFromHeader(c echo.Context) (string, error) {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return "", errors.New("missing authorization header")
	}
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", errors.New("invalid authorization header format")
	}
	return strings.TrimPrefix(auth, "Bearer "), nil
}

// validateToken validates the JWT token
func (m *JWTMiddleware) validateToken(token string) (jwt.MapClaims, error) {
	tokenObj, err := m.userService.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := tokenObj.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// setUserContext sets user information in the context
func (m *JWTMiddleware) setUserContext(c echo.Context, claims jwt.MapClaims) error {
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return errors.New("invalid user_id claim")
	}
	email, ok := claims["email"].(string)
	if !ok {
		return errors.New("invalid email claim")
	}
	role, ok := claims["role"].(string)
	if !ok {
		return errors.New("invalid role claim")
	}

	c.Set("user_id", uint(userID))
	c.Set("email", email)
	c.Set("role", role)
	return nil
}

// checkRoleAccess verifies role-based access
func (m *JWTMiddleware) checkRoleAccess(c echo.Context) error {
	role, ok := c.Get("role").(string)
	if !ok {
		return errors.New("invalid role type in context")
	}
	path := c.Path()

	// Admin routes
	if strings.HasPrefix(path, "/admin") && role != "admin" {
		return errors.New("unauthorized: admin access required")
	}

	// User routes
	if strings.HasPrefix(path, "/user") && role != "user" && role != "admin" {
		return errors.New("unauthorized: user access required")
	}

	return nil
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
