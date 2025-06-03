package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/a-h/templ"
	amw "github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/presentation/handlers"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	// CookieMaxAgeMinutes is the number of minutes before a cookie expires
	CookieMaxAgeMinutes = 15
)

var (
	// ErrNoCurrentUser is returned when no user is found in the current context
	ErrNoCurrentUser = errors.New("no current user found")
)

// WebHandlerOption defines a web handler option.
// This type is used to implement the functional options pattern
// for configuring the WebHandler.
type WebHandlerOption func(*WebHandler)

// WithRenderer sets the view renderer.
// This is a required option for the WebHandler as it needs
// the renderer to display web pages.
//
// Example:
//
//	handler := NewWebHandler(logger, WithRenderer(renderer))
func WithRenderer(renderer *view.Renderer) WebHandlerOption {
	return func(h *WebHandler) {
		h.renderer = renderer
	}
}

// WithMiddlewareManager sets the middleware manager for the web handler.
func WithMiddlewareManager(manager *amw.Manager) WebHandlerOption {
	return func(h *WebHandler) {
		h.middlewareManager = manager
	}
}

// WithConfig sets the config for the web handler.
func WithConfig(cfg *config.Config) WebHandlerOption {
	return func(h *WebHandler) {
		h.config = cfg
	}
}

// WebHandler handles web requests
type WebHandler struct {
	*handlers.BaseHandler
	renderer          *view.Renderer
	middlewareManager *amw.Manager
	config            *config.Config
	userService       user.Service
	sessionManager    *amw.SessionManager
}

// NewWebHandler creates a new web handler
func NewWebHandler(
	baseHandler *handlers.BaseHandler,
	userService user.Service,
	sessionManager *amw.SessionManager,
) *WebHandler {
	return &WebHandler{
		BaseHandler:    baseHandler,
		userService:    userService,
		sessionManager: sessionManager,
	}
}

// Validate validates that required dependencies are set.
// Returns an error if any required dependency is missing.
//
// Required dependencies:
//   - renderer
//   - middlewareManager
//   - config
func (h *WebHandler) Validate() error {
	if err := h.BaseHandler.Validate(); err != nil {
		return fmt.Errorf("WebHandler validation failed: %w", err)
	}
	if h.renderer == nil {
		return errors.New("WebHandler validation failed: renderer is required")
	}
	if h.middlewareManager == nil {
		return errors.New("WebHandler validation failed: middleware manager is required")
	}
	if h.config == nil {
		return errors.New("WebHandler validation failed: config is required")
	}
	return nil
}

// getCSRFToken retrieves the CSRF token from the context
func (h *WebHandler) getCSRFToken(c echo.Context) (string, error) {
	if token, ok := c.Get("csrf").(string); ok && token != "" {
		return token, nil
	}
	return "", errors.New("CSRF token not found or invalid")
}

// buildPageData constructs the shared page data for rendering
func (h *WebHandler) buildPageData(c echo.Context, title string) (shared.PageData, error) {
	csrfToken, err := h.getCSRFToken(c)
	if err != nil {
		return shared.PageData{}, echo.NewHTTPError(http.StatusForbidden, "CSRF token validation failed")
	}

	var data shared.PageData
	if userData, ok := c.Get("user").(*user.User); ok {
		data.User = userData
	}

	data.Title = title
	data.CSRFToken = csrfToken
	data.IsDevelopment = h.config.App.IsDevelopment()

	return data, nil
}

// renderPage renders a page with the given title and content
func (h *WebHandler) renderPage(c echo.Context, title string, template func(shared.PageData) templ.Component) error {
	data, err := h.buildPageData(c, title)
	if err != nil {
		return err
	}

	if renderErr := template(data).Render(c.Request().Context(), c.Response().Writer); renderErr != nil {
		return fmt.Errorf("failed to render template: %w", renderErr)
	}

	return nil
}

// registerRoutes registers the web routes
func (h *WebHandler) registerRoutes(e *echo.Echo) {
	// Static files
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "public",
		Browse: false,
	}))

	// Public routes
	e.GET("/", h.handleHome)
	e.GET("/login", h.handleLogin)
	e.GET("/signup", h.handleSignup)
	e.GET("/forgot-password", h.handleForgotPassword)
	e.GET("/contact", h.handleContact)
	e.GET("/demo", h.handleDemo)

	// Auth routes
	e.POST("/login", h.handleLoginPost)
	e.POST("/signup", h.handleSignupPost)
	e.POST("/logout", h.handleLogout)

	// Protected routes
	protected := e.Group("")
	protected.Use(h.sessionManager.SessionMiddleware())
	protected.GET("/dashboard", h.handleDashboard)
	protected.GET("/profile", h.handleProfile)
	protected.GET("/settings", h.handleSettings)
	protected.GET("/forms", h.handleForms)
	protected.GET("/forms/new", h.handleNewForm)
}

// validateDependencies validates required dependencies for the handler
func (h *WebHandler) validateDependencies() {
	if err := h.Validate(); err != nil {
		h.LogError("failed to validate web handler", err)
		panic(fmt.Sprintf("failed to validate web handler: %v", err))
	}
}

// Register registers the web routes
func (h *WebHandler) Register(e *echo.Echo) {
	h.validateDependencies()
	if h.config.App.IsDevelopment() {
		h.LogDebug("registering web routes")
	}

	h.middlewareManager.Setup(e) // Ensure middleware is loaded properly
	h.registerRoutes(e)

	if h.config.App.IsDevelopment() {
		h.LogDebug("web routes registration complete")
	}
}

// handleHome renders the home page
func (h *WebHandler) handleHome(c echo.Context) error {
	return h.renderPage(c, "Home", pages.Home)
}

// handleDemo renders the demo page
func (h *WebHandler) handleDemo(c echo.Context) error {
	return h.renderPage(c, "Demo", pages.Demo)
}

// handleSignup renders the signup page
func (h *WebHandler) handleSignup(c echo.Context) error {
	return h.renderPage(c, "Sign Up", pages.Signup)
}

// handleLogin renders the login page
func (h *WebHandler) handleLogin(c echo.Context) error {
	return h.renderPage(c, "Login", pages.Login)
}

// handleLoginPost handles the login form submission
func (h *WebHandler) handleLoginPost(c echo.Context) error {
	// Parse form data
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Create login request
	login := &user.Login{
		Email:    email,
		Password: password,
	}

	// Attempt login
	loginResp, loginErr := h.userService.Login(c.Request().Context(), login)
	if loginErr != nil {
		if errors.Is(loginErr, user.ErrInvalidCredentials) {
			data, buildErr := h.buildPageData(c, "Login")
			if buildErr != nil {
				return buildErr
			}
			return c.Render(
				http.StatusUnauthorized,
				"login",
				pages.LoginWithError(data, "Invalid email or password"),
			)
		}
		h.LogError("failed to login", loginErr)
		data, buildErr := h.buildPageData(c, "Login")
		if buildErr != nil {
			return buildErr
		}
		return c.Render(
			http.StatusInternalServerError,
			"login",
			pages.LoginWithError(data, "An error occurred. Please try again."),
		)
	}

	// Create session
	sessionID, sessionErr := h.sessionManager.CreateSession(
		loginResp.User.ID, loginResp.User.Email, loginResp.User.Role,
	)
	if sessionErr != nil {
		h.LogError("failed to create session", sessionErr)
		data, buildErr := h.buildPageData(c, "Login")
		if buildErr != nil {
			return buildErr
		}
		return c.Render(
			http.StatusInternalServerError,
			"login",
			pages.LoginWithError(data, "An error occurred. Please try again."),
		)
	}

	// Set session cookie
	h.sessionManager.SetSessionCookie(c, sessionID)

	// Set refresh token cookie
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    loginResp.Token.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(CookieMaxAgeMinutes * time.Minute.Seconds()), // 15 minutes
	})

	// Redirect to dashboard
	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// handleSignupPost handles the signup form submission
func (h *WebHandler) handleSignupPost(c echo.Context) error {
	// Parse form data
	email := c.FormValue("email")
	password := c.FormValue("password")
	confirmPassword := c.FormValue("confirm_password")
	firstName := c.FormValue("first_name")
	lastName := c.FormValue("last_name")

	// Validate password confirmation
	if password != confirmPassword {
		data, buildErr := h.buildPageData(c, "Sign Up")
		if buildErr != nil {
			return buildErr
		}
		return c.Render(
			http.StatusBadRequest,
			"signup",
			pages.SignupWithError(data, "Passwords do not match"),
		)
	}

	// Create signup request
	signup := &user.Signup{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	}

	// Attempt signup
	u, signupErr := h.userService.SignUp(c.Request().Context(), signup)
	if signupErr != nil {
		if errors.Is(signupErr, user.ErrUserExists) || errors.Is(signupErr, user.ErrEmailAlreadyExists) {
			data, buildErr := h.buildPageData(c, "Sign Up")
			if buildErr != nil {
				return buildErr
			}
			return c.Render(
				http.StatusBadRequest,
				"signup",
				pages.SignupWithError(data, "Email already exists"),
			)
		}
		h.LogError("failed to signup", signupErr)
		data, buildErr := h.buildPageData(c, "Sign Up")
		if buildErr != nil {
			return buildErr
		}
		return c.Render(
			http.StatusInternalServerError,
			"signup",
			pages.SignupWithError(data, "An error occurred. Please try again."),
		)
	}

	// Create session
	sessionID, sessionErr := h.sessionManager.CreateSession(u.ID, u.Email, u.Role)
	if sessionErr != nil {
		h.LogError("failed to create session", sessionErr)
		data, buildErr := h.buildPageData(c, "Sign Up")
		if buildErr != nil {
			return buildErr
		}
		return c.Render(
			http.StatusInternalServerError,
			"signup",
			pages.SignupWithError(data, "An error occurred. Please try again."),
		)
	}

	// Set session cookie
	h.sessionManager.SetSessionCookie(c, sessionID)

	// Redirect to dashboard
	return c.Redirect(http.StatusSeeOther, "/dashboard")
}

// handleLogout handles the logout request
func (h *WebHandler) handleLogout(c echo.Context) error {
	// Get session ID from cookie
	cookie, err := c.Cookie("session_id")
	if err == nil {
		// Delete session
		h.sessionManager.DeleteSession(cookie.Value)
	}

	// Clear session cookie
	h.sessionManager.ClearSessionCookie(c)

	// Redirect to home
	return c.Redirect(http.StatusSeeOther, "/")
}

// handleForgotPassword renders the forgot password page
func (h *WebHandler) handleForgotPassword(c echo.Context) error {
	return h.renderPage(c, "Forgot Password", pages.ForgotPassword)
}

// handleContact renders the contact page
func (h *WebHandler) handleContact(c echo.Context) error {
	return h.renderPage(c, "Contact", pages.Contact)
}

// handleDashboard renders the dashboard page
func (h *WebHandler) handleDashboard(c echo.Context) error {
	return h.renderPage(c, "Dashboard", pages.Dashboard)
}

// handleProfile renders the profile page
func (h *WebHandler) handleProfile(c echo.Context) error {
	return h.renderPage(c, "Profile", pages.Profile)
}

// handleSettings renders the settings page
func (h *WebHandler) handleSettings(c echo.Context) error {
	return h.renderPage(c, "Settings", pages.Settings)
}

// handleForms renders the forms page
func (h *WebHandler) handleForms(c echo.Context) error {
	return h.renderPage(c, "Forms", pages.Forms)
}

// handleNewForm renders the new form page
func (h *WebHandler) handleNewForm(c echo.Context) error {
	return h.renderPage(c, "New Form", pages.NewForm)
}
