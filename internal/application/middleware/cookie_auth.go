package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"errors"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// CookieAuthMiddleware handles cookie-based authentication
type CookieAuthMiddleware struct {
	userService user.Service
	logger      logging.Logger
}

// NewCookieAuthMiddleware creates a new cookie auth middleware
func NewCookieAuthMiddleware(userService user.Service, logger logging.Logger) *CookieAuthMiddleware {
	return &CookieAuthMiddleware{
		userService: userService,
		logger:      logger,
	}
}

// validateToken performs all token validation steps and returns the user, validation status, and any error
func (m *CookieAuthMiddleware) validateToken(c echo.Context, tokenStr string) (*user.User, bool, error) {
	// Validate the token
	if err := m.userService.ValidateToken(c.Request().Context(), tokenStr); err != nil {
		return nil, false, err
	}

	// Get user ID from token
	userID, err := m.userService.GetUserIDFromToken(c.Request().Context(), tokenStr)
	if err != nil {
		return nil, false, err
	}

	// Retrieve user using request context
	userObj, err := m.userService.GetByID(c.Request().Context(), strconv.FormatUint(uint64(userID), 10))
	if err != nil {
		return nil, false, err
	}

	return userObj, true, nil
}

// getAuthToken retrieves and validates the authentication token from cookies
func getAuthToken(c echo.Context) (string, error) {
	cookie, err := c.Cookie("token")
	if err != nil || cookie.Value == "" {
		return "", errors.New("missing or empty token")
	}
	return cookie.Value, nil
}

// publicRoutePrefixes defines common prefixes for public routes
var publicRoutePrefixes = []string{
	"/api/v1",
	"/auth/",
	"/health",
}

// isPublicRoute checks if the route is public using optimized matching
func isPublicRoute(path string) bool {
	// Check exact match first
	if publicRoutes[path] {
		return true
	}

	// Check common prefixes
	for _, prefix := range publicRoutePrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// RequireAuth middleware ensures the user is authenticated
func (m *CookieAuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Skip auth for public routes
		if isPublicRoute(c.Path()) {
			return next(c)
		}

		token, err := getAuthToken(c)
		if err != nil {
			m.logger.Error("missing token",
				logging.StringField("path", c.Path()),
				logging.StringField("method", c.Request().Method),
				logging.ErrorField("error", err))
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		userObj, valid, err := m.validateToken(c, token)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				m.logger.Warn("token expired",
					logging.StringField("path", c.Path()),
					logging.StringField("method", c.Request().Method))
			} else {
				m.logger.Error("invalid token",
					logging.StringField("path", c.Path()),
					logging.StringField("method", c.Request().Method),
					logging.ErrorField("error", err))
			}
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		if !valid {
			m.logger.Error("token validation failed",
				logging.StringField("path", c.Path()),
				logging.StringField("method", c.Request().Method))
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		c.Set("user", userObj)
		return next(c)
	}
}

// RequireNoAuth middleware ensures the user is not authenticated
func (m *CookieAuthMiddleware) RequireNoAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, err := getAuthToken(c)
		if err == nil {
			_, valid, validateErr := m.validateToken(c, token)
			if validateErr == nil && valid {
				isBlacklisted, err := m.userService.IsTokenBlacklisted(c.Request().Context(), token)
				if err == nil && !isBlacklisted {
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

		token, err := getAuthToken(c)
		if err != nil {
			return next(c)
		}

		userObj, valid, err := m.validateToken(c, token)
		if err != nil || !valid {
			return next(c)
		}

		c.Set("user", userObj)
		return next(c)
	}
}

// RefreshToken middleware attempts to refresh the token
func (m *CookieAuthMiddleware) RefreshToken(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, err := getAuthToken(c)
		if err != nil {
			return next(c)
		}

		userObj, valid, err := m.validateToken(c, token)
		if err == nil && valid {
			// Renew secure cookie
			cookie := new(http.Cookie)
			cookie.Name = "token"
			cookie.Value = token
			cookie.HttpOnly = true
			cookie.Secure = true
			cookie.Path = "/"
			cookie.SameSite = http.SameSiteStrictMode
			c.SetCookie(cookie)

			c.Set("user", userObj)
		}

		return next(c)
	}
}

// publicRoutes defines routes that don't require authentication
var publicRoutes = map[string]bool{
	"/health":              true,
	"/login":               true,
	"/signup":              true,
	"/api/v1/contact":      true,
	"/api/v1/subscription": true,
}
