package handlers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMarketingHandler(t *testing.T) {
	// Create a logger instance
	logger, _ := zap.NewDevelopment()

	// Parse the templates
	templates, err := template.ParseGlob("../../templates/*.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	// Create a new marketing handler
	handler := NewMarketingHandler(logger, templates)

	// Set up Echo
	e := echo.New()
	handler.Register(e)

	// Test the home page
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Welcome to the Home Page")

	// Test the contact page
	req = httptest.NewRequest(http.MethodGet, "/contact", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Contact Us")
}
