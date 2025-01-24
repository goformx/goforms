package services_test

import (
	"context"
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
)

func TestSubscriptionHandler_HandleSubscribe(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setupMocks     func(*storemock.MockStore, *mocklogging.MockLogger)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful subscription",
			requestBody: `{
                "email": "test@example.com",
                "name": "Test User"
            }`,
			setupMocks: func(store *storemock.MockStore, logger *mocklogging.MockLogger) {
				sub := &subscription.Subscription{
					Email:  "test@example.com",
					Name:   "Test User",
					Status: subscription.StatusPending,
				}
				// Handler directly calls Create
				store.ExpectCreate(context.Background(), sub, nil)
				logger.ExpectInfo("subscription created",
					logging.String("email", "test@example.com"),
				)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"success":true,"data":{"id":0,"email":"test@example.com","name":"Test User","status":"pending","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name:        "invalid request body",
			requestBody: `invalid json`,
			setupMocks: func(store *storemock.MockStore, logger *mocklogging.MockLogger) {
				logger.ExpectError("failed to bind subscription request",
					logging.Error(errors.New("invalid json")),
				)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":"Invalid request body"}`,
		},
		{
			name: "missing required fields",
			requestBody: `{
                "email": "",
                "name": ""
            }`,
			setupMocks:     func(store *storemock.MockStore, logger *mocklogging.MockLogger) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"success":false,"error":"email is required"}`,
		},
		{
			name: "store error",
			requestBody: `{
                "email": "test@example.com",
                "name": "Test User"
            }`,
			setupMocks: func(store *storemock.MockStore, logger *mocklogging.MockLogger) {
				sub := &subscription.Subscription{
					Email:  "test@example.com",
					Name:   "Test User",
					Status: subscription.StatusPending,
				}
				// Handler directly calls Create which fails
				storeErr := errors.New("store error")
				store.ExpectCreate(context.Background(), sub, storeErr)
				logger.ExpectError("failed to create subscription",
					logging.Error(storeErr),
					logging.String("email", "test@example.com"),
				)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"success":false,"error":"Failed to create subscription"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storemock.NewMockStore()
			logger := mocklogging.NewMockLogger()
			tt.setupMocks(store, logger)

			handler := services.NewSubscriptionHandler(store, logger)
			assert.NotNil(t, handler)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/subscribe", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.HandleSubscribe(c)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.JSONEq(t, tt.expectedBody, rec.Body.String())

			if err := logger.Verify(); err != nil {
				t.Errorf("logger expectations not met: %v", err)
			}
			if err := store.Verify(); err != nil {
				t.Errorf("store expectations not met: %v", err)
			}
		})
	}
}

func TestSubscriptionService(t *testing.T) {
	t.Run("create subscription", func(t *testing.T) {
		mockStore := storemock.NewMockStore()
		mockLogger := mocklogging.NewMockLogger()
		service := subscription.NewService(mockLogger, mockStore)
		assert.NotNil(t, service)

		sub := &subscription.Subscription{
			Email: "test@example.com",
			Name:  "Test User",
		}

		// First, expect GetByEmail check for existing subscription
		mockStore.ExpectGetByEmail(context.Background(), "test@example.com", nil, subscription.ErrSubscriptionNotFound)
		// Then, expect Create call
		mockStore.ExpectCreate(context.Background(), sub, nil)

		err := service.CreateSubscription(context.Background(), sub)
		assert.NoError(t, err)

		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("create subscription error", func(t *testing.T) {
		mockStore := storemock.NewMockStore()
		mockLogger := mocklogging.NewMockLogger()
		service := subscription.NewService(mockLogger, mockStore)
		assert.NotNil(t, service)

		sub := &subscription.Subscription{
			Email:  "test@example.com",
			Name:   "Test User",
			Status: subscription.StatusPending,
		}

		storeErr := errors.New("store error")
		// First, expect GetByEmail check
		mockStore.ExpectGetByEmail(context.Background(), "test@example.com", nil, subscription.ErrSubscriptionNotFound)
		// Then, expect Create call with error
		mockStore.ExpectCreate(context.Background(), sub, storeErr)
		mockLogger.ExpectError("failed to create subscription",
			logging.Error(storeErr),
		)

		err := service.CreateSubscription(context.Background(), sub)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create subscription")
		assert.Contains(t, err.Error(), storeErr.Error())

		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("list subscriptions", func(t *testing.T) {
		mockStore := storemock.NewMockStore()
		mockLogger := mocklogging.NewMockLogger()
		service := subscription.NewService(mockLogger, mockStore)
		assert.NotNil(t, service)

		expected := []subscription.Subscription{}
		mockStore.ExpectList(context.Background(), expected, nil)

		subs, err := service.ListSubscriptions(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, subs)

		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})

	t.Run("list subscriptions error", func(t *testing.T) {
		mockStore := storemock.NewMockStore()
		mockLogger := mocklogging.NewMockLogger()
		service := subscription.NewService(mockLogger, mockStore)
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

		if err := mockStore.Verify(); err != nil {
			t.Errorf("store expectations not met: %v", err)
		}
		if err := mockLogger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})
}
