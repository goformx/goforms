package middleware

import (
	"net/http"
	"strings"

	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// CookieAuthMiddleware handles cookie-based authentication
type CookieAuthMiddleware struct {
	userService user.Service
	logger      logging.Logger
}

// NewCookieAuthMiddleware creates a new cookie auth middleware
func NewCookieAuthMiddleware(userService user.Service) (*CookieAuthMiddleware, error) {
	logger, err := logging.NewLogger(false, "cookie_auth")
	if err != nil {
		return nil, err
	}

	return &CookieAuthMiddleware{
		userService: userService,
		logger:      logger,
	}, nil
}

// RequireAuth middleware ensures the user is authenticated
func (m *CookieAuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Skip auth for public routes
		if isPublicRoute(c.Path()) {
			return next(c)
		}

		// Get token from cookie
		token, err := getTokenFromCookie(c)
		if err != nil {
			m.logger.Error("missing token", logging.Error(err))
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing token")
		}

		// Validate token
		if _, validateErr := m.userService.ValidateToken(token); validateErr != nil {
			m.logger.Error("invalid token", logging.Error(validateErr))
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}

		// Get user ID from token
		userID, idErr := m.userService.GetUserIDFromToken(token)
		if idErr != nil {
			m.logger.Error("invalid token", logging.Error(idErr))
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}

		// Get user
		currentUser, userErr := m.userService.GetByID(c.Request().Context(), userID)
		if userErr != nil {
			m.logger.Error("user not found", logging.Error(userErr))
			return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
		}

		c.Set("user", currentUser)
		return next(c)
	}
}

// RequireNoAuth middleware ensures the user is not authenticated
func (m *CookieAuthMiddleware) RequireNoAuth(next echo.HandlerFunc) echo.HandlerFunc {
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

// Authenticate middleware attempts to authenticate the user
func (m *CookieAuthMiddleware) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
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

// RefreshToken middleware attempts to refresh the token
func (m *CookieAuthMiddleware) RefreshToken(next echo.HandlerFunc) echo.HandlerFunc {
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