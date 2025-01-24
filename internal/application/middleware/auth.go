package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/user"
)

// JWTClaims is a type alias for jwt.MapClaims
type JWTClaims = jwt.MapClaims

// JWTConfig holds the configuration for JWT middleware
type JWTConfig struct {
	Secret      []byte
	UserService user.Service
}

// NewJWTMiddleware creates a new JWT authentication middleware
func NewJWTMiddleware(userService user.Service, secret string) echo.MiddlewareFunc {
	config := &JWTConfig{
		Secret:      []byte(secret),
		UserService: userService,
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip authentication for public endpoints
			if isPublicPath(c.Path()) {
				return next(c)
			}

			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.Split(auth, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			tokenString := parts[1]

			// Check if token is blacklisted
			if config.UserService.IsTokenBlacklisted(tokenString) {
				return echo.NewHTTPError(http.StatusUnauthorized, "token has been invalidated")
			}

			// Validate token
			token, err := config.UserService.ValidateToken(tokenString)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			claims, ok := token.Claims.(*JWTClaims)
			if !ok || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}

			// Check token type
			tokenType, ok := (*claims)["type"].(string)
			if !ok || tokenType != "access" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token type")
			}

			// Set user information in context
			c.Set("user_id", uint((*claims)["user_id"].(float64)))
			c.Set("email", (*claims)["email"].(string))
			c.Set("role", (*claims)["role"].(string))

			return next(c)
		}
	}
}

// RequireRole creates middleware that checks if the user has the required role
func RequireRole(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole := c.Get("role").(string)
			if userRole != role {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
			}
			return next(c)
		}
	}
}

// isPublicPath checks if the path is a public endpoint that doesn't require authentication
func isPublicPath(path string) bool {
	publicPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/signup",
		"/api/v1/auth/refresh",
	}

	for _, p := range publicPaths {
		if path == p {
			return true
		}
	}

	return false
}
