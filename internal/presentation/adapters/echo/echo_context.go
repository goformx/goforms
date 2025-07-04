package echo

import (
	"context"
	"net/url"

	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
	echo "github.com/labstack/echo/v4"
)

// EchoContextAdapter adapts echo.Context to the httpiface.Context interface.
type EchoContextAdapter struct {
	echoCtx echo.Context
	user    *httpiface.UserContext
	session httpiface.Session
}

// NewEchoContextAdapter wraps an echo.Context.
func NewEchoContextAdapter(c echo.Context) *EchoContextAdapter {
	return &EchoContextAdapter{echoCtx: c}
}

// Request returns the request (stub: returns echo.Context for now)
func (a *EchoContextAdapter) Request() httpiface.Request {
	return a.echoCtx // stub
}

// Response returns the response (stub: returns echo.Context for now)
func (a *EchoContextAdapter) Response() httpiface.Response {
	return a.echoCtx // stub
}

// Context returns the underlying context.Context
func (a *EchoContextAdapter) Context() context.Context {
	return a.echoCtx.Request().Context()
}

// WithContext returns a new adapter with the given context (not implemented)
func (a *EchoContextAdapter) WithContext(ctx context.Context) httpiface.Context {
	// Not implemented: echo.Context does not support WithContext directly
	return a
}

// Get retrieves a value from echo.Context
func (a *EchoContextAdapter) Get(key string) (any, bool) {
	v := a.echoCtx.Get(key)

	return v, v != nil
}

// Set sets a value in echo.Context
func (a *EchoContextAdapter) Set(key string, value any) {
	a.echoCtx.Set(key, value)
}

// Delete removes a value (not supported by echo.Context, so set to nil)
func (a *EchoContextAdapter) Delete(key string) {
	a.echoCtx.Set(key, nil)
}

// User returns the user context (stub)
func (a *EchoContextAdapter) User() *httpiface.UserContext {
	return a.user
}

// SetUser sets the user context
func (a *EchoContextAdapter) SetUser(user *httpiface.UserContext) {
	a.user = user
}

// Session returns the session (stub)
func (a *EchoContextAdapter) Session() httpiface.Session {
	return a.session
}

// SetSession sets the session
func (a *EchoContextAdapter) SetSession(session httpiface.Session) {
	a.session = session
}

// Param returns a route param
func (a *EchoContextAdapter) Param(name string) string {
	return a.echoCtx.Param(name)
}

// QueryParam returns a query param
func (a *EchoContextAdapter) QueryParam(name string) string {
	return a.echoCtx.QueryParam(name)
}

// FormValue returns a form value
func (a *EchoContextAdapter) FormValue(name string) string {
	return a.echoCtx.FormValue(name)
}

// Form returns all form values
func (a *EchoContextAdapter) Form() (url.Values, error) {
	if err := a.echoCtx.Request().ParseForm(); err != nil {
		return nil, err
	}

	return a.echoCtx.Request().Form, nil
}

// PathParam returns a route param (same as Param)
func (a *EchoContextAdapter) PathParam(name string) string {
	return a.echoCtx.Param(name)
}

// JSON sends a JSON response
func (a *EchoContextAdapter) JSON(code int, i any) error {
	return a.echoCtx.JSON(code, i)
}

// JSONBlob sends a JSON blob response
func (a *EchoContextAdapter) JSONBlob(code int, b []byte) error {
	return a.echoCtx.JSONBlob(code, b)
}

// String sends a string response
func (a *EchoContextAdapter) String(code int, s string) error {
	return a.echoCtx.String(code, s)
}

// HTML sends an HTML response
func (a *EchoContextAdapter) HTML(code int, html string) error {
	return a.echoCtx.HTML(code, html)
}

// NoContent sends a no content response
func (a *EchoContextAdapter) NoContent(code int) error {
	return a.echoCtx.NoContent(code)
}

// Redirect redirects the request
func (a *EchoContextAdapter) Redirect(code int, url string) error {
	return a.echoCtx.Redirect(code, url)
}

// Error sends an error response
func (a *EchoContextAdapter) Error(err error) error {
	return a.echoCtx.JSON(500, map[string]any{
		"error": err.Error(),
	})
}
