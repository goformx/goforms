package response

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

func TestResponse(t *testing.T) {
	e := echo.New()
	logger := logging.NewTestLogger()

	setupTest := func() (echo.Context, *httptest.ResponseRecorder) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("logger", logger)
		return c, rec
	}

	t.Run("NotFound", func(t *testing.T) {
		c, rec := setupTest()
		err := NotFound(c, "test not found")
		if err != nil {
			t.Errorf("NotFound() error = %v, want nil", err)
		}
		if rec.Code != http.StatusNotFound {
			t.Errorf("NotFound() status = %v, want %v", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("Success with data", func(t *testing.T) {
		c, rec := setupTest()
		data := map[string]string{"key": "value"}
		err := Success(c, data)
		if err != nil {
			t.Errorf("Success() error = %v, want nil", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("Success() status = %v, want %v", rec.Code, http.StatusOK)
		}
	})

	t.Run("Success without data", func(t *testing.T) {
		c, rec := setupTest()
		err := Success(c, nil)
		if err != nil {
			t.Errorf("Success() error = %v, want nil", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("Success() status = %v, want %v", rec.Code, http.StatusOK)
		}
	})

	t.Run("Created with data", func(t *testing.T) {
		c, rec := setupTest()
		data := map[string]string{"id": "123"}
		err := Created(c, data)
		if err != nil {
			t.Errorf("Created() error = %v, want nil", err)
		}
		if rec.Code != http.StatusCreated {
			t.Errorf("Created() status = %v, want %v", rec.Code, http.StatusCreated)
		}
	})

	t.Run("BadRequest with message", func(t *testing.T) {
		c, rec := setupTest()
		err := BadRequest(c, "invalid input")
		if err != nil {
			t.Errorf("BadRequest() error = %v, want nil", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("BadRequest() status = %v, want %v", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("InternalError with error", func(t *testing.T) {
		c, rec := setupTest()
		testErr := errors.New("test error")
		err := InternalError(c, testErr.Error())
		if err != nil {
			t.Errorf("InternalError() error = %v, want nil", err)
		}
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("InternalError() status = %v, want %v", rec.Code, http.StatusInternalServerError)
		}
	})

	t.Run("getLogger with context logger", func(t *testing.T) {
		c, _ := setupTest()
		l := getLogger(c)
		if l == nil {
			t.Error("getLogger() returned nil, want logger")
		}
		if _, ok := l.(logging.Logger); !ok {
			t.Error("getLogger() returned logger that does not implement logging.Logger interface")
		}
	})

	t.Run("getLogger without context logger", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		l := getLogger(c)
		if l == nil {
			t.Error("getLogger() returned nil, want logger")
		}
		if _, ok := l.(logging.Logger); !ok {
			t.Error("getLogger() returned logger that does not implement logging.Logger interface")
		}
	})
}
