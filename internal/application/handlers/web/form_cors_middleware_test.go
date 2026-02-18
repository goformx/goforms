package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/config"
	mockform "github.com/goformx/goforms/test/mocks/form"
)

func TestFormCORSMiddleware_AllowsConfiguredOrigin(t *testing.T) {
	ctrl := gomock.NewController(t)
	formService := mockform.NewMockService(ctrl)

	formService.EXPECT().
		GetForm(gomock.Any(), "form-123").
		Return(&model.Form{
			CorsOrigins: model.JSON{"origins": []any{"https://allowed.example"}},
		}, nil)

	corsConfig := config.CORSConfig{
		AllowCredentials: true,
		ExposedHeaders:   []string{"X-Csrf-Token"},
		AllowedMethods:   []string{http.MethodGet, http.MethodOptions},
		AllowedHeaders:   []string{"Content-Type"},
		MaxAge:           600,
	}

	e := echo.New()
	formsAPI := e.Group(constants.PathAPIForms)
	formsAPI.Use(web.NewFormCORSMiddleware(formService, corsConfig))
	formsAPI.GET("/:id/schema", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/forms/form-123/schema", http.NoBody)
	req.Header.Set("Origin", "https://allowed.example")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "https://allowed.example", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "Origin", rec.Header().Get("Vary"))
	assert.Contains(t, rec.Header().Get("Access-Control-Expose-Headers"), "X-Csrf-Token")
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), http.MethodGet)
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	assert.Equal(t, "600", rec.Header().Get("Access-Control-Max-Age"))
}

func TestFormCORSMiddleware_DeniesUnknownOrigin(t *testing.T) {
	ctrl := gomock.NewController(t)
	formService := mockform.NewMockService(ctrl)

	formService.EXPECT().
		GetForm(gomock.Any(), "form-123").
		Return(&model.Form{
			CorsOrigins: model.JSON{"origins": []any{"https://allowed.example"}},
		}, nil)

	e := echo.New()
	formsAPI := e.Group(constants.PathAPIForms)
	formsAPI.Use(web.NewFormCORSMiddleware(formService, config.CORSConfig{}))
	formsAPI.GET("/:id/schema", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/forms/form-123/schema", http.NoBody)
	req.Header.Set("Origin", "https://blocked.example")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestFormCORSMiddleware_AllowsPreflightForSubmit(t *testing.T) {
	ctrl := gomock.NewController(t)
	formService := mockform.NewMockService(ctrl)

	formService.EXPECT().
		GetForm(gomock.Any(), "form-123").
		Return(&model.Form{
			CorsOrigins: model.JSON{"origins": []any{"https://allowed.example"}},
		}, nil)

	e := echo.New()
	formsAPI := e.Group(constants.PathAPIForms)
	formsAPI.Use(web.NewFormCORSMiddleware(formService, config.CORSConfig{}))
	formsAPI.OPTIONS("/:id/submit", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/forms/form-123/submit", http.NoBody)
	req.Header.Set("Origin", "https://allowed.example")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "https://allowed.example", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), http.MethodPost)
}

func TestFormCORSMiddleware_SkipsWhenOriginMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	formService := mockform.NewMockService(ctrl)

	e := echo.New()
	formsAPI := e.Group(constants.PathAPIForms)
	formsAPI.Use(web.NewFormCORSMiddleware(formService, config.CORSConfig{}))
	formsAPI.GET("/:id/schema", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/forms/form-123/schema", http.NoBody)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestFormCORSMiddleware_AllowsPublicFormsPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	formService := mockform.NewMockService(ctrl)

	formService.EXPECT().
		GetForm(gomock.Any(), "form-456").
		Return(&model.Form{
			CorsOrigins: model.JSON{"origins": []any{"https://embed.example"}},
		}, nil)

	e := echo.New()
	formsPublic := e.Group(constants.PathFormsPublic)
	formsPublic.Use(web.NewFormCORSMiddleware(formService, config.CORSConfig{}))
	formsPublic.GET("/:id/schema", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	formsPublic.GET("/:id/embed", func(c echo.Context) error {
		return c.String(http.StatusOK, "embed")
	})

	req := httptest.NewRequest(http.MethodGet, "/forms/form-456/schema", http.NoBody)
	req.Header.Set("Origin", "https://embed.example")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "https://embed.example", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "ok", rec.Body.String())
}
