package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/application/services"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
	storemock "github.com/jonesrussell/goforms/test/mocks/store/subscription"
)

func TestSubscriptionHandler_HandleSubscribe(t *testing.T) {
	// Setup
	mockStore := storemock.NewMockStore(t)
	mockLogger := mocklogging.NewMockLogger()
	handler := services.NewSubscriptionHandler(mockStore, mockLogger)

	t.Run("successful subscription", func(t *testing.T) {
		sub := &subscription.Subscription{
			Email:  "test@example.com",
			Name:   "Test User",
			Status: subscription.StatusPending,
		}
		mockStore.ExpectGetByEmail(context.Background(), "test@example.com", nil, subscription.ErrSubscriptionNotFound)
		mockStore.ExpectCreate(context.Background(), sub, nil)
		mockLogger.ExpectInfo("subscription created").WithFields(map[string]interface{}{
			"email": "test@example.com",
		})

		req := httptest.NewRequest(http.MethodPost, "/subscribe", strings.NewReader(`{"email":"test@example.com","name":"Test User"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		err := handler.HandleSubscribe(c)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if rec.Code != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, rec.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if success, ok := resp["success"].(bool); !ok || !success {
			t.Errorf("expected success to be true, got %v", resp["success"])
		}

		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockLogger.ExpectError("failed to bind subscription request").WithFields(map[string]interface{}{
			"error": mocklogging.AnyValue{},
		})

		req := httptest.NewRequest(http.MethodPost, "/subscribe", strings.NewReader(`invalid json`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		err := handler.HandleSubscribe(c)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if success, ok := resp["success"].(bool); !ok || success {
			t.Errorf("expected success to be false, got %v", resp["success"])
		}
		if errMsg, ok := resp["error"].(string); !ok || errMsg != "Invalid request body" {
			t.Errorf("expected error message %q, got %q", "Invalid request body", errMsg)
		}

		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})
}

func TestSubscriptionService(t *testing.T) {
	t.Run("create subscription", func(t *testing.T) {
		mockStore := storemock.NewMockStore(t)
		mockLogger := mocklogging.NewMockLogger()
		service := subscription.NewService(mockStore, mockLogger)
		if service == nil {
			t.Fatal("expected service to be created")
		}

		sub := &subscription.Subscription{
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockStore.ExpectGetByEmail(context.Background(), "test@example.com", nil, subscription.ErrSubscriptionNotFound)
		mockStore.ExpectCreate(context.Background(), sub, nil)

		err := service.CreateSubscription(context.Background(), sub)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("create subscription error", func(t *testing.T) {
		mockStore := storemock.NewMockStore(t)
		mockLogger := mocklogging.NewMockLogger()
		service := subscription.NewService(mockStore, mockLogger)
		if service == nil {
			t.Fatal("expected service to be created")
		}

		sub := &subscription.Subscription{
			Email:  "test@example.com",
			Name:   "Test User",
			Status: subscription.StatusPending,
		}

		storeErr := errors.New("store error")
		mockStore.ExpectGetByEmail(context.Background(), "test@example.com", nil, subscription.ErrSubscriptionNotFound)
		mockStore.ExpectCreate(context.Background(), sub, storeErr)
		mockLogger.ExpectError("failed to create subscription").WithFields(map[string]interface{}{
			"error": mocklogging.AnyValue{},
		})

		err := service.CreateSubscription(context.Background(), sub)
		if err == nil {
			t.Error("expected error, got nil")
		}

		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("list subscriptions", func(t *testing.T) {
		mockStore := storemock.NewMockStore(t)
		mockLogger := mocklogging.NewMockLogger()
		service := subscription.NewService(mockStore, mockLogger)
		if service == nil {
			t.Fatal("expected service to be created")
		}

		expected := []subscription.Subscription{{ID: 1}}
		mockStore.ExpectList(context.Background(), expected, nil)

		got, err := service.ListSubscriptions(context.Background())
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(got) != len(expected) {
			t.Errorf("expected %d subscriptions, got %d", len(expected), len(got))
		}

		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("list subscriptions error", func(t *testing.T) {
		mockStore := storemock.NewMockStore(t)
		mockLogger := mocklogging.NewMockLogger()
		service := subscription.NewService(mockStore, mockLogger)
		if service == nil {
			t.Fatal("expected service to be created")
		}

		storeErr := errors.New("store error")
		mockStore.ExpectList(context.Background(), nil, storeErr)
		mockLogger.ExpectError("failed to list subscriptions").WithFields(map[string]interface{}{
			"error": mocklogging.AnyValue{},
		})

		got, err := service.ListSubscriptions(context.Background())
		if err == nil {
			t.Error("expected error, got nil")
		}
		if got != nil {
			t.Errorf("expected nil result, got %v", got)
		}

		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})
}
