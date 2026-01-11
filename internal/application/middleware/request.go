// Package middleware provides framework-agnostic middleware interfaces and abstractions
// for the GoForms application.
package middleware

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

// Request represents an HTTP request abstraction that is framework-agnostic.
// This interface provides access to request data without depending on specific
// HTTP framework implementations like Echo.
type Request interface {
	// Method returns the HTTP method (GET, POST, PUT, DELETE, etc.)
	Method() string

	// URL returns the request URL
	URL() *url.URL

	// Path returns the request path
	Path() string

	// Query returns the query parameters
	Query() url.Values

	// Headers returns the request headers
	Headers() http.Header

	// Body returns the request body as an io.Reader
	Body() io.Reader

	// ContentLength returns the length of the request body
	ContentLength() int64

	// ContentType returns the Content-Type header value
	ContentType() string

	// RemoteAddr returns the client's network address
	RemoteAddr() string

	// UserAgent returns the User-Agent header value
	UserAgent() string

	// Referer returns the Referer header value
	Referer() string

	// Host returns the Host header value
	Host() string

	// IsSecure returns true if the request was made over HTTPS
	IsSecure() bool

	// Context returns the request context
	Context() context.Context

	// WithContext returns a new request with the given context
	WithContext(ctx context.Context) Request

	// Get retrieves a value from the request context
	Get(key string) any

	// Set stores a value in the request context
	Set(key string, value any)

	// Param returns a path parameter by name
	Param(name string) string

	// Params returns all path parameters
	Params() map[string]string

	// Cookie returns a cookie by name
	Cookie(name string) (*http.Cookie, error)

	// Cookies returns all cookies
	Cookies() []*http.Cookie

	// FormValue returns a form value by name
	FormValue(name string) string

	// Form returns the parsed form data
	Form() (url.Values, error)

	// MultipartForm returns the parsed multipart form data
	MultipartForm() (*multipart.Form, error)

	// IsAJAX returns true if the request is an AJAX request
	IsAJAX() bool

	// IsWebSocket returns true if the request is a WebSocket upgrade request
	IsWebSocket() bool

	// IsJSON returns true if the request expects JSON response
	IsJSON() bool

	// IsXML returns true if the request expects XML response
	IsXML() bool

	// Accepts returns true if the request accepts the given content type
	Accepts(contentType string) bool

	// AcceptsEncoding returns true if the request accepts the given encoding
	AcceptsEncoding(encoding string) bool

	// AcceptsLanguage returns true if the request accepts the given language
	AcceptsLanguage(language string) bool

	// RealIP returns the real IP address of the client
	RealIP() string

	// ForwardedFor returns the X-Forwarded-For header value
	ForwardedFor() string

	// RequestID returns the request ID if present
	RequestID() string

	// Timestamp returns when the request was received
	Timestamp() time.Time
}

// RequestBuilder provides a fluent interface for building Request objects.
// This is useful for testing and creating mock requests.
type RequestBuilder interface {
	// Method sets the HTTP method
	Method(method string) RequestBuilder

	// URL sets the request URL
	URL(u *url.URL) RequestBuilder

	// Path sets the request path
	Path(path string) RequestBuilder

	// Query sets query parameters
	Query(query url.Values) RequestBuilder

	// AddQuery adds a query parameter
	AddQuery(key, value string) RequestBuilder

	// Header sets a header
	Header(key, value string) RequestBuilder

	// Headers sets multiple headers
	Headers(headers http.Header) RequestBuilder

	// Body sets the request body
	Body(body io.Reader) RequestBuilder

	// ContentType sets the Content-Type header
	ContentType(contentType string) RequestBuilder

	// RemoteAddr sets the client address
	RemoteAddr(addr string) RequestBuilder

	// UserAgent sets the User-Agent header
	UserAgent(userAgent string) RequestBuilder

	// Referer sets the Referer header
	Referer(referer string) RequestBuilder

	// Host sets the Host header
	Host(host string) RequestBuilder

	// Secure sets whether the request is secure
	Secure(secure bool) RequestBuilder

	// Context sets the request context
	Context(ctx context.Context) RequestBuilder

	// Param sets a path parameter
	Param(key, value string) RequestBuilder

	// Params sets all path parameters
	Params(params map[string]string) RequestBuilder

	// Cookie adds a cookie
	Cookie(cookie *http.Cookie) RequestBuilder

	// FormValue sets a form value
	FormValue(key, value string) RequestBuilder

	// AJAX marks the request as AJAX
	AJAX() RequestBuilder

	// WebSocket marks the request as WebSocket
	WebSocket() RequestBuilder

	// JSON marks the request as expecting JSON
	JSON() RequestBuilder

	// XML marks the request as expecting XML
	XML() RequestBuilder

	// RealIP sets the real IP address
	RealIP(ip string) RequestBuilder

	// RequestID sets the request ID
	RequestID(id string) RequestBuilder

	// Timestamp sets the request timestamp
	Timestamp(timestamp time.Time) RequestBuilder

	// Build creates the final Request object
	Build() Request
}

// NewRequestBuilder creates a new RequestBuilder instance.
func NewRequestBuilder() RequestBuilder {
	return &requestBuilder{
		request: &request{
			method:     "GET",
			url:        &url.URL{},
			headers:    make(http.Header),
			query:      make(url.Values),
			params:     make(map[string]string),
			cookies:    make([]*http.Cookie, 0),
			formValues: make(url.Values),
			context:    context.Background(),
			timestamp:  time.Now(),
		},
	}
}

// request is the default implementation of the Request interface.
type request struct {
	method      string
	url         *url.URL
	headers     http.Header
	query       url.Values
	body        io.Reader
	contentType string
	remoteAddr  string
	userAgent   string
	referer     string
	host        string
	secure      bool
	context     context.Context
	params      map[string]string
	cookies     []*http.Cookie
	formValues  url.Values
	realIP      string
	requestID   string
	timestamp   time.Time
	values      map[string]any
}

// Method returns the HTTP method
func (r *request) Method() string {
	return r.method
}

// URL returns the request URL
func (r *request) URL() *url.URL {
	return r.url
}

// Path returns the request path
func (r *request) Path() string {
	return r.url.Path
}

// Query returns the query parameters
func (r *request) Query() url.Values {
	return r.query
}

// Headers returns the request headers
func (r *request) Headers() http.Header {
	return r.headers
}

// Body returns the request body
func (r *request) Body() io.Reader {
	return r.body
}

// ContentLength returns the length of the request body
func (r *request) ContentLength() int64 {
	if r.body == nil {
		return 0
	}
	// This is a simplified implementation
	// In a real implementation, you might want to cache this value
	return -1
}

// ContentType returns the Content-Type header value
func (r *request) ContentType() string {
	return r.contentType
}

// RemoteAddr returns the client's network address
func (r *request) RemoteAddr() string {
	return r.remoteAddr
}

// UserAgent returns the User-Agent header value
func (r *request) UserAgent() string {
	return r.userAgent
}

// Referer returns the Referer header value
func (r *request) Referer() string {
	return r.referer
}

// Host returns the Host header value
func (r *request) Host() string {
	return r.host
}

// IsSecure returns true if the request was made over HTTPS
func (r *request) IsSecure() bool {
	return r.secure
}

// Context returns the request context
func (r *request) Context() context.Context {
	return r.context
}

// WithContext returns a new request with the given context
func (r *request) WithContext(ctx context.Context) Request {
	newReq := *r
	newReq.context = ctx

	return &newReq
}

// Get retrieves a value from the request context
func (r *request) Get(key string) any {
	if r.values == nil {
		return nil
	}

	return r.values[key]
}

// Set stores a value in the request context
func (r *request) Set(key string, value any) {
	if r.values == nil {
		r.values = make(map[string]any)
	}

	r.values[key] = value
}

// Param returns a path parameter by name
func (r *request) Param(name string) string {
	return r.params[name]
}

// Params returns all path parameters
func (r *request) Params() map[string]string {
	return r.params
}

// Cookie returns a cookie by name
func (r *request) Cookie(name string) (*http.Cookie, error) {
	for _, cookie := range r.cookies {
		if cookie.Name == name {
			return cookie, nil
		}
	}

	return nil, http.ErrNoCookie
}

// Cookies returns all cookies
func (r *request) Cookies() []*http.Cookie {
	return r.cookies
}

// FormValue returns a form value by name
func (r *request) FormValue(name string) string {
	return r.formValues.Get(name)
}

// Form returns the parsed form data
func (r *request) Form() (url.Values, error) {
	return r.formValues, nil
}

// MultipartForm returns the parsed multipart form data
func (r *request) MultipartForm() (*multipart.Form, error) {
	// This is a simplified implementation
	// In a real implementation, you would parse the multipart form
	return nil, errors.New("multipart form parsing not implemented")
}

// IsAJAX returns true if the request is an AJAX request
func (r *request) IsAJAX() bool {
	return r.headers.Get("X-Requested-With") == "XMLHttpRequest"
}

// IsWebSocket returns true if the request is a WebSocket upgrade request
func (r *request) IsWebSocket() bool {
	return r.headers.Get("Upgrade") == "websocket"
}

// IsJSON returns true if the request expects JSON response
func (r *request) IsJSON() bool {
	return r.headers.Get("Accept") == "application/json"
}

// IsXML returns true if the request expects XML response
func (r *request) IsXML() bool {
	return r.headers.Get("Accept") == "application/xml"
}

// Accepts returns true if the request accepts the given content type
func (r *request) Accepts(contentType string) bool {
	accept := r.headers.Get("Accept")

	return accept == contentType || accept == "*/*"
}

// AcceptsEncoding returns true if the request accepts the given encoding
func (r *request) AcceptsEncoding(encoding string) bool {
	acceptEncoding := r.headers.Get("Accept-Encoding")

	return acceptEncoding == encoding || acceptEncoding == "*"
}

// AcceptsLanguage returns true if the request accepts the given language
func (r *request) AcceptsLanguage(language string) bool {
	acceptLanguage := r.headers.Get("Accept-Language")

	return acceptLanguage == language || acceptLanguage == "*"
}

// RealIP returns the real IP address of the client
func (r *request) RealIP() string {
	return r.realIP
}

// ForwardedFor returns the X-Forwarded-For header value
func (r *request) ForwardedFor() string {
	return r.headers.Get("X-Forwarded-For")
}

// RequestID returns the request ID if present
func (r *request) RequestID() string {
	return r.requestID
}

// Timestamp returns when the request was received
func (r *request) Timestamp() time.Time {
	return r.timestamp
}

// requestBuilder is the default implementation of the RequestBuilder interface.
type requestBuilder struct {
	request *request
}

// Method sets the HTTP method
func (rb *requestBuilder) Method(method string) RequestBuilder {
	rb.request.method = method

	return rb
}

// URL sets the request URL
func (rb *requestBuilder) URL(u *url.URL) RequestBuilder {
	rb.request.url = u

	return rb
}

// Path sets the request path
func (rb *requestBuilder) Path(path string) RequestBuilder {
	if rb.request.url == nil {
		rb.request.url = &url.URL{}
	}

	rb.request.url.Path = path

	return rb
}

// Query sets query parameters
func (rb *requestBuilder) Query(query url.Values) RequestBuilder {
	rb.request.query = query

	return rb
}

// AddQuery adds a query parameter
func (rb *requestBuilder) AddQuery(key, value string) RequestBuilder {
	rb.request.query.Add(key, value)

	return rb
}

// Header sets a header
func (rb *requestBuilder) Header(key, value string) RequestBuilder {
	rb.request.headers.Set(key, value)

	return rb
}

// Headers sets multiple headers
func (rb *requestBuilder) Headers(headers http.Header) RequestBuilder {
	rb.request.headers = headers

	return rb
}

// Body sets the request body
func (rb *requestBuilder) Body(body io.Reader) RequestBuilder {
	rb.request.body = body

	return rb
}

// ContentType sets the Content-Type header
func (rb *requestBuilder) ContentType(contentType string) RequestBuilder {
	rb.request.contentType = contentType
	rb.request.headers.Set("Content-Type", contentType)

	return rb
}

// RemoteAddr sets the client address
func (rb *requestBuilder) RemoteAddr(addr string) RequestBuilder {
	rb.request.remoteAddr = addr

	return rb
}

// UserAgent sets the User-Agent header
func (rb *requestBuilder) UserAgent(userAgent string) RequestBuilder {
	rb.request.userAgent = userAgent
	rb.request.headers.Set("User-Agent", userAgent)

	return rb
}

// Referer sets the Referer header
func (rb *requestBuilder) Referer(referer string) RequestBuilder {
	rb.request.referer = referer
	rb.request.headers.Set("Referer", referer)

	return rb
}

// Host sets the Host header
func (rb *requestBuilder) Host(host string) RequestBuilder {
	rb.request.host = host
	rb.request.headers.Set("Host", host)

	return rb
}

// Secure sets whether the request is secure
func (rb *requestBuilder) Secure(secure bool) RequestBuilder {
	rb.request.secure = secure

	return rb
}

// Context sets the request context
func (rb *requestBuilder) Context(ctx context.Context) RequestBuilder {
	rb.request.context = ctx

	return rb
}

// Param sets a path parameter
func (rb *requestBuilder) Param(key, value string) RequestBuilder {
	rb.request.params[key] = value

	return rb
}

// Params sets all path parameters
func (rb *requestBuilder) Params(params map[string]string) RequestBuilder {
	rb.request.params = params

	return rb
}

// Cookie adds a cookie
func (rb *requestBuilder) Cookie(cookie *http.Cookie) RequestBuilder {
	rb.request.cookies = append(rb.request.cookies, cookie)

	return rb
}

// FormValue sets a form value
func (rb *requestBuilder) FormValue(key, value string) RequestBuilder {
	rb.request.formValues.Set(key, value)

	return rb
}

// AJAX marks the request as AJAX
func (rb *requestBuilder) AJAX() RequestBuilder {
	rb.request.headers.Set("X-Requested-With", "XMLHttpRequest")

	return rb
}

// WebSocket marks the request as WebSocket
func (rb *requestBuilder) WebSocket() RequestBuilder {
	rb.request.headers.Set("Upgrade", "websocket")

	return rb
}

// JSON marks the request as expecting JSON
func (rb *requestBuilder) JSON() RequestBuilder {
	rb.request.headers.Set("Accept", "application/json")

	return rb
}

// XML marks the request as expecting XML
func (rb *requestBuilder) XML() RequestBuilder {
	rb.request.headers.Set("Accept", "application/xml")

	return rb
}

// RealIP sets the real IP address
func (rb *requestBuilder) RealIP(ip string) RequestBuilder {
	rb.request.realIP = ip

	return rb
}

// RequestID sets the request ID
func (rb *requestBuilder) RequestID(id string) RequestBuilder {
	rb.request.requestID = id

	return rb
}

// Timestamp sets the request timestamp
func (rb *requestBuilder) Timestamp(timestamp time.Time) RequestBuilder {
	rb.request.timestamp = timestamp

	return rb
}

// Build creates the final Request object
func (rb *requestBuilder) Build() Request {
	return rb.request
}
