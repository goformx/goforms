package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/goforms/internal/application/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		wantCode int
		wantBody map[string]any
	}{
		{
			name:     "success",
			data:     "test",
			wantCode: http.StatusOK,
			wantBody: map[string]any{
				"success": true,
				"data":    "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := response.Success(c, tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCode, rec.Code)

			var gotBody map[string]any
			err = json.NewDecoder(rec.Body).Decode(&gotBody)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantBody, gotBody)
		})
	}
}
