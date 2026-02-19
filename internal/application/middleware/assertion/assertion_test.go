package assertion_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goformx/goforms/internal/application/middleware/assertion"
	"github.com/goformx/goforms/internal/application/middleware/context"
	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerify_ValidSignature_Passes(t *testing.T) {
	secret := "test-secret-key"
	userID := "user-123"
	timestamp := time.Now().Format(time.RFC3339)
	payload := userID + ":" + timestamp
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	cfg := &appconfig.Config{
		Security: appconfig.SecurityConfig{
			Assertion: appconfig.AssertionConfig{
				Secret:               secret,
				TimestampSkewSeconds: 60,
			},
		},
	}
	mw := assertion.NewMiddleware(cfg, nil)
	e := echo.New()
	e.Use(mw.Verify())
	e.GET("/test", func(c echo.Context) error {
		uid, ok := context.GetUserID(c)
		require.True(t, ok)
		assert.Equal(t, userID, uid)

		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("X-User-Id", userID)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Signature", signature)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

func TestVerify_InvalidSignature_Returns401(t *testing.T) {
	cfg := &appconfig.Config{
		Security: appconfig.SecurityConfig{
			Assertion: appconfig.AssertionConfig{
				Secret:               "correct-secret",
				TimestampSkewSeconds: 60,
			},
		},
	}
	mw := assertion.NewMiddleware(cfg, nil)
	e := echo.New()
	e.Use(mw.Verify())
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("X-User-Id", "user-123")
	req.Header.Set("X-Timestamp", time.Now().Format(time.RFC3339))
	req.Header.Set("X-Signature", "invalid-signature-hex")

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "unauthorized")
}

func TestVerify_MissingHeaders_Returns401(t *testing.T) {
	secret := "test-secret"
	userID := "user-456"
	timestamp := time.Now().Format(time.RFC3339)
	payload := userID + ":" + timestamp
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	cfg := &appconfig.Config{
		Security: appconfig.SecurityConfig{
			Assertion: appconfig.AssertionConfig{
				Secret:               secret,
				TimestampSkewSeconds: 60,
			},
		},
	}
	mw := assertion.NewMiddleware(cfg, nil)
	e := echo.New()
	e.Use(mw.Verify())
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	tests := []struct {
		name      string
		userID    string
		timestamp string
		signature string
	}{
		{"missing user id", "", timestamp, signature},
		{"missing timestamp", userID, "", signature},
		{"missing signature", userID, timestamp, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			if tt.userID != "" {
				req.Header.Set("X-User-Id", tt.userID)
			}
			if tt.timestamp != "" {
				req.Header.Set("X-Timestamp", tt.timestamp)
			}
			if tt.signature != "" {
				req.Header.Set("X-Signature", tt.signature)
			}

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.Contains(t, rec.Body.String(), "unauthorized")
		})
	}
}

func TestVerify_StaleTimestamp_Returns401(t *testing.T) {
	secret := "test-secret"
	userID := "user-789"
	staleTime := time.Now().Add(-2 * time.Minute)
	timestamp := staleTime.Format(time.RFC3339)
	payload := userID + ":" + timestamp
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	cfg := &appconfig.Config{
		Security: appconfig.SecurityConfig{
			Assertion: appconfig.AssertionConfig{
				Secret:               secret,
				TimestampSkewSeconds: 60,
			},
		},
	}
	mw := assertion.NewMiddleware(cfg, nil)
	e := echo.New()
	e.Use(mw.Verify())
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("X-User-Id", userID)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Signature", signature)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "unauthorized")
}
