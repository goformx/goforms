# Project TODO List

## Current Sprint: Email Subscription MVP
- [x] 1. Project Setup
  - [x] Initialize Go module
  - [x] Create basic directory structure
  - [x] Add .gitignore
  - [x] Add .env.example
  - [x] Set up configuration management

- [x] 2. Database Setup
  - [x] Create migrations directory
  - [x] Write subscription table migration
    - [x] Up migration
    - [x] Down migration
  - [x] Set up database connection
  - [x] Implement connection pooling

- [x] 3. Core Application
  - [x] Set up Echo server
  - [x] Configure middleware
    - [x] Logging (zap)
    - [x] Error handling
    - [x] CORS
  - [x] Create app struct and initialization

- [x] 4. Subscription Handler
  - [x] Create subscription model
  - [x] Implement email validation
  - [x] Create POST endpoint for subscriptions
  - [x] Add basic rate limiting
  - [x] Add error responses

- [x] 5. Testing
  - [x] Write handler tests
  - [x] Write validation tests
  - [x] Write integration tests

## Phase 1: Core API Setup
- [x] 1. Project Structure
   - [x] Set up Go project with modules
   - [x] Configure MariaDB database
   - [x] Set up basic HTTP server using Echo
   - [x] Implement middleware for logging, error handling
   - [x] Set up configuration management

- [ ] 2. Database Schema
   - [ ] Forms table
   - [ ] Form fields table
   - [ ] Form submissions table
   - [ ] Basic migrations setup

- [ ] 3. Observability
   - [ ] Structured logging with Zap
   - [ ] Health check endpoints
   - [ ] Request tracing with request IDs
   - [ ] Error tracking and reporting

## Phase 2: Form Management
- [ ] 1. Form Creation API
   - [ ] Endpoint to create new forms
   - [ ] Support for basic templates
   - [ ] Field configuration options

- [ ] 2. Form Management Features
   - [ ] List forms
   - [ ] Update forms
   - [ ] Delete forms
   - [ ] Get form details

## Future Features
- [ ] 1. Multi-tenancy Support
   - [ ] Organization/user management
   - [ ] Role-based access control

- [ ] 2. Advanced Form Features
   - [ ] Conditional logic
   - [ ] Multi-page forms
   - [ ] File upload support
   - [ ] Custom form templates

- [ ] 3. Integrations
   - [ ] Email notifications
   - [ ] Payment processing
   - [ ] Webhooks

- [ ] 4. Real-time Capabilities
   - [ ] Live form previews
   - [ ] Real-time validation

- [ ] 5. Frontend Interface
   - [ ] Admin dashboard
   - [ ] Form builder UI
   - [ ] Form preview
