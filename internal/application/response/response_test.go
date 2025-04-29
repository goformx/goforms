package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test success response
	err := Success(c, "value")
	assert.NoError(t, err)

	var response Response
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)
	assert.Equal(t, "value", response.Data)

	// Test error response
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = BadRequest(c, "test error")
	assert.NoError(t, err)

	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "test error", response.Message)
}

func TestCreated(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := Created(c, "value")
	assert.NoError(t, err)

	var response Response
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)
	assert.Equal(t, "value", response.Data)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestBadRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := BadRequest(c, "invalid input")
	assert.NoError(t, err)

	var response Response
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "invalid input", response.Message)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestNotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := NotFound(c, "resource not found")
	assert.NoError(t, err)

	var response Response
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "resource not found", response.Message)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestInternalError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := InternalError(c, "server error")
	assert.NoError(t, err)

	var response Response
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "server error", response.Message)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUnauthorized(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := Unauthorized(c, "unauthorized access")
	assert.NoError(t, err)

	var response Response
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "unauthorized access", response.Message)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestForbidden(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := Forbidden(c, "access denied")
	assert.NoError(t, err)

	var response Response
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "access denied", response.Message)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
