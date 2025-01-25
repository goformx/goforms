# API Documentation

## Overview

GoForms provides a RESTful API for managing forms, submissions, and subscriptions. All new endpoints are versioned under `/v1`.

## Authentication

Currently, the API is open for testing. Authentication will be implemented in future releases.

## Base URL

```plaintext
http://localhost:8090/api/v1
```

## Available Endpoints

### Contact API

```http
POST /v1/contacts
GET  /v1/contacts
GET  /v1/contacts/{id}
PUT  /v1/contacts/{id}/status
```

### Subscription API

```http
POST /v1/subscriptions
GET  /v1/subscriptions
GET  /v1/subscriptions/{id}
PUT  /v1/subscriptions/{id}
DELETE /v1/subscriptions/{id}
```

## Contact Form API

### Endpoints

#### Create Contact Submission
- **POST** `/api/v1/contacts`
- Creates a new contact form submission
- Public endpoint, no authentication required
- Request Body:
  ```json
  {
    "name": "string",
    "email": "string",
    "message": "string"
  }
  ```
- Response:
  ```json
  {
    "success": true,
    "data": {
      "id": "number",
      "name": "string",
      "email": "string",
      "message": "string",
      "status": "string",
      "created_at": "string",
      "updated_at": "string"
    }
  }
  ```

#### List Contact Submissions
- **GET** `/api/v1/contacts`
- Lists all contact form submissions
- Public endpoint for demo purposes
- Response:
  ```json
  {
    "success": true,
    "data": [
      {
        "id": "number",
        "name": "string",
        "email": "string",
        "message": "string",
        "status": "string",
        "created_at": "string",
        "updated_at": "string"
      }
    ]
  }
  ```

#### Get Contact Submission
- **GET** `/api/v1/contacts/:id`
- Gets a specific contact form submission
- Protected endpoint, requires authentication
- Response:
  ```json
  {
    "success": true,
    "data": {
      "id": "number",
      "name": "string",
      "email": "string",
      "message": "string",
      "status": "string",
      "created_at": "string",
      "updated_at": "string"
    }
  }
  ```

#### Update Contact Status
- **PUT** `/api/v1/contacts/:id/status`
- Updates the status of a contact form submission
- Protected endpoint, requires authentication
- Request Body:
  ```json
  {
    "status": "string" // "pending", "approved", "rejected"
  }
  ```
- Response:
  ```json
  {
    "success": true,
    "data": {
      "id": "number",
      "status": "string"
    }
  }
  ```

### Error Responses

All endpoints return standardized error responses:
```json
{
  "success": false,
  "error": "Error message"
}
```

Common error codes:
- 400: Bad Request (invalid input)
- 401: Unauthorized (missing/invalid auth)
- 404: Not Found
- 500: Internal Server Error

## Response Format

All responses follow a standard format:

```json
{
  "status": "success|error",
  "data": {},
  "message": "Optional message",
  "errors": []
}
```

## Rate Limiting

The API implements rate limiting based on IP address. Default limits:

- 100 requests per minute for public endpoints
- Configurable via environment variables

## Error Handling

Standard HTTP status codes are used:

- 200: Success
- 201: Created
- 400: Bad Request
- 404: Not Found
- 429: Too Many Requests
- 500: Internal Server Error

## Detailed Documentation

- [Contact API](./contact.md)
- [Subscription API](./subscription.md)
- [Error Codes](./errors.md)
- [Rate Limiting](./rate-limiting.md)

## Form Builder API

### Form Management

#### Create Form
```http
POST /api/v1/forms
```

Request body:
```json
{
  "name": "Contact Form",
  "description": "A simple contact form",
  "schema": {
    "type": "object",
    "properties": {
      "name": {
        "type": "string",
        "title": "Name"
      },
      "email": {
        "type": "string",
        "format": "email",
        "title": "Email"
      },
      "message": {
        "type": "string",
        "title": "Message"
      }
    },
    "required": ["name", "email", "message"]
  },
  "ui_schema": {
    "message": {
      "ui:widget": "textarea"
    }
  },
  "settings": {
    "submitButtonText": "Send Message",
    "successMessage": "Thank you for your message!",
    "allowedOrigins": ["*"],
    "notifications": {
      "email": {
        "to": "admin@example.com"
      }
    }
  }
}
```

Response:
```json
{
  "id": 1,
  "name": "Contact Form",
  "description": "A simple contact form",
  "schema": { ... },
  "ui_schema": { ... },
  "settings": { ... },
  "version": 1,
  "status": "draft",
  "created_at": "2024-01-24T12:00:00Z",
  "updated_at": "2024-01-24T12:00:00Z"
}
```

#### List Forms
```http
GET /api/v1/forms
```

Response:
```json
{
  "forms": [
    {
      "id": 1,
      "name": "Contact Form",
      "description": "A simple contact form",
      "status": "published",
      "version": 1,
      "created_at": "2024-01-24T12:00:00Z",
      "updated_at": "2024-01-24T12:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 10
}
```

#### Get Form
```http
GET /api/v1/forms/{id}
```

Response:
```json
{
  "id": 1,
  "name": "Contact Form",
  "description": "A simple contact form",
  "schema": { ... },
  "ui_schema": { ... },
  "settings": { ... },
  "version": 1,
  "status": "published",
  "created_at": "2024-01-24T12:00:00Z",
  "updated_at": "2024-01-24T12:00:00Z"
}
```

#### Update Form
```http
PUT /api/v1/forms/{id}
```

Request body: Same as Create Form

#### Delete Form
```http
DELETE /api/v1/forms/{id}
```

### Form Submissions

#### Submit Form Response
```http
POST /api/v1/forms/{id}/submissions
```

Request body:
```json
{
  "data": {
    "name": "John Doe",
    "email": "john@example.com",
    "message": "Hello, world!"
  },
  "metadata": {
    "userAgent": "Mozilla/5.0...",
    "ipAddress": "127.0.0.1"
  }
}
```

Response:
```json
{
  "id": 1,
  "form_id": 1,
  "data": { ... },
  "metadata": { ... },
  "status": "success",
  "created_at": "2024-01-24T12:00:00Z"
}
```

#### List Form Submissions
```http
GET /api/v1/forms/{id}/submissions
```

Response:
```json
{
  "submissions": [
    {
      "id": 1,
      "form_id": 1,
      "data": { ... },
      "metadata": { ... },
      "status": "success",
      "created_at": "2024-01-24T12:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 10
}
```

### Form Settings

#### Update Form Settings
```http
PUT /api/v1/forms/{id}/settings
```

Request body:
```json
{
  "submitButtonText": "Send Message",
  "successMessage": "Thank you for your message!",
  "allowedOrigins": ["example.com"],
  "notifications": {
    "email": {
      "to": "admin@example.com"
    },
    "webhook": {
      "url": "https://example.com/webhook",
      "secret": "your-webhook-secret"
    }
  },
  "security": {
    "captcha": true,
    "rateLimit": {
      "enabled": true,
      "max": 100,
      "window": "1h"
    }
  }
}
```

### Form Deployment

#### Get Form Embed Code
```http
GET /api/v1/forms/{id}/embed
```

Response:
```json
{
  "html": "<div id=\"form-1\"></div>\n<script src=\"https://cdn.goforms.io/v1/forms.js\"></script>\n<script>GoForms.render('form-1', { formId: '1' });</script>",
  "javascript": "GoForms.render('form-1', { formId: '1' })",
  "url": "https://cdn.goforms.io/v1/forms.js"
}
```
