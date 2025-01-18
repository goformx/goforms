# API Documentation

## Overview

GoForms provides a RESTful API for managing forms, submissions, and subscriptions. All new endpoints are versioned under `/v1`.

## Authentication

Currently, the API is open for testing. Authentication will be implemented in future releases.

## Base URL

```
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