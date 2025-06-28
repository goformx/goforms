// Package access provides access control middleware and utilities for the application.
package access

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Middleware creates a new access control middleware
func Middleware(manager *Manager, _ logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			method := c.Request().Method

			// Get required access level for this route
			requiredAccess := manager.GetRequiredAccess(path, method)

			// Check if user has required access
			switch requiredAccess {
			case Public:
				// No authentication required
				return next(c)

			case Authenticated:
				// Check if user is authenticated
				if !context.IsAuthenticated(c) {
					return c.Redirect(http.StatusSeeOther, constants.PathLogin)
				}

				return next(c)

			case Admin:
				// Check if user is authenticated and is an admin
				if !context.IsAuthenticated(c) {
					return c.Redirect(http.StatusSeeOther, constants.PathLogin)
				}

				if !context.IsAdmin(c) {
					return response.ErrorResponse(c, http.StatusForbidden, "Admin access required")
				}

				return next(c)

			default:
				// Default to requiring authentication
				if !context.IsAuthenticated(c) {
					return c.Redirect(http.StatusSeeOther, constants.PathLogin)
				}

				return next(c)
			}
		}
	}
}
