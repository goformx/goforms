package v1

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	userService user.Service
	logger      logging.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService user.Service, log logging.Logger) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		logger:      log,
	}
}

// Register registers the authentication routes
func (h *AuthHandler) Register(e *echo.Echo) {
	g := e.Group("/api/v1/auth")
	g.POST("/signup", h.SignUp)
	g.POST("/login", h.Login)
	g.POST("/logout", h.Logout)
	g.POST("/refresh", h.RefreshToken)
}

// SignUp godoc
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
func (h *AuthHandler) SignUp(c echo.Context) error {
	var signup user.Signup
	if err := c.Bind(&signup); err != nil {
		h.logger.Error("failed to bind signup request", logging.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(signup); err != nil {
		h.logger.Error("failed to validate signup request", logging.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	newUser, err := h.userService.SignUp(c.Request().Context(), &signup)
	if err != nil {
		if err == user.ErrEmailAlreadyExists {
			h.logger.Error("email already exists", logging.Error(err))
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		h.logger.Error("failed to create user", logging.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
	}

	return c.JSON(http.StatusCreated, newUser)
}

// Login godoc
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
func (h *AuthHandler) Login(c echo.Context) error {
	var login user.Login
	if err := c.Bind(&login); err != nil {
		h.logger.Error("failed to bind login request", logging.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(login); err != nil {
		h.logger.Error("failed to validate login request", logging.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tokens, err := h.userService.Login(c.Request().Context(), &login)
	if err != nil {
		if err == user.ErrInvalidCredentials {
			h.logger.Error("invalid credentials", logging.Error(err))
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
		}
		h.logger.Error("failed to authenticate user", logging.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to authenticate user")
	}

	return c.JSON(http.StatusOK, tokens)
}

// Logout godoc
// @Summary Logout user
// @Description Invalidate the current access token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} echo.HTTPError
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		h.logger.Error("missing authorization header")
		return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		h.logger.Error("invalid authorization header format")
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
	}

	token := parts[1]
	if err := h.userService.Logout(c.Request().Context(), token); err != nil {
		if err == user.ErrInvalidToken {
			h.logger.Error("invalid token", logging.Error(err))
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}
		h.logger.Error("failed to logout", logging.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to logout")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "successfully logged out",
	})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh_token body string true "Refresh token"
// @Success 200 {object} user.TokenPair
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		h.logger.Error("failed to bind refresh token request", logging.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(req); err != nil {
		h.logger.Error("failed to validate refresh token request", logging.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tokens, err := h.userService.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		switch err {
		case user.ErrInvalidToken:
			h.logger.Error("invalid refresh token", logging.Error(err))
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid refresh token")
		case user.ErrTokenBlacklisted:
			h.logger.Error("refresh token has been invalidated", logging.Error(err))
			return echo.NewHTTPError(http.StatusUnauthorized, "refresh token has been invalidated")
		default:
			h.logger.Error("failed to refresh token", logging.Error(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to refresh token")
		}
	}

	return c.JSON(http.StatusOK, tokens)
}
