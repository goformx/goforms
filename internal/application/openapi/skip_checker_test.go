package openapi_test

import (
	"testing"

	"github.com/goformx/goforms/internal/application/openapi"
	"github.com/stretchr/testify/assert"
)

func TestNewSkipConditionChecker(t *testing.T) {
	config := &openapi.Config{
		SkipPaths:   []string{"/health", "/metrics"},
		SkipMethods: []string{"OPTIONS", "HEAD"},
	}

	checker := openapi.NewSkipConditionChecker(config)

	assert.NotNil(t, checker)
	assert.Implements(t, (*openapi.SkipConditionChecker)(nil), checker)
}

func TestSkipConditionChecker_ShouldSkip_ByPath(t *testing.T) {
	config := &openapi.Config{
		SkipPaths: []string{"/health", "/metrics", "/api/v1/docs"},
	}

	checker := openapi.NewSkipConditionChecker(config)

	// Test exact path matches
	assert.True(t, checker.ShouldSkip("/health", "GET"))
	assert.True(t, checker.ShouldSkip("/metrics", "GET"))
	assert.True(t, checker.ShouldSkip("/api/v1/docs", "GET"))

	// Test path prefixes
	assert.True(t, checker.ShouldSkip("/health/status", "GET"))
	assert.True(t, checker.ShouldSkip("/metrics/prometheus", "GET"))
	assert.True(t, checker.ShouldSkip("/api/v1/docs/swagger.json", "GET"))

	// Test non-matching paths
	assert.False(t, checker.ShouldSkip("/api/users", "GET"))
	assert.False(t, checker.ShouldSkip("/healthcheck", "GET")) // Different path
	assert.False(t, checker.ShouldSkip("/", "GET"))
}

func TestSkipConditionChecker_ShouldSkip_ByMethod(t *testing.T) {
	config := &openapi.Config{
		SkipMethods: []string{"OPTIONS", "HEAD", "TRACE"},
	}

	checker := openapi.NewSkipConditionChecker(config)

	// Test method matches
	assert.True(t, checker.ShouldSkip("/api/users", "OPTIONS"))
	assert.True(t, checker.ShouldSkip("/api/users", "HEAD"))
	assert.True(t, checker.ShouldSkip("/api/users", "TRACE"))

	// Test non-matching methods
	assert.False(t, checker.ShouldSkip("/api/users", "GET"))
	assert.False(t, checker.ShouldSkip("/api/users", "POST"))
	assert.False(t, checker.ShouldSkip("/api/users", "PUT"))
	assert.False(t, checker.ShouldSkip("/api/users", "DELETE"))
}

func TestSkipConditionChecker_ShouldSkip_ByPathAndMethod(t *testing.T) {
	config := &openapi.Config{
		SkipPaths:   []string{"/health", "/metrics"},
		SkipMethods: []string{"OPTIONS", "HEAD"},
	}

	checker := openapi.NewSkipConditionChecker(config)

	// Test path match (should skip regardless of method)
	assert.True(t, checker.ShouldSkip("/health", "GET"))
	assert.True(t, checker.ShouldSkip("/health", "POST"))
	assert.True(t, checker.ShouldSkip("/health", "OPTIONS"))

	// Test method match (should skip regardless of path)
	assert.True(t, checker.ShouldSkip("/api/users", "OPTIONS"))
	assert.True(t, checker.ShouldSkip("/api/users", "HEAD"))
	assert.True(t, checker.ShouldSkip("/", "OPTIONS"))

	// Test both path and method match
	assert.True(t, checker.ShouldSkip("/health", "OPTIONS"))

	// Test no matches
	assert.False(t, checker.ShouldSkip("/api/users", "GET"))
	assert.False(t, checker.ShouldSkip("/api/users", "POST"))
}

func TestSkipConditionChecker_ShouldSkip_EmptyConfig(t *testing.T) {
	config := &openapi.Config{
		SkipPaths:   []string{},
		SkipMethods: []string{},
	}

	checker := openapi.NewSkipConditionChecker(config)

	// Should not skip anything
	assert.False(t, checker.ShouldSkip("/api/users", "GET"))
	assert.False(t, checker.ShouldSkip("/health", "GET"))
	assert.False(t, checker.ShouldSkip("/", "OPTIONS"))
}

func TestSkipConditionChecker_ShouldSkip_NilConfig(t *testing.T) {
	config := &openapi.Config{
		SkipPaths:   nil,
		SkipMethods: nil,
	}

	checker := openapi.NewSkipConditionChecker(config)

	// Should not skip anything
	assert.False(t, checker.ShouldSkip("/api/users", "GET"))
	assert.False(t, checker.ShouldSkip("/health", "GET"))
	assert.False(t, checker.ShouldSkip("/", "OPTIONS"))
}

func TestSkipConditionChecker_ShouldSkip_CaseSensitive(t *testing.T) {
	config := &openapi.Config{
		SkipPaths:   []string{"/Health", "/Metrics"},
		SkipMethods: []string{"Options", "Head"},
	}

	checker := openapi.NewSkipConditionChecker(config)

	// Test case sensitivity for paths
	assert.False(t, checker.ShouldSkip("/health", "GET"))  // Different case
	assert.False(t, checker.ShouldSkip("/metrics", "GET")) // Different case
	assert.True(t, checker.ShouldSkip("/Health", "GET"))   // Exact match
	assert.True(t, checker.ShouldSkip("/Metrics", "GET"))  // Exact match

	// Test case sensitivity for methods
	assert.False(t, checker.ShouldSkip("/api/users", "options")) // Different case
	assert.False(t, checker.ShouldSkip("/api/users", "head"))    // Different case
	assert.True(t, checker.ShouldSkip("/api/users", "Options"))  // Exact match
	assert.True(t, checker.ShouldSkip("/api/users", "Head"))     // Exact match
}

func TestSkipConditionChecker_ShouldSkip_EmptyPaths(t *testing.T) {
	config := &openapi.Config{
		SkipPaths:   []string{""},
		SkipMethods: []string{},
	}

	checker := openapi.NewSkipConditionChecker(config)

	// Empty path should match everything
	assert.True(t, checker.ShouldSkip("/api/users", "GET"))
	assert.True(t, checker.ShouldSkip("/health", "GET"))
	assert.True(t, checker.ShouldSkip("/", "GET"))
}

func TestSkipConditionChecker_ShouldSkip_EmptyMethods(t *testing.T) {
	config := &openapi.Config{
		SkipPaths:   []string{},
		SkipMethods: []string{""},
	}

	checker := openapi.NewSkipConditionChecker(config)

	// Empty method should not match anything (empty string != any method)
	assert.False(t, checker.ShouldSkip("/api/users", "GET"))
	assert.False(t, checker.ShouldSkip("/api/users", "POST"))
	assert.False(t, checker.ShouldSkip("/api/users", ""))
}

func TestSkipConditionChecker_ShouldSkip_ComplexPaths(t *testing.T) {
	config := &openapi.Config{
		SkipPaths: []string{
			"/api/v1",
			"/admin",
			"/static",
			"/docs",
		},
	}

	checker := openapi.NewSkipConditionChecker(config)

	// Test various path scenarios
	assert.True(t, checker.ShouldSkip("/api/v1/users", "GET"))
	assert.True(t, checker.ShouldSkip("/api/v1/users/123", "GET"))
	assert.True(t, checker.ShouldSkip("/admin/dashboard", "GET"))
	assert.True(t, checker.ShouldSkip("/static/css/style.css", "GET"))
	assert.True(t, checker.ShouldSkip("/docs/api", "GET"))

	// Test non-matching paths
	assert.False(t, checker.ShouldSkip("/api/v2/users", "GET"))
	assert.False(t, checker.ShouldSkip("/api/users", "GET"))
	assert.False(t, checker.ShouldSkip("/user/admin", "GET"))
	assert.False(t, checker.ShouldSkip("/staticfiles/css/style.css", "GET"))
}

func TestSkipConditionChecker_ShouldSkip_EdgeCases(t *testing.T) {
	config := &openapi.Config{
		SkipPaths:   []string{"/", "/api"},
		SkipMethods: []string{"GET", "POST"},
	}

	checker := openapi.NewSkipConditionChecker(config)

	// Test root path
	assert.True(t, checker.ShouldSkip("/", "GET"))
	assert.True(t, checker.ShouldSkip("/", "POST"))

	// Test API prefix
	assert.True(t, checker.ShouldSkip("/api", "GET"))
	assert.True(t, checker.ShouldSkip("/api/users", "GET"))
	assert.True(t, checker.ShouldSkip("/api/v1/users", "GET"))

	// Test methods
	result1 := checker.ShouldSkip("/users", "GET")
	t.Logf("ShouldSkip('/users', 'GET') = %v", result1)
	assert.True(t, result1)

	result2 := checker.ShouldSkip("/users", "POST")
	t.Logf("ShouldSkip('/users', 'POST') = %v", result2)
	assert.True(t, result2)

	assert.False(t, checker.ShouldSkip("/users", "PUT"))
	assert.False(t, checker.ShouldSkip("/users", "DELETE"))
}

func TestSkipConditionChecker_Debug(t *testing.T) {
	config := &openapi.Config{
		SkipPaths:   []string{"/", "/api"},
		SkipMethods: []string{"GET", "POST"},
	}

	checker := openapi.NewSkipConditionChecker(config)

	// Debug: Check what happens with /users path and GET method
	result := checker.ShouldSkip("/users", "GET")
	t.Logf("ShouldSkip('/users', 'GET') = %v", result)

	// Debug: Check what happens with /users path and POST method
	result2 := checker.ShouldSkip("/users", "POST")
	t.Logf("ShouldSkip('/users', 'POST') = %v", result2)

	// Debug: Check what happens with / path and GET method
	result3 := checker.ShouldSkip("/", "GET")
	t.Logf("ShouldSkip('/', 'GET') = %v", result3)
}
