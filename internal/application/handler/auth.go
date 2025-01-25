package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	Base
	userService user.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(logger logging.Logger, userService user.Service) *AuthHandler {
	return &AuthHandler{
		Base:        Base{Logger: logger},
		userService: userService,
	}
}

// Register registers the auth routes
func (h *AuthHandler) Register(e *echo.Echo) {
	g := e.Group("/api/v1/auth")
	g.POST("/signup", h.handleSignup)
	g.POST("/login", h.handleLogin)
	g.POST("/logout", h.handleLogout)
}

// handleSignup handles user registration
// @Summary Register a new user
// @Description Register a new user with the provided information
// @Tags auth
// @Accept json
// @Produce json
// @Param signup body user.Signup true "User signup information"
// @Success 201 {object} user.User
// @Failure 400 {object} echo.HTTPError
// @Failure 409 {object} echo.HTTPError
// @Router /api/v1/auth/signup [post]
func (h *AuthHandler) handleSignup(c echo.Context) error {
	var signup user.Signup
	if err := c.Bind(&signup); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(signup); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	newUser, err := h.userService.SignUp(c.Request().Context(), &signup)
	if err != nil {
		h.LogError("failed to create user", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}

	return c.JSON(http.StatusCreated, newUser)
}

// handleLogin handles user authentication
// @Summary Authenticate user
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param login body user.Login true "User login credentials"
// @Success 200 {object} user.TokenPair
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) handleLogin(c echo.Context) error {
	var login user.Login
	if err := c.Bind(&login); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(login); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tokens, err := h.userService.Login(c.Request().Context(), &login)
	if err != nil {
		h.LogError("failed to authenticate user", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	return c.JSON(http.StatusOK, tokens)
}

// handleLogout handles user logout
// @Summary Logout user
// @Description Invalidate user's tokens
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} echo.HTTPError
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) handleLogout(c echo.Context) error {
	token := c.Request().Header.Get("Authorization")
	if token == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Missing authorization token")
	}

	if err := h.userService.Logout(c.Request().Context(), token); err != nil {
		h.LogError("failed to logout user", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to logout")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Successfully logged out"})
}
