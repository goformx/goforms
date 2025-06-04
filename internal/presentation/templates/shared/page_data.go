package shared

import (
	"github.com/a-h/templ"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	webassets "github.com/goformx/goforms/internal/infrastructure/web"
)

// PageData contains common data used across all pages
type PageData struct {
	Title                string
	User                 *user.User
	Forms                []*form.Form
	Form                 *form.Form
	Submissions          []*model.FormSubmission
	CSRFToken            string
	IsDevelopment        bool
	AssetPath            func(string) string
	Content              templ.Component
	FormBuilderAssetPath string // Path to the form builder JS asset
}

// BuildPageData centralizes construction of PageData for handlers
func BuildPageData(cfg *config.Config, title string) PageData {
	return PageData{
		Title:         title,
		IsDevelopment: cfg != nil && cfg.App.IsDevelopment(),
		AssetPath:     webassets.GetAssetPath,
	}
}
