package handlers

import (
	"net/http"

	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/presentation/middleware"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
	"github.com/jonesrussell/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

type DashboardHandler struct {
	authMiddleware *middleware.AuthMiddleware
	formService    form.Service
}

func NewDashboardHandler(userService user.Service, formService form.Service) *DashboardHandler {
	return &DashboardHandler{
		authMiddleware: middleware.NewAuthMiddleware(userService),
		formService:    formService,
	}
}

func (h *DashboardHandler) Register(e *echo.Echo) {
	// Dashboard routes
	dashboard := e.Group("/dashboard")
	dashboard.Use(h.authMiddleware.RequireAuth) // Middleware to ensure user is authenticated

	dashboard.GET("", h.ShowDashboard)
}

func (h *DashboardHandler) ShowDashboard(c echo.Context) error {
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	// Get user's forms
	forms, err := h.formService.GetUserForms(currentUser.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch forms")
	}

	// Create page data
	data := shared.PageData{
		Title: "Dashboard - GoForms",
		User:  currentUser,
		Forms: forms,
	}

	// Render dashboard page
	return pages.Dashboard(data).Render(c.Request().Context(), c.Response().Writer)
}
