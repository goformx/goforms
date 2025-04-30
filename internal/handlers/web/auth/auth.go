package auth

import (
	"net/http"

	"github.com/jonesrussell/goforms/internal/application/validation"
	"github.com/jonesrussell/goforms/internal/domain/entities"
	"github.com/jonesrussell/goforms/internal/domain/repositories"
	"github.com/jonesrussell/goforms/internal/handlers"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	base handlers.Base
	userRepo repositories.UserRepository
}

func NewAuthHandler(
	logger logging.Logger,
	userRepo repositories.UserRepository,
) *AuthHandler {
	return &AuthHandler{
		base: handlers.Base{
			Logger: logger,
		},
		userRepo: userRepo,
	}
}

func (h *AuthHandler) Register(e *echo.Echo) {
	h.base.RegisterRoute(e, "POST", "/auth/signup", h.Signup)
	h.base.RegisterRoute(e, "POST", "/auth/login", h.Login)
}

func (h *AuthHandler) Signup(c echo.Context) error {
	var req validation.SignupRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if err := validation.ValidateSignup(&req); err != nil {
		return c.JSON(http.StatusBadRequest, validation.GetValidationErrors(err))
	}

	// Check if user already exists
	existingUser, err := h.userRepo.FindByEmail(req.Email)
	if err == nil && existingUser != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email already registered"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
	}

	// Create new user
	user := &entities.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	createErr := h.userRepo.Create(user)
	if createErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "User created successfully"})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req validation.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if err := validation.ValidateLogin(&req); err != nil {
		return c.JSON(http.StatusBadRequest, validation.GetValidationErrors(err))
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