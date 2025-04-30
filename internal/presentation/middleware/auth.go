package middleware

import (
	"net/http"
	"strings"

	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/labstack/echo/v4"
)

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	userService user.Service
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(userService user.Service) *AuthMiddleware {
	return &AuthMiddleware{
		userService: userService,
	}
}

// RequireAuth middleware ensures the user is authenticated
func (m *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Skip auth for public routes
		if isPublicRoute(c.Path()) {
			return next(c)
		}

		// Get token from cookie
		token, err := getTokenFromCookie(c)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		// Validate token
		if _, err := m.userService.ValidateToken(token); err != nil {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		// Check if token is blacklisted
		if m.userService.IsTokenBlacklisted(token) {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		// Get user ID from token
		userID, err := m.userService.GetUserIDFromToken(token)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		// Get user by ID
		user, err := m.userService.GetByID(c.Request().Context(), userID)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		// Set user in context
		c.Set("user", user)
		return next(c)
	}
}

// RequireNoAuth middleware ensures the user is not authenticated
func (m *AuthMiddleware) RequireNoAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get token from cookie
		token, err := getTokenFromCookie(c)
		if err == nil && token != "" {
			// Validate token
			if _, err := m.userService.ValidateToken(token); err == nil {
				// Check if token is blacklisted
				if !m.userService.IsTokenBlacklisted(token) {
					return c.Redirect(http.StatusSeeOther, "/dashboard")
				}
			}
		}
		return next(c)
	}
}

// isPublicRoute checks if the route is public
func isPublicRoute(path string) bool {
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

// getTokenFromCookie extracts token from cookie
func getTokenFromCookie(c echo.Context) (string, error) {
	cookie, err := c.Cookie("token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
} 