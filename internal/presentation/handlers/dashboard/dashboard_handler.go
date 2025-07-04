package dashboard

import (
	"fmt"

	"github.com/goformx/goforms/internal/presentation/handlers"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
)

// DashboardHandler handles the /dashboard route
// Implements httpiface.Handler
type DashboardHandler struct {
	handlers.BaseHandler
}

// NewDashboardHandler creates a new DashboardHandler and registers the /dashboard route
func NewDashboardHandler() *DashboardHandler {
	h := &DashboardHandler{
		BaseHandler: *handlers.NewBaseHandler("dashboard"),
	}

	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/dashboard",
		Handler: h.Dashboard,
	})

	return h
}

// Dashboard handles GET /dashboard
func (h *DashboardHandler) Dashboard(ctx httpiface.Context) error {
	return fmt.Errorf("Dashboard page (placeholder)")
}
