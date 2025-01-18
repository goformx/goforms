package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/jonesrussell/goforms/internal/logger"
)

func TestMarketingHandler_Register(t *testing.T) {
	// Setup
	mockLogger := logger.NewMockLogger()
	handler := NewMarketingHandler(mockLogger)
	e := echo.New()

	// Test
	handler.Register(e)

	// Verify routes are registered
	routes := e.Routes()
	var foundHome, foundContact bool

	for _, route := range routes {
		switch {
		case route.Path == "/" && route.Method == http.MethodGet:
			foundHome = true
		case route.Path == "/contact" && route.Method == http.MethodGet:
			foundContact = true
		}
	}

	assert.True(t, foundHome, "Home route should be registered")
	assert.True(t, foundContact, "Contact route should be registered")
	assert.Contains(t, mockLogger.DebugCalls[0].Message, "Registering marketing routes")
}

func TestMarketingHandler_HandleHome(t *testing.T) {
	// Setup
	mockLogger := logger.NewMockLogger()
	handler := NewMarketingHandler(mockLogger)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test
	err := handler.HandleHome(c)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, len(mockLogger.DebugCalls) > 0, "Expected debug log for home request")
}

func TestMarketingHandler_HandleContact(t *testing.T) {
	// Setup
	mockLogger := logger.NewMockLogger()
	handler := NewMarketingHandler(mockLogger)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/contact", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test
	err := handler.HandleContact(c)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, len(mockLogger.DebugCalls) > 0, "Expected debug log for contact request")
}
