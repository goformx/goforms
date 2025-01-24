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
