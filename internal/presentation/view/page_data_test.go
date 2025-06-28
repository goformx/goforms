package view_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

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
				ID:    "user-123",
				Email: "test@example.com",
				Role:  "admin",
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
				ID:    "user-123",
				Email: "",
				Role:  "",
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

func TestBuildPageData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create test dependencies
	cfg := &config.Config{
		App: config.AppConfig{
			Environment: "development",
		},
	}

	// Create a mock asset manager
	mockManager := webmocks.NewMockAssetManagerInterface(ctrl)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set up context with user and CSRF token
	contextmw.SetUserID(c, "user-123")
	contextmw.SetEmail(c, "test@example.com")
	c.Set("csrf", "csrf-token-123")

	// Test building page data
	pageData := view.BuildPageData(cfg, mockManager, c, "Test Page")

	// Verify the page data
	assert.Equal(t, "Test Page", pageData.Title)
	assert.NotNil(t, pageData.User)
	assert.Equal(t, "user-123", pageData.User.ID)
	assert.Equal(t, "test@example.com", pageData.User.Email)
	assert.Equal(t, "csrf-token-123", pageData.CSRFToken)
	assert.True(t, pageData.IsDevelopment)
	assert.NotNil(t, pageData.AssetPath)
	assert.Equal(t, cfg, pageData.Config)
}

func TestNewPageData(t *testing.T) {
	user := &entities.User{
		ID:    "user-123",
		Email: "test@example.com",
		Role:  "admin",
	}

	pageData := view.NewPageData("Test Page", "Test Description", user)

	assert.Equal(t, "Test Page", pageData.Title)
	assert.Equal(t, "Test Description", pageData.Description)
	assert.Equal(t, user, pageData.User)
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
				ID:    "user-123",
				Email: "test@example.com",
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
				User: tt.user,
			}
			got := pageData.IsAuthenticated()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPageData_GetUser(t *testing.T) {
	user := &entities.User{
		ID:    "user-123",
		Email: "test@example.com",
	}

	pageData := &view.PageData{
		User: user,
	}

	got := pageData.GetUser()
	assert.Equal(t, user, got)
}

func TestPageData_SetUser(t *testing.T) {
	pageData := &view.PageData{}

	user := &entities.User{
		ID:    "user-123",
		Email: "test@example.com",
	}

	pageData.SetUser(user)
	assert.Equal(t, user, pageData.User)
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

func TestPageData_Construction(_ *testing.T) {
	_ = &view.PageData{}
}
