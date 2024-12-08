package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/goforms/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockSubscriptionStore struct {
	CreateSubscriptionFunc func(ctx context.Context, sub *models.Subscription) error
}

func (m *MockSubscriptionStore) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	return m.CreateSubscriptionFunc(ctx, sub)
}

func TestCreateSubscription(t *testing.T) {
	e := echo.New()
	// Add the error handling middleware
	e.Use(middleware.Recover())
	logger, _ := zap.NewDevelopment()
	store := &MockSubscriptionStore{}
	handler := NewSubscriptionHandler(logger, store)

	// Test successful creation
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions",
		bytes.NewReader([]byte(`{"email":"test@example.com"}`)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	store.CreateSubscriptionFunc = func(_ context.Context, sub *models.Subscription) error {
		return nil
	}

	err := handler.CreateSubscription(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Test duplicate subscription
	store.CreateSubscriptionFunc = func(_ context.Context, sub *models.Subscription) error {
		return echo.NewHTTPError(http.StatusConflict, "Email already subscribed")
	}

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Request().Body = httptest.NewRequest(http.MethodPost, "/api/subscriptions",
		bytes.NewReader([]byte(`{"email":"test@example.com"}`))).Body

	err = handler.CreateSubscription(c)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusConflict, he.Code)
	assert.Equal(t, "Email already subscribed", he.Message)

	// Test invalid email format
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Request().Body = httptest.NewRequest(http.MethodPost, "/api/subscriptions",
		bytes.NewReader([]byte(`{"email":"invalid-email"}`))).Body

	err = handler.CreateSubscription(c)
	he, ok = err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	assert.Equal(t, "invalid email format", he.Message)
}

func TestInvalidPayload(t *testing.T) {
	e := echo.New()
	e.Use(middleware.Recover())
	logger, _ := zap.NewDevelopment()
	store := &MockSubscriptionStore{}
	handler := NewSubscriptionHandler(logger, store)

	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions",
		bytes.NewReader([]byte(`{"invalid":"data"}`)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreateSubscription(c)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	assert.Equal(t, "email is required", he.Message)
}

func TestHandlerRegister(t *testing.T) {
	e := echo.New()
	logger, _ := zap.NewDevelopment()
	store := &MockSubscriptionStore{}
	handler := NewSubscriptionHandler(logger, store)
	handler.Register(e)

	// Test that the route is registered
	routes := e.Routes()
	found := false
	for _, route := range routes {
		if route.Path == "/api/subscriptions" && route.Method == http.MethodPost {
			found = true
			break
		}
	}

	require.True(t, found, "Route /api/subscriptions should be registered")
}
