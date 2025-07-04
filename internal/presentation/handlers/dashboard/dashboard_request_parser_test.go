package dashboard_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goformx/goforms/internal/presentation/handlers/dashboard"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestDashboardRequestParser_ParseDashboardFilters(t *testing.T) {
	parser := dashboard.NewDashboardRequestParser()

	tests := []struct {
		name            string
		queryParams     map[string]string
		expectedFilters map[string]string
		expectError     bool
	}{
		{
			name:            "no filters",
			queryParams:     map[string]string{},
			expectedFilters: map[string]string{},
			expectError:     false,
		},
		{
			name: "valid filters",
			queryParams: map[string]string{
				"search":  "test form",
				"status":  "published",
				"sort_by": "created_at",
				"order":   "DESC",
			},
			expectedFilters: map[string]string{
				"search":  "test form",
				"status":  "published",
				"sort_by": "created_at",
				"order":   "DESC",
			},
			expectError: false,
		},
		{
			name: "invalid order parameter",
			queryParams: map[string]string{
				"order": "INVALID",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with query parameters
			req := httptest.NewRequest(http.MethodGet, "/dashboard", http.NoBody)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			// Create Echo context
			e := echo.New()
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Parse filters
			filters, err := parser.ParseDashboardFilters(c)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFilters, filters)
			}
		})
	}
}

func TestDashboardRequestParser_ValidateDashboardFilters(t *testing.T) {
	parser := dashboard.NewDashboardRequestParser()

	tests := []struct {
		name        string
		filters     map[string]string
		expectError bool
	}{
		{
			name: "valid filters",
			filters: map[string]string{
				"sort_by": "created_at",
				"status":  "published",
			},
			expectError: false,
		},
		{
			name: "invalid sort_by",
			filters: map[string]string{
				"sort_by": "invalid_field",
			},
			expectError: true,
		},
		{
			name: "invalid status",
			filters: map[string]string{
				"status": "invalid_status",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateDashboardFilters(tt.filters)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDashboardRequestParser_ParsePaginationParams(t *testing.T) {
	parser := dashboard.NewDashboardRequestParser()

	tests := []struct {
		name        string
		queryParams map[string]string
		expectError bool
	}{
		{
			name:        "no pagination params",
			queryParams: map[string]string{},
			expectError: false,
		},
		{
			name: "valid pagination params",
			queryParams: map[string]string{
				"page":  "1",
				"limit": "20",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with query parameters
			req := httptest.NewRequest(http.MethodGet, "/dashboard", http.NoBody)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			// Create Echo context
			e := echo.New()
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Parse pagination params
			page, limit, err := parser.ParsePaginationParams(c)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, page, 1)
				assert.GreaterOrEqual(t, limit, 1)
			}
		})
	}
}
