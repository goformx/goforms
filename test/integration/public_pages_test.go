package integration_test

import (
	"embed"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/goformx/goforms/internal/application"
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/infrastructure"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/presentation"
)

//go:embed dist
var testDistFS embed.FS

// createTestAppWithEcho creates a test fx app and returns the app and the captured *echo.Echo instance
func createTestAppWithEcho(t *testing.T) (*fxtest.App, *echo.Echo) {
	var echoInstance *echo.Echo

	testConfig := &config.Config{
		App: config.AppConfig{
			Name:            "test-app",
			Version:         "0.0.1-test",
			Environment:     "test",
			Debug:           true,
			LogLevel:        "debug",
			URL:             "http://localhost:8080",
			Scheme:          "http",
			Port:            8080,
			Host:            "localhost",
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			IdleTimeout:     5 * time.Second,
			RequestTimeout:  5 * time.Second,
			ShutdownTimeout: 5 * time.Second,
		},
		Database: config.DatabaseConfig{
			Driver:          "postgres",
			Host:            "localhost",
			Port:            5432,
			Name:            "testdb",
			Username:        "testuser",
			Password:        "testpass",
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 5 * time.Minute,
			SSLMode:         "disable",
		},
		Security: config.SecurityConfig{
			CSRF: config.CSRFConfig{
				Enabled:        false, // Disable CSRF for testing
				Secret:         "test-secret-key-for-testing-only",
				TokenName:      "_csrf",
				HeaderName:     "X-Csrf-Token",
				TokenLength:    32,
				ContextKey:     "csrf",
				CookieName:     "_csrf",
				CookiePath:     "/",
				CookieHTTPOnly: true,
				CookieSameSite: "Lax",
				CookieMaxAge:   86400,
			},
			CORS: config.CORSConfig{
				Enabled:          true,
				AllowedOrigins:   []string{"http://localhost"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
				AllowedHeaders:   []string{"*"},
				AllowCredentials: false,
			},
			RateLimit: config.RateLimitConfig{
				Enabled: false, // Disable rate limiting for testing
				RPS:     100,
				Burst:   10,
				Window:  60,
			},
		},
		Session: config.SessionConfig{
			Type:   "memory", // Use memory sessions for testing
			Secret: "test-session-secret-key-for-testing-only",
		},
	}

	testModules := []fx.Option{
		fx.Provide(func() *config.Config { return testConfig }),
		fx.Provide(func() embed.FS { return testDistFS }),
		infrastructure.Module,
		domain.Module,
		application.Module,
		middleware.Module, // Add middleware module
		presentation.Module,
		// Add the presentation lifecycle setup that registers routes
		fx.Invoke(presentation.RegisterRoutes),
		fx.Invoke(func(e *echo.Echo) {
			echoInstance = e
		}),
	}

	app := fxtest.New(t, testModules...)
	return app, echoInstance
}

// TestPublicPagesAccessibility tests that all public pages are accessible
// and return proper HTML content with correct status codes
func TestPublicPagesAccessibility(t *testing.T) {
	app, echoInstance := createTestAppWithEcho(t)
	app.RequireStart()
	defer app.RequireStop()

	// Add a minimal test route to check if Echo is working
	echoInstance.GET("/test-alive", func(c echo.Context) error {
		return c.String(200, "alive")
	})
	req := httptest.NewRequest("GET", "/test-alive", nil)
	rec := httptest.NewRecorder()
	echoInstance.ServeHTTP(rec, req)
	t.Logf("Test route status: %d, body: %s", rec.Code, rec.Body.String())
	require.Equal(t, 200, rec.Code, "Echo instance should respond to /test-alive")
	require.Equal(t, "alive", rec.Body.String())

	// Define all public pages that should be accessible without authentication
	publicPages := []struct {
		name            string
		path            string
		method          string
		expectedStatus  int
		expectedContent string
		description     string
		critical        bool
	}{
		{
			name:            "home page",
			path:            "/",
			method:          "GET",
			expectedStatus:  http.StatusOK,
			expectedContent: "text/html",
			description:     "Home page must be accessible and return HTML",
			critical:        true,
		},
		{
			name:            "login page",
			path:            "/login",
			method:          "GET",
			expectedStatus:  http.StatusOK,
			expectedContent: "text/html",
			description:     "Login page must be accessible and return HTML",
			critical:        true,
		},
		{
			name:            "signup page",
			path:            "/signup",
			method:          "GET",
			expectedStatus:  http.StatusOK,
			expectedContent: "text/html",
			description:     "Signup page must be accessible and return HTML",
			critical:        true,
		},
		{
			name:            "demo page",
			path:            "/demo",
			method:          "GET",
			expectedStatus:  http.StatusOK,
			expectedContent: "text/html",
			description:     "Demo page must be accessible and return HTML",
			critical:        false,
		},
		{
			name:            "health check endpoint",
			path:            "/health",
			method:          "GET",
			expectedStatus:  http.StatusOK,
			expectedContent: "application/json",
			description:     "Health check endpoint must be accessible",
			critical:        true,
		},
		{
			name:            "favicon",
			path:            "/favicon.ico",
			method:          "GET",
			expectedStatus:  http.StatusOK,
			expectedContent: "image/x-icon",
			description:     "Favicon must be accessible",
			critical:        false,
		},
		{
			name:            "robots.txt",
			path:            "/robots.txt",
			method:          "GET",
			expectedStatus:  http.StatusOK,
			expectedContent: "text/plain",
			description:     "Robots.txt must be accessible",
			critical:        false,
		},
		{
			name:            "API health endpoint",
			path:            "/api/v1/health",
			method:          "GET",
			expectedStatus:  http.StatusOK,
			expectedContent: "application/json",
			description:     "API health endpoint must be accessible",
			critical:        true,
		},
	}

	// Test each public page
	for _, page := range publicPages {
		t.Run(page.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(page.method, page.path, http.NoBody)
			rec := httptest.NewRecorder()

			// Handle the request
			echoInstance.ServeHTTP(rec, req)

			// Assert status code
			assert.Equal(t, page.expectedStatus, rec.Code,
				"Expected status %d for %s, got %d", page.expectedStatus, page.path, rec.Code)

			// Assert content type
			contentType := rec.Header().Get("Content-Type")
			if page.expectedContent != "" {
				assert.Contains(t, contentType, page.expectedContent,
					"Expected content type to contain '%s' for %s, got '%s'",
					page.expectedContent, page.path, contentType)
			}

			// For HTML pages, verify they contain basic HTML structure
			if strings.Contains(page.expectedContent, "text/html") {
				body := rec.Body.String()
				assert.Contains(t, body, "<!DOCTYPE html>",
					"HTML page should contain DOCTYPE declaration")
				assert.Contains(t, body, "<html",
					"HTML page should contain html tag")
				assert.Contains(t, body, "</html>",
					"HTML page should contain closing html tag")
			}

			// For JSON endpoints, verify they return valid JSON
			if strings.Contains(page.expectedContent, "application/json") {
				body := rec.Body.String()
				assert.True(t, strings.HasPrefix(body, "{") || strings.HasPrefix(body, "["),
					"JSON endpoint should return valid JSON structure")
			}

			// Log test results
			t.Logf("✓ %s: %s %s", page.description, page.method, page.path)
			t.Logf("  Status: %d", rec.Code)
			t.Logf("  Content-Type: %s", contentType)

			if page.critical {
				t.Logf("  Critical: Yes")
			}
		})
	}
}

// TestPublicPagesSecurityHeaders tests that public pages return appropriate security headers
func TestPublicPagesSecurityHeaders(t *testing.T) {
	app, echoInstance := createTestAppWithEcho(t)
	app.RequireStart()
	defer app.RequireStop()

	// Test security headers on public pages
	securityTests := []struct {
		name     string
		path     string
		headers  map[string]string
		critical bool
	}{
		{
			name: "home page security headers",
			path: "/",
			headers: map[string]string{
				"X-Content-Type-Options": "nosniff",
			},
			critical: true,
		},
		{
			name: "login page security headers",
			path: "/login",
			headers: map[string]string{
				"X-Content-Type-Options": "nosniff",
			},
			critical: true,
		},
		{
			name: "health endpoint security headers",
			path: "/health",
			headers: map[string]string{
				"X-Content-Type-Options": "nosniff",
			},
			critical: true,
		},
	}

	for _, test := range securityTests {
		t.Run(test.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", test.path, http.NoBody)
			rec := httptest.NewRecorder()

			// Handle the request
			echoInstance.ServeHTTP(rec, req)

			// Check security headers
			for headerName, expectedValue := range test.headers {
				actualValue := rec.Header().Get(headerName)
				assert.Equal(t, expectedValue, actualValue,
					"Expected header %s to be '%s' for %s, got '%s'",
					headerName, expectedValue, test.path, actualValue)
			}

			t.Logf("✓ Security headers verified for %s", test.path)
		})
	}
}

// TestPublicPagesErrorHandling tests error handling for public pages
func TestPublicPagesErrorHandling(t *testing.T) {
	app, echoInstance := createTestAppWithEcho(t)
	app.RequireStart()
	defer app.RequireStop()

	// Test error handling
	errorTests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
		description    string
	}{
		{
			name:           "404 for non-existent page",
			path:           "/non-existent-page",
			method:         "GET",
			expectedStatus: http.StatusNotFound,
			description:    "Non-existent pages should return 404",
		},
		{
			name:           "404 for non-existent API endpoint",
			path:           "/api/v1/non-existent",
			method:         "GET",
			expectedStatus: http.StatusNotFound,
			description:    "Non-existent API endpoints should return 404",
		},
	}

	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(test.method, test.path, http.NoBody)
			rec := httptest.NewRecorder()

			// Handle the request
			echoInstance.ServeHTTP(rec, req)

			// Assert status code
			assert.Equal(t, test.expectedStatus, rec.Code,
				"Expected status %d for %s %s, got %d",
				test.expectedStatus, test.method, test.path, rec.Code)

			t.Logf("✓ %s: %s %s", test.description, test.method, test.path)
		})
	}
}

// TestPublicPagesPerformance tests basic performance characteristics of public pages
func TestPublicPagesPerformance(t *testing.T) {
	app, echoInstance := createTestAppWithEcho(t)
	app.RequireStart()
	defer app.RequireStop()

	// Test performance for critical pages
	performanceTests := []struct {
		name        string
		path        string
		maxDuration time.Duration
		description string
	}{
		{
			name:        "home page performance",
			path:        "/",
			maxDuration: 100 * time.Millisecond,
			description: "Home page should respond quickly",
		},
		{
			name:        "login page performance",
			path:        "/login",
			maxDuration: 100 * time.Millisecond,
			description: "Login page should respond quickly",
		},
		{
			name:        "health endpoint performance",
			path:        "/health",
			maxDuration: 50 * time.Millisecond,
			description: "Health endpoint should respond very quickly",
		},
	}

	for _, test := range performanceTests {
		t.Run(test.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", test.path, http.NoBody)
			rec := httptest.NewRecorder()

			// Measure response time
			start := time.Now()
			echoInstance.ServeHTTP(rec, req)
			duration := time.Since(start)

			// Assert response time
			assert.LessOrEqual(t, duration, test.maxDuration,
				"Expected response time <= %v for %s, got %v",
				test.maxDuration, test.path, duration)

			// Assert successful response
			assert.Equal(t, http.StatusOK, rec.Code,
				"Expected status 200 for %s, got %d", test.path, rec.Code)

			t.Logf("✓ %s: %s (response time: %v)", test.description, test.path, duration)
		})
	}
}
