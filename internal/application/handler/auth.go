package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/handlers"
)

const (
	// CookieExpiryMinutes is the number of minutes before a cookie expires
	CookieExpiryMinutes = 15
	// SecondsInMinute is the number of seconds in a minute
	SecondsInMinute = 60
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
	*handlers.BaseHandler
	userService user.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(logger logging.Logger, opts ...AuthHandlerOption) *AuthHandler {
	h := &AuthHandler{
		BaseHandler: handlers.NewBaseHandler(nil, nil, logger),
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Validate validates that required dependencies are set
func (h *AuthHandler) Validate() error {
	if err := h.BaseHandler.Validate(); err != nil {
		h.LogError("failed to validate handler", err)
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
		h.LogError("failed to validate handler", err)
		return
	}

	// API routes
	g := e.Group("/api/v1/auth")
	g.POST("/signup", h.handleSignup)
	g.POST("/login", h.handleLogin)
	g.POST("/logout", h.handleLogout)

	// Web routes - logout only via POST for security
	e.POST("/logout", h.handleWebLogout)
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
		h.LogError("failed to bind signup request", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	h.LogDebug("received signup request",
		logging.StringField("email", signup.Email),
		logging.StringField("first_name", signup.FirstName),
		logging.StringField("last_name", signup.LastName),
	)

	if err := c.Validate(signup); err != nil {
		h.LogError("signup validation failed", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	h.LogDebug("calling user service SignUp")
	newUser, err := h.userService.SignUp(c.Request().Context(), &signup)
	if err != nil {
		h.LogError("SignUp returned error", err)

		switch {
		case errors.Is(err, user.ErrUserExists):
			return echo.NewHTTPError(http.StatusConflict, "Email already exists")
		default:
			h.LogError("unexpected error during signup", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
		}
	}

	h.LogInfo("User created successfully",
		logging.UintField("user_id", newUser.ID),
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
		h.LogError("failed to bind login request", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	h.LogDebug("login attempt",
		logging.StringField("email", login.Email),
		logging.BoolField("has_password", login.Password != ""),
	)

	if err := c.Validate(login); err != nil {
		h.LogError("login validation failed", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tokenPair, err := h.userService.Login(c.Request().Context(), &login)
	if err != nil {
		h.LogError("login failed", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	// Set refresh token cookie
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    tokenPair.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(CookieExpiryMinutes * SecondsInMinute),
	})

	h.LogInfo("User logged in successfully",
		logging.StringField("email", login.Email),
	)
	return c.JSON(http.StatusOK, tokenPair)
}

// handleLogout handles user logout
// @Summary Logout user
// @Description Logout user and invalidate tokens
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} echo.HTTPError
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) handleLogout(c echo.Context) error {
	// Get refresh token from cookie
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		h.LogError("failed to get refresh token cookie", err)
		return echo.NewHTTPError(http.StatusBadRequest, "No refresh token found")
	}

	// Blacklist the refresh token
	logoutErr := h.userService.Logout(c.Request().Context(), cookie.Value)
	if logoutErr != nil {
		h.LogError("failed to logout", logoutErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to logout")
	}

	// Clear the refresh token cookie
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	h.LogInfo("User logged out successfully")
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Successfully logged out",
	})
}

// handleWebLogout handles web logout requests
func (h *AuthHandler) handleWebLogout(c echo.Context) error {
	// Get refresh token from cookie
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		h.LogError("failed to get refresh token cookie", err)
		return echo.NewHTTPError(http.StatusBadRequest, "No refresh token found")
	}

	// Blacklist the refresh token
	logoutErr := h.userService.Logout(c.Request().Context(), cookie.Value)
	if logoutErr != nil {
		h.LogError("failed to logout", logoutErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to logout")
	}

	// Clear the refresh token cookie
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	h.LogInfo("User logged out successfully via web")
	return c.Redirect(http.StatusSeeOther, "/login")
}
