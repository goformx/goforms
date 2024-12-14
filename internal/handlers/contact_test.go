package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/goforms/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockContactStore struct {
	CreateContactFunc  func(ctx context.Context, contact *models.ContactSubmission) error
	createContactCalls []struct {
		Ctx     context.Context
		Contact *models.ContactSubmission
	}
}

func (m *MockContactStore) CreateContact(ctx context.Context, contact *models.ContactSubmission) error {
	if m.CreateContactFunc == nil {
		return nil
	}
	m.createContactCalls = append(m.createContactCalls, struct {
		Ctx     context.Context
		Contact *models.ContactSubmission
	}{ctx, contact})
	return m.CreateContactFunc(ctx, contact)
}

func setupContactTestHandler() (*echo.Echo, *ContactHandler) {
	e := echo.New()
	logger, _ := zap.NewDevelopment()
	store := &MockContactStore{}
	handler := NewContactHandler(logger, store)
	return e, handler
}

func TestCreateContact(t *testing.T) {
	e, handler := setupContactTestHandler()

	validPayload := `{
		"name": "John Doe",
		"email": "john@example.com",
		"message": "Test message"
	}`

	t.Run("successful submission", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/app/contact",
			bytes.NewReader([]byte(validPayload)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler.store.(*MockContactStore).CreateContactFunc = func(_ context.Context, _ *models.ContactSubmission) error {
			return nil
		}

		err := handler.CreateContact(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("missing required fields", func(t *testing.T) {
		invalidPayloads := []string{
			`{"email":"john@example.com", "message":"Test"}`,            // missing name
			`{"name":"John Doe", "message":"Test"}`,                     // missing email
			`{"name":"John Doe", "email":"john@example.com"}`,           // missing message
			`{"name":"", "email":"john@example.com", "message":"Test"}`, // empty name
		}

		for _, payload := range invalidPayloads {
			req := httptest.NewRequest(http.MethodPost, "/app/contact",
				bytes.NewReader([]byte(payload)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.CreateContact(c)
			if he, ok := err.(*echo.HTTPError); ok {
				assert.Equal(t, http.StatusBadRequest, he.Code)
			} else {
				t.Error("Expected HTTPError")
			}
		}
	})

	t.Run("invalid email format", func(t *testing.T) {
		invalidPayload := `{
			"name": "John Doe",
			"email": "invalid-email",
			"message": "Test message"
		}`

		req := httptest.NewRequest(http.MethodPost, "/app/contact",
			bytes.NewReader([]byte(invalidPayload)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.CreateContact(c)
		he, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, he.Code)
		assert.Equal(t, "invalid email format", he.Message)
	})
}

func TestContactHandlerRegister(t *testing.T) {
	e, handler := setupContactTestHandler()
	handler.Register(e)

	routes := e.Routes()
	found := false
	for _, route := range routes {
		if route.Path == "/v1/contact" && route.Method == http.MethodPost {
			found = true
			break
		}
	}

	require.True(t, found, "Route /v1/contact should be registered")
}
