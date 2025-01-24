package v1

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/core/user"
	"github.com/jonesrussell/goforms/internal/models"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	userService user.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService user.Service) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// RegisterRoutes registers the authentication routes
func (h *AuthHandler) RegisterRoutes(e *echo.Echo) {
	auth := e.Group("/api/v1/auth")
	auth.POST("/signup", h.SignUp)
	auth.POST("/login", h.Login)
	auth.POST("/logout", h.Logout)
	auth.POST("/refresh", h.RefreshToken)
}

// SignUp godoc
// @Summary Register a new user
// @Description Register a new user with the provided information
// @Tags auth
// @Accept json
// @Produce json
// @Param signup body models.UserSignup true "User signup information"
// @Success 201 {object} models.User
// @Failure 400 {object} echo.HTTPError
// @Failure 409 {object} echo.HTTPError
// @Router /api/v1/auth/signup [post]
func (h *AuthHandler) SignUp(c echo.Context) error {
	var signup models.UserSignup
	if err := c.Bind(&signup); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(signup); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	newUser, err := h.userService.SignUp(c.Request().Context(), &signup)
	if err != nil {
		if err == user.ErrEmailAlreadyExists {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
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
// @Param login body models.UserLogin true "User login credentials"
// @Success 200 {object} user.TokenPair
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var login models.UserLogin
	if err := c.Bind(&login); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(login); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tokens, err := h.userService.Login(c.Request().Context(), &login)
	if err != nil {
		if err == user.ErrInvalidCredentials {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
		}
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
		return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
	}

	token := parts[1]
	if err := h.userService.Logout(c.Request().Context(), token); err != nil {
		if err == user.ErrInvalidToken {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}
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
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	tokens, err := h.userService.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		switch err {
		case user.ErrInvalidToken:
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid refresh token")
		case user.ErrTokenBlacklisted:
			return echo.NewHTTPError(http.StatusUnauthorized, "refresh token has been invalidated")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to refresh token")
		}
	}

	return c.JSON(http.StatusOK, tokens)
}
