package handler

import (
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
	"github.com/jonesrussell/goforms/internal/presentation/view"
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

// WebHandler handles web page requests.
// It requires a renderer, contact service, and subscription service to function properly.
// Use the functional options pattern to configure these dependencies.
//
// Dependencies:
//   - renderer: Required for rendering web pages
//   - contactService: Required for contact form functionality
//   - subscriptionService: Required for newsletter subscription functionality
type WebHandler struct {
	Base
	contactService      contact.Service
	subscriptionService subscription.Service
	renderer            *view.Renderer
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
func NewWebHandler(logger logging.Logger, opts ...WebHandlerOption) *WebHandler {
	h := &WebHandler{
		Base: NewBase(WithLogger(logger)),
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
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
		return fmt.Errorf("renderer is required")
	}
	if h.contactService == nil {
		return fmt.Errorf("contact service is required")
	}
	if h.subscriptionService == nil {
		return fmt.Errorf("subscription service is required")
	}
	return nil
}

// Register registers the web routes.
// This method sets up all web page routes and static file serving.
// It validates that all required dependencies are available before
// registering any routes.
func (h *WebHandler) Register(e *echo.Echo) {
	if err := h.Validate(); err != nil {
		h.Logger.Error("failed to validate handler", logging.Error(err))
		return
	}

	h.Logger.Debug("registering web routes")

	// Web pages
	e.GET("/", h.handleHome)
	h.Logger.Debug("registered route", logging.String("method", "GET"), logging.String("path", "/"))

	e.GET("/newsletter", h.handleNewsletter)
	h.Logger.Debug("registered route", logging.String("method", "GET"), logging.String("path", "/newsletter"))

	e.GET("/signup", h.handleSignup)
	h.Logger.Debug("registered route", logging.String("method", "GET"), logging.String("path", "/signup"))

	e.GET("/login", h.handleLogin)
	h.Logger.Debug("registered route", logging.String("method", "GET"), logging.String("path", "/login"))

	// Static files - Note: paths must be relative to the project root
	e.Static("/static", "./static")
	h.Logger.Debug("registered static directory", logging.String("path", "/static"), logging.String("root", "./static"))

	e.File("/favicon.ico", "./static/favicon.ico")
	h.Logger.Debug("registered favicon", logging.String("path", "/favicon.ico"))

	h.Logger.Debug("web routes registration complete")
}

// handleHome renders the home page
func (h *WebHandler) handleHome(c echo.Context) error {
	h.Logger.Debug("handling home page request",
		logging.String("path", c.Path()),
		logging.String("method", c.Request().Method),
	)
	if err := h.renderer.Render(c, pages.Home()); err != nil {
		h.Logger.Error("failed to render home page",
			logging.String("path", c.Path()),
			logging.Error(err),
		)
		return fmt.Errorf("failed to render home page: %w", err)
	}
	return nil
}

// handleNewsletter renders the newsletter page
func (h *WebHandler) handleNewsletter(c echo.Context) error {
	h.Logger.Debug("handling newsletter page request",
		logging.String("path", c.Path()),
		logging.String("method", c.Request().Method),
	)
	if err := h.renderer.Render(c, pages.Newsletter()); err != nil {
		h.Logger.Error("failed to render newsletter page",
			logging.String("path", c.Path()),
			logging.Error(err),
		)
		return fmt.Errorf("failed to render newsletter page: %w", err)
	}
	return nil
}

// handleSignup renders the signup page
func (h *WebHandler) handleSignup(c echo.Context) error {
	h.Logger.Debug("handling signup page request",
		logging.String("path", c.Path()),
		logging.String("method", c.Request().Method),
	)
	if err := h.renderer.Render(c, pages.Signup()); err != nil {
		h.Logger.Error("failed to render signup page",
			logging.String("path", c.Path()),
			logging.Error(err),
		)
		return fmt.Errorf("failed to render signup page: %w", err)
	}
	return nil
}

// handleLogin renders the login page
func (h *WebHandler) handleLogin(c echo.Context) error {
	h.Logger.Debug("handling login page request",
		logging.String("path", c.Path()),
		logging.String("method", c.Request().Method),
	)
	if err := h.renderer.Render(c, pages.Login()); err != nil {
		h.Logger.Error("failed to render login page",
			logging.String("path", c.Path()),
			logging.Error(err),
		)
		return fmt.Errorf("failed to render login page: %w", err)
	}
	return nil
}
