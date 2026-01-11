// Package auth provides authentication middleware and utilities for
// protecting routes and managing user authentication state.
package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Middleware provides authentication utilities for handlers
type Middleware struct {
	logger       logging.Logger
	userService  user.Service
	errorHandler response.ErrorHandlerInterface
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(
	logger logging.Logger,
	userService user.Service,
	errorHandler response.ErrorHandlerInterface,
) *Middleware {
	return &Middleware{
		logger:       logger,
		userService:  userService,
		errorHandler: errorHandler,
	}
}

// RequireAuth middleware ensures user is authenticated
func (am *Middleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userEntity, err := am.RequireAuthenticatedUser(c)
		if err != nil {
			return err
		}

		// Store user in context for downstream handlers
		c.Set("user", userEntity)

		return next(c)
	}
}

// OptionalAuth middleware provides user if authenticated, but doesn't require it
func (am *Middleware) OptionalAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, ok := context.GetUserID(c)
		if ok {
			userEntity, err := am.userService.GetUserByID(c.Request().Context(), userID)
			if err == nil && userEntity != nil {
				c.Set("user", userEntity)
			}
		}

		return next(c)
	}
}

// GetUserFromContext safely retrieves user from context
func (am *Middleware) GetUserFromContext(c echo.Context) (*entities.User, bool) {
	userEntity, ok := c.Get("user").(*entities.User)

	return userEntity, ok
}

// RedirectIfAuthenticated redirects authenticated users
func (am *Middleware) RedirectIfAuthenticated(c echo.Context, redirectPath string) error {
	userEntity, err := am.RequireAuthenticatedUser(c)
	if err == nil && userEntity != nil {
		if redirectErr := c.Redirect(http.StatusFound, redirectPath); redirectErr != nil {
			return fmt.Errorf("redirect authenticated user: %w", redirectErr)
		}

		return nil
	}

	return nil
}

// RequireAuthenticatedUser ensures the user is authenticated and returns the user object
func (am *Middleware) RequireAuthenticatedUser(c echo.Context) (*entities.User, error) {
	userID, ok := context.GetUserID(c)
	if !ok {
		// No session found, redirect to login
		if redirectErr := c.Redirect(http.StatusSeeOther, constants.PathLogin); redirectErr != nil {
			return nil, fmt.Errorf("redirect to login: %w", redirectErr)
		}

		return nil, errors.New("user not authenticated")
	}

	userEntity, err := am.userService.GetUserByID(c.Request().Context(), userID)
	if err != nil || userEntity == nil {
		am.logger.Error("failed to get user", "error", err)

		if handleErr := am.errorHandler.HandleError(err, c, "Failed to get user"); handleErr != nil {
			return nil, fmt.Errorf("handle authentication error: %w", handleErr)
		}

		return nil, errors.New("user not found")
	}

	return userEntity, nil
}
