package handler

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
)

func TestNewBase(t *testing.T) {
	t.Run("creates base with logger", func(t *testing.T) {
		logger := mocklogging.NewMockLogger()
		base := NewBase(WithLogger(logger))
		assert.Equal(t, logger, base.Logger)
	})

	t.Run("creates base without logger", func(t *testing.T) {
		base := NewBase()
		assert.Nil(t, base.Logger)
	})
}

func TestBase_Validate(t *testing.T) {
	t.Run("valid when logger set", func(t *testing.T) {
		logger := mocklogging.NewMockLogger()
		base := NewBase(WithLogger(logger))
		err := base.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid when logger missing", func(t *testing.T) {
		base := NewBase()
		err := base.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "logger is required")
	})
}

func TestBase_WrapResponseError(t *testing.T) {
	logger := mocklogging.NewMockLogger()
	base := NewBase(WithLogger(logger))

	t.Run("wraps error with message", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := base.WrapResponseError(originalErr, "wrapped message")
		assert.Error(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "wrapped message")
		assert.Contains(t, wrappedErr.Error(), originalErr.Error())
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		wrappedErr := base.WrapResponseError(nil, "wrapped message")
		assert.NoError(t, wrappedErr)
	})
}

func TestBase_LogError(t *testing.T) {
	t.Run("logs error with fields", func(t *testing.T) {
		logger := mocklogging.NewMockLogger()
		base := NewBase(WithLogger(logger))

		err := errors.New("test error")

		// Set expectation that Error will be called with message and fields
		logger.ExpectError("test message").WithFields(map[string]interface{}{
			"key":   "value",
			"error": err.Error(),
		})

		base.LogError("test message", err, mocklogging.String("key", "value"))

		assert.NoError(t, logger.Verify())
	})
}
