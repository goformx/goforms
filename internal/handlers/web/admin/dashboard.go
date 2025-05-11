package admin

import (
	"net/http"

	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/handlers"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
	"github.com/jonesrussell/goforms/internal/presentation/templates/shared"
	"github.com/jonesrussell/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

const (
	// UnauthorizedErrorCode is the HTTP status code for unauthorized access
	UnauthorizedErrorCode = http.StatusUnauthorized
	// InternalServerErrorCode is the HTTP status code for internal server errors
	InternalServerErrorCode = http.StatusInternalServerError
)

// DashboardHandler handles the admin dashboard routes
type DashboardHandler struct {
	base        handlers.Base
	renderer    *view.Renderer
	UserService user.Service
	FormService form.Service
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(
	logger logging.Logger,
	renderer *view.Renderer,
	userService user.Service,
	formService form.Service,
) *DashboardHandler {
	return &DashboardHandler{
		base: handlers.Base{
			Logger: logger,
		},
		renderer:    renderer,
		UserService: userService,
		FormService: formService,
	}
}

// Register sets up the routes for the dashboard handler
func (h *DashboardHandler) Register(e *echo.Echo) {
	h.base.RegisterRoute(e, "GET", "/dashboard", h.showDashboard)
}

// showDashboard renders the dashboard page
func (h *DashboardHandler) showDashboard(c echo.Context) error {
	h.base.Logger.Debug("handling dashboard page request")

	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return echo.NewHTTPError(UnauthorizedErrorCode, "User not authenticated")
	}

	forms, err := h.FormService.GetUserForms(currentUser.ID)
	if err != nil {
		h.base.Logger.Error("failed to get user forms",
			logging.Error(err),
			logging.Uint("user_id", currentUser.ID),
		)
		return echo.NewHTTPError(InternalServerErrorCode, "Failed to get forms")
	}

	data := shared.PageData{
		Title: "Dashboard - GoForms",
		User:  currentUser,
		Forms: forms,
	}

	return h.renderer.Render(c, pages.Dashboard(data))
}
