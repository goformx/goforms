# GoFormX Forms Route Cleanup TODO

## Overview

This document outlines the cleanup tasks needed to fix the messy and inconsistent forms routes in GoFormX.

## Current Problems

1. **Mixed Responsibilities**: Web UI and API routes mixed in same handlers
2. **Duplicate Routes**: `POST /forms/:id/edit` and `PUT /forms/:id` both update forms
3. **Inconsistent HTTP Methods**: Some routes use POST, others use PUT for same operations
4. **Scattered Validation Routes**: Validation endpoints spread across multiple handlers
5. **Inconsistent Path Constants**: Mixed naming conventions for path constants
6. **Mixed Response Types**: Same routes serving both HTML and JSON

## Cleanup Tasks

### 1. Fix Path Constants (High Priority)

- [x] Standardize path constant naming in `internal/application/constants/constants.go`
- [x] Remove duplicate constants (`PathAPIV1` vs `PathAPIv1`)
- [x] Ensure consistent casing throughout the codebase

### 2. Remove Duplicate Routes (High Priority)

- [x] Remove `POST /forms/:id/edit` route (keep `PUT /forms/:id` for API)
- [x] Update form web handler to use `PUT /forms/:id` for updates
- [x] Update any frontend code that uses the old POST route

### 3. Separate Web UI from API Routes (High Priority)

- [x] Ensure `FormWebHandler` only serves HTML pages
- [x] Ensure `FormAPIHandler` only serves JSON responses
- [x] Remove mixed response types from handlers
- [x] Update route registration to clearly separate web and API routes

### 4. Consolidate Validation Routes (Medium Priority)

- [x] Create dedicated `ValidationHandler` for all validation schemas
- [x] Move validation routes from scattered handlers to centralized location
- [x] Update route registration to use new validation handler

### 5. Standardize Response Types (Medium Priority)

- [x] Web UI handlers should always return HTML (redirect or render)
- [x] API handlers should always return JSON
- [x] Remove conditional response logic based on request type
- [x] Update error handling to be consistent

### 6. Fix Route Organization (Medium Priority)

- [x] Centralize route registration in `module.go`
- [x] Group routes by purpose (web UI, API, validation)
- [x] Add clear comments for each route group
- [x] Ensure consistent middleware application

### 7. Update Handler Responsibilities (Medium Priority)

- [x] Clarify `FormWebHandler` responsibilities (HTML pages only)
- [x] Clarify `FormAPIHandler` responsibilities (JSON API only)
- [x] Create `ValidationHandler` for validation schemas
- [x] Update handler interfaces and documentation

### 8. Add Comprehensive Documentation (Low Priority)

- [ ] Document all routes with their purposes
- [ ] Add OpenAPI documentation for all API routes
- [ ] Create route architecture documentation
- [ ] Add examples for each route type

### 9. Add Route Validation Tests (Low Priority)

- [ ] Test that web routes return HTML
- [ ] Test that API routes return JSON
- [ ] Test route authentication requirements
- [ ] Test route parameter validation

## Implementation Order

1. Fix path constants (foundation)
2. Remove duplicate routes (immediate cleanup)
3. Separate web UI from API routes (core architecture)
4. Consolidate validation routes (organization)
5. Standardize response types (consistency)
6. Fix route organization (structure)
7. Update handler responsibilities (clarity)
8. Add documentation (maintainability)
9. Add tests (reliability)

## Success Criteria

- [x] No duplicate routes exist
- [x] All web routes serve HTML only
- [x] All API routes serve JSON only
- [x] Validation routes are centralized
- [x] Path constants are consistent
- [x] Route organization is clear
- [ ] All tests pass
- [ ] Documentation is complete

## Files to Modify

- `internal/application/constants/constants.go`
- `internal/application/handlers/web/form_web.go`
- `internal/application/handlers/web/form_api.go`
- `internal/application/handlers/web/module.go`
- `internal/application/handlers/web/form_handlers.go`
- `internal/application/handlers/web/form_response_builder.go`
- `internal/application/handlers/web/form_response_helper.go`
- `src/js/features/forms/services/form-api-service.ts` (frontend updates)
- `src/js/pages/form-builder.ts` (frontend updates)

## Notes

- This cleanup will improve code maintainability and reduce confusion
- Some frontend code may need updates to work with the new route structure
- All changes should be backward compatible where possible
- Comprehensive testing is required after implementation
