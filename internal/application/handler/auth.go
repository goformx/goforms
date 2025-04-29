package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// AuthHandlerOption defines an auth handler option
type AuthHandlerOption func(*AuthHandler)

// WithUserService sets the user service
func WithUserService(svc user.Service) AuthHandlerOption {
	return func(h *AuthHandler) {
		h.userService = svc
	}
}

// AuthHandler handles authentication related requests
type AuthHandler struct {
	Base
	userService user.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(logger logging.Logger, opts ...AuthHandlerOption) *AuthHandler {
	h := &AuthHandler{
		Base: NewBase(WithLogger(logger)),
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Validate validates that required dependencies are set
func (h *AuthHandler) Validate() error {
	if err := h.Base.Validate(); err != nil {
		return err
	}
	if h.userService == nil {
		return errors.New("user service is required")
	}
	return nil
}

// Register registers the auth routes
func (h *AuthHandler) Register(e *echo.Echo) {
	if err := h.Validate(); err != nil {
		h.Logger.Error("failed to validate handler", logging.Error(err))
		return
	}

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
		h.Logger.Error("failed to bind signup request", logging.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	h.Logger.Debug("received signup request",
		logging.String("email", signup.Email),
		logging.String("first_name", signup.FirstName),
		logging.String("last_name", signup.LastName),
	)

	if err := c.Validate(signup); err != nil {
		h.Logger.Error("signup validation failed", logging.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	h.Logger.Debug("calling user service SignUp")
	newUser, err := h.userService.SignUp(c.Request().Context(), &signup)
	if err != nil {
		h.Logger.Debug("SignUp returned error", logging.Error(err))

		switch {
		case errors.Is(err, user.ErrUserExists):
			return echo.NewHTTPError(http.StatusConflict, "Email already exists")
		default:
			h.Logger.Error("unexpected error during signup", logging.Error(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
		}
	}

	h.Logger.Debug("signup successful",
		logging.Uint("user_id", newUser.ID),
		logging.String("email", newUser.Email),
	)
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
		h.Logger.Error("failed to bind login request", logging.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	h.Logger.Debug("login attempt",
		logging.String("email", login.Email),
		logging.Bool("has_password", login.Password != ""),
	)

	if err := c.Validate(login); err != nil {
		h.Logger.Error("login validation failed", logging.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tokens, err := h.userService.Login(c.Request().Context(), &login)
	if err != nil {
		h.Logger.Error("login failed",
			logging.Error(err),
			logging.String("email", login.Email),
		)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	h.Logger.Debug("login successful", logging.String("email", login.Email))
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
