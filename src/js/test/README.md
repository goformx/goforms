# Frontend Testing Guide

This directory contains the testing setup and examples for the GoForms frontend application.

## Testing Framework

We use **Vitest** as our testing framework, which provides:
- Fast execution with Vite
- TypeScript support out of the box
- DOM testing with jsdom
- Coverage reporting
- Watch mode for development

## Setup

The testing environment is configured in:
- `vitest.config.ts` - Main configuration
- `setup.ts` - Test environment setup and mocks

## Running Tests

### All Tests
```bash
task test
```

### Frontend Tests Only
```bash
task test:frontend
```

### Backend Tests Only
```bash
task test:backend
```

### Watch Mode
```bash
npm run test:watch
```

### Coverage Report
```bash
npm run test:coverage
```

### UI Mode
```bash
npm run test:ui
```

## Test Structure

### Example Test File
```typescript
import { describe, it, expect, beforeEach, vi } from "vitest";
import { FormBuilderError } from "@/core/errors/form-builder-error";

describe("FormBuilderError", () => {
  it("should create validation errors correctly", () => {
    const error = FormBuilderError.validationError("Field is required");
    expect(error.code).toBe(ErrorCode.VALIDATION_FAILED);
  });
});
```

### Test Organization
- Group related tests with `describe` blocks
- Use descriptive test names with `it` blocks
- Use `beforeEach` for setup
- Use `vi` for mocking

## Available Mocks

The test setup provides mocks for:
- `fetch` - HTTP requests
- `localStorage` - Browser storage
- `sessionStorage` - Session storage
- `FormData` - Form data handling
- `Headers` - HTTP headers
- `Response` - HTTP responses
- `Request` - HTTP requests
- `console` - Console methods

## Testing Patterns

### DOM Testing
```typescript
it("should update DOM elements", () => {
  const element = document.createElement("div");
  document.body.appendChild(element);
  
  // Test DOM manipulation
  dom.showError("Error message", element);
  
  const errorElement = element.querySelector(".gf-error-message");
  expect(errorElement?.textContent).toBe("Error message");
});
```

### API Testing
```typescript
it("should make API calls", async () => {
  const mockFetch = vi.fn().mockResolvedValue({
    ok: true,
    status: 200,
    json: () => Promise.resolve({ data: "test" }),
  });
  
  global.fetch = mockFetch;
  
  const result = await HttpClient.get("/api/test");
  
  expect(mockFetch).toHaveBeenCalledWith("/api/test", expect.any(Object));
  expect(result).toEqual({ data: "test" });
});
```

### Error Testing
```typescript
it("should handle errors correctly", () => {
  const error = FormBuilderError.validationError("Test error");
  
  expect(error.code).toBe(ErrorCode.VALIDATION_FAILED);
  expect(error.userMessage).toBe("Test error");
  expect(error.isCode(ErrorCode.VALIDATION_FAILED)).toBe(true);
});
```

## Coverage

Coverage reports are generated using V8 coverage provider and include:
- Line coverage
- Branch coverage
- Function coverage
- Statement coverage

Coverage excludes:
- Node modules
- Generated files
- Test files
- Configuration files

## Best Practices

1. **Test Structure**
   - Use descriptive test names
   - Group related tests
   - Keep tests focused and simple

2. **Mocking**
   - Mock external dependencies
   - Use `vi.fn()` for function mocks
   - Reset mocks between tests

3. **Assertions**
   - Use specific assertions
   - Test both success and failure cases
   - Verify error messages and codes

4. **DOM Testing**
   - Clean up DOM elements after tests
   - Use realistic DOM structures
   - Test user interactions

5. **Async Testing**
   - Use `async/await` for async tests
   - Mock async operations
   - Test error handling

## Common Issues

### Module Resolution
If you encounter module resolution issues, ensure:
- Path mapping is configured in `vitest.config.ts`
- Import paths use the `@/` prefix
- TypeScript configuration is correct

### DOM Testing
For DOM-related issues:
- Use `jsdom` environment
- Mock DOM APIs as needed
- Clean up DOM elements

### Mocking
For mocking issues:
- Use `vi.fn()` for function mocks
- Reset mocks with `vi.clearAllMocks()`
- Mock global objects in setup

## Continuous Integration

Tests are automatically run in CI with:
- `task test` - All tests
- `task test:cover` - Coverage report
- `task lint` - Code quality checks

## Resources

- [Vitest Documentation](https://vitest.dev/)
- [Testing Library](https://testing-library.com/)
- [jsdom Documentation](https://github.com/jsdom/jsdom) 