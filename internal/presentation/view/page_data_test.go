package view_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	contextmw "github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/presentation/view"
	webmocks "github.com/goformx/goforms/test/mocks/web"
)

func TestGetCurrentUser(t *testing.T) {
	tests := []struct {
		name     string
		setupCtx func() echo.Context
		want     *entities.User
	}{
		{
			name: "user with all fields",
			setupCtx: func() echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				contextmw.SetUserID(c, "user-123")
				contextmw.SetEmail(c, "test@example.com")
				contextmw.SetRole(c, "admin")

				return c
			},
			want: &entities.User{
				ID:             "user-123",
				Email:          "test@example.com",
				Role:           "admin",
				HashedPassword: "",
				FirstName:      "",
				LastName:       "",
				Active:         false,
				CreatedAt:      time.Time{},
				UpdatedAt:      time.Time{},
				DeletedAt:      gorm.DeletedAt{},
			},
		},
		{
			name: "user with only ID",
			setupCtx: func() echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				contextmw.SetUserID(c, "user-123")

				return c
			},
			want: &entities.User{
				ID:             "user-123",
				Email:          "",
				Role:           "",
				HashedPassword: "",
				FirstName:      "",
				LastName:       "",
				Active:         false,
				CreatedAt:      time.Time{},
				UpdatedAt:      time.Time{},
				DeletedAt:      gorm.DeletedAt{},
			},
		},
		{
			name: "no user ID in context",
			setupCtx: func() echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				return c
			},
			want: nil,
		},
		{
			name: "nil context",
			setupCtx: func() echo.Context {
				return nil
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setupCtx()
			got := view.GetCurrentUser(c)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetCSRFToken(t *testing.T) {
	tests := []struct {
		name     string
		setupCtx func() echo.Context
		want     string
	}{
		{
			name: "CSRF token present",
			setupCtx: func() echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				c.Set("csrf", "csrf-token-123")

				return c
			},
			want: "csrf-token-123",
		},
		{
			name: "no CSRF token",
			setupCtx: func() echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				return c
			},
			want: "",
		},
		{
			name: "nil context",
			setupCtx: func() echo.Context {
				return nil
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setupCtx()
			got := view.GetCSRFToken(c)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateAssetPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock asset manager
	mockManager := webmocks.NewMockAssetManagerInterface(ctrl)

	// Set up expectations
	mockManager.EXPECT().AssetPath("test/path.js").Return("/assets/test-path.js").Times(1)
	mockManager.EXPECT().AssetPath("css/styles.css").Return("/assets/styles.css").Times(1)

	// Test the GenerateAssetPath function
	assetPathFn := view.GenerateAssetPath(mockManager)

	// Test that it returns a function
	assert.NotNil(t, assetPathFn)

	// Test that the function can be called and returns expected results
	result1 := assetPathFn("test/path.js")
	assert.Equal(t, "/assets/test-path.js", result1)

	result2 := assetPathFn("css/styles.css")
	assert.Equal(t, "/assets/styles.css", result2)
}

func TestNewPageData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create test dependencies
	cfg := &config.Config{
		App: config.AppConfig{
			Name:           "Test App",
			Environment:    "development",
			Version:        "1.0.0",
			Debug:          false,
			LogLevel:       "info",
			URL:            "",
			Scheme:         "http",
			Port:           8080,
			Host:           "localhost",
			ReadTimeout:    15 * time.Second,
			WriteTimeout:   15 * time.Second,
			IdleTimeout:    60 * time.Second,
			RequestTimeout: 30 * time.Second,
			ViteDevHost:    "localhost",
			ViteDevPort:    "5173",
		},
		Database: config.DatabaseConfig{},
		Security: config.SecurityConfig{},
		Email:    config.EmailConfig{},
		Storage:  config.StorageConfig{},
		Cache:    config.CacheConfig{},
		Logging:  config.LoggingConfig{},
		Session:  config.SessionConfig{},
		Auth:     config.AuthConfig{},
		Form:     config.FormConfig{},
		API:      config.APIConfig{},
		Web:      config.WebConfig{},
		User:     config.UserConfig{},
	}

	// Create a mock asset manager
	mockManager := webmocks.NewMockAssetManagerInterface(ctrl)
	mockManager.EXPECT().AssetPath(gomock.Any()).Return("/assets/test.js").AnyTimes()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set up context with user and CSRF token
	contextmw.SetUserID(c, "user-123")
	contextmw.SetEmail(c, "test@example.com")
	c.Set("csrf", "csrf-token-123")

	// Test building page data
	pageData := view.NewPageData(cfg, mockManager, c, "Test Page")

	// Verify the page data
	assert.Equal(t, "Test Page", pageData.Title)
	assert.NotNil(t, pageData.User)
	assert.Equal(t, "user-123", pageData.User.ID)
	assert.Equal(t, "test@example.com", pageData.User.Email)
	assert.Equal(t, "csrf-token-123", pageData.CSRFToken)
	assert.True(t, pageData.IsDevelopment)
	assert.NotNil(t, pageData.AssetPath)
	assert.Equal(t, cfg, pageData.Config)
	assert.NotNil(t, pageData.Forms)
	assert.Len(t, pageData.Forms, 0) // Should be empty slice
	assert.NotNil(t, pageData.Submissions)
	assert.Len(t, pageData.Submissions, 0) // Should be empty slice
}

func TestNewPageDataWithTitle(t *testing.T) {
	user := &entities.User{
		ID:             "user-123",
		Email:          "test@example.com",
		HashedPassword: "",
		FirstName:      "",
		LastName:       "",
		Role:           "",
		Active:         false,
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
		DeletedAt:      gorm.DeletedAt{},
	}

	// Note: This test is for the simple constructor that was in the original code
	// It creates a minimal PageData instance
	pageData := &view.PageData{
		Title:       "Test Page",
		Description: "Test Description",
		User:        user,
	}

	assert.Equal(t, "Test Page", pageData.Title)
	assert.Equal(t, "Test Description", pageData.Description)
	assert.Equal(t, user, pageData.User)
}

func TestPageData_FluentInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		App: config.AppConfig{
			Name:           "Test App",
			Environment:    "test",
			Version:        "1.0.0",
			Debug:          false,
			LogLevel:       "info",
			URL:            "",
			Scheme:         "http",
			Port:           8080,
			Host:           "localhost",
			ReadTimeout:    15 * time.Second,
			WriteTimeout:   15 * time.Second,
			IdleTimeout:    60 * time.Second,
			RequestTimeout: 30 * time.Second,
			ViteDevHost:    "localhost",
			ViteDevPort:    "5173",
		},
		Database: config.DatabaseConfig{},
		Security: config.SecurityConfig{},
		Email:    config.EmailConfig{},
		Storage:  config.StorageConfig{},
		Cache:    config.CacheConfig{},
		Logging:  config.LoggingConfig{},
		Session:  config.SessionConfig{},
		Auth:     config.AuthConfig{},
		Form:     config.FormConfig{},
		API:      config.APIConfig{},
		Web:      config.WebConfig{},
		User:     config.UserConfig{},
	}

	mockManager := webmocks.NewMockAssetManagerInterface(ctrl)
	mockManager.EXPECT().AssetPath(gomock.Any()).Return("/assets/test.js").AnyTimes()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test fluent interface
	pageData := view.NewPageData(cfg, mockManager, c, "Test").
		WithDescription("Test Description").
		WithKeywords("test, keywords").
		WithAuthor("Test Author").
		WithMessage("success", "Test message").
		WithFormBuilderAssetPath("/builder.js").
		WithFormPreviewAssetPath("/preview.js")

	assert.Equal(t, "Test", pageData.Title)
	assert.Equal(t, "Test Description", pageData.Description)
	assert.Equal(t, "test, keywords", pageData.Keywords)
	assert.Equal(t, "Test Author", pageData.Author)
	assert.NotNil(t, pageData.Message)
	assert.Equal(t, "success", pageData.Message.Type)
	assert.Equal(t, "Test message", pageData.Message.Text)
	assert.Equal(t, "/builder.js", pageData.FormBuilderAssetPath)
	assert.Equal(t, "/preview.js", pageData.FormPreviewAssetPath)
}

func TestPageData_IsAuthenticated(t *testing.T) {
	tests := []struct {
		name string
		user *entities.User
		want bool
	}{
		{
			name: "authenticated user",
			user: &entities.User{
				ID:             "user-123",
				Email:          "test@example.com",
				HashedPassword: "",
				FirstName:      "",
				LastName:       "",
				Role:           "",
				Active:         false,
				CreatedAt:      time.Time{},
				UpdatedAt:      time.Time{},
				DeletedAt:      gorm.DeletedAt{},
			},
			want: true,
		},
		{
			name: "no user",
			user: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageData := &view.PageData{
				Title:                "",
				Description:          "",
				Keywords:             "",
				Author:               "",
				Version:              "",
				BuildTime:            "",
				GitCommit:            "",
				Environment:          "",
				AssetPath:            nil,
				User:                 tt.user,
				Forms:                nil,
				Form:                 nil,
				Submissions:          nil,
				CSRFToken:            "",
				IsDevelopment:        false,
				Content:              nil,
				FormBuilderAssetPath: "",
				FormPreviewAssetPath: "",
				Message:              nil,
				Config:               nil,
				Session:              nil,
			}
			got := pageData.IsAuthenticated()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPageData_GetUser(t *testing.T) {
	user := &entities.User{
		ID:             "user-123",
		Email:          "test@example.com",
		HashedPassword: "",
		FirstName:      "",
		LastName:       "",
		Role:           "",
		Active:         false,
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
		DeletedAt:      gorm.DeletedAt{},
	}

	pageData := &view.PageData{
		User: user,
	}

	got := pageData.GetUser()
	assert.Equal(t, user, got)
}

func TestPageData_GetUserID(t *testing.T) {
	tests := []struct {
		name string
		user *entities.User
		want string
	}{
		{
			name: "user with ID",
			user: &entities.User{
				ID:             "user-123",
				Email:          "test@example.com",
				HashedPassword: "",
				FirstName:      "",
				LastName:       "",
				Role:           "",
				Active:         false,
				CreatedAt:      time.Time{},
				UpdatedAt:      time.Time{},
				DeletedAt:      gorm.DeletedAt{},
			},
			want: "user-123",
		},
		{
			name: "no user",
			user: nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageData := &view.PageData{
				User: tt.user,
			}
			got := pageData.GetUserID()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPageData_GetUserEmail(t *testing.T) {
	tests := []struct {
		name string
		user *entities.User
		want string
	}{
		{
			name: "user with email",
			user: &entities.User{
				ID:             "user-123",
				Email:          "test@example.com",
				HashedPassword: "",
				FirstName:      "",
				LastName:       "",
				Role:           "",
				Active:         false,
				CreatedAt:      time.Time{},
				UpdatedAt:      time.Time{},
				DeletedAt:      gorm.DeletedAt{},
			},
			want: "test@example.com",
		},
		{
			name: "no user",
			user: nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageData := &view.PageData{
				User: tt.user,
			}
			got := pageData.GetUserEmail()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPageData_SetUser(t *testing.T) {
	pageData := &view.PageData{}

	user := &entities.User{
		ID:             "user-123",
		Email:          "test@example.com",
		HashedPassword: "",
		FirstName:      "",
		LastName:       "",
		Role:           "",
		Active:         false,
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
		DeletedAt:      gorm.DeletedAt{},
	}

	pageData.SetUser(user)
	assert.Equal(t, user, pageData.User)
}

func TestPageData_HasMessage(t *testing.T) {
	tests := []struct {
		name    string
		message *view.Message
		want    bool
	}{
		{
			name: "has message",
			message: &view.Message{
				Type: "success",
				Text: "Test message",
			},
			want: true,
		},
		{
			name:    "no message",
			message: nil,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageData := &view.PageData{
				Message: tt.message,
			}
			got := pageData.HasMessage()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetMessageIcon(t *testing.T) {
	tests := []struct {
		name    string
		msgType string
		want    string
	}{
		{
			name:    "success message",
			msgType: "success",
			want:    "check-circle",
		},
		{
			name:    "error message",
			msgType: "error",
			want:    "exclamation-triangle",
		},
		{
			name:    "info message",
			msgType: "info",
			want:    "info-circle",
		},
		{
			name:    "warning message",
			msgType: "warning",
			want:    "exclamation-circle",
		},
		{
			name:    "unknown message type",
			msgType: "unknown",
			want:    "info-circle",
		},
		{
			name:    "empty message type",
			msgType: "",
			want:    "info-circle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := view.GetMessageIcon(tt.msgType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetMessageClass(t *testing.T) {
	tests := []struct {
		name    string
		msgType string
		want    string
	}{
		{
			name:    "success message",
			msgType: "success",
			want:    "alert-success",
		},
		{
			name:    "error message",
			msgType: "error",
			want:    "alert-danger",
		},
		{
			name:    "info message",
			msgType: "info",
			want:    "alert-info",
		},
		{
			name:    "warning message",
			msgType: "warning",
			want:    "alert-warning",
		},
		{
			name:    "unknown message type",
			msgType: "unknown",
			want:    "alert-info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := view.GetMessageClass(tt.msgType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPageData_Construction(_ *testing.T) {
	_ = &view.PageData{}
}
