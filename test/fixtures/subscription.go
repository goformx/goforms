package fixtures

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/models"
)

// SubscriptionFixture contains test data and helpers for subscription tests
type SubscriptionFixture struct {
	Echo    *echo.Echo
	Handler echo.HandlerFunc
}

const ValidOrigin = "https://jonesrussell.github.io/me"

// NewSubscriptionFixture creates a new subscription test fixture
func NewSubscriptionFixture(handler echo.HandlerFunc) *SubscriptionFixture {
	return &SubscriptionFixture{
		Echo:    echo.New(),
		Handler: handler,
	}
}

// CreateSubscriptionRequest creates a test request for subscription creation
func (f *SubscriptionFixture) CreateSubscriptionRequest(email string) (*httptest.ResponseRecorder, error) {
	requestBody := map[string]string{
		"email": email,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderOrigin, ValidOrigin)
	rec := httptest.NewRecorder()
	c := f.Echo.NewContext(req, rec)

	if err := f.Handler(c); err != nil {
		he, ok := err.(*echo.HTTPError)
		if ok {
			rec.Code = he.Code
			_ = json.NewEncoder(rec.Body).Encode(map[string]string{
				"error": he.Message.(string),
			})
		}
	}

	return rec, nil
}

// CreateSubscriptionRequestWithOrigin creates a test request with custom origin
func (f *SubscriptionFixture) CreateSubscriptionRequestWithOrigin(email, origin string) (*httptest.ResponseRecorder, error) {
	requestBody := map[string]string{
		"email": email,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderOrigin, origin)
	rec := httptest.NewRecorder()
	c := f.Echo.NewContext(req, rec)

	if err := f.Handler(c); err != nil {
		he, ok := err.(*echo.HTTPError)
		if ok {
			rec.Code = he.Code
			if err := json.NewEncoder(rec.Body).Encode(map[string]string{
				"error": he.Message.(string),
			}); err != nil {
				return nil, fmt.Errorf("failed to encode error response: %w", err)
			}
		} else {
			// Handle non-HTTP errors
			rec.Code = http.StatusInternalServerError
			if err := json.NewEncoder(rec.Body).Encode(map[string]string{
				"error": "Internal server error",
			}); err != nil {
				return nil, fmt.Errorf("failed to encode error response: %w", err)
			}
		}
	}

	return rec, nil
}

// ValidTestSubscription returns a valid subscription for testing
func ValidTestSubscription() *models.Subscription {
	return &models.Subscription{
		Email: "test@example.com",
	}
}

// ParseResponse parses the response body into the given interface
func ParseResponse(rec *httptest.ResponseRecorder, v interface{}) error {
	return json.NewDecoder(rec.Body).Decode(v)
}
