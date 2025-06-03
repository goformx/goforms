package pagedata

import (
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// Service defines the interface for page data operations
type Service interface {
	// PrepareDashboardData prepares data for the dashboard page
	PrepareDashboardData(c echo.Context, usr *user.User, forms []*form.Form) shared.PageData
	// PrepareFormData prepares data for the form page (with submissions)
	PrepareFormData(
		c echo.Context,
		usr *user.User,
		frm *form.Form,
		submissions []*model.FormSubmission,
	) shared.PageData
	// PrepareNewFormData prepares data for the new form page
	PrepareNewFormData(c echo.Context, usr *user.User) shared.PageData
	// PrepareProfileData prepares data for the profile page
	PrepareProfileData(c echo.Context, usr *user.User) shared.PageData
	// PrepareSettingsData prepares data for the settings page
	PrepareSettingsData(c echo.Context, usr *user.User) shared.PageData
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
func (s *service) PrepareDashboardData(c echo.Context, usr *user.User, forms []*form.Form) shared.PageData {
	if usr == nil {
		s.logger.Error("PrepareDashboardData called with nil user; this should not happen!", nil)
		return shared.PageData{
			Title:     "Dashboard",
			User:      &user.User{FirstName: "User"},
			Forms:     forms,
			AssetPath: web.GetAssetPath,
		}
	}
	return shared.PageData{
		Title:     "Dashboard",
		User:      usr,
		Forms:     forms,
		AssetPath: web.GetAssetPath,
	}
}

// PrepareFormData prepares data for the form page
func (s *service) PrepareFormData(
	c echo.Context,
	usr *user.User,
	frm *form.Form,
	submissions []*model.FormSubmission,
) shared.PageData {
	return shared.PageData{
		Title:       frm.Title,
		User:        usr,
		Form:        frm,
		Submissions: submissions,
		AssetPath:   web.GetAssetPath,
	}
}

// PrepareNewFormData prepares data for the new form page
func (s *service) PrepareNewFormData(c echo.Context, usr *user.User) shared.PageData {
	return shared.PageData{
		Title:     "Create New Form - GoFormX",
		User:      usr,
		AssetPath: web.GetAssetPath,
	}
}

// PrepareProfileData prepares data for the profile page
func (s *service) PrepareProfileData(c echo.Context, usr *user.User) shared.PageData {
	return shared.PageData{
		Title:     "Profile - GoFormX",
		User:      usr,
		AssetPath: web.GetAssetPath,
	}
}

// PrepareSettingsData prepares data for the settings page
func (s *service) PrepareSettingsData(c echo.Context, usr *user.User) shared.PageData {
	return shared.PageData{
		Title:     "Settings - GoFormX",
		User:      usr,
		AssetPath: web.GetAssetPath,
	}
}
