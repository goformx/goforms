package openapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goformx/goforms/internal/application/openapi"
)

// Test data - minimal OpenAPI spec for route cache testing
const routeCacheTestOpenAPISpec = `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      summary: Test endpoint
      responses:
        '200':
          description: OK
  /test/{id}:
    get:
      summary: Test endpoint with path parameter
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
`

func createRouteCacheTestRouter(t *testing.T) routers.Router {
	t.Helper()

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(routeCacheTestOpenAPISpec))
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	router, err := gorillamux.NewRouter(doc)
	require.NoError(t, err)

	return router
}

func createRouteCacheTestEchoContext(t *testing.T) echo.Context {
	t.Helper()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()

	return e.NewContext(req, rec)
}

func TestNewRouteCache(t *testing.T) {
	cache := openapi.NewRouteCache()

	assert.NotNil(t, cache)
	assert.Implements(t, (*openapi.RouteCache)(nil), cache)
}

func TestRouteCache_Get_Empty(t *testing.T) {
	cache := openapi.NewRouteCache()
	c := createRouteCacheTestEchoContext(t)

	route, pathParams, ok := cache.Get(c)

	assert.Nil(t, route)
	assert.Nil(t, pathParams)
	assert.False(t, ok)
}

func TestRouteCache_SetAndGet(t *testing.T) {
	cache := openapi.NewRouteCache()
	c := createRouteCacheTestEchoContext(t)
	router := createRouteCacheTestRouter(t)

	// Find a route to cache
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	// Set the route in cache
	cache.Set(c, route, pathParams)

	// Get the route from cache
	retrievedRoute, retrievedPathParams, ok := cache.Get(c)

	assert.True(t, ok)
	assert.Equal(t, route, retrievedRoute)
	assert.Equal(t, pathParams, retrievedPathParams)
}

func TestRouteCache_SetAndGet_WithPathParams(t *testing.T) {
	cache := openapi.NewRouteCache()
	c := createRouteCacheTestEchoContext(t)
	router := createRouteCacheTestRouter(t)

	// Find a route with path parameters
	req := httptest.NewRequest(http.MethodGet, "/test/123", http.NoBody)
	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	// Set the route in cache
	cache.Set(c, route, pathParams)

	// Get the route from cache
	retrievedRoute, retrievedPathParams, ok := cache.Get(c)

	assert.True(t, ok)
	assert.Equal(t, route, retrievedRoute)
	assert.Equal(t, pathParams, retrievedPathParams)
	assert.Equal(t, "123", retrievedPathParams["id"])
}

func TestRouteCache_Get_DifferentContext(t *testing.T) {
	cache := openapi.NewRouteCache()
	c1 := createRouteCacheTestEchoContext(t)
	c2 := createRouteCacheTestEchoContext(t)
	router := createRouteCacheTestRouter(t)

	// Find a route
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	// Set the route in cache for context 1
	cache.Set(c1, route, pathParams)

	// Try to get the route from context 2 (should fail)
	retrievedRoute, retrievedPathParams, ok := cache.Get(c2)

	assert.False(t, ok)
	assert.Nil(t, retrievedRoute)
	assert.Nil(t, retrievedPathParams)
}

func TestRouteCache_Overwrite(t *testing.T) {
	cache := openapi.NewRouteCache()
	c := createRouteCacheTestEchoContext(t)
	router := createRouteCacheTestRouter(t)

	// Find first route
	req1 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	route1, pathParams1, err := router.FindRoute(req1)
	require.NoError(t, err)

	// Find second route
	req2 := httptest.NewRequest(http.MethodGet, "/test/456", http.NoBody)
	route2, pathParams2, err := router.FindRoute(req2)
	require.NoError(t, err)

	// Set first route
	cache.Set(c, route1, pathParams1)

	// Overwrite with second route
	cache.Set(c, route2, pathParams2)

	// Get the route (should be the second one)
	retrievedRoute, retrievedPathParams, ok := cache.Get(c)

	assert.True(t, ok)
	assert.Equal(t, route2, retrievedRoute)
	assert.Equal(t, pathParams2, retrievedPathParams)
	assert.Equal(t, "456", retrievedPathParams["id"])
}

func TestRouteCache_Set_NilValues(t *testing.T) {
	cache := openapi.NewRouteCache()
	c := createRouteCacheTestEchoContext(t)

	// Set nil values
	cache.Set(c, nil, nil)

	// Get the values
	retrievedRoute, retrievedPathParams, ok := cache.Get(c)

	assert.True(t, ok)
	assert.Nil(t, retrievedRoute)
	assert.Nil(t, retrievedPathParams)
}

func TestRouteCache_Get_AfterNilSet(t *testing.T) {
	cache := openapi.NewRouteCache()
	c := createRouteCacheTestEchoContext(t)

	// Set nil values
	cache.Set(c, nil, nil)

	// Get the values
	retrievedRoute, retrievedPathParams, ok := cache.Get(c)

	assert.True(t, ok)
	assert.Nil(t, retrievedRoute)
	assert.Nil(t, retrievedPathParams)
}

func TestRouteCache_MultipleContexts(t *testing.T) {
	cache := openapi.NewRouteCache()
	router := createRouteCacheTestRouter(t)

	// Create multiple contexts
	c1 := createRouteCacheTestEchoContext(t)
	c2 := createRouteCacheTestEchoContext(t)
	c3 := createRouteCacheTestEchoContext(t)

	// Find different routes
	req1 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	route1, pathParams1, err := router.FindRoute(req1)
	require.NoError(t, err)

	req2 := httptest.NewRequest(http.MethodGet, "/test/789", http.NoBody)
	route2, pathParams2, err := router.FindRoute(req2)
	require.NoError(t, err)

	// Set routes for different contexts
	cache.Set(c1, route1, pathParams1)
	cache.Set(c2, route2, pathParams2)

	// Verify each context has its own cached data
	retrievedRoute1, retrievedPathParams1, ok1 := cache.Get(c1)
	assert.True(t, ok1)
	assert.Equal(t, route1, retrievedRoute1)
	assert.Equal(t, pathParams1, retrievedPathParams1)

	retrievedRoute2, retrievedPathParams2, ok2 := cache.Get(c2)
	assert.True(t, ok2)
	assert.Equal(t, route2, retrievedRoute2)
	assert.Equal(t, pathParams2, retrievedPathParams2)

	// Context 3 should have no cached data
	retrievedRoute3, retrievedPathParams3, ok3 := cache.Get(c3)
	assert.False(t, ok3)
	assert.Nil(t, retrievedRoute3)
	assert.Nil(t, retrievedPathParams3)
}
