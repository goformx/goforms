package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	formmodel "github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFormService is a mock implementation of the form service
type MockFormService struct {
	mock.Mock
}

func (m *MockFormService) CreateForm(ctx context.Context, form *formmodel.Form) error {
	args := m.Called(ctx, form)
	return args.Error(0)
}

func (m *MockFormService) UpdateForm(ctx context.Context, form *formmodel.Form) error {
	args := m.Called(ctx, form)
	return args.Error(0)
}

func (m *MockFormService) DeleteForm(ctx context.Context, formID string) error {
	args := m.Called(ctx, formID)
	return args.Error(0)
}

func (m *MockFormService) GetForm(ctx context.Context, formID string) (*formmodel.Form, error) {
	args := m.Called(ctx, formID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*formmodel.Form), args.Error(1)
}

func (m *MockFormService) ListForms(ctx context.Context, userID string) ([]*formmodel.Form, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*formmodel.Form), args.Error(1)
}

func (m *MockFormService) SubmitForm(ctx context.Context, submission *formmodel.FormSubmission) error {
	args := m.Called(ctx, submission)
	return args.Error(0)
}

func (m *MockFormService) GetFormSubmission(ctx context.Context, submissionID string) (*formmodel.FormSubmission, error) {
	args := m.Called(ctx, submissionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*formmodel.FormSubmission), args.Error(1)
}

func (m *MockFormService) ListFormSubmissions(ctx context.Context, formID string) ([]*formmodel.FormSubmission, error) {
	args := m.Called(ctx, formID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*formmodel.FormSubmission), args.Error(1)
}

func (m *MockFormService) UpdateFormState(ctx context.Context, formID, state string) error {
	args := m.Called(ctx, formID, state)
	return args.Error(0)
}

func (m *MockFormService) TrackFormAnalytics(ctx context.Context, formID, eventType string) error {
	args := m.Called(ctx, formID, eventType)
	return args.Error(0)
}

// createTestLogger creates a test logger for testing
func createTestLogger() logging.Logger {
	factory := logging.NewFactory(logging.FactoryConfig{
		AppName:     "test-app",
		Version:     "1.0.0",
		Environment: "test",
	}, sanitization.NewService())
	logger, _ := factory.CreateLogger()
	return logger
}

func TestPerFormCORS_FormRoute(t *testing.T) {
	// Setup
	e := echo.New()
	mockFormService := new(MockFormService)
	logger := createTestLogger()

	globalCORS := &config.SecurityConfig{
		CorsAllowedOrigins:   []string{"http://localhost:3000"},
		CorsAllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		CorsAllowedHeaders:   []string{"Content-Type", "Authorization"},
		CorsAllowCredentials: true,
		CorsMaxAge:           3600,
	}

	config := NewPerFormCORSConfig(mockFormService, logger, globalCORS)
	middleware := PerFormCORS(config)

	// Create a test form with custom CORS settings
	testForm := &formmodel.Form{
		ID:          "test-form-id",
		CorsOrigins: []string{"https://example.com"},
		CorsMethods: []string{"GET", "POST", "PUT", "OPTIONS"},
		CorsHeaders: []string{"Content-Type", "X-Custom-Header"},
	}

	// Setup mock expectations
	mockFormService.On("GetForm", mock.Anything, "test-form-id").Return(testForm, nil)

	// Test form route
	req := httptest.NewRequest(http.MethodGet, "/forms/test-form-id", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	err := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET,POST,PUT,OPTIONS", rec.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type,X-Custom-Header", rec.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))

	mockFormService.AssertExpectations(t)
}

func TestPerFormCORS_NonFormRoute(t *testing.T) {
	// Setup
	e := echo.New()
	mockFormService := new(MockFormService)
	logger := createTestLogger()

	globalCORS := &config.SecurityConfig{
		CorsAllowedOrigins:   []string{"http://localhost:3000"},
		CorsAllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		CorsAllowedHeaders:   []string{"Content-Type", "Authorization"},
		CorsAllowCredentials: true,
		CorsMaxAge:           3600,
	}

	config := NewPerFormCORSConfig(mockFormService, logger, globalCORS)
	middleware := PerFormCORS(config)

	// Test non-form route (should not call form service)
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	err := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET,POST,OPTIONS", rec.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type,Authorization", rec.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))

	// Verify form service was not called
	mockFormService.AssertNotCalled(t, "GetForm")
}

func TestPerFormCORS_PreflightRequest(t *testing.T) {
	// Setup
	e := echo.New()
	mockFormService := new(MockFormService)
	logger := createTestLogger()

	globalCORS := &config.SecurityConfig{
		CorsAllowedOrigins:   []string{"http://localhost:3000"},
		CorsAllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		CorsAllowedHeaders:   []string{"Content-Type", "Authorization"},
		CorsAllowCredentials: true,
		CorsMaxAge:           3600,
	}

	config := NewPerFormCORSConfig(mockFormService, logger, globalCORS)
	middleware := PerFormCORS(config)

	// Create a test form with custom CORS settings
	testForm := &formmodel.Form{
		ID:          "test-form-id",
		CorsOrigins: []string{"https://example.com"},
		CorsMethods: []string{"GET", "POST", "PUT", "OPTIONS"},
		CorsHeaders: []string{"Content-Type", "X-Custom-Header"},
	}

	// Setup mock expectations
	mockFormService.On("GetForm", mock.Anything, "test-form-id").Return(testForm, nil)

	// Test preflight request
	req := httptest.NewRequest(http.MethodOptions, "/forms/test-form-id", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	err := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET,POST,PUT,OPTIONS", rec.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type,X-Custom-Header", rec.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "3600", rec.Header().Get("Access-Control-Max-Age"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))

	mockFormService.AssertExpectations(t)
}

func TestPerFormCORS_FormNotFound(t *testing.T) {
	// Setup
	e := echo.New()
	mockFormService := new(MockFormService)
	logger := createTestLogger()

	globalCORS := &config.SecurityConfig{
		CorsAllowedOrigins:   []string{"http://localhost:3000"},
		CorsAllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		CorsAllowedHeaders:   []string{"Content-Type", "Authorization"},
		CorsAllowCredentials: true,
		CorsMaxAge:           3600,
	}

	config := NewPerFormCORSConfig(mockFormService, logger, globalCORS)
	middleware := PerFormCORS(config)

	// Setup mock expectations - form not found
	mockFormService.On("GetForm", mock.Anything, "non-existent-form").Return(nil, nil)

	// Test form route with non-existent form
	req := httptest.NewRequest(http.MethodGet, "/forms/non-existent-form", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	err := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})(c)

	// Assertions - should fallback to global CORS
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET,POST,OPTIONS", rec.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type,Authorization", rec.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))

	mockFormService.AssertExpectations(t)
}

func TestExtractFormID(t *testing.T) {
	config := NewPerFormCORSConfig(nil, nil, nil)

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"web form route", "/forms/test-form-id", "test-form-id"},
		{"api form route", "/api/v1/forms/test-form-id", "test-form-id"},
		{"form with subpath", "/forms/test-form-id/submissions", "test-form-id"},
		{"api form with subpath", "/api/v1/forms/test-form-id/schema", "test-form-id"},
		{"non-form route", "/dashboard", ""},
		{"root path", "/", ""},
		{"empty path", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFormID(tt.path, config.FormRouteRegex)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expected       bool
	}{
		{"wildcard allows all", "https://example.com", []string{"*"}, true},
		{"exact match", "https://example.com", []string{"https://example.com"}, true},
		{"no match", "https://example.com", []string{"https://other.com"}, false},
		{"empty origin allowed", "", []string{"https://example.com"}, true},
		{"multiple origins", "https://example.com", []string{"https://other.com", "https://example.com"}, true},
		{"case sensitive", "https://Example.com", []string{"https://example.com"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOriginAllowed(tt.origin, tt.allowedOrigins)
			assert.Equal(t, tt.expected, result)
		})
	}
}
