# TODO

## High Priority

### Code Quality & Architecture
- [x] **Fix linting issues in form handlers**
  - [x] Fix variable shadowing in form_handlers.go
  - [x] Fix paramTypeCombine issues in form_response_builder.go
  - [x] Fix line length issues in form_error_handler.go
  - [x] Replace magic numbers with constants in form_request_processor.go
  - [x] Replace fmt.Errorf with errors.New for simple errors
  - [x] Fix type assertion error in schema validation
- [x] **Refactor form_web.go to improve separation of concerns**
  - [x] Extract `FormService` to handle business logic and reduce duplication
  - [x] Extract `AuthHelper` to handle authentication and authorization patterns
  - [x] Move business logic from handlers to dedicated service methods
  - [x] Reduce code duplication in authentication and form ownership checks
  - [x] Improve separation of concerns and maintainability
- [x] **Refactor form_api.go to improve route organization and error handling**
  - [x] Separate authenticated and public route registration
  - [x] Implement centralized error handling using `FormErrorHandler`
  - [x] Use `FormRequestProcessor` for input processing
  - [x] Use `FormResponseBuilder` for standardized responses
  - [x] Remove duplicate route registrations
  - [x] Improve dependency injection with proper sanitizer injection
- [x] **Implement comprehensive form validation system**
  - [x] Use go-playground/validator/v10 for static struct validation
  - [x] Use custom logic only for dynamic schemas (user-defined forms)
  - [x] Remove/replace custom comprehensive validator for static cases
  - [x] Document when to use validator/v10 vs custom logic
  - [ ] Implement client-side validation generation for dynamic schemas
  - [ ] Add server-side validation for dynamic form submissions
  - [ ] Create validation error response standardization for dynamic forms
  - [ ] Add form field type validation for dynamic schemas
  - [ ] Implement conditional validation rules for dynamic schemas
  - [ ] Test the new validation endpoint from the frontend (ensure the client fetches /api/v1/forms/:id/validation for dynamic forms)
  - [ ] Implement or enhance client-side code to use the new validation schema for real-time validation
- [ ] **Add comprehensive logging for debugging**
  - [ ] Add structured logging with context
  - [ ] Implement request/response logging middleware
  - [ ] Add performance metrics logging
- [ ] **Review and update to Go 1.24 best practices**
  - [ ] Use `any` type consistently (already implemented)
  - [ ] Implement proper context propagation
  - [ ] Use structured error handling

### Security
- [ ] Implement CSRF protection for all forms
- [ ] Add rate limiting for form submissions
- [ ] Validate and sanitize all user inputs
- [ ] Implement proper session management
- [ ] Add security headers middleware

### Testing
- [ ] Add unit tests for form handlers
- [ ] Add integration tests for form workflows
- [ ] Add test coverage reporting
- [ ] Mock external dependencies in tests

## Medium Priority

### Performance
- [ ] Implement caching for form schemas
- [ ] Optimize database queries
- [ ] Add pagination for form submissions
- [ ] Implement lazy loading for large forms

### User Experience
- [ ] Add form preview functionality
- [ ] Implement form templates
- [ ] Add form sharing capabilities
- [ ] Improve form builder UI/UX

### API Improvements
- [ ] Add API versioning
- [ ] Implement proper API documentation
- [ ] Add API rate limiting
- [ ] Implement API authentication

## Low Priority

### Features
- [ ] Add form analytics
- [ ] Implement form notifications
- [ ] Add form export functionality
- [ ] Implement form collaboration features

### Infrastructure
- [ ] Add monitoring and alerting
- [ ] Implement backup strategies
- [ ] Add deployment automation
- [ ] Implement CI/CD pipeline improvements

## Completed

- [x] Initial project setup
- [x] Basic form CRUD operations
- [x] User authentication system
- [x] Database migrations
- [x] Asset serving system
- [x] Interface segregation for form handlers
- [x] Form error handler implementation
- [x] Fix linting issues in form handlers
- [x] Resolve duplicate method declarations in form handlers
- [x] Fix variable shadowing issues
- [x] Replace magic numbers with constants
- [x] Improve error handling with proper error types
- [x] Fix type assertion issues in schema validation
- [x] Refactor form_web.go to improve separation of concerns
  - [x] Extract FormService for business logic
  - [x] Extract AuthHelper for authentication patterns
  - [x] Reduce code duplication in handlers
  - [x] Improve separation of concerns
- [x] Refactor form_api.go to improve route organization and error handling
  - [x] Separate authenticated and public route registration
  - [x] Implement centralized error handling using FormErrorHandler
  - [x] Use FormRequestProcessor for input processing
  - [x] Use FormResponseBuilder for standardized responses
  - [x] Remove duplicate route registrations
  - [x] Improve dependency injection with proper sanitizer injection

## Code Review Notes

### Recent Fixes Applied:

#### Linting Issues Resolved:
1. **Variable Shadowing**: Fixed `err` variable shadowing in `handleUpdate` and `handleDelete` methods
2. **ParamTypeCombine**: Combined `field string, message string` into `field, message string`
3. **Line Length**: Broke long lines to stay within 120 character limit
4. **Magic Numbers**: Added constants `MaxTitleLength = 255` and `MaxDescriptionLength = 1000`
5. **Error Formatting**: Replaced `fmt.Errorf` with `errors.New` for simple error messages
6. **Type Assertion**: Fixed invalid type assertion on `model.JSON` type

#### Code Quality Improvements:
1. **Constants**: Defined validation constants for better maintainability
2. **Error Handling**: Improved error messages and consistency
3. **Type Safety**: Fixed type assertion issues
4. **Code Organization**: Removed duplicate method declarations

#### Architecture Improvements:
1. **Separation of Concerns**: Extracted business logic into dedicated services
2. **DRY Principle**: Reduced code duplication in authentication and form operations
3. **Single Responsibility**: Each component now has a clear, focused purpose
4. **Dependency Injection**: Improved dependency management and testability

### form_web.go Refactoring Results:

#### Before Refactoring:
- Business logic mixed with HTTP handling
- Repeated authentication patterns
- Duplicated form creation/update logic
- Helper methods scattered throughout handler

#### After Refactoring:
- **FormService**: Handles all form-related business logic
- **AuthHelper**: Centralizes authentication and authorization patterns
- **Clean Handlers**: Focus only on HTTP request/response handling
- **Reduced Duplication**: Common patterns extracted to reusable services
- **Better Testability**: Business logic separated from HTTP concerns

### form_api.go Issues Found:

#### Route Organization Issues:
1. **Duplicate Route Registration**: Both authenticated and public routes register the same schema endpoint
2. **Mixed Route Groups**: Authenticated and public routes are mixed in the same registration method
3. **Inconsistent Middleware Application**: Some routes have middleware, others don't
4. **Poor Separation of Concerns**: Route registration logic is not clearly separated

#### Error Handling Issues:
1. **Inconsistent Error Responses**: Different error handling patterns across endpoints
2. **Direct Error Handling**: Handlers directly handle errors instead of using centralized error handling
3. **Missing Error Context**: Errors lack proper context and categorization

#### Code Duplication:
1. **Repeated Form Retrieval**: Same form retrieval logic in multiple handlers
2. **Repeated Error Logging**: Similar error logging patterns across handlers
3. **Repeated Response Formatting**: Similar response formatting logic

### Recommended Refactoring:

#### For form_web.go:
1. Extract `FormRequestProcessor` for input handling
2. Extract `FormResponseBuilder` for response formatting
3. Create typed request structs (`FormCreateRequest`, `FormUpdateRequest`)
4. Move business logic to dedicated service methods
5. Implement helper methods to reduce duplication
6. Use standardized error handling patterns

#### For form_api.go:
1. **Route Organization**:
   ```go
   func (h *FormAPIHandler) RegisterRoutes(e *echo.Echo) {
       api := e.Group(constants.PathAPIv1)
       formsAPI := api.Group(constants.PathForms)
       
       // Register authenticated routes
       h.RegisterAuthenticatedRoutes(formsAPI)
       
       // Register public routes
       h.RegisterPublicRoutes(formsAPI)
   }
   ```

2. **Error Handling**:
   ```go
   func (h *FormAPIHandler) handleFormSchema(c echo.Context) error {
       form, err := h.GetFormByID(c)
       if err != nil {
           return h.ErrorHandler.HandleFormAccessError(c, err)
       }
       return h.ResponseBuilder.BuildSchemaResponse(c, form.Schema)
   }
   ```

3. **Request Processing**:
   ```go
   func (h *FormAPIHandler) handleFormSchemaUpdate(c echo.Context) error {
       schema, err := h.RequestProcessor.ProcessSchemaUpdateRequest(c)
       if err != nil {
           return h.ErrorHandler.HandleSchemaError(c, err)
       }
       // ... rest of logic
   }
   ```

## Next Priority Task

**🎯 RECOMMENDED NEXT TASK: Implement comprehensive form validation system**

**Why this should be next:**
1. **High Impact**: Will significantly improve form reliability and user experience
2. **Foundation**: Better validation will prevent data integrity issues
3. **User Experience**: Proper validation feedback improves form usability
4. **Security**: Server-side validation prevents malicious submissions
5. **Consistency**: Standardized validation across all form types

**Specific steps:**
1. Create form schema validation rules and constraints
2. Implement client-side validation generation from schema
3. Add server-side validation for form submissions
4. Create validation error response standardization
5. Add form field type validation (text, email, number, etc.)
6. Implement conditional validation rules (required_if, etc.)
7. Add validation testing and error handling

**Estimated time:** 3-4 hours

**Benefits:**
- Improved data quality and form reliability
- Better user experience with clear validation feedback
- Enhanced security through server-side validation
- Consistent validation behavior across the application

## Notes

- Review code regularly for DRY, SRP, and SoC violations
- Keep dependencies updated
- Monitor performance metrics
- Regular security audits
- All linting issues have been resolved - maintain this standard 