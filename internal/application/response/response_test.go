package response_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/goforms/internal/application/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestResponse(t *testing.T) {
	t.Run("success response", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(nil, rec)

		err := response.Success(c, map[string]interface{}{
			"message": "success",
			"data": map[string]interface{}{
				"key": "value",
			},
		})
		require.NoError(t, err)

		var resp map[string]interface{}
		err = json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)

		require.Equal(t, "success", resp["message"])
		require.Equal(t, "value", resp["data"].(map[string]interface{})["key"])
	})

	t.Run("error response", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(nil, rec)

		err := response.BadRequest(c, "error")
		require.NoError(t, err)

		var resp map[string]interface{}
		err = json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)

		require.Equal(t, "error", resp["error"])
		require.False(t, resp["success"].(bool))
	})
}
