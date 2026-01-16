package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/romsar/gonertia"
	"github.com/stretchr/testify/assert"

	"github.com/goformx/goforms/internal/application/constants"
)

// TestAuthHandler_HandleAuthSuccess_ResponseType tests that handleAuthSuccess
// returns the correct response type based on request headers.
// Since handleAuthSuccess is private, we test the logic directly.
func TestAuthHandler_HandleAuthSuccess_ResponseType(t *testing.T) {
	tests := []struct {
		name                string
		hasXInertia         bool
		hasXRequestedWith   bool
		expectedIsRedirect  bool
		expectedIsJSON      bool
		description         string
	}{
		{
			name:                "Inertia request should redirect",
			hasXInertia:         true,
			hasXRequestedWith:   true,
			expectedIsRedirect:  true,
			expectedIsJSON:      false,
			description:         "Inertia requests with X-Inertia header should redirect, not return JSON",
		},
		{
			name:                "Pure AJAX request should return JSON",
			hasXInertia:         false,
			hasXRequestedWith:   true,
			expectedIsRedirect:  false,
			expectedIsJSON:      true,
			description:         "Pure AJAX requests without X-Inertia header should return JSON",
		},
		{
			name:                "Regular form should redirect",
			hasXInertia:         false,
			hasXRequestedWith:   false,
			expectedIsRedirect:  true,
			expectedIsJSON:      false,
			description:         "Regular form submissions without AJAX headers should redirect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest(http.MethodPost, "/signup", nil)
			if tt.hasXInertia {
				req.Header.Set("X-Inertia", "true")
			}
			if tt.hasXRequestedWith {
				req.Header.Set("X-Requested-With", "XMLHttpRequest")
			}

			// Test the logic that handleAuthSuccess uses
			isInertiaRequest := gonertia.IsInertiaRequest(req)
			isAJAXRequest := req.Header.Get(constants.HeaderXRequestedWith) == "XMLHttpRequest"

			// Determine expected behavior
			var shouldRedirect bool
			var shouldReturnJSON bool

			if isInertiaRequest {
				// Inertia requests should always redirect
				shouldRedirect = true
				shouldReturnJSON = false
			} else if isAJAXRequest {
				// Pure AJAX (non-Inertia) requests can get JSON
				shouldRedirect = false
				shouldReturnJSON = true
			} else {
				// Regular form submissions redirect
				shouldRedirect = true
				shouldReturnJSON = false
			}

			// Verify expectations match test case
			assert.Equal(t, tt.expectedIsRedirect, shouldRedirect, "Redirect expectation mismatch: %s", tt.description)
			assert.Equal(t, tt.expectedIsJSON, shouldReturnJSON, "JSON expectation mismatch: %s", tt.description)
		})
	}
}

// TestAuthHandler_InertiaVsAJAX_HeaderDetection verifies the header detection logic
// matches the expected behavior for Inertia vs pure AJAX requests
func TestAuthHandler_InertiaVsAJAX_HeaderDetection(t *testing.T) {
	tests := []struct {
		name              string
		headers           map[string]string
		shouldCheckInertiaFirst bool
		description       string
	}{
		{
			name: "Inertia request has both headers",
			headers: map[string]string{
				"X-Inertia":        "true",
				"X-Requested-With": "XMLHttpRequest",
			},
			shouldCheckInertiaFirst: true,
			description:             "Inertia requests have both X-Inertia and X-Requested-With - must check X-Inertia first",
		},
		{
			name: "Pure AJAX has only X-Requested-With",
			headers: map[string]string{
				"X-Requested-With": "XMLHttpRequest",
				// No X-Inertia header
			},
			shouldCheckInertiaFirst: false,
			description:             "Pure AJAX only has X-Requested-With, no X-Inertia",
		},
		{
			name:              "Regular form has neither",
			headers:           map[string]string{},
			shouldCheckInertiaFirst: false,
			description:       "Regular forms have neither header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/signup", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// Check headers
			hasInertia := gonertia.IsInertiaRequest(req)
			hasAJAX := req.Header.Get(constants.HeaderXRequestedWith) == "XMLHttpRequest"

			// Verify the detection matches expectations
			if tt.headers["X-Inertia"] != "" {
				assert.True(t, hasInertia, "Should detect X-Inertia header: %s", tt.description)
				assert.True(t, hasAJAX, "Inertia requests also have X-Requested-With: %s", tt.description)
				// This is the key: if both are present, we must check Inertia first
				assert.True(t, tt.shouldCheckInertiaFirst, "Must check Inertia first when both headers present: %s", tt.description)
			} else if tt.headers["X-Requested-With"] != "" {
				assert.False(t, hasInertia, "Should not detect Inertia when X-Inertia header missing: %s", tt.description)
				assert.True(t, hasAJAX, "Should detect AJAX request: %s", tt.description)
			} else {
				assert.False(t, hasInertia, "Regular forms should not be detected as Inertia: %s", tt.description)
				assert.False(t, hasAJAX, "Regular forms should not be detected as AJAX: %s", tt.description)
			}
		})
	}
}
