package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/goforms/internal/application/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestSuccess(t *testing.T) {
	t.Run("success response", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := response.Success(c, map[string]any{
			"status": "success",
			"data": map[string]any{
				"key": "value",
			},
		})

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]any
		err = json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)

		require.Equal(t, "success", resp["status"])
		require.Equal(t, "value", resp["data"].(map[string]any)["key"])
	})

	t.Run("error response", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := response.BadRequest(c, "test error")
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)

		var resp map[string]any
		err = json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)

		require.Equal(t, "error", resp["status"])
		require.Equal(t, "test error", resp["message"])
	})
}
