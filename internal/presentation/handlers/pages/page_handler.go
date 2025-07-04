package pages

import (
	"github.com/goformx/goforms/internal/presentation/handlers"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
)

// PageHandler handles public page routes (e.g., home page)
type PageHandler struct {
	handlers.BaseHandler
}

// NewPageHandler creates a new PageHandler with a single home route.
func NewPageHandler() *PageHandler {
	h := &PageHandler{
		BaseHandler: *handlers.NewBaseHandler("pages"),
	}

	// Register the home page route
	h.AddRoute(httpiface.Route{
		Method:      "GET",
		Path:        "/",
		Handler:     h.handleHome,
		Name:        "home",
		Description: "Home page",
	})

	return h
}

// handleHome is the handler for GET /
func (h *PageHandler) handleHome(ctx httpiface.Context) error {
	// For now, just a placeholder (no response logic)
	return nil
}
