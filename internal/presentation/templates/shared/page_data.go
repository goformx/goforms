package shared

import (
	"github.com/a-h/templ"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
)

// PageData contains common data used across all pages
type PageData struct {
	Title                string
	User                 *user.User
	Forms                []*model.Form
	Form                 *model.Form
	Submissions          []*model.FormSubmission
	CSRFToken            string
	IsDevelopment        bool
	AssetPath            func(string) string
	Content              templ.Component
	FormBuilderAssetPath string // Path to the form builder JS asset
}

// BuildPageData centralizes construction of PageData for handlers
func BuildPageData(cfg *config.Config, title string) PageData {
	assetBase := "/assets/"
	if cfg != nil && cfg.App.IsDevelopment() {
		assetBase = "http://localhost:3000/src/"
	}
	return PageData{
		Title:         title,
		IsDevelopment: cfg != nil && cfg.App.IsDevelopment(),
		AssetPath:     func(path string) string { return assetBase + path },
	}
}
