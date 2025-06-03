package auth

import (
	"github.com/goformx/goforms/internal/domain/repositories"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/handlers"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	*handlers.BaseHandler
	userRepo repositories.UserRepository
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(logger logging.Logger, userRepo repositories.UserRepository) *AuthHandler {
	return &AuthHandler{
		BaseHandler: handlers.NewBaseHandler(nil, logger),
		userRepo:    userRepo,
	}
}

// Register registers the auth routes
func (h *AuthHandler) Register(e *echo.Echo) {
	// No routes to register - all handled by web handlers
}
