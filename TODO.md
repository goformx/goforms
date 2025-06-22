# TODO

## High Priority

### Code Quality & Architecture
- [ ] **Refactor form_web.go to improve separation of concerns**
  - [ ] Extract `FormRequestProcessor` to handle input validation and sanitization
  - [ ] Extract `FormResponseBuilder` to standardize response formatting
  - [ ] Create `FormCreateRequest` and `FormUpdateRequest` structs for type safety
  - [ ] Move business logic from handlers to dedicated service methods
  - [ ] Reduce code duplication in authentication and form ownership checks
- [ ] **Implement proper error handling patterns across handlers**
  - [ ] Use Go 1.24's `errors.Join` for error composition
  - [ ] Standardize error response format using `response.APIResponse`
  - [ ] Create dedicated error handling service
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

## Code Review Notes

### form_web.go Issues Found:

#### DRY Violations:
1. **Repeated Authentication Pattern**: Every handler starts with `RequireAuthenticatedUser(c)` check
2. **Repeated Form Ownership Check**: Multiple handlers use `GetFormWithOwnership(c)` pattern
3. **Repeated CORS Parsing Logic**: Same CORS parsing appears in `handleCreate` and `handleUpdate`
4. **Repeated Success Response Pattern**: Same JSON response structure in multiple handlers

#### SRP Violations:
1. **Mixed Responsibilities**: Handlers handle authentication, validation, business logic, and response formatting
2. **Handler Methods Doing Too Much**: Each method handles multiple concerns

#### SoC Violations:
1. **Business Logic in Handlers**: Form creation/update logic embedded in handlers
2. **Response Formatting in Handlers**: Direct JSON formatting instead of using response service
3. **Input Processing in Handlers**: Sanitization mixed with request handling

#### Go 1.24 Issues:
1. **Error Handling**: Not using `errors.Join` for error composition
2. **Context Usage**: Some methods don't properly propagate context
3. **Type Safety**: Could use more specific types instead of `any` in some places

### Recommended Refactoring:
1. Extract `FormRequestProcessor` for input handling
2. Extract `FormResponseBuilder` for response formatting
3. Create typed request structs (`FormCreateRequest`, `FormUpdateRequest`)
4. Move business logic to dedicated service methods
5. Implement helper methods to reduce duplication
6. Use standardized error handling patterns

## Notes

- Review code regularly for DRY, SRP, and SoC violations
- Keep dependencies updated
- Monitor performance metrics
- Regular security audits 