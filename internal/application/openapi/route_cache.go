package openapi

import (
	"github.com/getkin/kin-openapi/routers"
	"github.com/labstack/echo/v4"
)

// routeCache implements RouteCache
type routeCache struct{}

// NewRouteCache creates a new route cache
func NewRouteCache() *routeCache {
	return &routeCache{}
}

// Get retrieves cached route and pathParams from context
func (rc *routeCache) Get(c echo.Context) (*routers.Route, map[string]string, bool) {
	r, rok := c.Get(openapiRouteKey).(*routers.Route)
	p, pok := c.Get(openapiPathParamsKey).(map[string]string)

	return r, p, rok && pok
}

// Set stores route and pathParams in context
func (rc *routeCache) Set(c echo.Context, route *routers.Route, pathParams map[string]string) {
	c.Set(openapiRouteKey, route)
	c.Set(openapiPathParamsKey, pathParams)
}
