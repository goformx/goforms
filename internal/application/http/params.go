package response

import (
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"
)

// ParseInt64Param parses an int64 parameter from the request
func ParseInt64Param(c echo.Context, name string) (int64, error) {
	param := c.Param(name)
	if param == "" {
		return 0, fmt.Errorf("missing %s parameter", name)
	}

	value, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter: %v", name, err)
	}

	return value, nil
}
