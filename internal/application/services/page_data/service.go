package page_data

import (
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
)

// Service defines the interface for page data preparation
type Service interface {
	// PrepareDashboardData prepares data for the dashboard page
	PrepareDashboardData(c echo.Context, user *user.User, forms []*form.Form) shared.PageData
	// PrepareFormData prepares data for the form page
	PrepareFormData(c echo.Context, user *user.User, form *form.Form) shared.PageData
}

// service implements the page data service
type service struct {
	logger logging.Logger
}

// NewService creates a new page data service
func NewService(logger logging.Logger) Service {
	return &service{
		logger: logger,
	}
}

// PrepareDashboardData prepares data for the dashboard page
func (s *service) PrepareDashboardData(c echo.Context, user *user.User, forms []*form.Form) shared.PageData {
	return shared.PageData{
		Title: "Dashboard",
		User:  user,
		Forms: forms,
	}
}

// PrepareFormData prepares data for the form page
func (s *service) PrepareFormData(c echo.Context, user *user.User, form *form.Form) shared.PageData {
	return shared.PageData{
		Title: form.Title,
		User:  user,
		Form:  form,
	}
}
