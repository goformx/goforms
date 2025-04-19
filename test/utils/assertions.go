package utils

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// AssertJSONResponse asserts common JSON response properties
func AssertJSONResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	require.Equal(t, expectedStatus, rec.Code)
	require.Contains(t, rec.Header().Get("Content-Type"), "application/json")
}

// AssertErrorResponse asserts error response properties
func AssertErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int, expectedError string) {
	t.Helper()
	AssertJSONResponse(t, rec, expectedStatus)
	var response map[string]any
	err := ParseJSONResponse(rec, &response)
	require.NoError(t, err)
	require.Contains(t, response, "error")
	if expectedError != "" {
		require.Equal(t, expectedError, response["error"])
	}
}

// AssertSuccessResponse asserts success response properties
func AssertSuccessResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	AssertJSONResponse(t, rec, expectedStatus)
	var response map[string]any
	err := ParseJSONResponse(rec, &response)
	require.NoError(t, err)
	require.Contains(t, response, "data")
}

func AssertResponseCode(t *testing.T, rec *httptest.ResponseRecorder, expectedCode int) {
	require.Equal(t, expectedCode, rec.Code)
}

func AssertResponseBody(t *testing.T, rec *httptest.ResponseRecorder, expected interface{}) {
	var actual interface{}
	err := json.NewDecoder(rec.Body).Decode(&actual)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func AssertResponseHeader(t *testing.T, rec *httptest.ResponseRecorder, key, expectedValue string) {
	require.Equal(t, expectedValue, rec.Header().Get(key))
}

func AssertNoError(t *testing.T, err error) {
	require.NoError(t, err)
}

func AssertError(t *testing.T, err error) {
	require.Error(t, err)
}
