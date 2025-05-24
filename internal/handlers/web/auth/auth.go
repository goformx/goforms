package auth

import (
	"net/http"

	"github.com/goformx/goforms/internal/domain/repositories"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/handlers"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	*handlers.BaseHandler
	userRepo repositories.UserRepository
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(logger logging.Logger, userRepo repositories.UserRepository) *AuthHandler {
	return &AuthHandler{
		BaseHandler: handlers.NewBaseHandler(nil, nil, logger),
		userRepo:    userRepo,
	}
}

// Register registers the auth routes
func (h *AuthHandler) Register(e *echo.Echo) {
	e.POST("/api/v1/auth/login", h.Login)
	e.POST("/api/v1/auth/logout", h.Logout)
}

// Login handles user login
func (h *AuthHandler) Login(c echo.Context) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// TODO: Add validation logic or use domain validator if needed
	// For now, just check for empty email or password
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email and password are required"})
	}

	// Get user by email
	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil || user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	// Check password
	compareErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if compareErr != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	// TODO: Generate JWT token and set cookie
	// For now, just return success
	return c.JSON(http.StatusOK, map[string]string{"message": "Login successful"})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Logout successful"})
}
