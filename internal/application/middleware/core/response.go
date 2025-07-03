// Package core provides the core interfaces and types for middleware functionality.
package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// httpResponse implements the Response interface
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

// Body returns the response body as an io.Reader
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
	if body != nil {
		// Try to read the body into bytes for easier handling
		if bodyData, err := io.ReadAll(body); err == nil {
			r.bodyBytes = bodyData
			r.contentLength = int64(len(bodyData))
		}
	}

	return r
}

// BodyBytes returns the response body as bytes
func (r *httpResponse) BodyBytes() []byte {
	if r.bodyBytes != nil {
		return r.bodyBytes
	}

	if r.body != nil {
		// Try to read the body into bytes
		if bodyData, err := io.ReadAll(r.body); err == nil {
			r.bodyBytes = bodyData
			r.contentLength = int64(len(bodyData))

			return bodyData
		}
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

	return 0
}

// SetContentLength sets the Content-Length header
func (r *httpResponse) SetContentLength(length int64) Response {
	r.contentLength = length
	r.headers.Set("Content-Length", strconv.FormatInt(length, 10))

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
	return strings.Contains(r.contentType, "application/octet-stream")
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
	if r.values != nil {
		return r.values[key]
	}

	return nil
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

// writeString writes a string to the writer and returns the number of bytes written
func (r *httpResponse) writeString(w io.Writer, s string) (int64, error) {
	n, err := w.Write([]byte(s))
	if err != nil {
		return 0, fmt.Errorf("failed to write string: %w", err)
	}

	return int64(n), nil
}

// writeHeaders writes all headers to the writer
func (r *httpResponse) writeHeaders(w io.Writer) (int64, error) {
	var written int64

	for key, values := range r.headers {
		for _, value := range values {
			headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
			if n, err := r.writeString(w, headerLine); err != nil {
				return written, err
			} else {
				written += n
			}
		}
	}

	return written, nil
}

// writeCookies writes all cookies to the writer
func (r *httpResponse) writeCookies(w io.Writer) (int64, error) {
	var written int64

	for _, cookie := range r.cookies {
		cookieLine := fmt.Sprintf("Set-Cookie: %s\r\n", cookie.String())
		if n, err := r.writeString(w, cookieLine); err != nil {
			return written, err
		} else {
			written += n
		}
	}

	return written, nil
}

// writeBody writes the response body to the writer
func (r *httpResponse) writeBody(w io.Writer) (int64, error) {
	if r.bodyBytes != nil {
		n, err := w.Write(r.bodyBytes)
		if err != nil {
			return 0, fmt.Errorf("failed to write body bytes: %w", err)
		}

		return int64(n), nil
	}

	if r.body != nil {
		n, err := io.Copy(w, r.body)
		if err != nil {
			return 0, fmt.Errorf("failed to copy body: %w", err)
		}

		return n, nil
	}

	return 0, nil
}

// WriteTo writes the response to the given io.Writer
func (r *httpResponse) WriteTo(w io.Writer) (int64, error) {
	var written int64

	// Write status line
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.statusCode, http.StatusText(r.statusCode))
	if n, err := r.writeString(w, statusLine); err != nil {
		return written, err
	} else {
		written += n
	}

	// Write headers
	if n, err := r.writeHeaders(w); err != nil {
		return written, err
	} else {
		written += n
	}

	// Write cookies
	if n, err := r.writeCookies(w); err != nil {
		return written, err
	} else {
		written += n
	}

	// Write blank line
	if n, err := r.writeString(w, "\r\n"); err != nil {
		return written, err
	} else {
		written += n
	}

	// Write body
	if n, err := r.writeBody(w); err != nil {
		return written, err
	} else {
		written += n
	}

	return written, nil
}

// Clone creates a copy of this response
func (r *httpResponse) Clone() Response {
	newResp := *r

	newResp.headers = make(http.Header)
	for key, values := range r.headers {
		for _, value := range values {
			newResp.headers.Add(key, value)
		}
	}

	newResp.cookies = make([]*http.Cookie, len(r.cookies))
	copy(newResp.cookies, r.cookies)
	newResp.bodyBytes = make([]byte, len(r.bodyBytes))
	copy(newResp.bodyBytes, r.bodyBytes)

	newResp.values = make(map[string]any)
	for key, value := range r.values {
		newResp.values[key] = value
	}

	return &newResp
}
