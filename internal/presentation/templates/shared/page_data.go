package shared

import (
	"fmt"
	"net"
	"strconv"

	"github.com/a-h/templ"
	"github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/middleware/session"
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
	Description          string
	Config               *config.Config
	Session              *session.Session
	UserID               string
	Email                string
	Role                 string
	Error                string
	Success              string
	Data                 interface{}
}

// Message represents a user-facing message
type Message struct {
	Type string
	Text string
}

// GetCurrentUser extracts user data from context
func GetCurrentUser(c echo.Context) *user.User {
	if c == nil {
		return nil
	}
	userID, ok := context.GetUserID(c)
	if !ok {
		return nil
	}
	email, _ := context.GetEmail(c)
	role, _ := context.GetRole(c)
	return &user.User{
		ID:    userID,
		Email: email,
		Role:  role,
	}
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
		Description:          "",
		Config:               cfg,
		Session:              nil,
		UserID:               "",
		Email:                "",
		Role:                 "",
		Error:                "",
		Success:              "",
		Data:                 nil,
	}
}
