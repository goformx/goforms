# GORM Migration Plan

## Overview
This document outlines the plan to migrate from sqlx to GORM for database operations in the GoFormX application.

## Phase 1: Initial Setup and Dependencies
- [x] Add GORM dependencies
- [x] Update database configuration
- [x] Create GORM database connection
- [x] Set up GORM logging

## Phase 2: Model Migration

### 2.1 User Model
- [x] Add GORM tags to User struct
- [x] Implement GORM hooks for User
- [x] Add validation tags
- [x] Update User repository implementation

### 2.2 Form Model
- [x] Add GORM tags to Form struct
- [x] Implement GORM hooks for Form
- [x] Add validation tags
- [x] Update Form repository implementation

### 2.3 Form Submission Model
- [x] Add GORM tags to FormSubmission struct
- [x] Implement GORM hooks for FormSubmission
- [x] Add validation tags
- [x] Update FormSubmission repository implementation

## Phase 3: Schema and Data Migration
- [x] Create GORM migrations
  - [x] Create initial schema migration using go-migrate
  - [x] Set up migration tracking with go-migrate
  - [x] Add up/down migrations for schema changes
- [ ] Migrate existing data
- [x] Update database indexes
- [x] Add foreign key constraints

## Phase 4: Testing and Validation

### 4.1 Unit Tests
- [ ] Update repository tests for GORM
- [ ] Add GORM-specific test cases
- [ ] Test transaction handling
- [ ] Test error handling

### 4.2 Integration Tests
- [ ] Update integration tests
- [ ] Test database operations
- [ ] Test concurrent operations
- [ ] Test performance

## Phase 5: Deployment

### 5.1 Staging
- [ ] Deploy to staging environment
- [ ] Monitor performance
- [ ] Test all features
- [ ] Gather metrics

### 5.2 Production
- [ ] Create deployment plan
- [ ] Schedule maintenance window
- [ ] Deploy changes

## Notes
- Keep existing error handling patterns
- Maintain domain-driven design principles
- Preserve existing transaction boundaries
- Keep existing logging patterns
- Maintain backward compatibility where possible

## Recent Updates
- [x] Set up database migrations with go-migrate
  - Created initial schema migration with tables for users, forms, and form submissions
  - Added up/down migrations for schema changes
  - Added indexes and foreign key constraints
  - Added triggers for updated_at timestamps
  - Removed custom migration runner in favor of go-migrate
- [x] Migrated Form model to use GORM
  - Added GORM tags for all fields
  - Implemented BeforeCreate and BeforeUpdate hooks
  - Added soft delete support
  - Updated schema validation to match JSON Schema format
  - Added explicit table name
- [x] Migrated User model to use GORM
  - Added GORM tags for all fields
  - Implemented BeforeCreate and BeforeUpdate hooks
  - Added soft delete support
  - Added validation tags
  - Added explicit table name
- [x] Migrated FormSubmission model to use GORM
  - Added GORM tags for all fields
  - Implemented BeforeCreate and BeforeUpdate hooks
  - Added soft delete support
  - Added validation tags
  - Added explicit table name
- [x] Removed sqlx dependency
- [x] Updated database connection code to use GORM exclusively
- [x] Updated infrastructure module guide to reflect GORM usage
- [x] Updated dependency injection guide to reflect GORM usage 