package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/jonesrussell/goforms/internal/domain/subscription"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/client"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	logger := mocklogging.NewMockLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/signup":
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(&user.User{ID: 1, Email: "test@example.com"})
		case "/api/v1/auth/login":
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&user.TokenPair{AccessToken: "token"})
		case "/api/v1/auth/logout":
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.WriteHeader(http.StatusOK)
		case "/api/v1/contact":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
			} else if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode([]contact.Submission{{ID: 1}})
			}
		case "/api/v1/contact/1":
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&contact.Submission{ID: 1})
			} else if r.Method == http.MethodPut {
				w.WriteHeader(http.StatusOK)
			}
		case "/api/v1/subscriptions":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
			} else if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode([]subscription.Subscription{{ID: 1}})
			}
		case "/api/v1/subscriptions/1":
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&subscription.Subscription{ID: 1})
			} else if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusNoContent)
			}
		case "/api/v1/subscriptions/1/status":
			if r.Method == http.MethodPut {
				w.WriteHeader(http.StatusOK)
			}
		case "/v1/version":
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&client.VersionInfo{
					Version:   "1.0.0",
					BuildTime: time.Now().Format(time.RFC3339),
					GitCommit: "abc123",
					GoVersion: "1.24",
				})
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	c := client.NewClient(ts.URL, logger)

	t.Run("Auth API", func(t *testing.T) {
		t.Run("SignUp", func(t *testing.T) {
			user, err := c.SignUp(context.Background(), &user.Signup{
				Email:    "test@example.com",
				Password: "password",
			})
			require.NoError(t, err)
			assert.Equal(t, int64(1), user.ID)
			assert.Equal(t, "test@example.com", user.Email)
		})

		t.Run("Login", func(t *testing.T) {
			tokens, err := c.Login(context.Background(), &user.Login{
				Email:    "test@example.com",
				Password: "password",
			})
			require.NoError(t, err)
			assert.Equal(t, "token", tokens.AccessToken)
		})

		t.Run("Logout", func(t *testing.T) {
			err := c.Logout(context.Background(), "token")
			require.NoError(t, err)
		})
	})

	t.Run("Contact API", func(t *testing.T) {
		t.Run("SubmitContactForm", func(t *testing.T) {
			err := c.SubmitContactForm(context.Background(), &contact.Submission{
				ID: 1,
			})
			require.NoError(t, err)
		})

		t.Run("ListContactSubmissions", func(t *testing.T) {
			submissions, err := c.ListContactSubmissions(context.Background())
			require.NoError(t, err)
			assert.Len(t, submissions, 1)
			assert.Equal(t, int64(1), submissions[0].ID)
		})

		t.Run("GetContactSubmission", func(t *testing.T) {
			submission, err := c.GetContactSubmission(context.Background(), 1)
			require.NoError(t, err)
			assert.Equal(t, int64(1), submission.ID)
		})

		t.Run("UpdateContactSubmissionStatus", func(t *testing.T) {
			err := c.UpdateContactSubmissionStatus(context.Background(), 1, contact.StatusApproved)
			require.NoError(t, err)
		})
	})

	t.Run("Subscription API", func(t *testing.T) {
		t.Run("CreateSubscription", func(t *testing.T) {
			err := c.CreateSubscription(context.Background(), &subscription.Subscription{
				ID: 1,
			})
			require.NoError(t, err)
		})

		t.Run("ListSubscriptions", func(t *testing.T) {
			subs, err := c.ListSubscriptions(context.Background())
			require.NoError(t, err)
			assert.Len(t, subs, 1)
			assert.Equal(t, int64(1), subs[0].ID)
		})

		t.Run("GetSubscription", func(t *testing.T) {
			sub, err := c.GetSubscription(context.Background(), 1)
			require.NoError(t, err)
			assert.Equal(t, int64(1), sub.ID)
		})

		t.Run("UpdateSubscriptionStatus", func(t *testing.T) {
			err := c.UpdateSubscriptionStatus(context.Background(), 1, subscription.StatusActive)
			require.NoError(t, err)
		})

		t.Run("DeleteSubscription", func(t *testing.T) {
			err := c.DeleteSubscription(context.Background(), 1)
			require.NoError(t, err)
		})
	})

	t.Run("Version API", func(t *testing.T) {
		info, err := c.GetVersion(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", info.Version)
		assert.Equal(t, "abc123", info.GitCommit)
		assert.Equal(t, "1.24", info.GoVersion)
	})
} 