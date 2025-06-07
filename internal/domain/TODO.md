# Domain Layer Improvements TODO

## High Priority
1. [ ] Consolidate Domain Models
   - Move all form models to `/domain/form/model/`
   - Remove duplicate `Form` struct from root domain
   - Ensure consistent type usage (e.g., `uint` for IDs)

2. [ ] Standardize Context Handling
   - Add context to all service methods
   - Update interface definitions
   - Ensure consistent context propagation

3. [ ] Implement Domain Events
   - Create event interfaces
   - Add event publishing for form operations
   - Implement event handlers

## Medium Priority
4. [ ] Create Validation Layer
   - Extract validation logic from service layer
   - Create dedicated validators
   - Implement value objects for complex validations

5. [ ] Standardize Repository Pattern
   - Rename `Store` to `Repository` for consistency
   - Update interface definitions
   - Ensure proper separation of concerns

6. [ ] Add Domain Services
   - Extract complex business logic
   - Create dedicated domain services
   - Implement proper dependency injection

## Low Priority
7. [ ] Improve Error Handling
   - Create domain-specific errors
   - Implement proper error wrapping
   - Add error context

8. [ ] Add Documentation
   - Document domain models
   - Add interface documentation
   - Create usage examples

9. [ ] Add Tests
   - Unit tests for domain services
   - Integration tests for repositories
   - Event handling tests

## Implementation Order
1. Start with model consolidation (High Priority #1)
2. Follow with context standardization (High Priority #2)
3. Implement domain events (High Priority #3)
4. Move to validation layer (Medium Priority #4)
5. Standardize repository pattern (Medium Priority #5)
6. Add domain services (Medium Priority #6)
7. Complete remaining tasks in order of priority

## Notes
- Each task should be implemented in a separate branch
- Include tests for each change
- Update documentation as changes are made
- Ensure backward compatibility where possible 