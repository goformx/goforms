# Code Cleanup and Best Practices TODO

## To Discuss
- Refactor in-memory event publisher: The new event system's Subscriber interface only supports Subscribe, not Handle. The publisher should maintain a map of event names to handler functions and invoke them on Publish. Confirm design before refactor.

## Completed Tasks

- ✅ Removed old user service definition (`internal/domain/user_service.go`).
- ✅ Consolidated validator interfaces: now only in `internal/domain/common/interfaces/validator.go` and all references updated.
- ✅ Unified event system: core event interfaces are now in `internal/domain/common/events/`, infrastructure and form events updated, and deleted `internal/domain/common/events/events.go` and `internal/infrastructure/event/publisher.go`.
- ✅ Repository consolidation: moved all repositories from `internal/infrastructure/persistence/store/*` to `internal/infrastructure/repository/*` and updated all references.

## Redundant Code

### Service Definitions
- [x] Consolidate duplicate user service definitions:
  - `internal/domain/user_service.go` (old)
  - `internal/domain/user/service.go` (new)
  - Action: Remove `user_service.go` after verifying all functionality is in `service.go` (**Done**)

### Validator Interfaces
- [x] Consolidate validator interfaces:
  - `internal/infrastructure/validation/validator.go`
  - `internal/domain/common/interfaces/validator.go`
  - Action: Create a single validator interface in `internal/domain/common/interfaces/validator.go` and update all references (**Done**)

### Event System
- [x] Consolidate event-related interfaces:
  - `internal/infrastructure/event/publisher.go`
  - `internal/domain/common/events/events.go`
  - `internal/domain/form/event/event.go`
  - Action: Create a unified event system in `internal/domain/common/events/` (**Done**)

### Repository Pattern
- [x] Review repository implementations:
  - Some repositories are in `internal/infrastructure/persistence`
  - Some are in `internal/infrastructure/repository`
  - Action: Consolidate all repositories in `internal/infrastructure/repository` (**Done**)

## Code Organization

### Middleware
- [x] Review middleware organization:
  - Some middleware is in `internal/application/middleware`
  - Some is in `internal/infrastructure/web/middleware`
  - Action: Consolidate all middleware in `internal/application/middleware`
  - Note: All middleware is already properly organized in `internal/application/middleware/` with no duplicates found in infrastructure layer.

## Best Practices Violations

### Error Handling
- [x] Review error handling patterns:
  - Some errors are returned directly
  - Some are wrapped with context
  - Action: Standardize error handling with proper context and wrapping
  - Note: Variable shadowing issues in form/service.go have been fixed.

### Logging
- [x] Review logging patterns:
  - Some logs include sensitive information
  - Some logs lack proper context
  - Action: Standardize logging with proper context and security
  - Note: Explicit masking for sensitive fields is now restored in the logger. The go-sanitize library is used only for input cleaning, not for masking sensitive data in logs.

### Configuration
- [x] Review configuration management:
  - Some config is hardcoded
  - Some config is in environment variables
  - Action: Standardize configuration management
  - Tasks:
    1. Remove sensitive defaults (DB credentials, CSRF secrets, API keys) ✅
    2. Consolidate duplicate config (AppConfig and ServerConfig) ✅
    3. Add configuration documentation
    4. Add environment variable validation
    5. Review and update security-related defaults

### Dependency Injection
- [x] Review dependency injection:
  - Some components use constructor injection
  - Some use fx dependency injection
  - Action: Standardize on fx dependency injection (keep it simple)
  - Tasks:
    1. ✅ Use fx.Provide for all injectable components
    2. ✅ Use fx.In for grouping dependencies only when it improves clarity
    3. ✅ Use fx.Out only when a constructor must provide multiple values
    4. ✅ Use fx.Annotate and fx.As only when interface casting or grouping is needed
    5. ✅ Add error handling for fx.Provide functions
    6. ✅ Add OnStart/OnStop hooks only for components that need resource management
    7. ✅ Group related providers in modules, but avoid unnecessary modules
    8. ✅ Use clear, descriptive names for modules and providers
    9. ✅ Keep interface decoupling simple: use fx.As only when needed
  - New Strategy:
    1. [x] Centralize core providers in infrastructure.Module
    2. [x] Keep app-specific wiring in main.go
    3. [x] Use fx.In/fx.Out for complex constructors
    4. [x] Document DI graph with comments
    5. [x] Test and iterate on wiring

### Input Validation
- [ ] Use go-sanitize for all user input:
  - Ensure all user-provided data is sanitized using go-sanitize before processing or storing.
  - Action: Audit all input points and apply go-sanitize as appropriate.

## Testing

### Test Coverage
- [ ] Review test coverage:
  - Some packages lack tests
  - Some tests are incomplete
  - Action: Add missing tests and improve existing ones

### Test Organization
- [ ] Review test organization:
  - Some tests are in `_test.go` files
  - Some are in separate test packages
  - Action: Standardize test organization

## Documentation

### API Documentation
- [ ] Review API documentation:
  - Some endpoints lack documentation
  - Some documentation is outdated
  - Action: Update and complete API documentation

### Code Comments
- [ ] Review code comments:
  - Some code lacks comments
  - Some comments are outdated
  - Action: Update and add missing comments

## Security

### Authentication
- [ ] Review authentication:
  - Some endpoints lack proper authentication
  - Some authentication is inconsistent
  - Action: Standardize authentication across all endpoints

### Authorization
- [ ] Review authorization:
  - Some endpoints lack proper authorization
  - Some authorization is inconsistent
  - Action: Standardize authorization across all endpoints

## Performance

### Database Operations
- [ ] Review database operations:
  - Some queries lack proper indexing
  - Some operations are inefficient
  - Action: Optimize database operations

### Caching
- [ ] Review caching:
  - Some operations lack caching
  - Some caching is inconsistent
  - Action: Implement proper caching strategy

## Frontend

### JavaScript Organization
- [ ] Review JavaScript organization:
  - Some code is in global scope
  - Some modules lack proper exports
  - Action: Standardize JavaScript module organization

### CSS Organization
- [ ] Review CSS organization:
  - Some styles are duplicated
  - Some styles lack proper scoping
  - Action: Standardize CSS organization

## Infrastructure

### Docker Configuration
- [ ] Review Docker configuration:
  - Some configurations are hardcoded
  - Some lack proper security settings
  - Action: Standardize Docker configuration

### CI/CD
- [ ] Review CI/CD pipeline:
  - Some steps are missing
  - Some configurations are outdated
  - Action: Update and complete CI/CD pipeline

## Critical Issues

### Security Vulnerabilities
- [ ] Fix hardcoded database passwords in `internal/infrastructure/config/config.go`
- [ ] Implement proper token generation and validation in `internal/domain/user/service.go`
- [ ] Add proper token blacklisting in `internal/domain/user/service.go`
- [ ] Implement proper JWT validation and parsing in `internal/domain/user/service.go`

### Incomplete Features
- [ ] Complete form details parsing in `internal/application/handlers/web/form.go`
- [ ] Add form cards in `internal/presentation/templates/pages/forms.templ`
- [ ] Implement proper token blacklist check in `internal/domain/user/service.go`

### Code Quality
- [ ] Remove redundant error checks in `internal/application/handlers/web/form.go`
- [ ] Consolidate duplicate error handling patterns
- [ ] Standardize error wrapping and context
- [ ] Remove hardcoded values and move to configuration

### Performance Issues
- [ ] Review and optimize database queries
- [ ] Implement proper caching for frequently accessed data
- [ ] Add database indexes for common queries
- [ ] Optimize form submission handling

### Security Best Practices
- [ ] Implement proper CSRF protection
- [ ] Add rate limiting for authentication endpoints
- [ ] Implement proper session management
- [ ] Add input validation for all endpoints
- [ ] Implement proper password hashing and validation
- [ ] Add proper error messages for security-related failures

### Code Organization
- [ ] Move all middleware to `internal/application/middleware`
- [ ] Consolidate repository implementations
- [ ] Standardize service interfaces
- [ ] Implement proper dependency injection
- [ ] Add proper error types and handling

### Testing
- [ ] Add unit tests for all services
- [ ] Add integration tests for all endpoints
- [ ] Add proper test coverage reporting
- [ ] Implement proper test fixtures
- [ ] Add proper test documentation

### Documentation
- [ ] Add proper API documentation
- [ ] Add proper code comments
- [ ] Add proper README files
- [ ] Add proper setup instructions
- [ ] Add proper contribution guidelines

# GoForms TODO List

## Dependency Injection (DI) Implementation

### Completed
- [x] Centralize core providers in infrastructure.Module
- [x] Keep app-specific wiring in main.go
- [x] Document DI graph with comments
- [x] Move event publisher to infrastructure layer
- [x] Fix import conflicts with proper aliases
- [x] Add error handling for provider initialization
- [x] Implement proper store initialization with error handling
- [x] Add structured logging for infrastructure components
- [x] Update application layer DI with proper error handling
- [x] Implement proper handler dependency management
- [x] Add validation for handler dependencies

### In Progress
- [ ] Add tests for infrastructure providers
- [ ] Add tests for application providers
- [ ] Document DI patterns in architecture guide

### Next Steps
1. Testing:
   - [ ] Unit tests for providers
   - [ ] Integration tests for DI graph
   - [ ] Error handling tests

2. Documentation:
   - [ ] Update architecture guide with DI patterns
   - [ ] Add examples for common DI scenarios
   - [ ] Document error handling strategies

## Infrastructure Module

### Completed
- [x] Implement proper store initialization
- [x] Add error handling for database operations
- [x] Implement structured logging
- [x] Fix import conflicts
- [x] Move event publisher to infrastructure layer
- [x] Add proper error handling for providers

### In Progress
- [ ] Add tests for infrastructure components
- [ ] Document infrastructure patterns

### Next Steps
1. Testing:
   - [ ] Add unit tests for stores
   - [ ] Add integration tests for database operations
   - [ ] Test error handling scenarios

2. Documentation:
   - [ ] Document infrastructure patterns
   - [ ] Add examples for common scenarios
   - [ ] Update architecture guide

## Domain Layer

### Completed
- [x] Keep domain layer free of infrastructure details
- [x] Define clear interfaces for repositories
- [x] Implement proper error handling
- [x] Add structured logging

### In Progress
- [ ] Add tests for domain services
- [ ] Document domain patterns

### Next Steps
1. Testing:
   - [ ] Add unit tests for services
   - [ ] Add integration tests for repositories
   - [ ] Test error handling scenarios

2. Documentation:
   - [ ] Document domain patterns
   - [ ] Add examples for common scenarios
   - [ ] Update architecture guide

## Application Layer

### Completed
- [x] Implement proper error handling
- [x] Add structured logging
- [x] Keep handlers focused and single-purpose
- [x] Implement proper DI for handlers
- [x] Add validation for handler dependencies
- [x] Centralize handler dependency management

### In Progress
- [ ] Add tests for handlers
- [ ] Document handler patterns

### Next Steps
1. Testing:
   - [ ] Add unit tests for handlers
   - [ ] Add integration tests for handlers
   - [ ] Test error handling scenarios

2. Documentation:
   - [ ] Document handler patterns
   - [ ] Add examples for common scenarios
   - [ ] Update architecture guide

## Presentation Layer

### Completed
- [x] Implement proper error handling
- [x] Add structured logging
- [x] Keep views focused and single-purpose

### In Progress
- [ ] Add tests for views
- [ ] Document view patterns

### Next Steps
1. Testing:
   - [ ] Add unit tests for views
   - [ ] Add integration tests for views
   - [ ] Test error handling scenarios

2. Documentation:
   - [ ] Document view patterns
   - [ ] Add examples for common scenarios
   - [ ] Update architecture guide

## General

### Completed
- [x] Implement proper error handling
- [x] Add structured logging
- [x] Keep components focused and single-purpose

### In Progress
- [ ] Add tests for all components
- [ ] Document patterns and best practices

### Next Steps
1. Testing:
   - [ ] Add unit tests for all components
   - [ ] Add integration tests for all components
   - [ ] Test error handling scenarios

2. Documentation:
   - [ ] Document patterns and best practices
   - [ ] Add examples for common scenarios
   - [ ] Update architecture guide 