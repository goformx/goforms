package client_test

import (
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

func setupTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/signup":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(&user.User{ID: 1, Email: "test@example.com"})
				return
			}
		case "/api/v1/auth/login":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&user.TokenPair{AccessToken: "token"})
				return
			}
		case "/api/v1/auth/logout":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusOK)
				return
			}
		case "/api/v1/contact":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
				return
			}
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode([]contact.Submission{{ID: 1}})
				return
			}
		case "/api/v1/contact/1":
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&contact.Submission{ID: 1})
				return
			}
			if r.Method == http.MethodPut {
				w.WriteHeader(http.StatusOK)
				return
			}
		case "/api/v1/subscriptions":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
				return
			}
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode([]subscription.Subscription{{ID: 1}})
				return
			}
		case "/api/v1/subscriptions/1":
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&subscription.Subscription{ID: 1})
				return
			}
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		case "/api/v1/subscriptions/1/status":
			if r.Method == http.MethodPut {
				w.WriteHeader(http.StatusOK)
				return
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
				return
			}
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}

func TestAuthAPI(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	c := client.NewClient(ts.URL, mocklogging.NewMockLogger())

	t.Run("SignUp", func(t *testing.T) {
		newUser, err := c.SignUp(t.Context(), &user.Signup{
			Email:    "test@example.com",
			Password: "password",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), newUser.ID)
		assert.Equal(t, "test@example.com", newUser.Email)
	})

	t.Run("Login", func(t *testing.T) {
		tokenPair, err := c.Login(t.Context(), &user.Login{
			Email:    "test@example.com",
			Password: "password",
		})
		require.NoError(t, err)
		assert.Equal(t, "token", tokenPair.AccessToken)
	})

	t.Run("Logout", func(t *testing.T) {
		err := c.Logout(t.Context(), "token")
		require.NoError(t, err)
	})
}

func TestContactAPI(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	c := client.NewClient(ts.URL, mocklogging.NewMockLogger())

	t.Run("SubmitContactForm", func(t *testing.T) {
		err := c.SubmitContactForm(t.Context(), &contact.Submission{
			ID: 1,
		})
		require.NoError(t, err)
	})

	t.Run("ListContactSubmissions", func(t *testing.T) {
		submissions, err := c.ListContactSubmissions(t.Context())
		require.NoError(t, err)
		assert.Len(t, submissions, 1)
		assert.Equal(t, int64(1), submissions[0].ID)
	})

	t.Run("GetContactSubmission", func(t *testing.T) {
		submission, err := c.GetContactSubmission(t.Context(), 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), submission.ID)
	})

	t.Run("UpdateContactSubmissionStatus", func(t *testing.T) {
		err := c.UpdateContactSubmissionStatus(t.Context(), 1, contact.StatusApproved)
		require.NoError(t, err)
	})
}

func TestSubscriptionAPI(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	c := client.NewClient(ts.URL, mocklogging.NewMockLogger())

	t.Run("CreateSubscription", func(t *testing.T) {
		err := c.CreateSubscription(t.Context(), &subscription.Subscription{
			ID: 1,
		})
		require.NoError(t, err)
	})

	t.Run("ListSubscriptions", func(t *testing.T) {
		subs, err := c.ListSubscriptions(t.Context())
		require.NoError(t, err)
		assert.Len(t, subs, 1)
		assert.Equal(t, int64(1), subs[0].ID)
	})

	t.Run("GetSubscription", func(t *testing.T) {
		sub, err := c.GetSubscription(t.Context(), 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), sub.ID)
	})

	t.Run("UpdateSubscriptionStatus", func(t *testing.T) {
		err := c.UpdateSubscriptionStatus(t.Context(), 1, subscription.StatusActive)
		require.NoError(t, err)
	})

	t.Run("DeleteSubscription", func(t *testing.T) {
		err := c.DeleteSubscription(t.Context(), 1)
		require.NoError(t, err)
	})
}

func TestVersionAPI(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	c := client.NewClient(ts.URL, mocklogging.NewMockLogger())

	info, err := c.GetVersion(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "abc123", info.GitCommit)
	assert.Equal(t, "1.24", info.GoVersion)
}
