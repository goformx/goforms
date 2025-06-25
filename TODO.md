# API Response Standardization TODO

## Overview

This document outlines the work needed to standardize all API responses across the GoForms application to follow the established best practice pattern.

## Standard Response Format

All API responses should use this structure:

```json
{
  "success": true|false,
  "message": "Optional message",
  "data": { ... } // Optional payload
}
```

## Backend Changes Required

### 1. Authentication Handlers

**File: `internal/application/handlers/web/auth.go`**

- [x] **Line 86**: `return c.JSON(constants.StatusOK, map[string]string{...})` → Use `response.Success()`
- [x] **Line 149**: `return c.JSON(constants.StatusOK, map[string]string{...})` → Use `response.Success()`
- [x] **Line 213**: `return c.JSON(constants.StatusOK, map[string]string{...})` → Use `response.Success()`
- [x] **Line 243**: `return c.JSON(constants.StatusOK, schema)` → Use `response.Success()` with schema in data
- [x] **Line 251**: `return c.JSON(constants.StatusOK, schema)` → Use `response.Success()` with schema in data

### 2. Form Response Helper

**File: `internal/application/handlers/web/form_response_helper.go`**

- [x] **Line 39**: `return c.JSON(http.StatusBadRequest, &ErrorResponse{...})` → Use `response.ErrorResponse()`
- [x] **Line 43**: `return c.JSON(http.StatusBadRequest, &ErrorResponse{...})` → Use `response.ErrorResponse()`
- [x] **Line 47**: `return c.JSON(http.StatusInternalServerError, &ErrorResponse{...})` → Use `response.ErrorResponse()`
- [x] **Line 55**: `return c.JSON(http.StatusOK, &FormSuccessResponse{...})` → Use `response.Success()`
- [x] **Line 64**: `return c.JSON(http.StatusOK, &FormSuccessResponse{...})` → Use `response.Success()`

### 3. Form Web Handler

**File: `internal/application/handlers/web/form_web.go`**

- [x] **Line 153**: `return c.JSON(200, schema)` → Use `response.Success()` with schema in data

### 4. Form API Handler

**File: `internal/application/handlers/web/form_api.go`**

- [x] **Line 132**: `return c.JSON(constants.StatusOK, clientValidation)` → Use `response.Success()` with validation in data

### 5. Auth Response Builder

**File: `internal/application/handlers/web/auth_response_builder.go`**

- [x] **Line 24**: `return c.JSON(status, map[string]string{...})` → Use `response.ErrorResponse()`

### 6. Middleware

**File: `internal/application/middleware/access/middleware.go`**

- [x] **Line 43**: `return c.JSON(http.StatusForbidden, map[string]string{...})` → Use `response.ErrorResponse()`

**File: `internal/application/middleware/session/middleware.go`**

- [x] **Line 218**: `return c.JSON(http.StatusUnauthorized, map[string]string{...})` → Use `response.ErrorResponse()`

**File: `internal/application/middleware/recovery.go`**

- [x] **Line 71**: `if jsonErr := c.JSON(statusCode, domainErr);` → Use `response.ErrorResponse()`
- [x] **Line 87**: `if jsonErr := c.JSON(http.StatusInternalServerError, map[string]string{...});` → Use `response.ErrorResponse()`

### 7. Error Handler

**File: `internal/application/response/error_handler.go`**

- [x] **Line 99**: `if jsonErr := c.JSON(statusCode, map[string]any{...});` → Use `response.ErrorResponse()`
- [x] **Line 145**: `return c.JSON(statusCode, map[string]any{...});` → Use `response.ErrorResponse()`

### 8. Health Handler

**File: `internal/application/handlers/health/handler.go`**

- [x] **Line 37**: `if jsonErr := c.JSON(http.StatusServiceUnavailable, status);` → Use `response.ErrorResponse()`
- [x] **Line 43**: `if jsonErr := c.JSON(http.StatusOK, status);` → Use `response.Success()`

### 9. Server

**File: `internal/infrastructure/server/server.go`**

- [x] **Line 126**: `return c.JSON(http.StatusOK, map[string]string{...})` → Use `response.Success()`

## Frontend Changes Required

### 1. HTTP Client (COMPLETED)

**File: `src/js/core/http-client.ts`**

- [x] **Line 7**: Added `ApiResponse<T>` interface to match backend standard
- [x] **Line 84**: Updated `handleResponse` to handle standardized API response format
- [x] **Line 88**: Added success/error detection logic to check `response.success`
- [x] **Line 93**: Updated success handling to use `response.data` for payload
- [x] **Line 96**: Updated error handling to use `response.message` for errors

### 2. Response Handler (COMPLETED)

**File: `src/js/features/forms/handlers/response-handler.ts`**

- [x] **Line 16**: Updated to use `Partial<ServerResponse>` for better type safety
- [x] **Line 40**: Updated success detection logic to check `response.success === true`
- [x] **Line 42**: Updated error detection logic to check `response.success === false`
- [x] **Line 50**: Updated success handling to use `response.data` for payload
- [x] **Line 65**: Updated error handling to use `response.message` for errors
- [x] **Line 66**: Updated to get message from `data.message` first, then fallback to `message`

### 3. ServerResponse Type (COMPLETED)

**File: `src/js/shared/types/form-types.ts`**

- [x] **Line 153**: Updated `ServerResponse` interface to match backend standard:
  ```typescript
  export interface ServerResponse<T = unknown> {
    readonly success: boolean;
    readonly message?: string;
    readonly data?: T;
    readonly errors?: Readonly<Record<string, readonly string[]>>;
  }
  ```

### 4. Form Service (COMPLETED)

**File: `src/js/features/forms/services/form-service.ts`**

- [x] Already using FormApiService which benefits from HttpClient updates
- [x] No direct API calls that need updating

### 5. Form API Service (COMPLETED)

**File: `src/js/features/forms/services/form-api-service.ts`**

- [x] Updated `submitForm` method to use HttpClient instead of direct fetch
- [x] All other methods already using HttpClient
- [x] Now benefits from standardized response format handling

## Implementation Progress

### ✅ Phase 1: Core Response Package (COMPLETED)

1. ✅ Updated all backend handlers to use `response.Success()` and `response.ErrorResponse()`
2. ✅ Removed custom response structs that don't follow the standard
3. ✅ Updated middleware to use standardized responses
4. ✅ Updated error handler to use standardized responses

### ✅ Phase 2: Frontend Response Handling (COMPLETED)

1. ✅ Updated HTTP Client to handle new response format
2. ✅ Updated `ServerResponse` type definition
3. ✅ Updated `ResponseHandler` to expect new format
4. ✅ Updated all form services to handle new response format

### 🔄 Phase 3: Testing and Validation (COMPLETED)

1. ✅ Update all tests to expect new response format
2. ✅ Test all API endpoints to ensure they return correct format
3. ✅ Test frontend error handling with new format
4. ✅ Test signup/login flow to verify it works correctly

**Note**: All API endpoints are returning the correct standardized response format. The original signup issue has been resolved - the frontend now correctly handles the `{ success: true/false, message, data }` format.

### ⏳ Phase 4: Documentation and Cleanup (PENDING)

1. [ ] Update API documentation to reflect new response format
2. [ ] Remove any remaining custom response types
3. [ ] Add response format validation

## Files That Already Follow Standard

These files already use the correct response format and don't need changes:

- ✅ `internal/application/response/response.go` - Defines the standard
- ✅ `internal/application/handlers/web/form_response_builder.go` - Uses standard format
- ✅ `internal/application/services/health.go` - Uses `response.Success()` and `response.ErrorResponse()`

## Testing Checklist

- [x] Test signup endpoint returns `{ success: true, data: { message, redirect } }`
- [x] Test login endpoint returns `{ success: true, data: { message, redirect } }`
- [x] Test form creation returns `{ success: true, data: { form_id, ... } }`
- [x] Test form validation errors return `{ success: false, message: "...", data: { errors: [...] } }`
- [x] Test authentication errors return `{ success: false, message: "..." }`
- [x] Test frontend correctly handles both success and error responses
- [x] Test redirects work correctly with new format
- [x] Test health endpoint returns `{ success: true, data: { status: "ok", time: "..." } }`
- [x] Test validation endpoints return `{ success: true, data: { ... } }`

**Note**: All API endpoints are returning the correct standardized response format. The original signup issue has been resolved - the frontend now correctly handles the `{ success: true/false, message, data }` format.

## Next Steps

1. **Complete remaining backend handlers** (Health Handler, Server)
2. **Update frontend type definitions** (ServerResponse interface)
3. **Update frontend response handlers** (ResponseHandler, Form services)
4. **Test the complete flow** to ensure signup/login work correctly
5. **Update tests** to expect new response format

## Notes

- ✅ **Backend standardization is complete** - all handlers now use `response.Success()` and `response.ErrorResponse()`
- ✅ Frontend standardization is complete - HTTP Client, ResponseHandler, and all services updated
- ✅ HTTP Client has been updated to handle the standardized response format
- ✅ ServerResponse type definition updated to match backend standard
- ✅ ResponseHandler updated to handle both success and error cases correctly
- ✅ Form services updated to use HttpClient consistently
- ✅ **Frontend tests updated to use standardized response format** - All HttpClient and FormApiService tests now pass
- ✅ **All backend handlers updated** - Health handler and server endpoints now use standardized format
- ✅ **Phase 3 Testing Complete** - All API endpoints tested and returning correct standardized format
- ✅ **Original signup issue resolved** - Frontend now correctly handles standardized response format
- The signup issue should now be resolved with the standardized response format
- All responses now include the `success` field to make frontend handling consistent
- Error responses use `response.ErrorResponse()` with appropriate status codes
- Success responses use `response.Success()` with payload in the `data` field
