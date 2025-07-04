// Package dto provides data transfer objects for application layer
package dto

import (
	"time"
)

// ViewUser represents user data for view rendering
type ViewUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// ViewForm represents form data for view rendering
type ViewForm struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Schema      map[string]any `json:"schema"`
	UserID      string         `json:"user_id"`
	Status      string         `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// ViewFormSubmission represents form submission data for view rendering
type ViewFormSubmission struct {
	ID        string         `json:"id"`
	FormID    string         `json:"form_id"`
	Data      map[string]any `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
}

// ViewMessage represents a user-facing message
type ViewMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ViewPageData represents the data passed to templates
type ViewPageData struct {
	Title                string
	Description          string
	Keywords             string
	Author               string
	Version              string
	BuildTime            string
	GitCommit            string
	Environment          string
	AssetPath            func(string) string
	User                 *ViewUser
	Forms                []*ViewForm
	Form                 *ViewForm
	Submissions          []*ViewFormSubmission
	CSRFToken            string
	IsDevelopment        bool
	Content              any // templ.Component or similar
	FormBuilderAssetPath string
	FormPreviewAssetPath string
	Message              *ViewMessage
	// Application layer abstractions instead of infrastructure
	AppConfig   *ViewAppConfig
	SessionData *ViewSessionData
}

// ViewAppConfig represents application configuration for views
type ViewAppConfig struct {
	Version     string `json:"version"`
	Environment string `json:"environment"`
	IsDev       bool   `json:"is_dev"`
}

// ViewSessionData represents session data for views
type ViewSessionData struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

// ConvertUserToView converts domain user to view DTO
func ConvertUserToView(user any) *ViewUser {
	if user == nil {
		return nil
	}

	// Use reflection or type assertion based on the actual user type
	// This is a placeholder - implement based on your domain user structure
	return &ViewUser{
		ID:    "", // Extract from user
		Email: "", // Extract from user
		Role:  "", // Extract from user
	}
}

// ConvertFormToView converts domain form to view DTO
func ConvertFormToView(form any) *ViewForm {
	if form == nil {
		return nil
	}

	// Use reflection or type assertion based on the actual form type
	// This is a placeholder - implement based on your domain form structure
	return &ViewForm{
		ID:          "",          // Extract from form
		Title:       "",          // Extract from form
		Description: "",          // Extract from form
		Schema:      nil,         // Extract from form
		UserID:      "",          // Extract from form
		Status:      "",          // Extract from form
		CreatedAt:   time.Time{}, // Extract from form
		UpdatedAt:   time.Time{}, // Extract from form
	}
}

// ConvertFormListToView converts domain form list to view DTOs
func ConvertFormListToView(forms any) []*ViewForm {
	// Implementation depends on the actual form list type
	return []*ViewForm{}
}

// ConvertSubmissionToView converts domain submission to view DTO
func ConvertSubmissionToView(submission any) *ViewFormSubmission {
	if submission == nil {
		return nil
	}

	// Use reflection or type assertion based on the actual submission type
	return &ViewFormSubmission{
		ID:        "",          // Extract from submission
		FormID:    "",          // Extract from submission
		Data:      nil,         // Extract from submission
		CreatedAt: time.Time{}, // Extract from submission
	}
}

// ConvertSubmissionListToView converts domain submission list to view DTOs
func ConvertSubmissionListToView(submissions any) []*ViewFormSubmission {
	// Implementation depends on the actual submission list type
	return []*ViewFormSubmission{}
}
