package shared

import (
	"fmt"
	"net"

	"github.com/a-h/templ"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/labstack/echo/v4"
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
	FormBuilderAssetPath string   // Path to the form builder JS asset
	Message              *Message // Optional message to display
}

// Message represents a user-facing message with a type and text
type Message struct {
	Type string // "error", "success", "info", "warning"
	Text string
}

// BuildPageData centralizes construction of PageData for handlers
func BuildPageData(cfg *config.Config, c echo.Context, title string) PageData {
	assetBase := "/assets/"
	if cfg != nil && cfg.App.IsDevelopment() {
		hostPort := net.JoinHostPort(cfg.App.ViteDevHost, cfg.App.ViteDevPort)
		assetBase = fmt.Sprintf("http://%s/", hostPort)
	}

	csrfToken := ""
	if c != nil {
		if token, ok := c.Get("csrf").(string); ok {
			csrfToken = token
		}
	}

	return PageData{
		Title:         title,
		IsDevelopment: cfg != nil && cfg.App.IsDevelopment(),
		AssetPath:     func(path string) string { return assetBase + path },
		CSRFToken:     csrfToken,
	}
}
