package middleware_test

import (
	"bytes"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/goformx/goforms/internal/application/middleware"
)

// Test the response implementation
//
//nolint:gocognit // Test function with multiple subtests
func TestResponseImplementation(t *testing.T) {
	// Test JSON response
	t.Run("JSON Response", func(t *testing.T) {
		data := map[string]string{"message": "test"}
		resp := middleware.NewJSONResponse(data)

		if resp.StatusCode() != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode())
		}

		if !resp.IsJSON() {
			t.Error("Expected JSON content type")
		}

		// Test WriteTo
		var buf bytes.Buffer

		bytesWritten, err := resp.WriteTo(&buf)
		if err != nil {
			t.Errorf("WriteTo failed: %v", err)
		}

		if bytesWritten == 0 {
			t.Error("Expected bytes written to be greater than 0")
		}

		// Verify response contains expected content
		if !bytes.Contains(buf.Bytes(), []byte("HTTP/1.1 200 OK")) {
			t.Error("Expected status line not found")
		}

		if !bytes.Contains(buf.Bytes(), []byte("Content-Type: application/json")) {
			t.Error("Expected content type header not found")
		}

		if !bytes.Contains(buf.Bytes(), []byte(`"message":"test"`)) {
			t.Error("Expected JSON body not found")
		}
	})

	// Test error response
	t.Run("Error Response", func(t *testing.T) {
		testErr := errors.New("test error")
		resp := middleware.NewErrorResponse(http.StatusBadRequest, testErr)

		if resp.StatusCode() != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode())
		}

		if !resp.IsError() {
			t.Error("Expected error response")
		}

		if !errors.Is(resp.Error(), testErr) {
			t.Error("Expected error to match")
		}
	})

	// Test redirect response
	t.Run("Redirect Response", func(t *testing.T) {
		location := "/new-location"
		resp := middleware.NewRedirectResponse(http.StatusFound, location)

		if resp.StatusCode() != http.StatusFound {
			t.Errorf("Expected status 302, got %d", resp.StatusCode())
		}

		if !resp.IsRedirect() {
			t.Error("Expected redirect response")
		}

		if resp.Location() != location {
			t.Errorf("Expected location %s, got %s", location, resp.Location())
		}
	})

	// Test response builder
	t.Run("Response Builder", func(t *testing.T) {
		resp := middleware.NewResponseBuilder().
			StatusCode(http.StatusCreated).
			JSON().
			BodyBytes([]byte(`{"id": 123}`)).
			Header("X-Custom", "value").
			Build()

		if resp.StatusCode() != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", resp.StatusCode())
		}

		if !resp.IsJSON() {
			t.Error("Expected JSON content type")
		}

		if resp.Headers().Get("X-Custom") != "value" {
			t.Error("Expected custom header not found")
		}
	})

	// Test response cloning
	t.Run("Response Cloning", func(t *testing.T) {
		original := middleware.NewJSONResponse(map[string]string{"test": "data"})
		original.SetHeader("X-Original", "value")

		cloned := original.Clone()

		if cloned.StatusCode() != original.StatusCode() {
			t.Error("Cloned response should have same status code")
		}

		if cloned.Headers().Get("X-Original") != "value" {
			t.Error("Cloned response should have same headers")
		}

		// Modify original should not affect clone
		original.SetHeader("X-Original", "modified")

		if cloned.Headers().Get("X-Original") == "modified" {
			t.Error("Modifying original should not affect clone")
		}
	})
}

// Benchmark response creation
func BenchmarkNewJSONResponse(b *testing.B) {
	data := map[string]any{
		"message":   "Hello, World!",
		"timestamp": time.Now().Unix(),
		"items":     []string{"item1", "item2", "item3"},
	}

	b.ResetTimer()

	for range b.N {
		resp := middleware.NewJSONResponse(data)
		if resp == nil {
			b.Fatal("Response should not be nil")
		}
	}
}

// Benchmark response writing
func BenchmarkResponseWriteTo(b *testing.B) {
	data := map[string]any{
		"message":   "Hello, World!",
		"timestamp": time.Now().Unix(),
	}

	resp := middleware.NewJSONResponse(data)
	resp.SetHeader("X-Custom-Header", "custom-value")
	resp.SetRequestID("req-123")

	var buf bytes.Buffer

	b.ResetTimer()

	for range b.N {
		buf.Reset()

		_, err := resp.WriteTo(&buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}
