package handlers

import (
	"net/http"

	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/web"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
	"github.com/jonesrussell/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// DashboardHandler handles dashboard-related HTTP requests
type DashboardHandler struct {
	formService form.Service
	logger      logging.Logger
	Base        *BaseHandler
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(
	formService form.Service,
	logger logging.Logger,
	base *BaseHandler,
) *DashboardHandler {
	return &DashboardHandler{
		formService: formService,
		logger:      logger,
		Base:        base,
	}
}

// Register sets up the dashboard routes
func (h *DashboardHandler) Register(e *echo.Echo) {
	dashboard := e.Group("/dashboard")
	dashboard.Use(h.Base.authMiddleware.RequireAuth)

	dashboard.GET("", h.ShowDashboard)
}

// ShowDashboard displays the user's dashboard
func (h *DashboardHandler) ShowDashboard(c echo.Context) error {
	currentUser, err := h.Base.getAuthenticatedUser(c)
	if err != nil {
		return err
	}

	forms, err := h.formService.GetUserForms(currentUser.ID)
	if err != nil {
		return h.Base.handleError(err, http.StatusInternalServerError, "Failed to fetch forms")
	}

	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = ""
	}

	data := shared.PageData{
		Title:     "Dashboard - GoForms",
		User:      currentUser,
		Forms:     forms,
		CSRFToken: csrfToken,
		AssetPath: web.GetAssetPath,
	}

	data.Content = pages.DashboardContent(data)
	return pages.Dashboard(data).Render(c.Request().Context(), c.Response().Writer)
}
