package openapi

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
)

// responseCapture implements ResponseCapture
type responseCapture struct{}

// NewResponseCapture creates a new response capture
func NewResponseCapture() ResponseCapture {
	return &responseCapture{}
}

// Setup sets up response capture for validation
func (rc *responseCapture) Setup(c echo.Context) *CapturedResponse {
	responseBody := make([]byte, 0)
	originalWriter := c.Response().Writer

	responseWriter := &responseCaptureWriter{
		Response: c.Response(),
		Writer:   originalWriter,
		body:     &responseBody,
	}
	c.Response().Writer = responseWriter

	return &CapturedResponse{
		Body:           &responseBody,
		OriginalWriter: originalWriter,
	}
}

// Restore restores the original response writer
func (rc *responseCapture) Restore(c echo.Context, capture *CapturedResponse) {
	if capture != nil && capture.OriginalWriter != nil {
		c.Response().Writer = capture.OriginalWriter
	}
}

// responseCaptureWriter captures the response body for validation
type responseCaptureWriter struct {
	Response *echo.Response
	Writer   http.ResponseWriter
	body     *[]byte
}

// Write captures the response body
func (w *responseCaptureWriter) Write(b []byte) (int, error) {
	*w.body = append(*w.body, b...)

	n, err := w.Writer.Write(b)
	if err != nil {
		return n, fmt.Errorf("failed to write response: %w", err)
	}

	return n, nil
}

// WriteHeader captures the response status
func (w *responseCaptureWriter) WriteHeader(statusCode int) {
	w.Response.Status = statusCode
	w.Writer.WriteHeader(statusCode)
}

// Header returns the response headers
func (w *responseCaptureWriter) Header() http.Header {
	return w.Writer.Header()
}

// Flush flushes the underlying writer
func (w *responseCaptureWriter) Flush() {
	if flusher, ok := w.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack hijacks the connection
func (w *responseCaptureWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.Writer.(http.Hijacker); ok {
		conn, rw, err := hijacker.Hijack()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to hijack connection: %w", err)
		}

		return conn, rw, nil
	}

	return nil, nil, fmt.Errorf("underlying writer does not implement http.Hijacker")
}
