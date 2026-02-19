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
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

const (
	headerUserID    = "X-User-Id"
	headerTimestamp = "X-Timestamp"
	headerSignature = "X-Signature"

	// FailureReasonContextKey is the Echo context key set when assertion verification fails (value: reason string).
	// The request logging middleware can include it in the "request completed with client error" log.
	FailureReasonContextKey = "assertion_failure_reason"
)

// Middleware verifies Laravel signed assertion headers and sets user_id in Echo context.
type Middleware struct {
	config *appconfig.Config
	logger logging.Logger
}

// NewMiddleware creates a new assertion verification middleware.
// logger may be nil; if set, it is used to log assertion failures so the 401 reason is always visible.
func NewMiddleware(config *appconfig.Config, logger logging.Logger) *Middleware {
	return &Middleware{config: config, logger: logger}
}

// Verify returns an Echo middleware that verifies X-User-Id, X-Timestamp, X-Signature headers.
func (m *Middleware) Verify() echo.MiddlewareFunc {
	cfg := m.config.Security.Assertion

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, failReason := verifyAssertionHeaders(c.Request().Header, cfg)
			if failReason != "" {
				m.logFailure(c, failReason)

				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			context.SetUserID(c, userID)

			return next(c)
		}
	}
}

// verifyAssertionHeaders checks headers and config; returns (userID, "") on success or ("", reason) on failure.
func verifyAssertionHeaders(headers http.Header, cfg appconfig.AssertionConfig) (userID, failureReason string) {
	userID = strings.TrimSpace(headers.Get(headerUserID))
	timestamp := strings.TrimSpace(headers.Get(headerTimestamp))
	signature := strings.TrimSpace(headers.Get(headerSignature))

	if userID == "" || timestamp == "" || signature == "" {
		return "", "missing_headers"
	}

	if cfg.Secret == "" {
		return "", "empty_secret"
	}

	ts, err := parseTimestamp(timestamp)
	if err != nil {
		return "", "timestamp_parse_error"
	}

	skew := time.Duration(cfg.TimestampSkewSeconds) * time.Second
	if time.Since(ts) > skew {
		return "", "timestamp_too_old"
	}

	payload := userID + ":" + timestamp
	expected := computeHMAC(cfg.Secret, payload)

	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return "", "signature_not_hex"
	}

	if !hmacEqual(sigBytes, expected) {
		return "", "signature_mismatch"
	}

	return userID, ""
}

func (m *Middleware) logFailure(c echo.Context, reason string) {
	c.Set(FailureReasonContextKey, reason)
	msg := "assertion verification failed"
	var logFn func(string, ...any)
	if m.logger != nil {
		logFn = m.logger.Warn
	} else if logger := context.GetLogger(c.Request().Context()); logger != nil {
		logFn = logger.Warn
	} else {
		c.Logger().Warn(msg, "reason", reason, "path", c.Path())

		return
	}
	logFn(msg, "reason", reason, "path", c.Path())
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
