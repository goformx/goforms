package auth

import (
	"net/http"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/handlers"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

const (
	// CookieMaxAgeMinutes is the number of minutes before a cookie expires
	CookieMaxAgeMinutes = 15
	// SecondsInMinute is the number of seconds in a minute
	SecondsInMinute = 60
)

// WebLoginHandler handles web login requests
type WebLoginHandler struct {
	*handlers.BaseHandler
	userService user.Service
}

// NewWebLoginHandler creates a new web login handler
func NewWebLoginHandler(logger logging.Logger, userService user.Service) *WebLoginHandler {
	return &WebLoginHandler{
		BaseHandler: handlers.NewBaseHandler(nil, nil, logger),
		userService: userService,
	}
}

// Register registers the web login routes
func (h *WebLoginHandler) Register(e *echo.Echo) {
	e.GET("/login", h.Login)
	e.POST("/login", h.LoginPost)
}

// Login handles the login page request
func (h *WebLoginHandler) Login(c echo.Context) error {
	// Get CSRF token from context
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = "" // Set empty string if token not found
	}

	// Create page data
	data := shared.PageData{
		Title:     "Login - GoFormX",
		CSRFToken: csrfToken,
		AssetPath: web.GetAssetPath,
	}

	// Render login page
	return pages.Login(data).Render(c.Request().Context(), c.Response().Writer)
}

// LoginPost handles the login form submission
func (h *WebLoginHandler) LoginPost(c echo.Context) error {
	// Parse form data
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Create login request
	login := &user.Login{
		Email:    email,
		Password: password,
	}

	// Attempt login
	loginResp, err := h.userService.Login(c.Request().Context(), login)
	if err != nil {
		// Get CSRF token for re-rendering the form
		csrfToken, _ := c.Get("csrf").(string)
		data := shared.PageData{
			Title:     "Login - GoFormX",
			CSRFToken: csrfToken,
			AssetPath: web.GetAssetPath,
		}

		// Re-render login page with error
		return pages.Login(data).Render(c.Request().Context(), c.Response().Writer)
	}

	// Set refresh token cookie
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    loginResp.Token.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   CookieMaxAgeMinutes * SecondsInMinute, // 15 minutes
	})

	// Redirect to dashboard on success
	return c.Redirect(http.StatusSeeOther, "/dashboard")
}
