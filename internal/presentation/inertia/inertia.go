// Package inertia provides Gonertia (Inertia.js) integration for Vue 3 SPA rendering.
package inertia

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/romsar/gonertia"

	"github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Manager handles Inertia.js rendering and configuration.
type Manager struct {
	inertia *gonertia.Inertia
	config  *config.Config
	logger  logging.Logger
}

// Props represents the properties passed to Inertia pages.
type Props = gonertia.Props

// NewManager creates a new Inertia manager with the given configuration.
func NewManager(cfg *config.Config, logger logging.Logger) (*Manager, error) {
	// Determine the root template path
	rootTemplate := buildRootTemplate(cfg)

	// Create Gonertia options
	opts := []gonertia.Option{
		gonertia.WithVersion(cfg.App.Version),
	}

	// In development, use Vite dev server manifest
	if cfg.App.IsDevelopment() {
		opts = append(opts, gonertia.WithVersionFromFile(filepath.Join("dist", ".vite", "manifest.json")))
	} else {
		// In production, use the built manifest
		manifestPath := filepath.Join("dist", ".vite", "manifest.json")
		if _, err := os.Stat(manifestPath); err == nil {
			opts = append(opts, gonertia.WithVersionFromFile(manifestPath))
		}
	}

	// Create the Inertia instance
	i, err := gonertia.New(rootTemplate, opts...)
	if err != nil {
		return nil, err
	}

	return &Manager{
		inertia: i,
		config:  cfg,
		logger:  logger,
	}, nil
}

// buildRootTemplate creates the root HTML template for Inertia.
func buildRootTemplate(cfg *config.Config) string {
	if cfg.App.IsDevelopment() {
		return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="GoFormX - A self-hosted form backend service built with Go">
    <meta name="color-scheme" content="light dark">
    {{ .inertiaHead }}
    <script type="module" src="http://localhost:5173/@vite/client"></script>
    <script type="module" src="http://localhost:5173/src/main.ts"></script>
</head>
<body>
    {{ .inertia }}
</body>
</html>`
	}

	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="GoFormX - A self-hosted form backend service built with Go">
    <meta name="color-scheme" content="light dark">
    {{ .inertiaHead }}
    {{ .viteAssets }}
</head>
<body>
    {{ .inertia }}
</body>
</html>`
}

// Render renders an Inertia page with the given component and props.
func (m *Manager) Render(c echo.Context, component string, props Props) error {
	// Add shared props
	sharedProps := m.getSharedProps(c)
	for key, value := range sharedProps {
		if _, exists := props[key]; !exists {
			props[key] = value
		}
	}

	return m.inertia.Render(c.Response(), c.Request(), component, props)
}

// Location performs an Inertia redirect (external URL redirect).
func (m *Manager) Location(c echo.Context, url string) error {
	m.inertia.Location(c.Response(), c.Request(), url)
	return nil
}

// getSharedProps returns props that should be shared across all pages.
func (m *Manager) getSharedProps(c echo.Context) Props {
	props := Props{
		"csrf":          m.getCSRFToken(c),
		"isDevelopment": m.config.App.IsDevelopment(),
		"appVersion":    m.config.App.Version,
	}

	// Add authenticated user if present
	if user := m.getCurrentUser(c); user != nil {
		props["auth"] = map[string]any{
			"user": map[string]any{
				"id":        user.ID,
				"email":     user.Email,
				"firstName": user.FirstName,
				"lastName":  user.LastName,
				"role":      user.Role,
			},
		}
	} else {
		props["auth"] = map[string]any{
			"user": nil,
		}
	}

	// Add flash messages if present
	if flash := m.getFlashMessages(c); flash != nil {
		props["flash"] = flash
	}

	return props
}

// getCurrentUser extracts the current user from context.
func (m *Manager) getCurrentUser(c echo.Context) *entities.User {
	userID, ok := context.GetUserID(c)
	if !ok {
		return nil
	}

	email, _ := context.GetEmail(c)
	role, _ := context.GetRole(c)
	firstName, _ := context.GetFirstName(c)
	lastName, _ := context.GetLastName(c)

	return &entities.User{
		ID:        userID,
		Email:     email,
		Role:      role,
		FirstName: firstName,
		LastName:  lastName,
	}
}

// getCSRFToken retrieves the CSRF token from context.
func (m *Manager) getCSRFToken(c echo.Context) string {
	contextKey := "csrf"
	if m.config.Security.CSRF.ContextKey != "" {
		contextKey = m.config.Security.CSRF.ContextKey
	}

	if token, ok := c.Get(contextKey).(string); ok {
		return token
	}

	// Fallback to cookie
	cookieName := "_csrf"
	if m.config.Security.CSRF.CookieName != "" {
		cookieName = m.config.Security.CSRF.CookieName
	}

	cookie, err := c.Cookie(cookieName)
	if err == nil && cookie != nil {
		return cookie.Value
	}

	return ""
}

// getFlashMessages retrieves flash messages from the session.
func (m *Manager) getFlashMessages(c echo.Context) map[string]string {
	// Check for flash messages in context
	if flash, ok := c.Get("flash").(map[string]string); ok {
		return flash
	}
	return nil
}

// Middleware returns an Echo middleware that handles Inertia requests.
func (m *Manager) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if this is an Inertia request
			if c.Request().Header.Get("X-Inertia") != "" {
				// Set proper content type for Inertia responses
				c.Response().Header().Set("Vary", "X-Inertia")
			}
			return next(c)
		}
	}
}

// ShareProps adds props that will be shared with all Inertia responses.
func (m *Manager) ShareProps(props Props) {
	m.inertia.ShareProp("shared", props)
}

// InertiaPageData represents the data structure for an Inertia page.
type InertiaPageData struct {
	Component string         `json:"component"`
	Props     map[string]any `json:"props"`
	URL       string         `json:"url"`
	Version   string         `json:"version"`
}

// MarshalInertiaPage marshals page data for the data-page attribute.
// The JSON is marshaled from trusted internal data structures, not user input.
func MarshalInertiaPage(data InertiaPageData) template.JS {
	bytes, err := json.Marshal(data)
	if err != nil {
		return template.JS("{}") // #nosec G203 - empty object is safe
	}
	return template.JS(bytes) // #nosec G203 - data is from trusted internal sources
}

// IsInertiaRequest checks if the current request is an Inertia request.
func IsInertiaRequest(r *http.Request) bool {
	return r.Header.Get("X-Inertia") != ""
}

// EchoHandler wraps the Inertia manager to provide a cleaner API for Echo handlers.
type EchoHandler struct {
	*Manager
}

// NewEchoHandler creates a new EchoHandler wrapper.
func NewEchoHandler(m *Manager) *EchoHandler {
	return &EchoHandler{Manager: m}
}
