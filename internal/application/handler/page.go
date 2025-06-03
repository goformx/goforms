package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	amw "github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/handlers"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// PageHandler handles page rendering
type PageHandler struct {
	*handlers.BaseHandler
	renderer          *view.Renderer
	middlewareManager *amw.Manager
	cfg               *config.Config
	userService       user.Service
	sessionManager    *amw.SessionManager
}

// NewPageHandler creates a new page handler
func NewPageHandler(
	baseHandler *handlers.BaseHandler,
	userService user.Service,
	sessionManager *amw.SessionManager,
	renderer *view.Renderer,
	middlewareManager *amw.Manager,
	cfg *config.Config,
) *PageHandler {
	return &PageHandler{
		BaseHandler:       baseHandler,
		userService:       userService,
		sessionManager:    sessionManager,
		renderer:          renderer,
		middlewareManager: middlewareManager,
		cfg:               cfg,
	}
}

// Register registers the page routes
func (h *PageHandler) Register(e *echo.Echo) {
	// Public pages
	e.GET("/", h.handleHome)
	e.GET("/login", h.handleLogin)
	e.GET("/signup", h.handleSignup)
	e.GET("/forgot-password", h.handleForgotPassword)
	e.GET("/contact", h.handleContact)
	e.GET("/demo", h.handleDemo)

	// Protected pages
	protected := e.Group("")
	protected.GET("/dashboard", h.handleDashboard)
	protected.GET("/profile", h.handleProfile)
	protected.GET("/settings", h.handleSettings)
	protected.GET("/forms", h.handleForms)
	protected.GET("/forms/new", h.handleNewForm)
}

// getCSRFToken retrieves the CSRF token from the context
func (h *PageHandler) getCSRFToken(c echo.Context) (string, error) {
	if token, ok := c.Get("csrf").(string); ok && token != "" {
		return token, nil
	}
	return "", errors.New("CSRF token not found or invalid")
}

// buildPageData constructs the shared page data for rendering
func (h *PageHandler) buildPageData(c echo.Context, title string) (shared.PageData, error) {
	csrfToken, err := h.getCSRFToken(c)
	if err != nil {
		h.LogError("CSRF token missing or invalid", nil)
		return shared.PageData{}, echo.NewHTTPError(http.StatusForbidden, "CSRF token validation failed")
	}

	var data shared.PageData
	userData, userOk := c.Get("user").(*user.User)
	if userOk {
		h.LogDebug("User found in context", "email", userData.Email, "route", c.Path())
		data.User = userData
	} else {
		h.LogError("No user found in context", nil)
	}

	data.Title = title
	data.CSRFToken = csrfToken
	data.IsDevelopment = h.cfg.App.IsDevelopment()
	data.AssetPath = web.GetAssetPath

	// Avoid logging the entire struct due to function fields
	userEmail := "<nil>"
	if data.User != nil {
		userEmail = data.User.Email
	}
	assetPathSet := data.AssetPath != nil
	h.LogDebug(
		"PageData built",
		"title", data.Title,
		"user_email", userEmail,
		"is_dev", data.IsDevelopment,
		"asset_path_set", assetPathSet,
		"route", c.Path(),
	)

	return data, nil
}

// renderPage renders a page with the given title and content
func (h *PageHandler) renderPage(c echo.Context, title string, template func(shared.PageData) templ.Component) error {
	data, err := h.buildPageData(c, title)
	if err != nil {
		return err
	}

	if renderErr := template(data).Render(c.Request().Context(), c.Response().Writer); renderErr != nil {
		return fmt.Errorf("failed to render template: %w", renderErr)
	}

	return nil
}

// handleHome renders the home page
func (h *PageHandler) handleHome(c echo.Context) error {
	return h.renderPage(c, "Home", pages.Home)
}

// handleDemo renders the demo page
func (h *PageHandler) handleDemo(c echo.Context) error {
	return h.renderPage(c, "Demo", pages.Demo)
}

// handleSignup renders the signup page
func (h *PageHandler) handleSignup(c echo.Context) error {
	return h.renderPage(c, "Sign Up", pages.Signup)
}

// handleLogin renders the login page
func (h *PageHandler) handleLogin(c echo.Context) error {
	return h.renderPage(c, "Login", pages.Login)
}

// handleForgotPassword renders the forgot password page
func (h *PageHandler) handleForgotPassword(c echo.Context) error {
	return h.renderPage(c, "Forgot Password", pages.ForgotPassword)
}

// handleContact renders the contact page
func (h *PageHandler) handleContact(c echo.Context) error {
	return h.renderPage(c, "Contact", pages.Contact)
}

// handleDashboard renders the dashboard page
func (h *PageHandler) handleDashboard(c echo.Context) error {
	data, err := h.buildPageData(c, "Dashboard")
	if err != nil {
		return err
	}
	data.Content = pages.DashboardContent(data)
	return pages.Dashboard(data).Render(c.Request().Context(), c.Response().Writer)
}

// handleProfile renders the profile page
func (h *PageHandler) handleProfile(c echo.Context) error {
	return h.renderPage(c, "Profile", pages.Profile)
}

// handleSettings renders the settings page
func (h *PageHandler) handleSettings(c echo.Context) error {
	return h.renderPage(c, "Settings", pages.Settings)
}

// handleForms renders the forms page
func (h *PageHandler) handleForms(c echo.Context) error {
	return h.renderPage(c, "Forms", pages.Forms)
}

// handleNewForm renders the new form page
func (h *PageHandler) handleNewForm(c echo.Context) error {
	return h.renderPage(c, "New Form", pages.NewForm)
}
