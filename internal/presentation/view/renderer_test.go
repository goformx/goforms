package view_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/goformx/goforms/internal/presentation/view"
	mocklogging "github.com/goformx/goforms/test/mocks/logging"
)

func TestNewRenderer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)
	renderer := view.NewRenderer(logger)

	assert.NotNil(t, renderer)
	assert.Implements(t, (*view.Renderer)(nil), renderer)
}

func TestRenderer_Render_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)
	renderer := view.NewRenderer(logger)

	// Create a simple test component
	component := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte("<div>Test Content</div>"))
		return err
	})

	// Create Echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test successful rendering
	err := renderer.Render(c, component)
	assert.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "<div>Test Content</div>")
}

func TestRenderer_Render_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)
	logger.EXPECT().Error("failed to render template", "error", gomock.Any(), "template", gomock.Any()).Return()

	renderer := view.NewRenderer(logger)

	// Create a component that returns an error
	component := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return errors.New("render error")
	})

	// Create Echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test error handling
	err := renderer.Render(c, component)
	assert.Error(t, err)

	// Verify it's an HTTP error
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.Equal(t, "Failed to render page", httpErr.Message)
}

func TestRenderer_Render_NilComponent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)
	logger.EXPECT().Error("failed to render template", "error", "nil component", "template", nil).Return()

	renderer := view.NewRenderer(logger)

	// Create Echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test with nil component
	err := renderer.Render(c, nil)
	assert.Error(t, err)

	// Verify it's an HTTP error
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
}

func TestRenderer_Render_NilContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)
	logger.EXPECT().Error("failed to render template", "error", "nil context", "template", nil).Return()

	renderer := view.NewRenderer(logger)

	// Create a simple test component
	component := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte("<div>Test Content</div>"))
		return err
	})

	// Test with nil context
	err := renderer.Render(nil, component)
	assert.Error(t, err)

	// Verify it's an HTTP error
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
}

func TestRenderer_InterfaceCompliance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)
	renderer := view.NewRenderer(logger)

	// Verify the renderer implements the Renderer interface
	var _ view.Renderer = renderer
}

func TestRenderer_Construction(t *testing.T) {
	var _ view.Renderer
}
