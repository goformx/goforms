package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	amw "github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
	"github.com/jonesrussell/goforms/internal/presentation/templates/shared"
	"github.com/jonesrussell/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

var (
	ErrNoCurrentUser = errors.New("no current user found")
)

// WebHandlerOption defines a web handler option.
// This type is used to implement the functional options pattern
// for configuring the WebHandler.
type WebHandlerOption func(*WebHandler)

// WithContactService sets the contact service.
// This is a required option for the WebHandler as it needs
// the contact service to function properly.
//
// Example:
//
//	handler := NewWebHandler(logger, WithContactService(contactService))
func WithContactService(svc contact.Service) WebHandlerOption {
	return func(h *WebHandler) {
		h.contactService = svc
	}
}

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

// WebHandler handles web page requests.
// It requires a renderer, contact service, and subscription service to function properly.
// Use the functional options pattern to configure these dependencies.
//
// Dependencies:
//   - renderer: Required for rendering web pages
//   - contactService: Required for contact form functionality
//   - middlewareManager: Required for security and request processing
//   - config: Required for configuration
type WebHandler struct {
	Base
	contactService    contact.Service
	renderer          *view.Renderer
	middlewareManager *amw.Manager
	config            *config.Config
}

// NewWebHandler creates a new web handler.
// It uses the functional options pattern to configure the handler.
// The logger is required as a direct parameter, while other dependencies
// are provided through options.
//
// Example:
//
//	handler := NewWebHandler(logger,
//	    WithRenderer(renderer),
//	    WithContactService(contactService),
//	    WithConfig(config),
//	)
func NewWebHandler(logger logging.Logger, opts ...WebHandlerOption) (*WebHandler, error) {
	h := &WebHandler{
		Base: NewBase(WithLogger(logger)),
	}

	for _, opt := range opts {
		opt(h)
	}

	// Validate critical dependencies during construction
	if h.renderer == nil {
		return nil, errors.New("WebHandler initialization failed: renderer is required")
	}
	if h.contactService == nil {
		return nil, errors.New("WebHandler initialization failed: contact service is required")
	}
	if h.middlewareManager == nil {
		return nil, errors.New("WebHandler initialization failed: middleware manager is required")
	}
	if h.config == nil {
		return nil, errors.New("WebHandler initialization failed: config is required")
	}

	return h, nil
}

// Validate validates that required dependencies are set.
// Returns an error if any required dependency is missing.
//
// Required dependencies:
//   - renderer
//   - contactService
//   - middlewareManager
//   - config
func (h *WebHandler) Validate() error {
	if err := h.Base.Validate(); err != nil {
		return fmt.Errorf("WebHandler validation failed: %w", err)
	}
	if h.renderer == nil {
		return errors.New("WebHandler validation failed: renderer is required")
	}
	if h.contactService == nil {
		return errors.New("WebHandler validation failed: contact service is required")
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
	token := c.Get("csrf")
	if token == nil {
		return "", errors.New("CSRF token not found in context")
	}

	tokenStr, ok := token.(string)
	if !ok {
		return "", errors.New("invalid CSRF token type")
	}

	if tokenStr == "" {
		return "", errors.New("empty CSRF token")
	}

	return tokenStr, nil
}

// renderPage renders a page with the given title and content
func (h *WebHandler) renderPage(c echo.Context, title string, template func(shared.PageData) templ.Component) error {
	// Get CSRF token from context
	csrfToken, err := h.getCSRFToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "CSRF token validation failed")
	}

	// Get user from context if available
	var userData *user.User
	if u := c.Get("user"); u != nil {
		if userObj, ok := u.(*user.User); ok {
			userData = userObj
		}
	}

	// Create page data
	data := shared.PageData{
		Title:         title,
		CSRFToken:     csrfToken,
		User:          userData,
		IsDevelopment: h.config.App.IsDevelopment(),
	}

	// Render page
	if renderErr := template(data).Render(c.Request().Context(), c.Response().Writer); renderErr != nil {
		h.Logger.Error("failed to render page",
			logging.String("title", title),
			logging.String("path", c.Request().URL.Path),
			logging.Error(renderErr))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to render page")
	}

	return nil
}

// registerAndLogRoute registers a GET route and logs the registration
func (h *WebHandler) registerAndLogRoute(e *echo.Echo, path string, handler echo.HandlerFunc) {
	e.GET(path, handler)
	if h.config.App.IsDevelopment() {
		h.Logger.Debug("registered route",
			logging.String("method", http.MethodGet),
			logging.String("path", path))
	}
}

// registerRoutes registers all web routes
func (h *WebHandler) registerRoutes(e *echo.Echo) {
	// Web pages
	h.registerAndLogRoute(e, "/", h.handleHome)
	h.registerAndLogRoute(e, "/demo", h.handleDemo)
	h.registerAndLogRoute(e, "/signup", h.handleSignup)
	h.registerAndLogRoute(e, "/login", h.handleLogin)

	// API endpoints
	h.registerAndLogRoute(e, "/api/validation/:schema", h.handleValidationSchema)

	// Static files
	e.Static("/public", "./public")
	e.Static("/dist", h.config.Static.DistDir)
	e.File("/favicon.ico", "./public/favicon.ico")
}

// Register registers the web routes
func (h *WebHandler) Register(e *echo.Echo) {
	// Validate base dependencies
	if err := h.Validate(); err != nil {
		h.Logger.Error("failed to validate web handler", logging.Error(err))
		return
	}

	if h.config.App.IsDevelopment() {
		h.Logger.Debug("registering web routes")
	}
	h.registerRoutes(e)
	if h.config.App.IsDevelopment() {
		h.Logger.Debug("web routes registration complete")
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

// handleValidationSchema returns the validation schema for a given form
func (h *WebHandler) handleValidationSchema(c echo.Context) error {
	schemaName := c.Param("schema")
	if schemaName == "signup" {
		// Return a JSON schema for signup fields
		return c.JSON(200, map[string]any{
			"first_name": map[string]any{
				"type":    "string",
				"min":     2,
				"max":     50,
				"message": "First name must be between 2 and 50 characters",
			},
			"last_name": map[string]any{
				"type":    "string",
				"min":     2,
				"max":     50,
				"message": "Last name must be between 2 and 50 characters",
			},
			"email": map[string]any{
				"type":    "email",
				"min":     5,
				"max":     100,
				"message": "Please enter a valid email address",
			},
			"password": map[string]any{
				"type":    "password",
				"min":     8,
				"max":     100,
				"message": "Password must be at least 8 characters and contain upper, lower, number, special",
			},
			"confirm_password": map[string]any{
				"type":       "match",
				"matchField": "password",
				"message":    "Passwords don't match",
			},
		})
	}
	return c.JSON(404, map[string]string{"error": "validation schema not found"})
}
