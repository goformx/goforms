package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/form/model"
	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
	mockform "github.com/goformx/goforms/test/mocks/form"
	mocklogging "github.com/goformx/goforms/test/mocks/logging"
)

func TestPerFormCORS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFormService := mockform.NewMockService(ctrl)
	mockLogger := mocklogging.NewMockLogger(ctrl)

	// Create global CORS config
	globalCORS := &appconfig.SecurityConfig{
		CORS: appconfig.CORSConfig{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Csrf-Token", "X-Requested-With"},
			AllowCredentials: false,
			MaxAge:           86400,
		},
	}

	// Create middleware instance
	config := middleware.NewPerFormCORSConfig(mockFormService, mockLogger, globalCORS)
	mw := middleware.PerFormCORS(config)

	// Test form with CORS configuration
	formWithCORS := &model.Form{
		ID:          "test-form-123",
		Title:       "Test Form",
		CorsOrigins: model.JSON{"origins": []string{"https://example.com", "https://app.example.com"}},
		CorsMethods: model.JSON{"methods": []string{"GET", "POST", "PUT"}},
		CorsHeaders: model.JSON{"headers": []string{"Content-Type", "Authorization"}},
	}

	// Test form without CORS configuration
	formWithoutCORS := &model.Form{
		ID:    "test-form-456",
		Title: "Test Form No CORS",
	}

	tests := []struct {
		name            string
		path            string
		method          string
		origin          string
		setupMock       func()
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name:   "form route with CORS - actual request",
			path:   "/forms/test-form-123",
			method: "POST",
			origin: "https://example.com",
			setupMock: func() {
				mockFormService.EXPECT().
					GetForm(gomock.Any(), "test-form-123").
					Return(formWithCORS, nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "https://example.com",
				"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Csrf-Token,X-Requested-With",
			},
		},
		{
			name:   "form route with CORS - preflight request",
			path:   "/forms/test-form-123",
			method: "OPTIONS",
			origin: "https://example.com",
			setupMock: func() {
				mockFormService.EXPECT().
					GetForm(gomock.Any(), "test-form-123").
					Return(formWithCORS, nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "https://example.com",
				"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Csrf-Token,X-Requested-With",
			},
		},
		{
			name:   "form route without CORS - fallback to global",
			path:   "/forms/test-form-456",
			method: "POST",
			origin: "https://example.com",
			setupMock: func() {
				mockFormService.EXPECT().
					GetForm(gomock.Any(), "test-form-456").
					Return(formWithoutCORS, nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "https://example.com",
				"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Csrf-Token,X-Requested-With",
			},
		},
		{
			name:   "form route - form not found - fallback to global",
			path:   "/forms/non-existent",
			method: "POST",
			origin: "https://example.com",
			setupMock: func() {
				mockFormService.EXPECT().
					GetForm(gomock.Any(), "non-existent").
					Return(nil, errors.New("form not found"))
				mockLogger.EXPECT().
					SanitizeField("form_id", "non-existent").
					Return("non-existent")
				mockLogger.EXPECT().
					Debug(
						"failed to load form for CORS",
						"form_id", "non-existent",
						"error", gomock.Any(),
						"falling_back_to_global_cors", true,
					)
			},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "https://example.com",
				"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Csrf-Token,X-Requested-With",
			},
		},
		{
			name:   "form route - form returns nil without error - fallback to global",
			path:   "/forms/nil-form",
			method: "POST",
			origin: "https://example.com",
			setupMock: func() {
				mockFormService.EXPECT().
					GetForm(gomock.Any(), "nil-form").
					Return(nil, nil)
				mockLogger.EXPECT().
					SanitizeField("form_id", "nil-form").
					Return("nil-form")
				mockLogger.EXPECT().
					Debug(
						"form not found for CORS",
						"form_id", "nil-form",
						"falling_back_to_global_cors", true,
					)
			},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "https://example.com",
				"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Csrf-Token,X-Requested-With",
			},
		},
		{
			name:   "non-form route - global CORS",
			path:   "/api/health",
			method: "GET",
			origin: "https://example.com",
			setupMock: func() {
				// No mock expectations for non-form routes
			},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "https://example.com",
				"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Csrf-Token,X-Requested-With",
			},
		},
		{
			name:   "API form route with CORS",
			path:   "/api/v1/forms/test-form-123",
			method: "POST",
			origin: "https://example.com",
			setupMock: func() {
				mockFormService.EXPECT().
					GetForm(gomock.Any(), "test-form-123").
					Return(formWithCORS, nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "https://example.com",
				"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Csrf-Token,X-Requested-With",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			tt.setupMock()

			// Create Echo context
			e := echo.New()
			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Create a simple handler that returns success
			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "success")
			}

			// Apply middleware
			err := mw(handler)(c)

			// Assertions
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Check CORS headers
			for header, expectedValue := range tt.expectedHeaders {
				actualValue := rec.Header().Get(header)
				assert.Equal(t, expectedValue, actualValue, "Header %s mismatch", header)
			}
		})
	}
}

func TestExtractFormID(t *testing.T) {
	// Create a test regex
	formRouteRegex := regexp.MustCompile(`^/(?:forms|api/v1/forms)/([^/]+)(?:/.*)?$`)

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "web form route",
			path:     "/forms/test-form-123",
			expected: "test-form-123",
		},
		{
			name:     "API form route",
			path:     "/api/v1/forms/test-form-456",
			expected: "test-form-456",
		},
		{
			name:     "form route with trailing slash",
			path:     "/forms/test-form-789/",
			expected: "test-form-789",
		},
		{
			name:     "API form route with trailing slash",
			path:     "/api/v1/forms/test-form-abc/",
			expected: "test-form-abc",
		},
		{
			name:     "non-form route",
			path:     "/api/health",
			expected: "",
		},
		{
			name:     "root path",
			path:     "/",
			expected: "",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formID := middleware.ExtractFormID(tt.path, formRouteRegex)
			assert.Equal(t, tt.expected, formID)
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
		{
			name:           "wildcard allows all",
			origin:         "https://example.com",
			allowedOrigins: []string{"*"},
			expected:       true,
		},
		{
			name:           "exact match",
			origin:         "https://example.com",
			allowedOrigins: []string{"https://example.com"},
			expected:       true,
		},
		{
			name:           "no match",
			origin:         "https://example.com",
			allowedOrigins: []string{"https://other.com"},
			expected:       false,
		},
		{
			name:           "empty origin allowed",
			origin:         "",
			allowedOrigins: []string{"https://example.com"},
			expected:       true,
		},
		{
			name:           "multiple origins",
			origin:         "https://example.com",
			allowedOrigins: []string{"https://other.com", "https://example.com"},
			expected:       true,
		},
		{
			name:           "case sensitive",
			origin:         "https://Example.com",
			allowedOrigins: []string{"https://example.com"},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware.IsOriginAllowed(tt.origin, tt.allowedOrigins)
			assert.Equal(t, tt.expected, result)
		})
	}
}
