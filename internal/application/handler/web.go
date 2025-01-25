package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
	"github.com/jonesrussell/goforms/internal/presentation/view"
)

// WebHandler handles web page requests
type WebHandler struct {
	Base
	contactService contact.Service
	renderer       *view.Renderer
}

// NewWebHandler creates a new web handler
func NewWebHandler(logger logging.Logger, contactService contact.Service, renderer *view.Renderer) *WebHandler {
	return &WebHandler{
		Base:           Base{Logger: logger},
		contactService: contactService,
		renderer:       renderer,
	}
}

// Register registers the web routes
func (h *WebHandler) Register(e *echo.Echo) {
	// Web pages
	e.GET("/", h.handleHome)
	e.GET("/contact", h.handleContact)
	e.GET("/signup", h.handleSignup)
	e.GET("/login", h.handleLogin)

	// Static files
	e.Static("/static", "static")
	e.File("/favicon.ico", "static/favicon.ico")
}

// handleHome renders the home page
func (h *WebHandler) handleHome(c echo.Context) error {
	return h.renderer.Render(c, pages.Home())
}

// handleContact renders the contact page
func (h *WebHandler) handleContact(c echo.Context) error {
	return h.renderer.Render(c, pages.Contact())
}

// handleSignup renders the signup page
func (h *WebHandler) handleSignup(c echo.Context) error {
	return h.renderer.Render(c, pages.Signup())
}

// handleLogin renders the login page
func (h *WebHandler) handleLogin(c echo.Context) error {
	return h.renderer.Render(c, pages.Login())
}
