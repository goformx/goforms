package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/v4"
)

// NewJSONRequest creates a new JSON request for testing
func NewJSONRequest(method, path string, body interface{}) (*http.Request, error) {
	var jsonBody []byte
	var err error
	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return req, nil
}

// ParseJSONResponse parses a JSON response from a test recorder
func ParseJSONResponse(rec *httptest.ResponseRecorder, v interface{}) error {
	return json.NewDecoder(rec.Body).Decode(v)
}

// NewTestContext creates a new Echo context for testing
func NewTestContext(e *echo.Echo, req *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}
