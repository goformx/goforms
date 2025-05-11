package handler_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jonesrussell/goforms/internal/application/handler"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBase(t *testing.T) {
	t.Run("with logger", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		logger := mocklogging.NewMockLogger(ctrl)
		base := handler.NewBase(handler.WithLogger(logger))
		assert.NotNil(t, base)
		assert.Equal(t, logger, base.Logger)
	})

	t.Run("without logger", func(t *testing.T) {
		base := handler.NewBase()
		assert.NotNil(t, base)
		assert.Nil(t, base.Logger)
	})
}

func TestBase_Validate(t *testing.T) {
	t.Run("valid when logger set", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		logger := mocklogging.NewMockLogger(ctrl)
		base := handler.NewBase(handler.WithLogger(logger))
		err := base.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid when logger missing", func(t *testing.T) {
		base := handler.NewBase()
		err := base.Validate()
		require.Error(t, err)
		assert.Equal(t, "logger is required", err.Error())
	})
}

func TestBase_WrapResponseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := mocklogging.NewMockLogger(ctrl)
	base := handler.NewBase(handler.WithLogger(logger))

	t.Run("wraps error with message", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := base.WrapResponseError(originalErr, "wrapped message")
		require.Error(t, wrappedErr)
		require.ErrorIs(t, wrappedErr, originalErr)
		assert.Equal(t, "wrapped message: original error", wrappedErr.Error())
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		wrappedErr := base.WrapResponseError(nil, "wrapped message")
		assert.NoError(t, wrappedErr)
	})
}

func TestBase_LogError(t *testing.T) {
	t.Run("logs error with fields", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		logger := mocklogging.NewMockLogger(ctrl)
		base := handler.NewBase(handler.WithLogger(logger))

		err := errors.New("test error")
		logger.EXPECT().Error("test message", gomock.Any()).Times(1)

		base.LogError("test message", err, logging.StringField("key", "value"))
	})
}
