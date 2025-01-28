package response

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	logger := zap.NewExample().Sugar()

	t.Run("NotFound", func(t *testing.T) {
		err := NotFound(c, "test not found")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Success with data", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		err := Success(c, data)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Success without data", func(t *testing.T) {
		err := Success(c, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Created with data", func(t *testing.T) {
		data := map[string]string{"id": "123"}
		err := Created(c, data)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("BadRequest with message", func(t *testing.T) {
		err := BadRequest(c, "invalid input")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("InternalError with error", func(t *testing.T) {
		testErr := errors.New("test error")
		err := InternalError(c, testErr.Error())
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("getLogger with context logger", func(t *testing.T) {
		c.Set("logger", logger)
		l := getLogger(c)
		assert.NotNil(t, l)
	})

	t.Run("getLogger without context logger", func(t *testing.T) {
		c := e.NewContext(req, rec)
		l := getLogger(c)
		assert.NotNil(t, l)
	})
}
