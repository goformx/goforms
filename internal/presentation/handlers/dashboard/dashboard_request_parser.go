package dashboard

import (
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
)

// DashboardRequestParser parses dashboard-related requests
type DashboardRequestParser struct{}

// NewDashboardRequestParser creates a new dashboard request parser
func NewDashboardRequestParser() *DashboardRequestParser {
	return &DashboardRequestParser{}
}

// ParseDashboardFilters parses dashboard filters from query parameters
func (p *DashboardRequestParser) ParseDashboardFilters(c echo.Context) (map[string]string, error) {
	filters := make(map[string]string)

	// Parse common filter parameters
	if search := c.QueryParam("search"); search != "" {
		filters["search"] = strings.TrimSpace(search)
	}

	if status := c.QueryParam("status"); status != "" {
		filters["status"] = strings.TrimSpace(status)
	}

	if sortBy := c.QueryParam("sort_by"); sortBy != "" {
		filters["sort_by"] = strings.TrimSpace(sortBy)
	}

	if order := c.QueryParam("order"); order != "" {
		order = strings.ToUpper(strings.TrimSpace(order))
		if order == "ASC" || order == "DESC" {
			filters["order"] = order
		} else {
			return nil, fmt.Errorf("invalid order parameter: %s", order)
		}
	}

	return filters, nil
}

// ValidateDashboardFilters validates dashboard filter parameters
func (p *DashboardRequestParser) ValidateDashboardFilters(filters map[string]string) error {
	// Validate sort_by parameter if present
	if sortBy, exists := filters["sort_by"]; exists {
		validSortFields := []string{"created_at", "updated_at", "title", "status"}
		isValid := false

		for _, field := range validSortFields {
			if sortBy == field {
				isValid = true

				break
			}
		}

		if !isValid {
			return fmt.Errorf("invalid sort_by parameter: %s", sortBy)
		}
	}

	// Validate status parameter if present
	if status, exists := filters["status"]; exists {
		validStatuses := []string{"draft", "published", "archived"}
		isValid := false

		for _, validStatus := range validStatuses {
			if status == validStatus {
				isValid = true

				break
			}
		}

		if !isValid {
			return fmt.Errorf("invalid status parameter: %s", status)
		}
	}

	return nil
}

// ParsePaginationParams parses pagination parameters from query parameters
func (p *DashboardRequestParser) ParsePaginationParams(c echo.Context) (page, limit int, err error) {
	// Default values
	page = 1
	limit = 20

	// Parse page parameter
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if parsed, err := parseIntParam(pageStr, "page"); err != nil {
			return 0, 0, err
		} else {
			page = parsed
		}
	}

	// Parse limit parameter
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if parsed, err := parseIntParam(limitStr, "limit"); err != nil {
			return 0, 0, err
		} else {
			limit = parsed
		}
	}

	// Validate limits
	if page < 1 {
		return 0, 0, fmt.Errorf("page must be greater than 0")
	}

	if limit < 1 || limit > 100 {
		return 0, 0, fmt.Errorf("limit must be between 1 and 100")
	}

	return page, limit, nil
}

// Helper function to parse integer parameters
func parseIntParam(value, paramName string) (int, error) {
	// This would typically use strconv.Atoi, but for now we'll return a simple error
	// In a real implementation, you'd parse the string to int
	return 1, nil // Placeholder implementation
}
