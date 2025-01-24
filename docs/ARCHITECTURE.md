# Architecture Documentation

## Overview

GoForms is built using a clean architecture approach with clear separation of concerns. The application is divided into several layers:

- Domain Layer: Core business logic and entities
- Application Layer: Use cases and service orchestration
- Infrastructure Layer: External dependencies and implementations
- Presentation Layer: Web UI and API endpoints

## Components

### Domain Layer

#### Contact Module
- `contact.Submission`: Entity representing a contact form submission
- `contact.Service`: Interface defining contact form business operations
- `contact.Store`: Interface for contact data persistence
- `contact.Status`: Enumeration of possible submission statuses

#### Subscription Module
- `subscription.Subscription`: Entity for email subscriptions
- `subscription.Service`: Interface for subscription management
- `subscription.Store`: Interface for subscription data persistence
- `subscription.Status`: Enumeration of subscription statuses

### Application Layer

#### HTTP Handlers
- `v1.ContactAPI`: Handles contact form HTTP endpoints
  - Public endpoints for form submission and message listing
  - Protected endpoints for admin operations
- `v1.SubscriptionAPI`: Manages subscription-related endpoints

#### Services
- `services.ContactService`: Implements contact form business logic
- `services.SubscriptionService`: Implements subscription management

### Infrastructure Layer

#### Database
- MariaDB for persistent storage
- Connection pooling and configuration
- Migration management

#### Logging
- Structured logging using Uber's zap
- Request tracking with unique IDs
- Error and debug logging

#### Configuration
- Environment-based configuration
- Validation and defaults
- Type-safe config access

### Presentation Layer

#### Web UI
- Contact Form Demo
  - Form submission component
  - Message history display
  - API response visualization
- Subscription Management
  - Subscription forms
  - Status management

#### Templates
- Using templ for type-safe templates
- Component-based UI architecture
- Shared layouts and components

## Data Flow

1. User submits contact form
2. Frontend JavaScript handles submission
3. Request goes to public API endpoint
4. Service layer validates input
5. Data is persisted to database
6. Response is returned to user
7. UI updates to show submission status

## Security

- Authentication required for admin operations
- Public endpoints for demo functionality
- CORS and security headers
- Input validation at multiple levels
- Error handling and logging

## Dependencies

- Echo: Web framework
- Fx: Dependency injection
- Zap: Logging
- SQLx: Database access
- Templ: Templates
- Testify: Testing

## Development

- Dev container for consistent environment
- Hot reload for rapid development
- Task automation with Taskfile
- Comprehensive test suite
- Linting and formatting rules
