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
	CreateSubscriptionFunc  func(ctx context.Context, sub *models.Subscription) error
	createSubscriptionCalls []struct {
		Ctx context.Context
		Sub *models.Subscription
	}
}

func (m *MockSubscriptionStore) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	if m.CreateSubscriptionFunc == nil {
		return nil
	}
	m.createSubscriptionCalls = append(m.createSubscriptionCalls, struct {
		Ctx context.Context
		Sub *models.Subscription
	}{ctx, sub})
	return m.CreateSubscriptionFunc(ctx, sub)
}

func (m *MockSubscriptionStore) CreateSubscriptionCalls() []struct {
	Ctx context.Context
	Sub *models.Subscription
} {
	return m.createSubscriptionCalls
}

const validOrigin = "https://jonesrussell.github.io/me"

func setupTestHandler() (*echo.Echo, *SubscriptionHandler) {
	e := echo.New()
	e.Use(middleware.Recover())
	logger, _ := zap.NewDevelopment()
	store := &MockSubscriptionStore{}
	handler := NewSubscriptionHandler(logger, store)
	return e, handler
}

func TestCreateSubscription(t *testing.T) {
	e, handler := setupTestHandler()

	// Test successful creation
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions",
		bytes.NewReader([]byte(`{"email":"test@example.com"}`)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderOrigin, validOrigin)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler.store.(*MockSubscriptionStore).CreateSubscriptionFunc = func(_ context.Context, _ *models.Subscription) error {
		return nil
	}

	err := handler.CreateSubscription(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Test duplicate subscription
	handler.store.(*MockSubscriptionStore).CreateSubscriptionFunc = func(_ context.Context, _ *models.Subscription) error {
		return echo.NewHTTPError(http.StatusConflict, "Email already subscribed")
	}

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Request().Header.Set(echo.HeaderOrigin, validOrigin)
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
	c.Request().Header.Set(echo.HeaderOrigin, validOrigin)
	c.Request().Body = httptest.NewRequest(http.MethodPost, "/api/subscriptions",
		bytes.NewReader([]byte(`{"email":"invalid-email"}`))).Body

	err = handler.CreateSubscription(c)
	he, ok = err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	assert.Equal(t, "invalid email format", he.Message)
}

func TestInvalidPayload(t *testing.T) {
	e, handler := setupTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions",
		bytes.NewReader([]byte(`{"invalid":"data"}`)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderOrigin, validOrigin)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreateSubscription(c)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	assert.Equal(t, "email is required", he.Message)
}

func TestHandlerRegister(t *testing.T) {
	e, handler := setupTestHandler()
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
