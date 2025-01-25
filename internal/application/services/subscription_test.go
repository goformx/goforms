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
	"github.com/stretchr/testify/assert"

	"github.com/jonesrussell/goforms/internal/application/services"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
	storemock "github.com/jonesrussell/goforms/test/mocks/store/subscription"
	subscriptionmock "github.com/jonesrussell/goforms/test/mocks/store/subscription"
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
		mockLogger.ExpectInfo("subscription created",
			logging.String("email", "test@example.com"),
		)

		req := httptest.NewRequest(http.MethodPost, "/subscribe", strings.NewReader(`{"email":"test@example.com","name":"Test User"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		err := handler.HandleSubscribe(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		err = json.NewDecoder(rec.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))

		assert.NoError(t, mockStore.Verify())
		assert.NoError(t, mockLogger.Verify())
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockLogger.ExpectError("failed to bind subscription request",
			logging.Error(errors.New("invalid character 'i' looking for beginning of value")),
		)

		req := httptest.NewRequest(http.MethodPost, "/subscribe", strings.NewReader(`invalid json`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		err := handler.HandleSubscribe(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp map[string]interface{}
		err = json.NewDecoder(rec.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.False(t, resp["success"].(bool))
		assert.Equal(t, "Invalid request body", resp["error"])

		assert.NoError(t, mockLogger.Verify())
	})
}

func TestSubscriptionService(t *testing.T) {
	t.Run("create subscription", func(t *testing.T) {
		mockStore := storemock.NewMockStore(t)
		mockLogger := mocklogging.NewMockLogger()
		service := subscription.NewService(mockStore, mockLogger)
		assert.NotNil(t, service)

		sub := &subscription.Subscription{
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockStore.ExpectGetByEmail(context.Background(), "test@example.com", nil, subscription.ErrSubscriptionNotFound)
		mockStore.ExpectCreate(context.Background(), sub, nil)

		err := service.CreateSubscription(context.Background(), sub)
		assert.NoError(t, err)

		assert.NoError(t, mockStore.Verify(), "store expectations not met")
		assert.NoError(t, mockLogger.Verify(), "logger expectations not met")
	})

	t.Run("create subscription error", func(t *testing.T) {
		mockStore := storemock.NewMockStore(t)
		mockLogger := mocklogging.NewMockLogger()
		service := subscription.NewService(mockStore, mockLogger)
		assert.NotNil(t, service)

		sub := &subscription.Subscription{
			Email:  "test@example.com",
			Name:   "Test User",
			Status: subscription.StatusPending,
		}

		storeErr := errors.New("store error")
		mockStore.ExpectGetByEmail(context.Background(), "test@example.com", nil, subscription.ErrSubscriptionNotFound)
		mockStore.ExpectCreate(context.Background(), sub, storeErr)
		mockLogger.ExpectError("failed to create subscription",
			logging.Error(storeErr),
		)

		err := service.CreateSubscription(context.Background(), sub)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create subscription")
		assert.Contains(t, err.Error(), storeErr.Error())

		assert.NoError(t, mockStore.Verify(), "store expectations not met")
		assert.NoError(t, mockLogger.Verify(), "logger expectations not met")
	})

	t.Run("list subscriptions", func(t *testing.T) {
		mockStore := storemock.NewMockStore(t)
		mockLogger := mocklogging.NewMockLogger()
		service := subscription.NewService(mockStore, mockLogger)
		assert.NotNil(t, service)

		expected := []subscription.Subscription{}
		mockStore.ExpectList(context.Background(), expected, nil)

		subs, err := service.ListSubscriptions(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, subs)

		assert.NoError(t, mockStore.Verify(), "store expectations not met")
		assert.NoError(t, mockLogger.Verify(), "logger expectations not met")
	})

	t.Run("list subscriptions error", func(t *testing.T) {
		mockStore := storemock.NewMockStore(t)
		mockLogger := mocklogging.NewMockLogger()
		service := subscription.NewService(mockStore, mockLogger)
		assert.NotNil(t, service)

		storeErr := errors.New("store error")
		mockStore.ExpectList(context.Background(), nil, storeErr)
		mockLogger.ExpectError("failed to list subscriptions",
			logging.Error(storeErr),
		)

		subs, err := service.ListSubscriptions(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list subscriptions")
		assert.Contains(t, err.Error(), storeErr.Error())
		assert.Nil(t, subs)

		assert.NoError(t, mockStore.Verify(), "store expectations not met")
		assert.NoError(t, mockLogger.Verify(), "logger expectations not met")
	})
}

func TestSubscriptionService_Create(t *testing.T) {
	t.Run("valid_subscription", func(t *testing.T) {
		mockStore := subscriptionmock.NewMockStore(t) // Create new mocks for each subtest
		mockLogger := mocklogging.NewMockLogger()
		service := subscription.NewService(mockStore, mockLogger)

		sub := &subscription.Subscription{
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockStore.ExpectGetByEmail(context.Background(), "test@example.com", nil, nil)
		mockStore.ExpectCreate(context.Background(), sub, nil)

		err := service.CreateSubscription(context.Background(), sub)
		assert.NoError(t, err)
		assert.NoError(t, mockStore.Verify())
		assert.NoError(t, mockLogger.Verify())
	})

	t.Run("duplicate_email", func(t *testing.T) {
		mockStore := subscriptionmock.NewMockStore(t) // Create new mocks for each subtest
		mockLogger := mocklogging.NewMockLogger()
		service := subscription.NewService(mockStore, mockLogger)

		sub := &subscription.Subscription{
			Email: "test@example.com",
			Name:  "Test User",
		}

		existingSub := &subscription.Subscription{
			ID:    1,
			Email: "test@example.com",
			Name:  "Existing User",
		}

		// Expect validation check first
		mockStore.ExpectGetByEmail(context.Background(), "test@example.com", existingSub, nil)

		err := service.CreateSubscription(context.Background(), sub)
		assert.ErrorIs(t, err, subscription.ErrEmailAlreadyExists)
		assert.NoError(t, mockStore.Verify())
		assert.NoError(t, mockLogger.Verify())
	})
}
