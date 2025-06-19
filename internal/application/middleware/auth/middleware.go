package auth

import (
	"net/http"

	"github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// Middleware provides authentication utilities for handlers
type Middleware struct {
	logger      logging.Logger
	userService user.Service
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(logger logging.Logger, userService user.Service) *Middleware {
	return &Middleware{
		logger:      logger,
		userService: userService,
	}
}

// RequireAuth middleware ensures user is authenticated
func (am *Middleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, err := am.RequireAuthenticatedUser(c)
		if err != nil {
			return err
		}

		// Store user in context for downstream handlers
		c.Set("user", user)
		return next(c)
	}
}

// OptionalAuth middleware provides user if authenticated, but doesn't require it
func (am *Middleware) OptionalAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, ok := context.GetUserID(c)
		if ok {
			user, err := am.userService.GetUserByID(c.Request().Context(), userID)
			if err == nil && user != nil {
				c.Set("user", user)
			}
		}
		return next(c)
	}
}

// GetUserFromContext safely retrieves user from context
func (am *Middleware) GetUserFromContext(c echo.Context) (*entities.User, bool) {
	user, ok := c.Get("user").(*entities.User)
	return user, ok
}

// RedirectIfAuthenticated redirects authenticated users
func (am *Middleware) RedirectIfAuthenticated(c echo.Context, redirectPath string) error {
	user, err := am.RequireAuthenticatedUser(c)
	if err == nil && user != nil {
		return c.Redirect(http.StatusFound, redirectPath)
	}
	return nil
}

// RequireAuthenticatedUser ensures the user is authenticated and returns the user object
func (am *Middleware) RequireAuthenticatedUser(c echo.Context) (*entities.User, error) {
	userID, ok := context.GetUserID(c)
	if !ok {
		return nil, c.Redirect(http.StatusSeeOther, "/login")
	}

	user, err := am.userService.GetUserByID(c.Request().Context(), userID)
	if err != nil || user == nil {
		am.logger.Error("failed to get user", "error", err)
		return nil, response.WebErrorResponse(c, nil, http.StatusInternalServerError, "Failed to get user")
	}

	return user, nil
}
