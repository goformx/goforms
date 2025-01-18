package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestMarketingHandler(t *testing.T) {
	// Create a new marketing handler
	handler := NewMarketingHandler()

	// Set up Echo
	e := echo.New()
	handler.Register(e)

	// Test the home page
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Welcome to Goforms")

	// Test the contact page
	req = httptest.NewRequest(http.MethodGet, "/contact", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Contact Form Demo")
}
