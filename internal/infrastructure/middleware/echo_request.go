// Package middleware provides infrastructure layer middleware adapters
// for integrating framework-agnostic middleware with specific HTTP frameworks.
package middleware

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/middleware/core"
)

// maxMultipartMemory is the maximum memory used for parsing multipart forms (32MB)
const maxMultipartMemory = 32 << 20

// EchoRequest wraps Echo's echo.Context to implement our Request interface.
// This adapter provides access to Echo's request data through our framework-agnostic interface.
type EchoRequest struct {
	context echo.Context
	form    url.Values
}

// NewEchoRequest creates a new Echo request wrapper.
func NewEchoRequest(c echo.Context) core.Request {
	return &EchoRequest{
		context: c,
		form:    make(url.Values),
	}
}

// Method returns the HTTP method
func (r *EchoRequest) Method() string {
	return r.context.Request().Method
}

// URL returns the request URL
func (r *EchoRequest) URL() *http.Request {
	return r.context.Request()
}

// Path returns the request path
func (r *EchoRequest) Path() string {
	return r.context.Request().URL.Path
}

// Query returns the query parameters
func (r *EchoRequest) Query() map[string][]string {
	return r.context.QueryParams()
}

// Headers returns the request headers
func (r *EchoRequest) Headers() http.Header {
	return r.context.Request().Header
}

// Body returns the request body as an io.Reader
func (r *EchoRequest) Body() io.Reader {
	return r.context.Request().Body
}

// ContentLength returns the length of the request body
func (r *EchoRequest) ContentLength() int64 {
	return r.context.Request().ContentLength
}

// ContentType returns the Content-Type header value
func (r *EchoRequest) ContentType() string {
	return r.context.Request().Header.Get("Content-Type")
}

// RemoteAddr returns the client's network address
func (r *EchoRequest) RemoteAddr() string {
	return r.context.RealIP()
}

// UserAgent returns the User-Agent header value
func (r *EchoRequest) UserAgent() string {
	return r.context.Request().UserAgent()
}

// Referer returns the Referer header value
func (r *EchoRequest) Referer() string {
	return r.context.Request().Header.Get("Referer")
}

// Host returns the Host header value
func (r *EchoRequest) Host() string {
	return r.context.Request().Host
}

// IsSecure returns true if the request was made over HTTPS
func (r *EchoRequest) IsSecure() bool {
	return r.context.IsTLS()
}

// Context returns the request context
func (r *EchoRequest) Context() context.Context {
	return r.context.Request().Context()
}

// WithContext returns a new request with the given context
func (r *EchoRequest) WithContext(ctx context.Context) core.Request {
	newReq := *r
	newReq.context = r.context // Echo context doesn't support context replacement

	return &newReq
}

// Get retrieves a value from the request context
func (r *EchoRequest) Get(key string) any {
	return r.context.Get(key)
}

// Set stores a value in the request context
func (r *EchoRequest) Set(key string, value any) {
	r.context.Set(key, value)
}

// Param returns a path parameter by name
func (r *EchoRequest) Param(name string) string {
	return r.context.Param(name)
}

// Params returns all path parameters
func (r *EchoRequest) Params() map[string]string {
	params := make(map[string]string)
	for _, name := range r.context.ParamNames() {
		params[name] = r.context.Param(name)
	}

	return params
}

// Cookie returns a cookie by name
func (r *EchoRequest) Cookie(name string) (*http.Cookie, error) {
	cookie, err := r.context.Cookie(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get cookie %s: %w", name, err)
	}

	return cookie, nil
}

// Cookies returns all cookies
func (r *EchoRequest) Cookies() []*http.Cookie {
	return r.context.Cookies()
}

// FormValue returns a form value by name
func (r *EchoRequest) FormValue(name string) string {
	// Try to get from parsed form first
	if r.form != nil {
		if value := r.form.Get(name); value != "" {
			return value
		}
	}

	// Fall back to Echo's form value method
	return r.context.FormValue(name)
}

// Form returns the parsed form data
func (r *EchoRequest) Form() (url.Values, error) {
	// Parse form if not already parsed
	if r.form == nil {
		if err := r.context.Request().ParseForm(); err != nil {
			return nil, fmt.Errorf("failed to parse form: %w", err)
		}

		r.form = r.context.Request().Form
	}

	return r.form, nil
}

// MultipartForm returns the parsed multipart form data
func (r *EchoRequest) MultipartForm() (*multipart.Form, error) {
	if r.context.Request().MultipartForm != nil {
		return r.context.Request().MultipartForm, nil
	}

	// Parse multipart form if not already parsed
	if err := r.context.Request().ParseMultipartForm(maxMultipartMemory); err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	return r.context.Request().MultipartForm, nil
}

// IsAJAX returns true if the request is an AJAX request
func (r *EchoRequest) IsAJAX() bool {
	return r.context.Request().Header.Get("X-Requested-With") == "XMLHttpRequest"
}

// IsWebSocket returns true if the request is a WebSocket upgrade request
func (r *EchoRequest) IsWebSocket() bool {
	return strings.ToLower(r.context.Request().Header.Get("Upgrade")) == "websocket"
}

// IsJSON returns true if the request expects JSON response
func (r *EchoRequest) IsJSON() bool {
	accept := r.context.Request().Header.Get("Accept")

	return strings.Contains(accept, "application/json") || accept == "*/*"
}

// IsXML returns true if the request expects XML response
func (r *EchoRequest) IsXML() bool {
	accept := r.context.Request().Header.Get("Accept")

	return strings.Contains(accept, "application/xml") || strings.Contains(accept, "text/xml")
}

// Accepts returns true if the request accepts the given content type
func (r *EchoRequest) Accepts(contentType string) bool {
	accept := r.context.Request().Header.Get("Accept")

	return accept == contentType || accept == "*/*"
}

// AcceptsEncoding returns true if the request accepts the given encoding
func (r *EchoRequest) AcceptsEncoding(encoding string) bool {
	acceptEncoding := r.context.Request().Header.Get("Accept-Encoding")

	return acceptEncoding == encoding || acceptEncoding == "*"
}

// AcceptsLanguage returns true if the request accepts the given language
func (r *EchoRequest) AcceptsLanguage(language string) bool {
	acceptLanguage := r.context.Request().Header.Get("Accept-Language")

	return acceptLanguage == language || acceptLanguage == "*"
}

// RealIP returns the real IP address of the client
func (r *EchoRequest) RealIP() string {
	return r.context.RealIP()
}

// ForwardedFor returns the X-Forwarded-For header value
func (r *EchoRequest) ForwardedFor() string {
	return r.context.Request().Header.Get("X-Forwarded-For")
}

// RequestID returns the request ID if present
func (r *EchoRequest) RequestID() string {
	if requestID := r.context.Get("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}

	return ""
}

// Timestamp returns when the request was received
func (r *EchoRequest) Timestamp() time.Time {
	if timestamp := r.context.Get("request_timestamp"); timestamp != nil {
		if ts, ok := timestamp.(time.Time); ok {
			return ts
		}
	}

	return time.Now()
}
