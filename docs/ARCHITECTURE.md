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

## Form Builder Architecture

### Overview
The Form Builder system allows users to create, customize, and deploy forms using a JSON Schema-based approach. The system is divided into several components:

### Components

#### Form Schema Layer
- JSON Schema-based form definitions
- UI Schema for rendering configuration
- Form settings and metadata storage
- Version control system for forms

#### Database Structure
```sql
-- Forms table
CREATE TABLE forms (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    schema JSON NOT NULL,
    ui_schema JSON,
    settings JSON,
    version INT NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Form submissions
CREATE TABLE form_submissions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    form_id BIGINT NOT NULL,
    data JSON NOT NULL,
    metadata JSON,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (form_id) REFERENCES forms(id)
);
```

#### Form Builder UI
- Schema Editor: Visual interface for form creation
- Live Preview: Real-time form rendering
- Settings Panel: Form configuration interface
- Deployment Guide: Integration instructions

#### JavaScript SDK
- Form Renderer: Client-side form display
- Validation Engine: JSON Schema validation
- Submission Handler: API integration
- Style System: Theme customization

### Security

- Origin Validation: Control allowed domains
- Rate Limiting: Per-form submission limits
- CAPTCHA: Bot protection
- XSS Protection: Input sanitization
- CORS: Cross-origin security

### Integration System

- Webhooks: Custom HTTP callbacks
- Email Notifications: Automated alerts
- Third-party Services: Slack, etc.
- Custom Actions: Extensible handlers

### Data Flow

1. Form Creation:
   - User creates form via builder UI
   - Schema validated and stored
   - Form settings configured
   - Deployment code generated

2. Form Deployment:
   - SDK loaded on client site
   - Form schema fetched
   - Form rendered with custom styling
   - Client-side validation enabled

3. Form Submission:
   - Data validated against schema
   - Submission processed and stored
   - Notifications triggered
   - Integrations executed

### Standards

The system adheres to the following standards:
- JSON Schema (json-schema.org)
- OpenAPI 3.0
- Web Content Accessibility Guidelines (WCAG)
- General Data Protection Regulation (GDPR)
