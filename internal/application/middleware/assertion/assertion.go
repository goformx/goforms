package assertion

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/goformx/goforms/internal/application/middleware/context"
	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/labstack/echo/v4"
)

const (
	headerUserID    = "X-User-Id"
	headerTimestamp = "X-Timestamp"
	headerSignature = "X-Signature"
)

// Middleware verifies Laravel signed assertion headers and sets user_id in Echo context.
type Middleware struct {
	config *appconfig.Config
}

// NewMiddleware creates a new assertion verification middleware.
func NewMiddleware(config *appconfig.Config) *Middleware {
	return &Middleware{config: config}
}

// Verify returns an Echo middleware that verifies X-User-Id, X-Timestamp, X-Signature headers.
func (m *Middleware) Verify() echo.MiddlewareFunc {
	cfg := m.config.Security.Assertion

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID := strings.TrimSpace(c.Request().Header.Get(headerUserID))
			timestamp := strings.TrimSpace(c.Request().Header.Get(headerTimestamp))
			signature := strings.TrimSpace(c.Request().Header.Get(headerSignature))

			if userID == "" || timestamp == "" || signature == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			ts, err := parseTimestamp(timestamp)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			skew := time.Duration(cfg.TimestampSkewSeconds) * time.Second
			if time.Since(ts) > skew {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			payload := userID + ":" + timestamp
			expected := computeHMAC(cfg.Secret, payload)

			sigBytes, err := hex.DecodeString(signature)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			if !hmacEqual(sigBytes, expected) {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			context.SetUserID(c, userID)

			return next(c)
		}
	}
}

func parseTimestamp(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		if sec, parseErr := strconv.ParseInt(s, 10, 64); parseErr == nil {
			return time.Unix(sec, 0), nil
		}

		return time.Time{}, err
	}

	return t, nil
}

func computeHMAC(secret, payload string) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))

	return h.Sum(nil)
}

func hmacEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	return subtle.ConstantTimeCompare(a, b) == 1
}
