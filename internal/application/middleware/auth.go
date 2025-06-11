package middleware

import (
	"net/http"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// RequireAuth is a middleware that checks if a user is logged in.
// If not, it redirects to the login page.
func RequireAuth(logger logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := c.Get("user_id").(string)
			if !ok || userID == "" {
				logger.Debug("authentication required",
					"path", c.Request().URL.Path,
					"method", c.Request().Method,
					"ip", c.RealIP(),
					"user_agent", c.Request().UserAgent(),
				)
				return c.Redirect(http.StatusSeeOther, "/login")
			}
			logger.Debug("user authenticated",
				"path", c.Request().URL.Path,
				"method", c.Request().Method,
				"user_id", userID,
				"ip", c.RealIP(),
				"user_agent", c.Request().UserAgent(),
			)
			return next(c)
		}
	}
}
