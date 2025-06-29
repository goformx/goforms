// Package view provides types and utilities for rendering page data and templates.
package view

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/version"
	"github.com/goformx/goforms/internal/infrastructure/web"
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
	FormPreviewAssetPath string
	Message              *Message
	Config               *config.Config
	Session              *session.Session
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
	IsEntry bool     `json:"is_entry"`
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

// GetSession retrieves the session from context
func GetSession(c echo.Context) *session.Session {
	if c == nil {
		return nil
	}

	if sess, ok := c.Get("session").(*session.Session); ok {
		return sess
	}

	return nil
}

// GenerateAssetPath creates asset paths using the provided AssetManager
func GenerateAssetPath(manager web.AssetManagerInterface) func(string) string {
	return func(path string) string {
		return manager.AssetPath(path)
	}
}

// NewPageData creates a new PageData instance with essential data
func NewPageData(cfg *config.Config, manager web.AssetManagerInterface, c echo.Context, title string) *PageData {
	return &PageData{
		Title:         title,
		Description:   "",
		Keywords:      "",
		Author:        "",
		Version:       cfg.App.Version,
		BuildTime:     version.BuildTime,
		GitCommit:     version.GitCommit,
		Environment:   cfg.App.Environment,
		AssetPath:     GenerateAssetPath(manager),
		User:          GetCurrentUser(c),
		Forms:         make([]*model.Form, 0),
		Form:          nil,
		Submissions:   make([]*model.FormSubmission, 0),
		CSRFToken:     GetCSRFToken(c),
		IsDevelopment: cfg.App.IsDevelopment(),
		Content:       nil,
		Message:       nil,
		Config:        cfg,
		Session:       GetSession(c),
	}
}

// WithTitle sets the page title
func (p *PageData) WithTitle(title string) *PageData {
	p.Title = title
	return p
}

// WithDescription sets the page description
func (p *PageData) WithDescription(description string) *PageData {
	p.Description = description
	return p
}

// WithKeywords sets the page keywords
func (p *PageData) WithKeywords(keywords string) *PageData {
	p.Keywords = keywords
	return p
}

// WithAuthor sets the page author
func (p *PageData) WithAuthor(author string) *PageData {
	p.Author = author
	return p
}

// WithContent sets the page content component
func (p *PageData) WithContent(content templ.Component) *PageData {
	p.Content = content
	return p
}

// WithMessage sets a message for the page
func (p *PageData) WithMessage(msgType, text string) *PageData {
	p.Message = &Message{
		Type: msgType,
		Text: text,
	}

	return p
}

// WithForm sets a single form
func (p *PageData) WithForm(form *model.Form) *PageData {
	p.Form = form
	return p
}

// WithForms sets multiple forms
func (p *PageData) WithForms(forms []*model.Form) *PageData {
	p.Forms = forms
	return p
}

// WithSubmissions sets form submissions
func (p *PageData) WithSubmissions(submissions []*model.FormSubmission) *PageData {
	p.Submissions = submissions
	return p
}

// WithFormBuilderAssetPath sets the form builder asset path
func (p *PageData) WithFormBuilderAssetPath(path string) *PageData {
	p.FormBuilderAssetPath = path
	return p
}

// WithFormPreviewAssetPath sets the form preview asset path
func (p *PageData) WithFormPreviewAssetPath(path string) *PageData {
	p.FormPreviewAssetPath = path
	return p
}

// IsAuthenticated checks if the user is authenticated
func (p *PageData) IsAuthenticated() bool {
	return p.User != nil
}

// GetUser returns the current user
func (p *PageData) GetUser() *entities.User {
	return p.User
}

// GetUserID returns the current user ID or empty string if not authenticated
func (p *PageData) GetUserID() string {
	if p.User != nil {
		return p.User.ID
	}

	return ""
}

// GetUserEmail returns the current user email or empty string if not authenticated
func (p *PageData) GetUserEmail() string {
	if p.User != nil {
		return p.User.Email
	}

	return ""
}

// SetUser sets the current user
func (p *PageData) SetUser(user *entities.User) {
	p.User = user
}

// HasMessage checks if there's a message to display
func (p *PageData) HasMessage() bool {
	return p.Message != nil
}

// GetMessageIcon returns the appropriate Bootstrap icon name for a message type
func GetMessageIcon(msgType string) string {
	switch msgType {
	case "success":
		return "check-circle"
	case "error":
		return "exclamation-triangle"
	case "info":
		return "info-circle"
	case "warning":
		return "exclamation-circle"
	default:
		return "info-circle"
	}
}

// GetMessageClass returns the appropriate Bootstrap CSS class for a message type
func GetMessageClass(msgType string) string {
	switch msgType {
	case "success":
		return "alert-success"
	case "error":
		return "alert-danger"
	case "info":
		return "alert-info"
	case "warning":
		return "alert-warning"
	default:
		return "alert-info"
	}
}
