// Package middleware provides infrastructure layer middleware adapters
// for integrating framework-agnostic middleware with specific HTTP frameworks.
package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/middleware/core"
)

// EchoResponseWrapper wraps our Response interface to work with Echo's response system.
// This adapter provides a bridge between our framework-agnostic response and Echo's
// response handling mechanisms.
type EchoResponseWrapper struct {
	response core.Response
	context  echo.Context
}

// NewEchoResponseWrapper creates a new Echo response wrapper.
func NewEchoResponseWrapper(resp core.Response, c echo.Context) *EchoResponseWrapper {
	return &EchoResponseWrapper{
		response: resp,
		context:  c,
	}
}

// ApplyToEcho applies our response to Echo's context.
// This method handles status codes, headers, cookies, and body writing.
func (w *EchoResponseWrapper) ApplyToEcho() error {
	// Set status code
	w.context.Response().Status = w.response.StatusCode()

	// Apply headers
	for key, values := range w.response.Headers() {
		for _, value := range values {
			w.context.Response().Header().Add(key, value)
		}
	}

	// Apply cookies
	for _, cookie := range w.response.Cookies() {
		w.context.SetCookie(cookie)
	}

	// Handle redirects
	if w.response.IsRedirect() && w.response.Location() != "" {
		return w.context.Redirect(w.response.StatusCode(), w.response.Location())
	}

	// Handle errors
	if w.response.IsError() {
		if w.response.Error() != nil {
			return fmt.Errorf("response error: %w", w.response.Error())
		}
		// Create HTTP error if no specific error is set
		return echo.NewHTTPError(w.response.StatusCode(), http.StatusText(w.response.StatusCode()))
	}

	// Write response body
	return w.writeBody()
}

// writeBody writes the response body to Echo's response writer.
func (w *EchoResponseWrapper) writeBody() error {
	// Handle different body types
	if w.response.Body() != nil {
		// Copy body to Echo's response writer
		if _, err := io.Copy(w.context.Response().Writer, w.response.Body()); err != nil {
			return fmt.Errorf("failed to copy response body: %w", err)
		}
	} else if w.response.BodyBytes() != nil {
		// Write bytes directly
		if _, err := w.context.Response().Writer.Write(w.response.BodyBytes()); err != nil {
			return fmt.Errorf("failed to write response body: %w", err)
		}
	}

	return nil
}

// EchoResponseBuilder provides a builder pattern for creating responses
// that can be easily applied to Echo contexts.
type EchoResponseBuilder struct {
	response core.Response
	context  echo.Context
}

// NewEchoResponseBuilder creates a new Echo response builder.
func NewEchoResponseBuilder(c echo.Context) *EchoResponseBuilder {
	return &EchoResponseBuilder{
		response: core.NewResponse(http.StatusOK),
		context:  c,
	}
}

// Status sets the response status code.
func (b *EchoResponseBuilder) Status(code int) *EchoResponseBuilder {
	b.response.SetStatusCode(code)

	return b
}

// Header adds a header to the response.
func (b *EchoResponseBuilder) Header(key, value string) *EchoResponseBuilder {
	b.response.AddHeader(key, value)

	return b
}

// Headers adds multiple headers to the response.
func (b *EchoResponseBuilder) Headers(headers map[string]string) *EchoResponseBuilder {
	for key, value := range headers {
		b.response.AddHeader(key, value)
	}

	return b
}

// ContentType sets the Content-Type header.
func (b *EchoResponseBuilder) ContentType(contentType string) *EchoResponseBuilder {
	b.response.SetContentType(contentType)

	return b
}

// JSON sets the response to return JSON data.
func (b *EchoResponseBuilder) JSON(data any) *EchoResponseBuilder {
	b.response.SetContentType("application/json")
	// Convert data to JSON bytes
	if jsonBytes, err := json.Marshal(data); err == nil {
		b.response.SetBodyBytes(jsonBytes)
	}

	return b
}

// Text sets the response to return plain text.
func (b *EchoResponseBuilder) Text(text string) *EchoResponseBuilder {
	b.response.SetContentType("text/plain")
	b.response.SetBodyBytes([]byte(text))

	return b
}

// HTML sets the response to return HTML.
func (b *EchoResponseBuilder) HTML(html string) *EchoResponseBuilder {
	b.response.SetContentType("text/html")
	b.response.SetBodyBytes([]byte(html))

	return b
}

// Redirect sets the response to redirect to the given URL.
func (b *EchoResponseBuilder) Redirect(statusCode int, url string) *EchoResponseBuilder {
	b.response.SetStatusCode(statusCode)
	b.response.SetLocation(url)

	return b
}

// Error sets the response to return an error.
func (b *EchoResponseBuilder) Error(statusCode int, err error) *EchoResponseBuilder {
	b.response.SetStatusCode(statusCode)
	b.response.SetError(err)

	return b
}

// Cookie adds a cookie to the response.
func (b *EchoResponseBuilder) Cookie(cookie *http.Cookie) *EchoResponseBuilder {
	b.response.SetCookie(cookie)

	return b
}

// RequestID sets the request ID for the response.
func (b *EchoResponseBuilder) RequestID(requestID string) *EchoResponseBuilder {
	b.response.SetRequestID(requestID)

	return b
}

// Body sets the response body.
func (b *EchoResponseBuilder) Body(body io.Reader) *EchoResponseBuilder {
	b.response.SetBody(body)

	return b
}

// BodyBytes sets the response body as bytes.
func (b *EchoResponseBuilder) BodyBytes(data []byte) *EchoResponseBuilder {
	b.response.SetBodyBytes(data)

	return b
}

// BodyString sets the response body as a string.
func (b *EchoResponseBuilder) BodyString(data string) *EchoResponseBuilder {
	b.response.SetBodyBytes([]byte(data))

	return b
}

// Build returns the built response.
func (b *EchoResponseBuilder) Build() core.Response {
	return b.response
}

// Send applies the response to Echo and returns any error.
func (b *EchoResponseBuilder) Send() error {
	wrapper := NewEchoResponseWrapper(b.response, b.context)

	return wrapper.ApplyToEcho()
}

// EchoResponseConverter provides utility functions for converting between
// Echo responses and our framework-agnostic responses.
type EchoResponseConverter struct{}

// NewEchoResponseConverter creates a new Echo response converter.
func NewEchoResponseConverter() *EchoResponseConverter {
	return &EchoResponseConverter{}
}

// FromEchoResponse converts an Echo response to our Response interface.
// Note: This is a simplified conversion as Echo doesn't provide direct
// access to all response data in this context.
func (c *EchoResponseConverter) FromEchoResponse(echoResp *echo.Response) core.Response {
	resp := core.NewResponse(echoResp.Status)

	// Copy headers
	for key, values := range echoResp.Header() {
		for _, value := range values {
			resp.AddHeader(key, value)
		}
	}

	// Set content type
	if contentType := echoResp.Header().Get("Content-Type"); contentType != "" {
		resp.SetContentType(contentType)
	}

	// Set content length
	if contentLength := echoResp.Size; contentLength > 0 {
		resp.SetContentLength(contentLength)
	}

	return resp
}

// EchoResponseWriter provides a writer that can be used to write responses
// directly to Echo contexts while maintaining our framework-agnostic interface.
type EchoResponseWriter struct {
	context echo.Context
	buffer  *bytes.Buffer
}

// NewEchoResponseWriter creates a new Echo response writer.
func NewEchoResponseWriter(c echo.Context) *EchoResponseWriter {
	return &EchoResponseWriter{
		context: c,
		buffer:  &bytes.Buffer{},
	}
}

// Write implements io.Writer interface.
func (w *EchoResponseWriter) Write(p []byte) (n int, err error) {
	n, err = w.buffer.Write(p)
	if err != nil {
		return n, fmt.Errorf("failed to write to buffer: %w", err)
	}

	return n, nil
}

// WriteString writes a string to the response.
func (w *EchoResponseWriter) WriteString(s string) (int, error) {
	n, err := w.buffer.WriteString(s)
	if err != nil {
		return n, fmt.Errorf("failed to write string to buffer: %w", err)
	}

	return n, nil
}

// WriteJSON writes JSON data to the response.
func (w *EchoResponseWriter) WriteJSON(data any) error {
	w.context.Response().Header().Set("Content-Type", "application/json")

	return w.context.JSON(w.context.Response().Status, data)
}

// WriteText writes plain text to the response.
func (w *EchoResponseWriter) WriteText(text string) error {
	w.context.Response().Header().Set("Content-Type", "text/plain")

	if err := w.context.String(w.context.Response().Status, text); err != nil {
		return fmt.Errorf("failed to write text response: %w", err)
	}

	return nil
}

// WriteHTML writes HTML to the response.
func (w *EchoResponseWriter) WriteHTML(html string) error {
	w.context.Response().Header().Set("Content-Type", "text/html")

	if err := w.context.HTML(w.context.Response().Status, html); err != nil {
		return fmt.Errorf("failed to write HTML response: %w", err)
	}

	return nil
}

// Flush writes the buffered content to Echo's response writer.
func (w *EchoResponseWriter) Flush() error {
	if w.buffer.Len() > 0 {
		_, err := w.context.Response().Writer.Write(w.buffer.Bytes())
		w.buffer.Reset()

		if err != nil {
			return fmt.Errorf("failed to flush buffer: %w", err)
		}
	}

	return nil
}

// Bytes returns the buffered content as bytes.
func (w *EchoResponseWriter) Bytes() []byte {
	return w.buffer.Bytes()
}

// String returns the buffered content as a string.
func (w *EchoResponseWriter) String() string {
	return w.buffer.String()
}
