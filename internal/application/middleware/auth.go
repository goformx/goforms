package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
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
) (echo.MiddlewareFunc, error) {
	m := &JWTMiddleware{
		userService: userService,
		secret:      secret,
		logger:      logger,
		config:      cfg,
	}
	return m.Handle, nil
}

// isStaticPath checks if the path is for static content
func (m *JWTMiddleware) isStaticPath(path string) bool {
	return strings.HasPrefix(path, "/"+m.config.Static.DistDir+"/") ||
		path == "/favicon.ico" ||
		path == "/robots.txt"
}

// isValidationAPI checks if the path is for validation API
func (m *JWTMiddleware) isValidationAPI(path string) bool {
	return strings.HasPrefix(path, "/api/validation/")
}

// isPublicPage checks if the path is for a public page
func (m *JWTMiddleware) isPublicPage(path string) bool {
	return strings.HasPrefix(path, "/login") ||
		strings.HasPrefix(path, "/signup") ||
		strings.HasPrefix(path, "/forgot-password") ||
		strings.HasPrefix(path, "/contact") ||
		strings.HasPrefix(path, "/demo") ||
		strings.HasPrefix(path, "/api/v1/auth/login")
}

// Handle processes JWT authentication
func (m *JWTMiddleware) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Request().URL.Path
		method := c.Request().Method

		// Skip authentication for static files and special files
		if m.isStaticPath(path) {
			m.logger.Debug("skipping auth check",
				logging.StringField("path", path),
				logging.StringField("reason", "static content path"))
			return next(c)
		}

		// Skip authentication for validation API endpoints
		if m.isValidationAPI(path) {
			m.logger.Debug("skipping auth check",
				logging.StringField("path", path),
				logging.StringField("reason", "validation API endpoint"))
			return next(c)
		}

		// Skip authentication for public pages
		if m.isPublicPage(path) {
			m.logger.Debug("skipping auth check",
				logging.StringField("path", path),
				logging.StringField("method", method),
				logging.StringField("reason", "public page"))
			return next(c)
		}

		// Get token from header
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return m.handleAuthError(c, errors.New("missing authorization header"))
		}

		// Parse token
		token, err := m.parseToken(authHeader)
		if err != nil {
			return m.handleAuthError(c, errors.New("invalid token"))
		}

		// Validate token claims
		if !token.Valid {
			return m.handleAuthError(c, errors.New("invalid token claims"))
		}

		// Get user ID from claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return m.handleAuthError(c, errors.New("invalid token claims"))
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID in token")
		}

		// Get user from service
		userData, err := m.userService.GetByID(c.Request().Context(), fmt.Sprintf("%v", userID))
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid user")
		}

		// Set user in context
		c.Set("user", userData)

		return next(c)
	}
}

// handleAuthError handles authentication errors
func (m *JWTMiddleware) handleAuthError(c echo.Context, err error) error {
	m.logger.Error("auth check failed",
		logging.StringField("path", c.Path()),
		logging.StringField("method", c.Request().Method),
		logging.ErrorField("error", err))
	return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
}

// parseToken parses and validates a JWT token
func (m *JWTMiddleware) parseToken(authHeader string) (*jwt.Token, error) {
	// Parse token from authorization header
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errors.New("invalid authorization header format")
	}

	// Parse token
	token, err := jwt.Parse(parts[1], func(token *jwt.Token) (any, error) {
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
