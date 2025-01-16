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
	GetContactsFunc    func(ctx context.Context) ([]models.ContactSubmission, error)
	createContactCalls []struct {
		Ctx     context.Context
		Contact *models.ContactSubmission
	}
	getContactsCalls []context.Context
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

func (m *MockContactStore) GetContacts(ctx context.Context) ([]models.ContactSubmission, error) {
	if m.GetContactsFunc == nil {
		return []models.ContactSubmission{}, nil
	}
	m.getContactsCalls = append(m.getContactsCalls, ctx)
	return m.GetContactsFunc(ctx)
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
		req := httptest.NewRequest(http.MethodPost, "/api/contact",
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
			req := httptest.NewRequest(http.MethodPost, "/api/contact",
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

		req := httptest.NewRequest(http.MethodPost, "/api/contact",
			bytes.NewReader([]byte(invalidPayload)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.CreateContact(c)
		if he, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusBadRequest, he.Code)
			assert.Contains(t, he.Message, "invalid email format")
		} else {
			t.Error("Expected HTTPError")
		}
	})
}

func TestGetContacts(t *testing.T) {
	e, handler := setupContactTestHandler()

	t.Run("successful retrieval", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/contact", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler.store.(*MockContactStore).GetContactsFunc = func(_ context.Context) ([]models.ContactSubmission, error) {
			return []models.ContactSubmission{
				{
					ID:      1,
					Name:    "John Doe",
					Email:   "john@example.com",
					Message: "Test message",
				},
			}, nil
		}

		err := handler.GetContacts(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestContactHandlerRegister(t *testing.T) {
	e, handler := setupContactTestHandler()
	handler.Register(e)

	routes := e.Routes()
	foundPost := false
	foundGet := false
	for _, route := range routes {
		if route.Path == "/api/contact" {
			if route.Method == http.MethodPost {
				foundPost = true
			}
			if route.Method == http.MethodGet {
				foundGet = true
			}
		}
	}

	require.True(t, foundPost, "POST /api/contact route should be registered")
	require.True(t, foundGet, "GET /api/contact route should be registered")
}
