package openapi

// OpenAPISpec contains the OpenAPI 3.1.0 specification for the GoFormX API
const OpenAPISpec = `openapi: 3.0.3
info:
  title: GoFormX API
  description: Modern form management API with MariaDB backend
  version: 1.0.0
  contact:
    name: GoFormX Team
servers:
  - url: http://localhost:8090
    description: Development server
  - url: https://api.goformx.com
    description: Production server

paths:
  /api/v1/forms:
    get:
      summary: List all forms
      description: Retrieve a list of forms for the authenticated user
      operationId: listForms
      security:
        - SessionAuth: []
        - {} # Allow unauthenticated access for testing
      responses:
        '200':
          description: List of forms retrieved successfully
          content:
            application/json:
              schema:
                allOf:
                  - $ref: '#/components/schemas/APIResponse'
                  - type: object
                    properties:
                      data:
                        type: object
                        properties:
                          forms:
                            type: array
                            items:
                              $ref: '#/components/schemas/FormListItem'
                          count:
                            type: integer
                            example: 5
        '401':
          $ref: '#/components/responses/Unauthorized'
        '403':
          $ref: '#/components/responses/Forbidden'
        '500':
          $ref: '#/components/responses/InternalServerError'
    post:
      summary: Create a new form
      description: Create a new form with the provided data
      operationId: createForm
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateFormRequest'
      responses:
        '201':
          description: Form created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/FormResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '422':
          $ref: '#/components/responses/ValidationError'
        '500':
          $ref: '#/components/responses/InternalServerError'

  /api/v1/forms/{id}:
    parameters:
      - $ref: '#/components/parameters/FormId'
    get:
      summary: Get form by ID
      description: Retrieve a specific form by its ID
      operationId: getForm
      security:
        - SessionAuth: []
      responses:
        '200':
          description: Form retrieved successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/FormResponse'
        '404':
          $ref: '#/components/responses/NotFound'
        '403':
          $ref: '#/components/responses/Forbidden'
        '500':
          $ref: '#/components/responses/InternalServerError'
    put:
      summary: Update form
      description: Update an existing form with the provided data
      operationId: updateForm
      security:
        - SessionAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateFormRequest'
      responses:
        '200':
          description: Form updated successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/FormResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '404':
          $ref: '#/components/responses/NotFound'
        '422':
          $ref: '#/components/responses/ValidationError'
        '500':
          $ref: '#/components/responses/InternalServerError'
    delete:
      summary: Delete form
      description: Delete a form by its ID
      operationId: deleteForm
      security:
        - SessionAuth: []
      responses:
        '204':
          description: Form deleted successfully
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'

  /api/v1/forms/{id}/schema:
    parameters:
      - $ref: '#/components/parameters/FormId'
    get:
      summary: Get form schema
      description: Retrieve the JSON schema for a form
      operationId: getFormSchema
      responses:
        '200':
          description: Form schema retrieved successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/JSONSchema'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'
    put:
      summary: Update form schema
      description: Update the JSON schema for a form
      operationId: updateFormSchema
      security:
        - SessionAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/JSONSchema'
      responses:
        '200':
          description: Form schema updated successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/JSONSchema'
        '400':
          $ref: '#/components/responses/BadRequest'
        '403':
          $ref: '#/components/responses/Forbidden'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'

  /api/v1/forms/{id}/validation:
    parameters:
      - $ref: '#/components/parameters/FormId'
    get:
      summary: Get form validation schema
      description: Retrieve client-side validation rules for a form
      operationId: getFormValidation
      responses:
        '200':
          description: Validation schema retrieved successfully
          content:
            application/json:
              schema:
                allOf:
                  - $ref: '#/components/schemas/APIResponse'
                  - type: object
                    properties:
                      data:
                        $ref: '#/components/schemas/ValidationRules'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'

  /api/v1/forms/{id}/submit:
    parameters:
      - $ref: '#/components/parameters/FormId'
    post:
      summary: Submit form data
      description: Submit form data for processing
      operationId: submitForm
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/FormSubmissionData'
      responses:
        '200':
          description: Form submitted successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SubmissionResponse'
        '400':
          $ref: '#/components/responses/ValidationError'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'

  /health:
    get:
      summary: Health check
      description: Check the health status of the API
      operationId: healthCheck
      responses:
        '200':
          description: Service is healthy
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HealthResponse'

components:
  securitySchemes:
    SessionAuth:
      type: apiKey
      in: cookie
      name: session
      description: Session cookie for authentication

  parameters:
    FormId:
      name: id
      in: path
      required: true
      description: Form ID
      schema:
        type: string
        format: uuid

  schemas:
    # Base response schema
    APIResponse:
      type: object
      required:
        - success
      properties:
        success:
          type: boolean
          description: Whether the request was successful
        message:
          type: string
          description: Optional message describing the result
        data:
          type: object
          description: Response data (structure varies by endpoint)

    # Common field types
    UUID:
      type: string
      format: uuid
      description: Universally unique identifier

    Timestamp:
      type: string
      format: date-time
      description: ISO 8601 timestamp

    FormStatus:
      type: string
      enum: [draft, active, inactive]
      description: Current status of the form

    SubmissionStatus:
      type: string
      enum: [pending, processing, completed, failed]
      description: Current status of the submission

    FormTitle:
      type: string
      minLength: 3
      maxLength: 100
      description: Title of the form

    FormDescription:
      type: string
      maxLength: 500
      description: Optional description of the form

    FormName:
      type: string
      minLength: 1
      maxLength: 255
      description: Name of the form

    JSONSchema:
      type: object
      description: JSON schema defining the form structure

    ValidationRules:
      type: object
      description: Client-side validation rules

    FormSubmissionData:
      type: object
      description: Form submission data

    # Core entities
    Form:
      type: object
      required:
        - id
        - title
        - created_at
        - updated_at
      properties:
        id:
          $ref: '#/components/schemas/UUID'
        title:
          $ref: '#/components/schemas/FormTitle'
        description:
          $ref: '#/components/schemas/FormDescription'
        status:
          $ref: '#/components/schemas/FormStatus'
        schema:
          $ref: '#/components/schemas/JSONSchema'
        created_at:
          $ref: '#/components/schemas/Timestamp'
        updated_at:
          $ref: '#/components/schemas/Timestamp'

    FormListItem:
      type: object
      required:
        - id
        - title
        - created_at
        - updated_at
      properties:
        id:
          $ref: '#/components/schemas/UUID'
        title:
          $ref: '#/components/schemas/FormTitle'
        description:
          $ref: '#/components/schemas/FormDescription'
        status:
          $ref: '#/components/schemas/FormStatus'
        created_at:
          $ref: '#/components/schemas/Timestamp'
        updated_at:
          $ref: '#/components/schemas/Timestamp'

    FormSubmission:
      type: object
      required:
        - id
        - form_id
        - data
        - submitted_at
        - status
      properties:
        id:
          $ref: '#/components/schemas/UUID'
        form_id:
          $ref: '#/components/schemas/UUID'
        data:
          $ref: '#/components/schemas/FormSubmissionData'
        submitted_at:
          $ref: '#/components/schemas/Timestamp'
        status:
          $ref: '#/components/schemas/SubmissionStatus'
        metadata:
          type: object
          description: Additional metadata for the submission

    # Request/Response schemas
    CreateFormRequest:
      type: object
      required:
        - name
      properties:
        name:
          $ref: '#/components/schemas/FormName'
        description:
          type: string
          maxLength: 1000
          description: Optional description of the form
        schema:
          $ref: '#/components/schemas/JSONSchema'

    UpdateFormRequest:
      type: object
      properties:
        name:
          $ref: '#/components/schemas/FormName'
        description:
          type: string
          maxLength: 1000
          description: Optional description of the form
        schema:
          $ref: '#/components/schemas/JSONSchema'

    FormResponse:
      allOf:
        - $ref: '#/components/schemas/APIResponse'
        - type: object
          properties:
            data:
              type: object
              properties:
                form:
                  $ref: '#/components/schemas/Form'

    SubmissionResponse:
      allOf:
        - $ref: '#/components/schemas/APIResponse'
        - type: object
          properties:
            data:
              type: object
              properties:
                submission_id:
                  $ref: '#/components/schemas/UUID'
                status:
                  $ref: '#/components/schemas/SubmissionStatus'
                submitted_at:
                  $ref: '#/components/schemas/Timestamp'

    HealthResponse:
      type: object
      required:
        - status
        - timestamp
        - version
      properties:
        status:
          type: string
          enum: [healthy]
        timestamp:
          $ref: '#/components/schemas/Timestamp'
        version:
          type: string

    # Error schemas
    BaseError:
      type: object
      required:
        - success
        - message
      properties:
        success:
          type: boolean
          enum: [false]
        message:
          type: string
          description: Error message

    ValidationErrorDetail:
      type: object
      required:
        - field
        - message
      properties:
        field:
          type: string
          description: Field name that failed validation
        message:
          type: string
          description: Validation error message for the field
        rule:
          type: string
          description: Validation rule that failed

    ValidationError:
      allOf:
        - $ref: '#/components/schemas/BaseError'
        - type: object
          required:
            - data
          properties:
            message:
              type: string
              enum: ["Validation failed"]
            data:
              type: object
              properties:
                errors:
                  type: array
                  items:
                    $ref: '#/components/schemas/ValidationErrorDetail'

  responses:
    BadRequest:
      description: Bad request
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/BaseError'

    Unauthorized:
      description: Authentication required
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/BaseError'

    Forbidden:
      description: Access forbidden
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/BaseError'

    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/BaseError'

    ValidationError:
      description: Validation error
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ValidationError'

    InternalServerError:
      description: Internal server error
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/BaseError'
`
