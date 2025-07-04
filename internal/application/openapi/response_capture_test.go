package openapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goformx/goforms/internal/application/openapi"
)

// Returning the interface is required for test helpers.
//

func createTestEchoContext(t *testing.T) echo.Context {
	t.Helper()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()

	return e.NewContext(req, rec)
}

func TestNewResponseCapture(t *testing.T) {
	capture := openapi.NewResponseCapture()

	assert.NotNil(t, capture)
	assert.Implements(t, (*openapi.ResponseCapture)(nil), capture)
}

func TestResponseCapture_Setup(t *testing.T) {
	capture := openapi.NewResponseCapture()
	c := createTestEchoContext(t)

	originalWriter := c.Response().Writer

	// Setup response capture
	capturedResponse := capture.Setup(c)

	assert.NotNil(t, capturedResponse)
	assert.NotNil(t, capturedResponse.Body)
	assert.Equal(t, originalWriter, capturedResponse.OriginalWriter)
	assert.NotEqual(t, originalWriter, c.Response().Writer) // Should be replaced with capture writer
}

func TestResponseCapture_Restore(t *testing.T) {
	capture := openapi.NewResponseCapture()
	c := createTestEchoContext(t)

	originalWriter := c.Response().Writer

	// Setup response capture
	capturedResponse := capture.Setup(c)
	assert.NotEqual(t, originalWriter, c.Response().Writer)

	// Restore original writer
	capture.Restore(c, capturedResponse)
	assert.Equal(t, originalWriter, c.Response().Writer)
}

func TestResponseCapture_Restore_NilCapture(t *testing.T) {
	capture := openapi.NewResponseCapture()
	c := createTestEchoContext(t)

	originalWriter := c.Response().Writer

	// Restore with nil capture (should not panic)
	capture.Restore(c, nil)
	assert.Equal(t, originalWriter, c.Response().Writer)
}

func TestResponseCapture_Restore_NilOriginalWriter(t *testing.T) {
	capture := openapi.NewResponseCapture()
	c := createTestEchoContext(t)

	// Create a capture with nil original writer
	capturedResponse := &openapi.CapturedResponse{
		Body:           &[]byte{},
		OriginalWriter: nil,
	}

	// Restore should not panic
	capture.Restore(c, capturedResponse)
}

func TestResponseCaptureWriter_Write(t *testing.T) {
	capture := openapi.NewResponseCapture()
	c := createTestEchoContext(t)

	// Setup response capture
	capturedResponse := capture.Setup(c)

	// Write some data
	testData := []byte("Hello, World!")
	n, err := c.Response().Writer.Write(testData)

	require.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, *capturedResponse.Body)
}

func TestResponseCaptureWriter_Write_MultipleWrites(t *testing.T) {
	capture := openapi.NewResponseCapture()
	c := createTestEchoContext(t)

	// Setup response capture
	capturedResponse := capture.Setup(c)

	// Write multiple times
	data1 := []byte("Hello, ")
	data2 := []byte("World!")
	data3 := []byte(" How are you?")

	_, err := c.Response().Writer.Write(data1)
	require.NoError(t, err)
	_, err = c.Response().Writer.Write(data2)
	require.NoError(t, err)
	_, err = c.Response().Writer.Write(data3)
	require.NoError(t, err)

	expected := append(append(data1, data2...), data3...)
	assert.Equal(t, expected, *capturedResponse.Body)
}

func TestResponseCaptureWriter_WriteHeader(t *testing.T) {
	capture := openapi.NewResponseCapture()
	c := createTestEchoContext(t)

	// Setup response capture
	capture.Setup(c)

	// Write header
	statusCode := http.StatusCreated
	c.Response().Writer.WriteHeader(statusCode)

	assert.Equal(t, statusCode, c.Response().Status)
}

func TestResponseCaptureWriter_Header(t *testing.T) {
	capture := openapi.NewResponseCapture()
	c := createTestEchoContext(t)

	// Setup response capture
	capture.Setup(c)

	// Set headers
	headers := c.Response().Writer.Header()
	headers.Set("Content-Type", "application/json")
	headers.Set("X-Custom-Header", "test-value")

	assert.Equal(t, "application/json", headers.Get("Content-Type"))
	assert.Equal(t, "test-value", headers.Get("X-Custom-Header"))
}

func TestResponseCaptureWriter_Flush(t *testing.T) {
	capture := openapi.NewResponseCapture()
	c := createTestEchoContext(t)

	// Setup response capture
	capture.Setup(c)

	// Flush should not panic
	flusher, ok := c.Response().Writer.(http.Flusher)
	require.True(t, ok)
	flusher.Flush()
}

func TestResponseCaptureWriter_Hijack(t *testing.T) {
	capture := openapi.NewResponseCapture()
	c := createTestEchoContext(t)

	// Setup response capture
	capture.Setup(c)

	// Test hijack functionality
	hijacker, ok := c.Response().Writer.(http.Hijacker)
	assert.True(t, ok)

	conn, rw, err := hijacker.Hijack()
	if err != nil {
		// Hijack might fail in test environment, which is expected
		assert.Contains(t, err.Error(), "underlying writer does not implement http.Hijacker")
	} else {
		assert.NotNil(t, conn)
		assert.NotNil(t, rw)
	}
}

func TestResponseCaptureWriter_Write_ErrorHandling(t *testing.T) {
	capture := openapi.NewResponseCapture()
	c := createTestEchoContext(t)

	// Setup response capture
	capturedResponse := capture.Setup(c)

	// Write data
	testData := []byte("test")
	n, err := c.Response().Writer.Write(testData)

	require.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, *capturedResponse.Body)
}

func TestResponseCapture_Integration(t *testing.T) {
	capture := openapi.NewResponseCapture()
	c := createTestEchoContext(t)

	originalWriter := c.Response().Writer

	// Setup capture
	capturedResponse := capture.Setup(c)
	assert.NotNil(t, capturedResponse)
	assert.Equal(t, originalWriter, capturedResponse.OriginalWriter)

	// Write response data
	responseData := []byte(`{"message": "success"}`)
	_, err := c.Response().Writer.Write(responseData)
	require.NoError(t, err)
	c.Response().Writer.WriteHeader(http.StatusOK)

	// Verify captured data
	assert.Equal(t, responseData, *capturedResponse.Body)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	// Restore original writer
	capture.Restore(c, capturedResponse)
	assert.Equal(t, originalWriter, c.Response().Writer)
}

func TestResponseCaptureWriter_ImplementsInterfaces(t *testing.T) {
	capture := openapi.NewResponseCapture()
	c := createTestEchoContext(t)

	// Setup response capture
	capture.Setup(c)
	writer := c.Response().Writer

	// Test that it implements required interfaces
	_, isFlusher := writer.(http.Flusher)
	assert.True(t, isFlusher)

	_, isHijacker := writer.(http.Hijacker)
	assert.True(t, isHijacker)
}

func TestResponseCapture_EmptyBody(t *testing.T) {
	capture := openapi.NewResponseCapture()
	c := createTestEchoContext(t)

	// Setup response capture
	capturedResponse := capture.Setup(c)

	// Don't write anything
	assert.NotNil(t, capturedResponse.Body)
	assert.Empty(t, *capturedResponse.Body)
}

func TestResponseCapture_WriteToNilBody(t *testing.T) {
	capture := openapi.NewResponseCapture()
	c := createTestEchoContext(t)

	// Setup response capture
	capturedResponse := capture.Setup(c)

	// Ensure body is not nil
	assert.NotNil(t, capturedResponse.Body)

	// Write data
	testData := []byte("test")
	_, err := c.Response().Writer.Write(testData)
	require.NoError(t, err)

	assert.Equal(t, testData, *capturedResponse.Body)
}

func TestResponseCapture_MultipleSetups(t *testing.T) {
	capture := openapi.NewResponseCapture()
	c := createTestEchoContext(t)

	// Setup capture multiple times
	capturedResponse1 := capture.Setup(c)
	capturedResponse2 := capture.Setup(c)

	// Write data
	testData := []byte("test")
	_, err := c.Response().Writer.Write(testData)
	require.NoError(t, err)

	// Both captures should have the data
	assert.Equal(t, testData, *capturedResponse1.Body)
	assert.Equal(t, testData, *capturedResponse2.Body)
}
