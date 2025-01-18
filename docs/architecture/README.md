# Architecture Overview

## Clean Architecture

GoForms follows Clean Architecture principles, organizing code into distinct layers:

```ascii
┌──────────────────┐
│     Web Layer    │ 
│   (UI/Templates) │
├──────────────────┤
│    API Layer     │
│   (Controllers)  │
├──────────────────┤
│   Core Layer     │
│ (Business Logic) │
├──────────────────┤
│ Platform Layer   │
│(Infrastructure)  │
└──────────────────┘
```

### Layer Responsibilities

1. **Web Layer** (`internal/web/`)
   - Server-side rendering with templ
   - UI components and layouts
   - Static assets
   - User interface logic

2. **API Layer** (`internal/api/`)
   - REST endpoints
   - Request/response handling
   - Input validation
   - Route definitions

3. **Core Layer** (`internal/core/`)
   - Business logic
   - Domain models
   - Interface definitions
   - Use cases

4. **Platform Layer** (`internal/platform/`)
   - Database implementations
   - External services
   - Infrastructure concerns
   - Technical implementations

## Key Design Principles

### Dependency Rule

Dependencies flow inward:

- Outer layers can depend on inner layers
- Inner layers cannot depend on outer layers
- Core layer has no external dependencies

### Dependency Injection

Using Uber's fx for:

- Constructor injection
- Module organization
- Lifecycle management
- Dependency graph

### Interface Segregation

- Small, focused interfaces
- Clear separation of concerns
- Mockable for testing

## Component Architecture

### Web Components

- Built using templ
- Type-safe templates
- Component-based design
- Reusable layouts

### API Components

- Echo framework
- Versioned endpoints
- Middleware stack
- Standard response formats

### Core Components

- Domain models
- Business rules
- Interface definitions
- Pure business logic

### Platform Components

- Database access
- External services
- Infrastructure code
- Technical implementations

## Data Flow

1. Request enters through Web/API layer
2. Controllers prepare data
3. Use cases execute business logic
4. Core layer processes domain logic
5. Platform layer handles persistence
6. Response flows back through layers

## Further Reading

- [Detailed Layer Design](./layers.md)
- [Component Design](./components.md)
- [Data Flow](./data-flow.md)
- [Dependencies](./dependencies.md)
