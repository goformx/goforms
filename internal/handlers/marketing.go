package handlers

import (
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Template is a custom renderer for Echo
type Template struct {
	templates *template.Template
}

// Render implements echo.Renderer
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Remove .html suffix if present
	templateName := name
	if len(name) > 5 && name[len(name)-5:] == ".html" {
		templateName = name[:len(name)-5]
	}
	return t.templates.ExecuteTemplate(w, templateName, data)
}

// MarketingHandler handles marketing page requests
type MarketingHandler struct {
	logger    *zap.Logger
	templates *template.Template
}

// NewMarketingHandler creates a new marketing handler
func NewMarketingHandler(logger *zap.Logger, templates *template.Template) *MarketingHandler {
	return &MarketingHandler{
		logger:    logger,
		templates: templates,
	}
}

// HomePage renders the landing page
// @Summary Serves the landing page
// @Description Returns the main marketing page for Goforms
// @Tags marketing
// @Produce html
// @Success 200 {string} html
// @Router / [get]
func (h *MarketingHandler) HomePage(c echo.Context) error {
	return c.Render(http.StatusOK, "home", nil)
}

// ContactPage renders the contact form demo page
// @Summary Serves the contact form demo page
// @Description Returns the contact form demo page
// @Tags marketing
// @Produce html
// @Success 200 {string} html
// @Router /contact [get]
func (h *MarketingHandler) ContactPage(c echo.Context) error {
	return c.Render(http.StatusOK, "contact", nil)
}

// Register registers the marketing routes and sets up the template renderer
func (h *MarketingHandler) Register(e *echo.Echo) {
	// Set up the template renderer
	e.Renderer = &Template{templates: h.templates}

	// Register routes
	e.GET("/", h.HomePage)
	e.GET("/contact", h.ContactPage)

	// Serve static files
	e.Static("/static", "static")
}
