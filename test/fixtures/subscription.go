package fixtures

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/pkg/errors"
)

// SubscriptionFixture contains test data and helpers for subscription tests
type SubscriptionFixture struct {
	Echo    *echo.Echo
	Handler echo.HandlerFunc
	Method  string
	URL     string
	Body    *bytes.Buffer
}

const ValidOrigin = "https://jonesrussell.github.io/me"

// NewSubscriptionFixture creates a new subscription test fixture
func NewSubscriptionFixture(handler echo.HandlerFunc) *SubscriptionFixture {
	return &SubscriptionFixture{
		Echo:    echo.New(),
		Handler: handler,
		Method:  http.MethodPost,
		URL:     "/api/v1/subscriptions",
		Body:    bytes.NewBuffer(nil),
	}
}

// CreateSubscriptionRequest creates a test request for subscription creation
func (f *SubscriptionFixture) CreateSubscriptionRequest(email string) (*httptest.ResponseRecorder, error) {
	requestBody := map[string]string{
		"email": email,
	}

	body, marshalErr := json.Marshal(requestBody)
	if marshalErr != nil {
		return nil, marshalErr
	}

	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderOrigin, ValidOrigin)
	rec := httptest.NewRecorder()
	c := f.Echo.NewContext(req, rec)

	if handlerErr := f.Handler(c); handlerErr != nil {
		return nil, handleRequestError(rec, handlerErr)
	}

	return rec, nil
}

// CreateSubscriptionRequestWithOrigin creates a subscription request with origin
func (f *SubscriptionFixture) CreateSubscriptionRequestWithOrigin(
	email string,
	origin string,
) (*httptest.ResponseRecorder, error) {
	requestBody := map[string]string{
		"email": email,
	}

	body, marshalErr := json.Marshal(requestBody)
	if marshalErr != nil {
		return nil, marshalErr
	}

	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderOrigin, origin)
	rec := httptest.NewRecorder()
	c := f.Echo.NewContext(req, rec)

	if handlerErr := f.Handler(c); handlerErr != nil {
		return nil, handleRequestError(rec, handlerErr)
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

func handleRequestError(rec *httptest.ResponseRecorder, handlerErr error) error {
	var httpErr *echo.HTTPError
	if !errors.As(handlerErr, &httpErr) {
		rec.Code = http.StatusInternalServerError
		if encodeErr := json.NewEncoder(rec.Body).Encode(map[string]string{
			"error": "Internal server error",
		}); encodeErr != nil {
			return fmt.Errorf("failed to encode error response: %w", encodeErr)
		}
		return handlerErr
	}

	rec.Code = httpErr.Code
	errResp, handleErr := handleError(httpErr)
	if handleErr != nil {
		return handleErr
	}
	if encodeErr := json.NewEncoder(rec.Body).Encode(errResp); encodeErr != nil {
		return encodeErr
	}
	return handlerErr
}

func (f *SubscriptionFixture) CreateSubscription(email string) (*httptest.ResponseRecorder, error) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/subscriptions",
		strings.NewReader(fmt.Sprintf(`{"email":%q}`, email)),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := f.Echo.NewContext(req, rec)

	handlerErr := f.Handler(c)
	if handlerErr != nil {
		if err := handleRequestError(rec, handlerErr); err != nil {
			return nil, err
		}
		return rec, handlerErr
	}
	return rec, nil
}

func (f *SubscriptionFixture) CreateSubscriptionWithOrigin(email, origin string) (*httptest.ResponseRecorder, error) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/subscriptions",
		strings.NewReader(fmt.Sprintf(`{"email":%q}`, email)),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderOrigin, origin)
	c := f.Echo.NewContext(req, rec)

	handlerErr := f.Handler(c)
	if handlerErr != nil {
		if err := handleRequestError(rec, handlerErr); err != nil {
			return nil, err
		}
		return rec, handlerErr
	}
	return rec, nil
}

func handleError(he *echo.HTTPError) (map[string]string, error) {
	msg, ok := he.Message.(string)
	if !ok {
		return nil, errors.New("invalid error message type")
	}
	return map[string]string{
		"error": msg,
	}, nil
}

func (f *SubscriptionFixture) Execute() (*http.Response, error) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(f.Method, f.URL, f.Body)
	c := f.Echo.NewContext(req, rec)

	if handlerErr := f.Handler(c); handlerErr != nil {
		if err := handleRequestError(rec, handlerErr); err != nil {
			return nil, err
		}
		return rec.Result(), handlerErr
	}

	return rec.Result(), nil
}
