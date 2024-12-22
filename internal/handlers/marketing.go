package handlers

import (
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// MarketingHandler handles marketing page requests
type MarketingHandler struct {
	logger    *zap.Logger
	templates *template.Template
}

// NewMarketingHandler creates a new marketing handler
func NewMarketingHandler(logger *zap.Logger) *MarketingHandler {
	templates := template.Must(template.ParseGlob("templates/*.html"))
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
	return c.Render(http.StatusOK, "home.html", map[string]interface{}{
		"title":       "Goforms - Simple Form Backend",
		"description": "Self-hosted form backend solution with API support",
	})
}

// Register registers the marketing routes
func (h *MarketingHandler) Register(e *echo.Echo) {
	e.GET("/", h.HomePage)
}
