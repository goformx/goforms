package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	amw "github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/application/validation"
	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/domain/user"
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

// WithWebSubscriptionService sets the subscription service for the web handler.
// This is a required option for the WebHandler as it needs the subscription
// service to handle newsletter signups.
//
// Example:
//
//	handler := NewWebHandler(logger, WithWebSubscriptionService(subscriptionService))
func WithWebSubscriptionService(svc subscription.Service) WebHandlerOption {
	return func(h *WebHandler) {
		h.subscriptionService = svc
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

// WithWebDebug sets the debug flag for the web handler.
// When enabled, additional debug features like client-side debugging will be enabled.
func WithWebDebug(debug bool) WebHandlerOption {
	return func(h *WebHandler) {
		h.Debug = debug
	}
}

// WebHandler handles web page requests.
// It requires a renderer, contact service, and subscription service to function properly.
// Use the functional options pattern to configure these dependencies.
//
// Dependencies:
//   - renderer: Required for rendering web pages
//   - contactService: Required for contact form functionality
//   - subscriptionService: Required for demo form submission functionality
type WebHandler struct {
	Base
	contactService      contact.Service
	subscriptionService subscription.Service
	renderer            *view.Renderer
	Debug               bool
	userService         user.Service
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
//	    WithWebSubscriptionService(subscriptionService),
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
		return nil, errors.New("renderer is required")
	}
	if h.contactService == nil {
		return nil, errors.New("contact service is required")
	}
	if h.subscriptionService == nil {
		return nil, errors.New("subscription service is required")
	}

	return h, nil
}

// Validate validates that required dependencies are set.
// Returns an error if any required dependency is missing.
//
// Required dependencies:
//   - renderer
//   - contactService
//   - subscriptionService
func (h *WebHandler) Validate() error {
	if err := h.Base.Validate(); err != nil {
		return err
	}
	if h.renderer == nil {
		return errors.New("renderer is required")
	}
	if h.contactService == nil {
		return errors.New("contact service is required")
	}
	if h.subscriptionService == nil {
		return errors.New("subscription service is required")
	}
	return nil
}

// getCSRFToken retrieves the CSRF token from the context
func (h *WebHandler) getCSRFToken(c echo.Context) string {
	h.Logger.Debug("attempting to get CSRF token from context",
		logging.String("path", c.Request().URL.Path),
		logging.String("method", c.Request().Method),
		logging.String("content_type", c.Request().Header.Get("Content-Type")))

	// First try to get the token from the context key (set by middleware)
	token := c.Get(amw.CSRFContextKey)
	if token == nil {
		h.Logger.Debug("CSRF token not found in context", 
			logging.String("path", c.Request().URL.Path),
			logging.String("method", c.Request().Method),
			logging.String("context_keys", fmt.Sprintf("%v", c.Get(""))))
		return ""
	}

	tokenStr, ok := token.(string)
	if !ok {
		h.Logger.Error("CSRF token is not a string", 
			logging.String("path", c.Request().URL.Path),
			logging.String("method", c.Request().Method),
			logging.String("token_type", fmt.Sprintf("%T", token)))
		return ""
	}

	if tokenStr == "" {
		h.Logger.Debug("CSRF token is empty string", 
			logging.String("path", c.Request().URL.Path),
			logging.String("method", c.Request().Method))
		return ""
	}

	h.Logger.Debug("CSRF token found", 
		logging.String("path", c.Request().URL.Path),
		logging.String("method", c.Request().Method),
		logging.String("token_prefix", tokenStr[:8]),
		logging.String("token_length", fmt.Sprintf("%d", len(tokenStr))))
	return tokenStr
}

// getCurrentUser retrieves the current user from the context
func getCurrentUser(c echo.Context, userService user.Service) (*user.User, error) {
	if userID, exists := c.Get("user_id").(uint); exists {
		u, err := userService.GetUserByID(c.Request().Context(), userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
		return u, nil
	}
	return nil, ErrNoCurrentUser
}

// renderPage renders a page with the given template and data
func (h *WebHandler) renderPage(
	c echo.Context,
	title string,
	template func(shared.PageData) templ.Component,
) error {
	h.Logger.Debug("starting page render",
		logging.String("path", c.Request().URL.Path),
		logging.String("method", c.Request().Method),
		logging.String("title", title))

	token := h.getCSRFToken(c)
	if token == "" {
		h.Logger.Debug("no CSRF token found in context", 
			logging.String("path", c.Request().URL.Path),
			logging.String("method", c.Request().Method))
	}

	currentUser, err := getCurrentUser(c, h.userService)
	if err != nil {
		h.Logger.Debug("no current user found", 
			logging.Error(err),
			logging.String("path", c.Request().URL.Path))
	}

	data := shared.PageData{
		Title:     title,
		CSRFToken: token,
		User:      currentUser,
	}

	h.Logger.Debug("preparing page data",
		logging.String("path", c.Request().URL.Path),
		logging.String("title", title),
		logging.Bool("has_csrf_token", token != ""),
		logging.Bool("has_user", currentUser != nil))

	if token != "" {
		h.Logger.Debug("rendering page with CSRF token", 
			logging.String("path", c.Request().URL.Path),
			logging.String("token_prefix", token[:8]))
	} else {
		h.Logger.Debug("rendering page without CSRF token", 
			logging.String("path", c.Request().URL.Path))
	}

	return h.renderer.Render(c, template(data))
}

// logRoute logs route registration
func (h *WebHandler) logRoute(method, path string) {
	h.Logger.Debug("registered route",
		logging.String("method", method),
		logging.String("path", path),
	)
}

// registerRoute registers a route with logging
func (h *WebHandler) registerRoute(e *echo.Echo, path string, handler echo.HandlerFunc) {
	e.GET(path, handler)
}

// registerRoutes registers all web routes
func (h *WebHandler) registerRoutes(e *echo.Echo) {
	// Web pages
	h.registerRoute(e, "/", h.handleHome)
	h.registerRoute(e, "/demo", h.handleDemo)
	h.registerRoute(e, "/signup", h.handleSignup)
	h.registerRoute(e, "/login", h.handleLogin)

	// Validation endpoints
	h.registerRoute(e, "/api/validation/:schema", h.handleValidationSchema)

	// Static files
	e.Static("/static", "./static")
	h.logRoute("GET", "/static/*")

	e.Static("/static/dist", "./static/dist")
	h.logRoute("GET", "/static/dist/*")

	e.File("/favicon.ico", "./static/favicon.ico")
	h.logRoute("GET", "/favicon.ico")
}

// Register registers the web routes
func (h *WebHandler) Register(e *echo.Echo) {
	// Validate base dependencies
	if err := h.Base.Validate(); err != nil {
		h.Logger.Error("failed to validate base handler", logging.Error(err))
		return
	}

	h.Logger.Debug("registering web routes")
	h.registerRoutes(e)
	h.Logger.Debug("web routes registration complete")
}

// handleHome renders the home page
func (h *WebHandler) handleHome(c echo.Context) error {
	h.Logger.Debug("handling home page request",
		logging.String("path", c.Request().URL.Path),
		logging.String("method", c.Request().Method),
		logging.String("user_agent", c.Request().UserAgent()))

	return h.renderPage(c, "Home", pages.Home)
}

// handleDemo renders the demo page
func (h *WebHandler) handleDemo(c echo.Context) error {
	h.Logger.Debug("handling demo page request",
		logging.String("path", c.Request().URL.Path),
		logging.String("method", c.Request().Method),
		logging.String("user_agent", c.Request().UserAgent()))

	return h.renderPage(c, "Demo", pages.Demo)
}

// handleSignup renders the signup page
func (h *WebHandler) handleSignup(c echo.Context) error {
	h.Logger.Debug("handling signup page request",
		logging.String("path", c.Request().URL.Path),
		logging.String("method", c.Request().Method),
		logging.String("user_agent", c.Request().UserAgent()))

	return h.renderPage(c, "Sign Up", pages.Signup)
}

// handleLogin renders the login page
func (h *WebHandler) handleLogin(c echo.Context) error {
	h.Logger.Debug("handling login page request",
		logging.String("path", c.Request().URL.Path),
		logging.String("method", c.Request().Method),
		logging.String("user_agent", c.Request().UserAgent()))

	return h.renderPage(c, "Login", pages.Login)
}

// handleValidationSchema returns the validation schema for a given form
func (h *WebHandler) handleValidationSchema(c echo.Context) error {
	schemaName := c.Param("schema")
	schema, err := validation.GetSchema(schemaName)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, schema)
}
