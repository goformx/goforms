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
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing token")
		}

		// Validate token
		if _, validateErr := m.userService.ValidateToken(token); validateErr != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}

		// Get user ID from token
		userID, idErr := m.userService.GetUserIDFromToken(token)
		if idErr != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}

		// Get user
		currentUser, userErr := m.userService.GetByID(c.Request().Context(), userID)
		if userErr != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
		}

		c.Set("user", currentUser)
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
			_, validateErr := m.userService.ValidateToken(token)
			if validateErr == nil {
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

func (m *AuthMiddleware) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Skip auth for public routes
		if isPublicRoute(c.Path()) {
			return next(c)
		}

		token, err := c.Cookie("token")
		if err != nil {
			return next(c)
		}

		// Validate token
		if _, validateErr := m.userService.ValidateToken(token.Value); validateErr != nil {
			return next(c)
		}

		// Get user ID from token
		userID, idErr := m.userService.GetUserIDFromToken(token.Value)
		if idErr != nil {
			return next(c)
		}

		// Get user
		currentUser, userErr := m.userService.GetByID(c.Request().Context(), userID)
		if userErr != nil {
			return next(c)
		}

		c.Set("user", currentUser)
		return next(c)
	}
}

func (m *AuthMiddleware) RefreshToken(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, err := c.Cookie("token")
		if err != nil {
			return next(c)
		}

		// Validate token
		if _, validateErr := m.userService.ValidateToken(token.Value); validateErr == nil {
			// Token is still valid, get user ID
			userID, idErr := m.userService.GetUserIDFromToken(token.Value)
			if idErr == nil {
				// Get user
				currentUser, userErr := m.userService.GetByID(c.Request().Context(), userID)
				if userErr == nil {
					c.Set("user", currentUser)
				}
			}
		}

		return next(c)
	}
}
