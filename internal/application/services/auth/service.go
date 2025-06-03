package auth

import (
	"net/http"

	"github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// Service defines the interface for authentication operations
type Service interface {
	// GetAuthenticatedUser retrieves and validates the authenticated user from the context
	GetAuthenticatedUser(c echo.Context) (*user.User, error)
	// RequireAuth creates middleware that requires authentication
	RequireAuth(next echo.HandlerFunc) echo.HandlerFunc
}

// service implements the auth service
type service struct {
	userService user.Service
	logger      logging.Logger
}

// NewService creates a new auth service
func NewService(userService user.Service, logger logging.Logger) Service {
	return &service{
		userService: userService,
		logger:      logger,
	}
}

// GetAuthenticatedUser retrieves and validates the authenticated user from the context
func (s *service) GetAuthenticatedUser(c echo.Context) (*user.User, error) {
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		s.logger.Error("user not found in context",
			logging.StringField("path", c.Request().URL.Path),
			logging.StringField("method", c.Request().Method),
		)
		return nil, errors.New(errors.ErrCodeUnauthorized, "user not found", nil)
	}
	return currentUser, nil
}

// RequireAuth creates middleware that requires authentication
func (s *service) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		usr, err := s.GetAuthenticatedUser(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]any{
				"error": "Authentication required",
				"code":  errors.ErrCodeUnauthorized,
			})
		}

		// Set user in context for downstream handlers
		c.Set("user", usr)
		return next(c)
	}
}
