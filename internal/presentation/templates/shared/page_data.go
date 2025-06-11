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
	FormBuilderAssetPath string
	Message              *Message
}

// Message represents a user-facing message
type Message struct {
	Type string
	Text string
}

// GetCurrentUser extracts session data from Echo's context
func GetCurrentUser(c echo.Context) *user.User {
	if c == nil {
		return nil
	}
	sessionVal := c.Get(middleware.SessionKey)
	if sessionVal != nil {
		if session, ok := sessionVal.(*middleware.Session); ok && session != nil {
			return &user.User{
				ID:    session.UserID,
				Email: session.Email,
				Role:  session.Role,
			}
		}
	}
	return nil
}

// GetCSRFToken retrieves the CSRF token from context
func GetCSRFToken(c echo.Context) string {
	if c == nil {
		return ""
	}
	if token, ok := c.Get("csrf").(string); ok {
		return token
	}
	return ""
}

// GenerateAssetPath creates asset paths based on environment settings
func GenerateAssetPath(cfg *config.Config) func(string) string {
	return func(path string) string {
		if cfg != nil && cfg.App.IsDevelopment() {
			return fmt.Sprintf("%s://%s/assets/%s",
				cfg.App.Scheme, net.JoinHostPort(cfg.App.ViteDevHost, cfg.App.ViteDevPort), path)
		}
		return fmt.Sprintf("%s://%s/assets/%s",
			cfg.App.Scheme, net.JoinHostPort(cfg.App.Host, strconv.Itoa(cfg.App.Port)), path)
	}
}

// BuildPageData constructs PageData with extracted functions
func BuildPageData(cfg *config.Config, c echo.Context, title string) PageData {
	return PageData{
		Title:                title,
		User:                 GetCurrentUser(c),
		Forms:                []*model.Form{},           // Placeholder, should be populated elsewhere
		Form:                 nil,                       // Placeholder
		Submissions:          []*model.FormSubmission{}, // Placeholder
		CSRFToken:            GetCSRFToken(c),
		IsDevelopment:        cfg != nil && cfg.App.IsDevelopment(),
		AssetPath:            GenerateAssetPath(cfg),
		Content:              nil, // Should be set by a handler
		FormBuilderAssetPath: "",  // Placeholder
		Message:              nil, // Can be set dynamically when needed
	}
}
