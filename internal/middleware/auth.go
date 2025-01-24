package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWTConfig holds the configuration for JWT middleware
type JWTConfig struct {
	Secret []byte
}

// NewJWTMiddleware creates a new JWT authentication middleware
func NewJWTMiddleware(secret string) echo.MiddlewareFunc {
	config := &JWTConfig{
		Secret: []byte(secret),
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip authentication for login and signup endpoints
			if c.Path() == "/api/v1/auth/login" || c.Path() == "/api/v1/auth/signup" {
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
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return config.Secret, nil
			})

			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}

			// Set user information in context
			c.Set("user_id", uint(claims["user_id"].(float64)))
			c.Set("email", claims["email"].(string))
			c.Set("role", claims["role"].(string))

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
