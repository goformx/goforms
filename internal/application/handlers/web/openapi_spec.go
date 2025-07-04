package web

// OpenAPISpec contains the OpenAPI 3.1.0 specification for the GoFormX API
const OpenAPISpec = `openapi: 3.1.0
info:
  title: GoFormX API
  description: |
    A modern form management system with RESTful API.

    ## Features
    - Form creation and management
    - Form submissions and validation
    - User authentication and authorization
    - Real-time form validation
    - CSRF protection

    ## Authentication
    This API supports multiple authentication methods:
    - Session-based authentication for web interfaces
    - CSRF tokens for state-changing operations

    ## Rate Limiting
    API requests are rate-limited to ensure fair usage:
    - 20 requests per minute for authenticated users
    - 5 requests per minute for unauthenticated users

  version: 1.0.0
  contact:
    name: GoFormX Support
    url: https://github.com/goformx/goforms
    email: support@goformx.com
  license:
    name: MIT
    url: https://opensource.org/licenses/MIT

servers:
  - url: http://localhost:8090/api/v1
    description: Development server
  - url: https://api.goformx.com/api/v1
    description: Production server

security:
  - csrf: []
  - session: []

paths:
  /health:
    get:
      summary: Health check
      description: Check if the API is healthy and responsive
      tags:
        - System
      responses:
        '200':
          description: Service is healthy
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    example: "healthy"
                  timestamp:
                    type: string
                    format: date-time
                    example: "2024-01-01T00:00:00Z"
                  version:
                    type: string
                    example: "1.0.0"
        '503':
          description: Service is unhealthy
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /forms:
    get:
      summary: List forms
      description: Retrieve a list of forms for the authenticated user
      tags:
        - Forms
      security:
        - session: []
      parameters:
        - name: page
          in: query
          description: Page number for pagination
          required: false
          schema:
            type: integer
            minimum: 1
            default: 1
        - name: limit
          in: query
          description: Number of items per page
          required: false
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20
        - name: status
          in: query
          description: Filter forms by status
          required: false
          schema:
            type: string
            enum: [draft, published, archived]
      responses:
        '200':
          description: List of forms retrieved successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/Form'
                  meta:
                    $ref: '#/components/schemas/PaginationMeta'
        '401':
          description: Unauthorized - User not authenticated
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /forms/{id}:
    get:
      summary: Get form by ID
      description: Retrieve a specific form by its ID
      tags:
        - Forms
      parameters:
        - name: id
          in: path
          required: true
          description: Form ID
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Form retrieved successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/Form'
        '404':
          description: Form not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /forms/{id}/schema:
    get:
      summary: Get form schema
      description: Retrieve the JSON schema for a form
      tags:
        - Forms
      parameters:
        - name: id
          in: path
          required: true
          description: Form ID
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Form schema retrieved successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: object
                    description: JSON schema for the form
                    additionalProperties: true
        '404':
          description: Form not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
    put:
      summary: Update form schema
      description: Update the JSON schema for a form (requires ownership)
      tags:
        - Forms
      security:
        - session: []
        - csrf: []
      parameters:
        - name: id
          in: path
          required: true
          description: Form ID
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                schema:
                  type: object
                  description: New JSON schema for the form
                  additionalProperties: true
              required:
                - schema
      responses:
        '200':
          description: Form schema updated successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/Form'
                  message:
                    type: string
                    example: "Form schema updated successfully"
        '400':
          description: Invalid schema format
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '401':
          description: Unauthorized - User not authenticated
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '403':
          description: Forbidden - User doesn't own the form
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '404':
          description: Form not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /forms/{id}/validation:
    get:
      summary: Get form validation schema
      description: Retrieve client-side validation rules for a form
      tags:
        - Forms
      parameters:
        - name: id
          in: path
          required: true
          description: Form ID
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Validation schema retrieved successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: object
                    description: Client-side validation rules
                    additionalProperties: true
        '404':
          description: Form not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /forms/{id}/submit:
    post:
      summary: Submit form data
      description: Submit data for a form
      tags:
        - Forms
      parameters:
        - name: id
          in: path
          required: true
          description: Form ID
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              description: Form submission data
              additionalProperties: true
      responses:
        '201':
          description: Form submitted successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/FormSubmission'
                  message:
                    type: string
                    example: "Form submission received"
        '400':
          description: Invalid submission data
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '404':
          description: Form not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '422':
          description: Validation errors
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ValidationError'
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /validation/login:
    get:
      summary: Get login validation schema
      description: Retrieve validation schema for login form
      tags:
        - Authentication
      responses:
        '200':
          description: Validation schema retrieved successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: object
                    description: Login form validation rules
                    additionalProperties: true
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /validation/signup:
    get:
      summary: Get signup validation schema
      description: Retrieve validation schema for signup form
      tags:
        - Authentication
      responses:
        '200':
          description: Validation schema retrieved successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: object
                    description: Signup form validation rules
                    additionalProperties: true
        '500':
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

components:
  securitySchemes:
    session:
      type: apiKey
      in: cookie
      name: session
      description: Session cookie for authentication
    csrf:
      type: apiKey
      in: header
      name: X-Csrf-Token
      description: CSRF token for state-changing operations

  schemas:
    Form:
      type: object
      properties:
        id:
          type: string
          format: uuid
          description: Unique form identifier
          example: "123e4567-e89b-12d3-a456-426614174000"
        title:
          type: string
          description: Form title
          example: "Contact Form"
        description:
          type: string
          description: Form description
          example: "A simple contact form"
        schema:
          type: object
          description: JSON schema defining the form structure
          additionalProperties: true
        status:
          type: string
          enum: [draft, published, archived]
          description: Form status
          example: "published"
        user_id:
          type: string
          format: uuid
          description: ID of the form owner
          example: "123e4567-e89b-12d3-a456-426614174001"
        created_at:
          type: string
          format: date-time
          description: Form creation timestamp
          example: "2024-01-01T00:00:00Z"
        updated_at:
          type: string
          format: date-time
          description: Form last update timestamp
          example: "2024-01-01T00:00:00Z"
      required:
        - id
        - title
        - schema
        - status
        - user_id
        - created_at
        - updated_at

    FormSubmission:
      type: object
      properties:
        id:
          type: string
          format: uuid
          description: Unique submission identifier
          example: "123e4567-e89b-12d3-a456-426614174002"
        form_id:
          type: string
          format: uuid
          description: ID of the submitted form
          example: "123e4567-e89b-12d3-a456-426614174000"
        data:
          type: object
          description: Submitted form data
          additionalProperties: true
        status:
          type: string
          enum: [pending, completed, failed, processing]
          description: Submission status
          example: "completed"
        created_at:
          type: string
          format: date-time
          description: Submission timestamp
          example: "2024-01-01T00:00:00Z"
      required:
        - id
        - form_id
        - data
        - status
        - created_at

    PaginationMeta:
      type: object
      properties:
        total:
          type: integer
          description: Total number of items
          example: 100
        page:
          type: integer
          description: Current page number
          example: 1
        per_page:
          type: integer
          description: Number of items per page
          example: 20
        total_pages:
          type: integer
          description: Total number of pages
          example: 5
      required:
        - total
        - page
        - per_page
        - total_pages

    Error:
      type: object
      properties:
        error:
          type: object
          properties:
            code:
              type: string
              description: Error code
              example: "VALIDATION_ERROR"
            message:
              type: string
              description: Human-readable error message
              example: "Invalid form data"
            details:
              type: object
              description: Additional error details
              additionalProperties: true
            request_id:
              type: string
              format: uuid
              description: Request ID for tracking
              example: "123e4567-e89b-12d3-a456-426614174003"
          required:
            - code
            - message

    ValidationError:
      type: object
      properties:
        error:
          type: object
          properties:
            code:
              type: string
              example: "VALIDATION_ERROR"
            message:
              type: string
              example: "Validation failed"
            details:
              type: array
              items:
                type: object
                properties:
                  field:
                    type: string
                    description: Field name with validation error
                    example: "email"
                  issue:
                    type: string
                    description: Description of the validation issue
                    example: "Invalid email format"
                  value:
                    type: string
                    description: Invalid value that was provided
                    example: "invalid-email"
            request_id:
              type: string
              format: uuid
              example: "123e4567-e89b-12d3-a456-426614174003"
          required:
            - code
            - message
            - details

tags:
  - name: Forms
    description: Form management operations
  - name: Authentication
    description: User authentication and validation
  - name: System
    description: System health and status endpoints`
