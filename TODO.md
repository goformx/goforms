# Project TODO List

## Current Sprint: Email Subscription MVP
- [ ] 1. Project Setup
  - [ ] Initialize Go module
  - [ ] Create basic directory structure
  - [ ] Add .gitignore
  - [ ] Add .env.example
  - [ ] Set up configuration management

- [ ] 2. Database Setup
  - [ ] Create migrations directory
  - [ ] Write subscription table migration
    - [ ] Up migration
    - [ ] Down migration
  - [ ] Set up database connection
  - [ ] Implement connection pooling

- [ ] 3. Core Application
  - [ ] Set up Echo server
  - [ ] Configure middleware
    - [ ] Logging (zap)
    - [ ] Error handling
    - [ ] CORS
  - [ ] Create app struct and initialization

- [ ] 4. Subscription Handler
  - [ ] Create subscription model
  - [ ] Implement email validation
  - [ ] Create POST endpoint for subscriptions
  - [ ] Add basic rate limiting
  - [ ] Add error responses

- [ ] 5. Testing
  - [ ] Write handler tests
  - [ ] Write validation tests
  - [ ] Write integration tests

## Phase 1: Core API Setup
- [ ] 1. Project Structure
   - [ ] Set up Go project with modules
   - [ ] Configure PostgreSQL database
   - [ ] Set up basic HTTP server using Echo
   - [ ] Implement middleware for logging, error handling
   - [ ] Set up configuration management

- [ ] 2. Database Schema
   - [ ] Forms table
   - [ ] Form fields table
   - [ ] Form submissions table
   - [ ] Basic migrations setup

- [ ] 3. Basic Form Submission Features
   - [ ] Create endpoints for form submission
   - [ ] Implement basic field validations:
     - [ ] Text validation
     - [ ] Email validation
     - [ ] Number validation
     - [ ] Date validation
     - [ ] Checkbox validation
     - [ ] Dropdown validation
     - [ ] Multiple choice validation
   - [ ] Store submissions in database

- [ ] 4. API Documentation
   - [ ] Set up Swagger/OpenAPI documentation
   - [ ] Document all endpoints
   - [ ] Create basic usage examples

- [ ] 5. Testing
   - [ ] Unit tests for validation logic
   - [ ] Integration tests for API endpoints
   - [ ] Database interaction tests

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
