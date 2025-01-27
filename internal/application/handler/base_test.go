package handler

import (
	"errors"
	"testing"

	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
)

func TestNewBase(t *testing.T) {
	t.Run("creates base with logger", func(t *testing.T) {
		logger := mocklogging.NewMockLogger()
		base := NewBase(WithLogger(logger))
		if base.Logger != logger {
			t.Errorf("expected logger to be set, got nil")
		}
	})

	t.Run("creates base without logger", func(t *testing.T) {
		base := NewBase()
		if base.Logger != nil {
			t.Errorf("expected nil logger, got %v", base.Logger)
		}
	})
}

func TestBase_Validate(t *testing.T) {
	t.Run("valid when logger set", func(t *testing.T) {
		logger := mocklogging.NewMockLogger()
		base := NewBase(WithLogger(logger))
		err := base.Validate()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid when logger missing", func(t *testing.T) {
		base := NewBase()
		err := base.Validate()
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err != nil && err.Error() != "logger is required" {
			t.Errorf("expected 'logger is required' error, got %v", err)
		}
	})
}

func TestBase_WrapResponseError(t *testing.T) {
	logger := mocklogging.NewMockLogger()
	base := NewBase(WithLogger(logger))

	t.Run("wraps error with message", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := base.WrapResponseError(originalErr, "wrapped message")
		if wrappedErr == nil {
			t.Error("expected error, got nil")
		}
		if !errors.Is(wrappedErr, originalErr) {
			t.Errorf("expected wrapped error to contain original error")
		}
		expectedMsg := "wrapped message: original error"
		if wrappedErr.Error() != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, wrappedErr.Error())
		}
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		wrappedErr := base.WrapResponseError(nil, "wrapped message")
		if wrappedErr != nil {
			t.Errorf("expected nil error, got %v", wrappedErr)
		}
	})
}

func TestBase_LogError(t *testing.T) {
	t.Run("logs error with fields", func(t *testing.T) {
		logger := mocklogging.NewMockLogger()
		base := NewBase(WithLogger(logger))

		err := errors.New("test error")
		logger.ExpectError("test message").WithFields(map[string]interface{}{
			"key":   "value",
			"error": err.Error(),
		})

		base.LogError("test message", err, mocklogging.String("key", "value"))

		if err := logger.Verify(); err != nil {
			t.Errorf("logger expectations not met: %v", err)
		}
	})
}
