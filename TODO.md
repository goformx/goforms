# Project TODO List

## Current Sprint: Email Subscription MVP ✅
- [x] 1. Project Setup
  - [x] Initialize Go module
  - [x] Create basic directory structure
  - [x] Add .gitignore
  - [x] Add .env.example
  - [x] Set up configuration management

- [x] 2. Database Setup
  - [x] Create migrations directory
  - [x] Write subscription table migration
  - [x] Set up database connection
  - [x] Implement connection pooling

- [x] 3. Core Application
  - [x] Set up Echo server
  - [x] Configure middleware
  - [x] Create app struct and initialization
  - [x] Implement health check endpoint
  - [x] Add database health monitoring

- [x] 4. Subscription Handler
  - [x] Create subscription model
  - [x] Implement email validation
  - [x] Create POST endpoint
  - [x] Add rate limiting
  - [x] Add error responses

- [x] 5. Testing
  - [x] Write handler tests
  - [x] Write validation tests
  - [x] Write integration tests
  - [x] Add health check tests

## Current Sprint: Form Management API
- [ ] 1. Database Schema
  - [ ] Create forms table migration
    - [ ] ID, title, description, created_at, updated_at
    - [ ] Status (draft, published, archived)
    - [ ] Validation rules
  - [ ] Create form_fields table migration
    - [ ] Field types (text, number, email, etc.)
    - [ ] Validation rules
    - [ ] Required/optional status
  - [ ] Create form_submissions table migration
    - [ ] Submission data (JSON)
    - [ ] Metadata (IP, timestamp, etc.)
  - [ ] Add form_access_control table migration
    - [ ] Owner ID, form ID, permissions
    - [ ] API key associations
    - [ ] Rate limit configurations

- [ ] 2. Core Form API
  - [ ] Form model and validation
  - [ ] CRUD endpoints for forms
  - [ ] Field configuration
  - [ ] Form submission handling
  - [ ] Input sanitization
  - [ ] Implement API versioning (v1)
  - [ ] Add OpenAPI/Swagger annotations
  - [ ] Add request/response validation middleware
  - [ ] Implement bulk operations

- [ ] 3. Testing & Documentation
  - [ ] Unit tests for form models
  - [ ] Integration tests for form API
  - [ ] API documentation with OpenAPI/Swagger
  - [ ] Update README with new endpoints

## Next Up: Enhanced Features
- [ ] 1. Security & Performance
  - [ ] Rate limiting per form
  - [ ] CAPTCHA integration
  - [ ] XSS protection
  - [ ] SQL injection prevention
  - [ ] Response caching

- [ ] 2. Analytics & Monitoring
  - [ ] Submission analytics
  - [ ] Performance metrics
  - [x] Health check endpoint
  - [x] Database connectivity monitoring
  - [ ] Error tracking
  - [ ] Usage statistics
  - [ ] Prometheus metrics integration
  - [ ] Uptime monitoring setup
  - [ ] Alert configuration for degraded status
  - [ ] Response time tracking
  - [ ] Resource usage monitoring (CPU, Memory)
  - [ ] Request rate monitoring
  - [ ] Component initialization metrics
    - [ ] FX dependency startup timing
    - [ ] Database connection timing
    - [ ] Server startup timing
    - [ ] Middleware initialization timing
  - [ ] Startup performance monitoring
    - [ ] Total startup time tracking
    - [ ] Component dependency graph timing
    - [ ] Startup sequence optimization
    - [ ] Cold start vs warm start metrics

## Future Phases
- [ ] 1. Advanced Form Features
  - [ ] Conditional logic
  - [ ] Multi-page forms
  - [ ] File uploads
  - [ ] Custom validation rules

- [ ] 2. Integration Features
  - [ ] Email notifications
  - [ ] Webhook support
  - [ ] Export capabilities (CSV, JSON)
  - [ ] API key management
  - [ ] Multi-tenant CORS management
    - [ ] Database-driven CORS configuration
    - [ ] Per-tenant origin management
    - [ ] Dynamic CORS updates without nginx changes
    - [ ] CORS audit logging

- [ ] 3. Administration
  - [ ] User management
  - [ ] Role-based access
  - [ ] Audit logging
  - [ ] Backup/restore functionality
  - [ ] Health check dashboard
  - [ ] System status page

## Current Sprint: Marketing Website API
- [ ] 1. Database Schema
  - [ ] Create marketing_pages table migration
    - [ ] ID, title, content, meta_description
    - [ ] Status (draft, published, archived)
    - [ ] SEO fields (meta_title, meta_description, og_image)
    - [ ] Created_at, updated_at, published_at
  - [ ] Create marketing_stats table migration
    - [ ] Page views, unique visitors
    - [ ] Conversion tracking
    - [ ] UTM parameter tracking

- [ ] 2. Core Marketing API
  - [ ] Create marketing module using fx.Module
  - [ ] Implement page model and validation
  - [ ] Add CRUD endpoints with OpenAPI annotations
  - [ ] Add rate limiting and caching
  - [ ] Implement stats collection middleware
  - [ ] Add input sanitization for content

- [ ] 3. Testing & Documentation
  - [ ] Unit tests for marketing models
  - [ ] Integration tests for marketing API
  - [ ] Update API documentation
  - [ ] Add performance benchmarks
