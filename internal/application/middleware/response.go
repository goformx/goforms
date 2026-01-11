// Package middleware provides framework-agnostic middleware interfaces and abstractions
// for the GoForms application.
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Response represents an HTTP response abstraction that is framework-agnostic.
// This interface provides access to response data without depending on specific
// HTTP framework implementations like Echo.
type Response interface {
	// StatusCode returns the HTTP status code
	StatusCode() int

	// SetStatusCode sets the HTTP status code
	SetStatusCode(code int) Response

	// Headers returns the response headers
	Headers() http.Header

	// SetHeader sets a response header
	SetHeader(key, value string) Response

	// AddHeader adds a response header (doesn't overwrite existing)
	AddHeader(key, value string) Response

	// Body returns the response body as an io.Reader
	Body() io.Reader

	// SetBody sets the response body
	SetBody(body io.Reader) Response

	// BodyBytes returns the response body as bytes
	BodyBytes() []byte

	// SetBodyBytes sets the response body from bytes
	SetBodyBytes(body []byte) Response

	// ContentType returns the Content-Type header value
	ContentType() string

	// SetContentType sets the Content-Type header
	SetContentType(contentType string) Response

	// ContentLength returns the length of the response body
	ContentLength() int64

	// SetContentLength sets the Content-Length header
	SetContentLength(length int64) Response

	// Location returns the Location header value (for redirects)
	Location() string

	// SetLocation sets the Location header (for redirects)
	SetLocation(location string) Response

	// SetCookie adds a cookie to the response
	SetCookie(cookie *http.Cookie) Response

	// Cookies returns all cookies that will be set
	Cookies() []*http.Cookie

	// Error returns the error associated with this response
	Error() error

	// SetError sets the error for this response
	SetError(err error) Response

	// IsError returns true if this response represents an error
	IsError() bool

	// IsRedirect returns true if this response is a redirect
	IsRedirect() bool

	// IsJSON returns true if the response content type is JSON
	IsJSON() bool

	// IsXML returns true if the response content type is XML
	IsXML() bool

	// IsHTML returns true if the response content type is HTML
	IsHTML() bool

	// IsText returns true if the response content type is plain text
	IsText() bool

	// IsBinary returns true if the response content type is binary
	IsBinary() bool

	// Context returns the response context
	Context() context.Context

	// WithContext returns a new response with the given context
	WithContext(ctx context.Context) Response

	// Get retrieves a value from the response context
	Get(key string) any

	// Set stores a value in the response context
	Set(key string, value any)

	// Timestamp returns when the response was created
	Timestamp() time.Time

	// SetTimestamp sets the response timestamp
	SetTimestamp(timestamp time.Time) Response

	// RequestID returns the request ID associated with this response
	RequestID() string

	// SetRequestID sets the request ID for this response
	SetRequestID(id string) Response

	// WriteTo writes the response to the given io.Writer
	WriteTo(io.Writer) (int64, error)

	// Clone creates a copy of this response
	Clone() Response
}

// ResponseBuilder provides a fluent interface for building Response objects.
// This is useful for testing and creating mock responses.
type ResponseBuilder interface {
	// StatusCode sets the HTTP status code
	StatusCode(code int) ResponseBuilder

	// Header sets a response header
	Header(key, value string) ResponseBuilder

	// Headers sets multiple response headers
	Headers(headers http.Header) ResponseBuilder

	// Body sets the response body
	Body(body io.Reader) ResponseBuilder

	// BodyBytes sets the response body from bytes
	BodyBytes(body []byte) ResponseBuilder

	// ContentType sets the Content-Type header
	ContentType(contentType string) ResponseBuilder

	// ContentLength sets the Content-Length header
	ContentLength(length int64) ResponseBuilder

	// Location sets the Location header (for redirects)
	Location(location string) ResponseBuilder

	// Cookie adds a cookie to the response
	Cookie(cookie *http.Cookie) ResponseBuilder

	// Error sets the error for this response
	Error(err error) ResponseBuilder

	// Context sets the response context
	Context(ctx context.Context) ResponseBuilder

	// Timestamp sets the response timestamp
	Timestamp(timestamp time.Time) ResponseBuilder

	// RequestID sets the request ID for this response
	RequestID(id string) ResponseBuilder

	// JSON marks the response as JSON
	JSON() ResponseBuilder

	// XML marks the response as XML
	XML() ResponseBuilder

	// HTML marks the response as HTML
	HTML() ResponseBuilder

	// Text marks the response as plain text
	Text() ResponseBuilder

	// Build creates the final Response object
	Build() Response
}

// Response constructors for common use cases

// NewResponse creates a new response with the given status code
func NewResponse(statusCode int) Response {
	return &httpResponse{
		statusCode:    statusCode,
		headers:       make(http.Header),
		body:          nil,
		bodyBytes:     nil,
		contentType:   "",
		contentLength: 0,
		location:      "",
		cookies:       make([]*http.Cookie, 0),
		err:           nil,
		context:       context.Background(),
		timestamp:     time.Now(),
		requestID:     "",
		values:        make(map[string]any),
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(statusCode int, err error) Response {
	if statusCode < http.StatusBadRequest {
		statusCode = http.StatusInternalServerError
	}

	resp := NewResponse(statusCode).SetError(err)

	// Set default error content type
	if err != nil {
		resp.SetContentType("application/json")

		errorBody := map[string]any{
			"error":   err.Error(),
			"code":    statusCode,
			"message": http.StatusText(statusCode),
		}
		if jsonData, jsonErr := json.Marshal(errorBody); jsonErr == nil {
			resp.SetBodyBytes(jsonData)
		}
	}

	return resp
}

// NewJSONResponse creates a new JSON response
func NewJSONResponse(data any) Response {
	resp := NewResponse(http.StatusOK)
	resp.SetContentType("application/json")

	if data != nil {
		if jsonData, err := json.Marshal(data); err == nil {
			resp.SetBodyBytes(jsonData)
		} else {
			return NewErrorResponse(http.StatusInternalServerError, fmt.Errorf("failed to marshal JSON: %w", err))
		}
	}

	return resp
}

// NewRedirectResponse creates a new redirect response
func NewRedirectResponse(statusCode int, location string) Response {
	if statusCode < 300 || statusCode >= 400 {
		statusCode = http.StatusFound
	}

	resp := NewResponse(statusCode)
	resp.SetLocation(location)
	resp.SetContentType("text/plain")
	resp.SetBodyBytes([]byte(fmt.Sprintf("Redirecting to %s", location)))

	return resp
}

// NewTextResponse creates a new text response
func NewTextResponse(statusCode int, text string) Response {
	resp := NewResponse(statusCode)
	resp.SetContentType("text/plain")
	resp.SetBodyBytes([]byte(text))

	return resp
}

// NewHTMLResponse creates a new HTML response
func NewHTMLResponse(statusCode int, html string) Response {
	resp := NewResponse(statusCode)
	resp.SetContentType("text/html")
	resp.SetBodyBytes([]byte(html))

	return resp
}

// NewResponseBuilder creates a new ResponseBuilder instance.
func NewResponseBuilder() ResponseBuilder {
	return &responseBuilder{
		response: &httpResponse{
			statusCode:    http.StatusOK,
			headers:       make(http.Header),
			body:          nil,
			bodyBytes:     nil,
			contentType:   "",
			contentLength: 0,
			location:      "",
			cookies:       make([]*http.Cookie, 0),
			err:           nil,
			context:       context.Background(),
			timestamp:     time.Now(),
			requestID:     "",
			values:        make(map[string]any),
		},
	}
}

// httpResponse is the default implementation of the Response interface.
type httpResponse struct {
	statusCode    int
	headers       http.Header
	body          io.Reader
	bodyBytes     []byte
	contentType   string
	contentLength int64
	location      string
	cookies       []*http.Cookie
	err           error
	context       context.Context
	timestamp     time.Time
	requestID     string
	values        map[string]any
}

// StatusCode returns the HTTP status code
func (r *httpResponse) StatusCode() int {
	return r.statusCode
}

// SetStatusCode sets the HTTP status code
func (r *httpResponse) SetStatusCode(code int) Response {
	if code < 100 || code > 599 {
		code = http.StatusInternalServerError
	}

	r.statusCode = code

	return r
}

// Headers returns the response headers
func (r *httpResponse) Headers() http.Header {
	return r.headers
}

// SetHeader sets a response header
func (r *httpResponse) SetHeader(key, value string) Response {
	r.headers.Set(key, value)

	return r
}

// AddHeader adds a response header (doesn't overwrite existing)
func (r *httpResponse) AddHeader(key, value string) Response {
	r.headers.Add(key, value)

	return r
}

// Body returns the response body
func (r *httpResponse) Body() io.Reader {
	if r.body != nil {
		return r.body
	}

	if r.bodyBytes != nil {
		return bytes.NewReader(r.bodyBytes)
	}

	return nil
}

// SetBody sets the response body
func (r *httpResponse) SetBody(body io.Reader) Response {
	r.body = body
	r.bodyBytes = nil

	return r
}

// BodyBytes returns the response body as bytes
func (r *httpResponse) BodyBytes() []byte {
	if r.bodyBytes != nil {
		return r.bodyBytes
	}

	if r.body != nil {
		return r.readBodyBytes()
	}

	return nil
}

// readBodyBytes reads bytes from the body reader
func (r *httpResponse) readBodyBytes() []byte {
	if reader, ok := r.body.(*bytes.Reader); ok {
		return r.readFromBytesReader(reader)
	}

	return r.readFromGenericReader()
}

// readFromBytesReader reads from a bytes.Reader
func (r *httpResponse) readFromBytesReader(reader *bytes.Reader) []byte {
	if _, err := reader.Seek(0, 0); err != nil {
		return nil
	}

	if data, err := io.ReadAll(reader); err == nil {
		r.bodyBytes = data

		return data
	}

	return nil
}

// readFromGenericReader reads from a generic io.Reader
func (r *httpResponse) readFromGenericReader() []byte {
	if data, err := io.ReadAll(r.body); err == nil {
		r.bodyBytes = data

		return data
	}

	return nil
}

// SetBodyBytes sets the response body from bytes
func (r *httpResponse) SetBodyBytes(body []byte) Response {
	r.bodyBytes = body
	r.body = nil
	r.contentLength = int64(len(body))

	return r
}

// ContentType returns the Content-Type header value
func (r *httpResponse) ContentType() string {
	return r.contentType
}

// SetContentType sets the Content-Type header
func (r *httpResponse) SetContentType(contentType string) Response {
	r.contentType = contentType
	r.headers.Set("Content-Type", contentType)

	return r
}

// ContentLength returns the length of the response body
func (r *httpResponse) ContentLength() int64 {
	if r.contentLength > 0 {
		return r.contentLength
	}

	if r.bodyBytes != nil {
		return int64(len(r.bodyBytes))
	}

	if r.body != nil {
		if reader, ok := r.body.(*bytes.Reader); ok {
			return reader.Size()
		}
	}

	return -1
}

// SetContentLength sets the Content-Length header
func (r *httpResponse) SetContentLength(length int64) Response {
	r.contentLength = length
	if length >= 0 {
		r.headers.Set("Content-Length", strconv.FormatInt(length, 10))
	}

	return r
}

// Location returns the Location header value (for redirects)
func (r *httpResponse) Location() string {
	return r.location
}

// SetLocation sets the Location header (for redirects)
func (r *httpResponse) SetLocation(location string) Response {
	r.location = location
	r.headers.Set("Location", location)

	return r
}

// SetCookie adds a cookie to the response
func (r *httpResponse) SetCookie(cookie *http.Cookie) Response {
	r.cookies = append(r.cookies, cookie)

	return r
}

// Cookies returns all cookies that will be set
func (r *httpResponse) Cookies() []*http.Cookie {
	return r.cookies
}

// Error returns the error associated with this response
func (r *httpResponse) Error() error {
	return r.err
}

// SetError sets the error for this response
func (r *httpResponse) SetError(err error) Response {
	r.err = err

	return r
}

// IsError returns true if this response represents an error
func (r *httpResponse) IsError() bool {
	return r.err != nil || r.statusCode >= 400
}

// IsRedirect returns true if this response is a redirect
func (r *httpResponse) IsRedirect() bool {
	return r.statusCode >= 300 && r.statusCode < 400
}

// IsJSON returns true if the response content type is JSON
func (r *httpResponse) IsJSON() bool {
	return strings.Contains(r.contentType, "application/json")
}

// IsXML returns true if the response content type is XML
func (r *httpResponse) IsXML() bool {
	return strings.Contains(r.contentType, "application/xml") || strings.Contains(r.contentType, "text/xml")
}

// IsHTML returns true if the response content type is HTML
func (r *httpResponse) IsHTML() bool {
	return strings.Contains(r.contentType, "text/html")
}

// IsText returns true if the response content type is plain text
func (r *httpResponse) IsText() bool {
	return strings.Contains(r.contentType, "text/plain")
}

// IsBinary returns true if the response content type is binary
func (r *httpResponse) IsBinary() bool {
	return r.contentType != "" && !r.IsJSON() && !r.IsXML() && !r.IsHTML() && !r.IsText()
}

// Context returns the response context
func (r *httpResponse) Context() context.Context {
	return r.context
}

// WithContext returns a new response with the given context
func (r *httpResponse) WithContext(ctx context.Context) Response {
	newResp := *r
	newResp.context = ctx

	return &newResp
}

// Get retrieves a value from the response context
func (r *httpResponse) Get(key string) any {
	if r.values == nil {
		return nil
	}

	return r.values[key]
}

// Set stores a value in the response context
func (r *httpResponse) Set(key string, value any) {
	if r.values == nil {
		r.values = make(map[string]any)
	}

	r.values[key] = value
}

// Timestamp returns when the response was created
func (r *httpResponse) Timestamp() time.Time {
	return r.timestamp
}

// SetTimestamp sets the response timestamp
func (r *httpResponse) SetTimestamp(timestamp time.Time) Response {
	r.timestamp = timestamp

	return r
}

// RequestID returns the request ID associated with this response
func (r *httpResponse) RequestID() string {
	return r.requestID
}

// SetRequestID sets the request ID for this response
func (r *httpResponse) SetRequestID(id string) Response {
	r.requestID = id

	return r
}

// WriteTo writes the response to the given io.Writer
func (r *httpResponse) WriteTo(w io.Writer) (int64, error) {
	var totalBytes int64

	// Write status line
	if n, err := r.writeStatusLine(w); err != nil {
		return totalBytes, err
	} else {
		totalBytes += n
	}

	// Set default headers
	r.setDefaultHeaders()

	// Write headers
	if n, err := r.writeHeaders(w); err != nil {
		return totalBytes, err
	} else {
		totalBytes += n
	}

	// Write cookies
	if n, err := r.writeCookies(w); err != nil {
		return totalBytes, err
	} else {
		totalBytes += n
	}

	// Write blank line separator
	if n, err := w.Write([]byte("\r\n")); err != nil {
		return totalBytes, fmt.Errorf("failed to write header separator: %w", err)
	} else {
		totalBytes += int64(n)
	}

	// Write body
	if n, err := r.writeBody(w); err != nil {
		return totalBytes, err
	} else {
		totalBytes += n
	}

	return totalBytes, nil
}

// writeStatusLine writes the HTTP status line
func (r *httpResponse) writeStatusLine(w io.Writer) (int64, error) {
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.statusCode, http.StatusText(r.statusCode))

	bytesWritten, err := w.Write([]byte(statusLine))
	if err != nil {
		return 0, fmt.Errorf("failed to write status line: %w", err)
	}

	return int64(bytesWritten), nil
}

// setDefaultHeaders sets default headers if not present
func (r *httpResponse) setDefaultHeaders() {
	if r.headers.Get("Date") == "" {
		r.headers.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	}

	if r.headers.Get("Server") == "" {
		r.headers.Set("Server", "GoForms/1.0")
	}

	// Set content length if we have body bytes
	if r.bodyBytes != nil && r.headers.Get("Content-Length") == "" {
		r.headers.Set("Content-Length", strconv.FormatInt(int64(len(r.bodyBytes)), 10))
	}
}

// writeHeaders writes all response headers
func (r *httpResponse) writeHeaders(w io.Writer) (int64, error) {
	var totalBytes int64

	for key, values := range r.headers {
		for _, value := range values {
			headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
			if bytesWritten, err := w.Write([]byte(headerLine)); err != nil {
				return totalBytes, fmt.Errorf("failed to write header %s: %w", key, err)
			} else {
				totalBytes += int64(bytesWritten)
			}
		}
	}

	return totalBytes, nil
}

// writeCookies writes all cookies as Set-Cookie headers
func (r *httpResponse) writeCookies(w io.Writer) (int64, error) {
	var totalBytes int64

	for _, cookie := range r.cookies {
		cookieHeader := fmt.Sprintf("Set-Cookie: %s\r\n", cookie.String())
		if bytesWritten, err := w.Write([]byte(cookieHeader)); err != nil {
			return totalBytes, fmt.Errorf("failed to write cookie: %w", err)
		} else {
			totalBytes += int64(bytesWritten)
		}
	}

	return totalBytes, nil
}

// writeBody writes the response body
func (r *httpResponse) writeBody(w io.Writer) (int64, error) {
	if r.bodyBytes != nil {
		bytesWritten, err := w.Write(r.bodyBytes)
		if err != nil {
			return 0, fmt.Errorf("failed to write body: %w", err)
		}

		return int64(bytesWritten), nil
	}

	if r.body != nil {
		bytesWritten, err := io.Copy(w, r.body)
		if err != nil {
			return 0, fmt.Errorf("failed to write body: %w", err)
		}

		return bytesWritten, nil
	}

	return 0, nil
}

// Clone creates a copy of this response
func (r *httpResponse) Clone() Response {
	newResp := *r

	newResp.headers = make(http.Header)
	for k, v := range r.headers {
		newResp.headers[k] = v
	}

	newResp.cookies = make([]*http.Cookie, len(r.cookies))
	copy(newResp.cookies, r.cookies)

	if r.bodyBytes != nil {
		newResp.bodyBytes = make([]byte, len(r.bodyBytes))
		copy(newResp.bodyBytes, r.bodyBytes)
	}

	if r.values != nil {
		newResp.values = make(map[string]any)
		for k, v := range r.values {
			newResp.values[k] = v
		}
	}

	return &newResp
}

// responseBuilder is the default implementation of the ResponseBuilder interface.
type responseBuilder struct {
	response *httpResponse
}

// StatusCode sets the HTTP status code
func (rb *responseBuilder) StatusCode(code int) ResponseBuilder {
	rb.response.statusCode = code

	return rb
}

// Header sets a response header
func (rb *responseBuilder) Header(key, value string) ResponseBuilder {
	rb.response.headers.Set(key, value)

	return rb
}

// Headers sets multiple response headers
func (rb *responseBuilder) Headers(headers http.Header) ResponseBuilder {
	rb.response.headers = headers

	return rb
}

// Body sets the response body
func (rb *responseBuilder) Body(body io.Reader) ResponseBuilder {
	rb.response.body = body
	rb.response.bodyBytes = nil

	return rb
}

// BodyBytes sets the response body from bytes
func (rb *responseBuilder) BodyBytes(body []byte) ResponseBuilder {
	rb.response.bodyBytes = body
	rb.response.body = nil
	rb.response.contentLength = int64(len(body))

	return rb
}

// ContentType sets the Content-Type header
func (rb *responseBuilder) ContentType(contentType string) ResponseBuilder {
	rb.response.contentType = contentType
	rb.response.headers.Set("Content-Type", contentType)

	return rb
}

// ContentLength sets the Content-Length header
func (rb *responseBuilder) ContentLength(length int64) ResponseBuilder {
	rb.response.contentLength = length
	if length >= 0 {
		rb.response.headers.Set("Content-Length", strconv.FormatInt(length, 10))
	}

	return rb
}

// Location sets the Location header (for redirects)
func (rb *responseBuilder) Location(location string) ResponseBuilder {
	rb.response.location = location
	rb.response.headers.Set("Location", location)

	return rb
}

// Cookie adds a cookie to the response
func (rb *responseBuilder) Cookie(cookie *http.Cookie) ResponseBuilder {
	rb.response.cookies = append(rb.response.cookies, cookie)

	return rb
}

// Error sets the error for this response
func (rb *responseBuilder) Error(err error) ResponseBuilder {
	rb.response.err = err

	return rb
}

// Context sets the response context
func (rb *responseBuilder) Context(ctx context.Context) ResponseBuilder {
	rb.response.context = ctx

	return rb
}

// Timestamp sets the response timestamp
func (rb *responseBuilder) Timestamp(timestamp time.Time) ResponseBuilder {
	rb.response.timestamp = timestamp

	return rb
}

// RequestID sets the request ID for this response
func (rb *responseBuilder) RequestID(id string) ResponseBuilder {
	rb.response.requestID = id

	return rb
}

// JSON marks the response as JSON
func (rb *responseBuilder) JSON() ResponseBuilder {
	rb.response.contentType = "application/json"
	rb.response.headers.Set("Content-Type", "application/json")

	return rb
}

// XML marks the response as XML
func (rb *responseBuilder) XML() ResponseBuilder {
	rb.response.contentType = "application/xml"
	rb.response.headers.Set("Content-Type", "application/xml")

	return rb
}

// HTML marks the response as HTML
func (rb *responseBuilder) HTML() ResponseBuilder {
	rb.response.contentType = "text/html"
	rb.response.headers.Set("Content-Type", "text/html")

	return rb
}

// Text marks the response as plain text
func (rb *responseBuilder) Text() ResponseBuilder {
	rb.response.contentType = "text/plain"
	rb.response.headers.Set("Content-Type", "text/plain")

	return rb
}

// Build creates the final Response object
func (rb *responseBuilder) Build() Response {
	return rb.response
}
