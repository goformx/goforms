package shared

import (
	"fmt"
	"net"
	"strconv"

	"github.com/a-h/templ"
	"github.com/goformx/goforms/internal/application/middleware"
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
	csrfToken := ""
	var currentUser *user.User
	if c != nil {
		if token, ok := c.Get("csrf").(string); ok {
			csrfToken = token
		}

		// Try to get session from context and populate currentUser
		if sessionVal := c.Get(middleware.SessionKey); sessionVal != nil {
			if session, ok := sessionVal.(*middleware.Session); ok && session != nil {
				currentUser = &user.User{
					ID:    session.UserID,
					Email: session.Email,
					Role:  session.Role,
				}
			}
		}
	}

	return PageData{
		Title:         title,
		User:          currentUser,
		IsDevelopment: cfg != nil && cfg.App.IsDevelopment(),
		AssetPath: func(path string) string {
			if cfg != nil && cfg.App.IsDevelopment() {
				// In development, serve source files directly from Vite
				return fmt.Sprintf("%s://%s/assets/%s", cfg.App.Scheme, net.JoinHostPort(cfg.App.ViteDevHost, cfg.App.ViteDevPort), path)
			}
			// In production, use the built assets
			return fmt.Sprintf("%s://%s/assets/%s", cfg.App.Scheme, net.JoinHostPort(cfg.App.Host, strconv.Itoa(cfg.App.Port)), path)
		},
		CSRFToken: csrfToken,
	}
}
