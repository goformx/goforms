package handlers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTemplate_Render(t *testing.T) {
	tmpl := template.Must(template.New("base").Parse(`{{ define "base" }}<h1>{{.Title}}</h1>{{ end }}`))
	renderer := &Template{templates: tmpl}
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(nil, rec)
	data := map[string]interface{}{
		"Title": "Test Title",
	}

	err := renderer.Render(rec, "test", data, c)
	assert.NoError(t, err)
	assert.Equal(t, "<h1>Test Title</h1>", strings.TrimSpace(rec.Body.String()))
}

func setupTestMarketingHandler(t *testing.T) (*MarketingHandler, *echo.Echo) {
	logger, _ := zap.NewDevelopment()
	templates := template.Must(template.ParseGlob("testdata/templates/*.html"))
	handler := &MarketingHandler{
		logger:    logger,
		templates: templates,
	}
	e := echo.New()
	e.Renderer = &Template{templates: templates}
	return handler, e
}

func TestNewMarketingHandler(t *testing.T) {
	handler, _ := setupTestMarketingHandler(t)
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.templates)
	assert.NotNil(t, handler.logger)
}

func TestMarketingHandler_HomePage(t *testing.T) {
	handler, e := setupTestMarketingHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.HomePage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Modern Form Handling")
}

func TestMarketingHandler_ContactPage(t *testing.T) {
	handler, e := setupTestMarketingHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/contact", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ContactPage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Contact Form Demo")
}

func TestMarketingHandler_Register(t *testing.T) {
	handler, e := setupTestMarketingHandler(t)
	handler.Register(e)

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

	require.True(t, foundHome, "Home route should be registered")
	require.True(t, foundContact, "Contact route should be registered")
}
