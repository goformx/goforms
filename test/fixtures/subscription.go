package fixtures

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
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
		if he, ok := err.(*echo.HTTPError); ok {
			msg, ok := he.Message.(string)
			if !ok {
				return nil, fmt.Errorf("invalid error message type")
			}
			rec.Body.WriteString(fmt.Sprintf(`{"error": "%s"}`, msg))
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
		if he, ok := err.(*echo.HTTPError); ok {
			msg, ok := he.Message.(string)
			if !ok {
				return nil, fmt.Errorf("invalid error message type")
			}
			rec.Body.WriteString(fmt.Sprintf(`{"error": "%s"}`, msg))
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
func ValidTestSubscription() *subscription.Subscription {
	return &subscription.Subscription{
		Email: "test@example.com",
		Name:  "Test User",
	}
}

// ParseResponse parses the response body into the given interface
func ParseResponse(rec *httptest.ResponseRecorder, v any) error {
	return json.NewDecoder(rec.Body).Decode(v)
}
