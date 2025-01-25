# Development Guide

## Getting Started

### Prerequisites

- Docker
- VS Code with Dev Containers extension
- Git

### Development Environment

1. Clone the repository:

   ```bash
   git clone https://github.com/jonesrussell/goforms.git
   cd goforms
   ```

2. Open in VS Code:

   ```bash
   code .
   ```

3. When prompted, click "Reopen in Container" or use Command Palette:

   ```plaintext
   Dev Containers: Reopen in Container
   ```

4. Set up environment:

   ```bash
   cp .env.example .env
   task install
   task migrate:up
   ```

## Development Workflow

### Running the Application

```bash
# Start development server with hot reload
task dev

# Run without hot reload
task run
```

### Testing

```bash
# Run all tests
task test

# Run integration tests
task test:integration

# View test coverage
task test:coverage
```

### Database Operations

```bash
# Create new migration
task migrate:create name=migration_name

# Run migrations
task migrate:up

# Rollback migrations
task migrate:down
```

## Code Organization

```plaintext
.
├── cmd/                  # Application entrypoints
├── internal/            
│   ├── api/             # API endpoints
│   ├── app/             # Application setup
│   ├── core/            # Domain logic
│   ├── platform/        # Infrastructure
│   └── web/             # Web UI
├── migrations/          # Database migrations
├── static/             # Static assets
└── test/               # Test helpers
```

## Development Guidelines

### Code Style

- Follow Go standard project layout
- Use interfaces for dependency inversion
- Keep functions focused and small
- Write idiomatic Go code

### Testing Requirements

- Write unit tests for core logic
- Write integration tests for APIs
- Mock external dependencies
- Use table-driven tests
- Aim for high coverage

### Git Workflow

1. Create feature branch:

   ```bash
   git checkout -b feature/name
   ```

2. Make changes and commit:

   ```bash
   git add .
   git commit -m "Description of changes"
   ```

3. Push and create PR:

   ```bash
   git push origin feature/name
   ```

### Documentation

- Update API docs for endpoint changes
- Document new features
- Keep README.md current
- Add code comments for complex logic

## Tooling

### Task Runner

[Task](https://taskfile.dev) commands are defined in `Taskfile.yml`:

```yaml
tasks:
  install:
    desc: Install dependencies
  
  dev:
    desc: Start development server
  
  test:
    desc: Run tests
  
  migrate:up:
    desc: Run migrations
```

### VS Code Extensions

Recommended extensions:

- Go
- Dev Containers
- GitLens
- Go Test Explorer

### Debugging

1. Set breakpoints in VS Code
2. Use "Run and Debug" panel
3. Select "Go: Launch Package"
4. Start debugging session

## Further Reading

- [API Development](./api.md)
- [Testing Guide](./testing.md)
- [Database Guide](./database.md)
- [Deployment Guide](./deployment.md)

## Form Builder Development

### Overview
The Form Builder system allows users to create, customize, and deploy forms. This guide covers the development setup and implementation details.

### Prerequisites
- Go 1.23 or later
- Node.js 20 or later (for UI development)
- MariaDB 10.11 or later
- Redis (for rate limiting)

### Project Structure

```
.
├── cmd/
│   └── goforms/
│       └── main.go           # Application entry point
├── internal/
│   ├── domain/
│   │   └── form/
│   │       ├── form.go       # Form domain model
│   │       ├── schema.go     # JSON Schema validation
│   │       └── settings.go   # Form settings
│   ├── application/
│   │   └── form/
│   │       ├── service.go    # Form business logic
│   │       └── repository.go # Form data access
│   └── infrastructure/
│       └── form/
│           ├── http/         # HTTP handlers
│           ├── storage/      # Database implementation
│           └── validation/   # Schema validation
├── static/
│   └── js/
│       └── form-builder/     # Form Builder UI
└── web/
    └── templates/
        └── form-builder/     # Form Builder templates
```

### Domain Models

```go
// Form represents a user-created form
type Form struct {
    ID          int64           `json:"id" db:"id"`
    UserID      int64           `json:"user_id" db:"user_id"`
    Name        string         `json:"name" db:"name"`
    Description string         `json:"description" db:"description"`
    Schema      schema.Schema  `json:"schema" db:"schema"`
    UISchema    schema.UI     `json:"ui_schema" db:"ui_schema"`
    Settings    FormSettings  `json:"settings" db:"settings"`
    Version     int           `json:"version" db:"version"`
    Status      string        `json:"status" db:"status"`
    CreatedAt   time.Time     `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}

// FormSubmission represents a form submission
type FormSubmission struct {
    ID        int64          `json:"id" db:"id"`
    FormID    int64          `json:"form_id" db:"form_id"`
    Data      json.RawMessage `json:"data" db:"data"`
    Metadata  json.RawMessage `json:"metadata" db:"metadata"`
    Status    string         `json:"status" db:"status"`
    CreatedAt time.Time      `json:"created_at" db:"created_at"`
}
```

### Database Migrations

```sql
-- Create forms table
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

-- Create form submissions table
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

### Implementation Steps

1. Domain Layer
   - Implement form and submission models
   - Add JSON Schema validation
   - Create form settings types

2. Application Layer
   - Implement form service
   - Add submission handling
   - Create form repository interface

3. Infrastructure Layer
   - Add HTTP handlers for API endpoints
   - Implement database storage
   - Add validation middleware

4. UI Development
   - Create form builder components
   - Add schema editor
   - Implement form preview
   - Add settings panel

### Testing

```go
func TestFormValidation(t *testing.T) {
    form := &Form{
        Name: "Test Form",
        Schema: schema.Schema{
            Type: "object",
            Properties: map[string]schema.Property{
                "name": {
                    Type:  "string",
                    Title: "Name",
                },
            },
            Required: []string{"name"},
        },
    }

    err := form.Validate()
    assert.NoError(t, err)
}

func TestFormSubmission(t *testing.T) {
    submission := &FormSubmission{
        FormID: 1,
        Data: json.RawMessage(`{
            "name": "John Doe"
        }`),
    }

    err := submission.Validate()
    assert.NoError(t, err)
}
```

### JavaScript SDK

```javascript
// Form Builder SDK
const GoForms = {
    render: (elementId, options) => {
        const element = document.getElementById(elementId);
        if (!element) {
            throw new Error(`Element ${elementId} not found`);
        }

        // Fetch form schema
        fetch(`/api/v1/forms/${options.formId}`)
            .then(response => response.json())
            .then(form => {
                // Render form using schema
                const formElement = createForm(form.schema, form.ui_schema);
                element.appendChild(formElement);

                // Add submit handler
                formElement.addEventListener('submit', async (event) => {
                    event.preventDefault();
                    const data = new FormData(formElement);
                    
                    try {
                        const response = await fetch(`/api/v1/forms/${options.formId}/submissions`, {
                            method: 'POST',
                            body: JSON.stringify(Object.fromEntries(data)),
                            headers: {
                                'Content-Type': 'application/json'
                            }
                        });
                        
                        if (response.ok) {
                            showSuccess(form.settings.successMessage);
                        } else {
                            showError('Submission failed');
                        }
                    } catch (error) {
                        showError(error.message);
                    }
                });
            });
    }
};
```

### Security Considerations

1. Input Validation
   - Validate all form submissions against schema
   - Sanitize HTML input
   - Validate file uploads

2. Rate Limiting
   - Implement per-form rate limits
   - Add IP-based rate limiting
   - Monitor submission patterns

3. CORS Security
   - Validate origin headers
   - Implement allowlist
   - Add CSRF protection

4. Data Protection
   - Encrypt sensitive data
   - Implement data retention
   - Add audit logging

### Performance Optimization

1. Caching
   - Cache form schemas
   - Use Redis for rate limiting
   - Implement CDN for SDK

2. Database
   - Index frequently queried fields
   - Optimize JSON columns
   - Implement connection pooling

3. API
   - Implement pagination
   - Add field filtering
   - Use compression

### Monitoring

1. Metrics
   - Form submission rate
   - Error rate
   - Response time

2. Logging
   - Form creation events
   - Submission events
   - Validation errors

3. Alerts
   - High error rates
   - Rate limit breaches
   - System issues
