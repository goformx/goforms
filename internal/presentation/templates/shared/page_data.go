package shared

import (
	"github.com/a-h/templ"
	"github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/middleware/session"
)

// PageData represents the data passed to templates
type PageData struct {
	Title                string
	Description          string
	Keywords             string
	Author               string
	Version              string
	BuildTime            string
	GitCommit            string
	Environment          string
	AssetPath            func(string) string
	User                 *entities.User
	Forms                []*model.Form
	Form                 *model.Form
	Submissions          []*model.FormSubmission
	CSRFToken            string
	IsDevelopment        bool
	Content              templ.Component
	FormBuilderAssetPath string
	Message              *Message
	Config               *config.Config
	Session              *session.Session
	UserID               string
	Email                string
}

// Message represents a user-facing message
type Message struct {
	Type string
	Text string
}

// ViteManifest represents the structure of the Vite manifest file
type ViteManifest struct {
	File    string   `json:"file"`
	Name    string   `json:"name"`
	Src     string   `json:"src,omitempty"`
	CSS     []string `json:"css,omitempty"`
	Assets  []string `json:"assets,omitempty"`
	IsEntry bool     `json:"isEntry"`
}

// GetCurrentUser extracts user data from context
func GetCurrentUser(c echo.Context) *entities.User {
	if c == nil {
		return nil
	}
	userID, ok := context.GetUserID(c)
	if !ok {
		return nil
	}
	email, _ := context.GetEmail(c)
	role, _ := context.GetRole(c)
	return &entities.User{
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

// GenerateAssetPath creates asset paths using the provided AssetManager
func GenerateAssetPath(manager *web.AssetManager) func(string) string {
	return func(path string) string {
		return manager.AssetPath(path)
	}
}

// BuildPageData constructs PageData with extracted functions
func BuildPageData(cfg *config.Config, manager *web.AssetManager, c echo.Context, title string) PageData {
	return PageData{
		Title:                title,
		User:                 GetCurrentUser(c),
		Forms:                []*model.Form{},           // Placeholder, should be populated elsewhere
		Form:                 nil,                       // Placeholder
		Submissions:          []*model.FormSubmission{}, // Placeholder
		CSRFToken:            GetCSRFToken(c),
		IsDevelopment:        cfg.App.IsDevelopment(),
		AssetPath:            GenerateAssetPath(manager),
		Content:              nil, // Should be set by a handler
		FormBuilderAssetPath: "",  // Placeholder
		Message:              nil, // Can be set dynamically when needed
		Description:          "",
		Config:               cfg,
		Session:              nil,
		UserID:               "",
		Email:                "",
	}
}

// NewPageData creates a new PageData instance
func NewPageData(title, description string, user *entities.User) *PageData {
	return &PageData{
		Title:       title,
		Description: description,
		User:        user,
	}
}

// IsAuthenticated checks if the user is authenticated
func (p *PageData) IsAuthenticated() bool {
	return p.User != nil
}

// GetUser returns the current user
func (p *PageData) GetUser() *entities.User {
	return p.User
}

// SetUser sets the current user
func (p *PageData) SetUser(user *entities.User) {
	p.User = user
}
