package openapi

// OpenAPISpec contains the OpenAPI 3.1 specification for the GoFormX API
const OpenAPISpec = `openapi: 3.1.0
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
        - BearerAuth: []
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
                $ref: '#/components/schemas/Form'
        '400':
          $ref: '#/components/responses/BadRequest'
        '422':
          $ref: '#/components/responses/ValidationError'
        '500':
          $ref: '#/components/responses/InternalServerError'
  /api/v1/forms/{id}:
    get:
      summary: Get form by ID
      description: Retrieve a specific form by its ID
      operationId: getForm
      security:
        - BearerAuth: []
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
                allOf:
                  - $ref: '#/components/schemas/APIResponse'
                  - type: object
                    properties:
                      data:
                        type: object
                        properties:
                          form:
                            $ref: '#/components/schemas/Form'
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
              $ref: '#/components/schemas/UpdateFormRequest'
      responses:
        '200':
          description: Form updated successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Form'
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
      parameters:
        - name: id
          in: path
          required: true
          description: Form ID
          schema:
            type: string
            format: uuid
      responses:
        '204':
          description: Form deleted successfully
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'
  /api/v1/forms/{id}/schema:
    get:
      summary: Get form schema
      description: Retrieve the JSON schema for a form
      operationId: getFormSchema
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
                description: JSON schema defining the form structure
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'
    put:
      summary: Update form schema
      description: Update the JSON schema for a form
      operationId: updateFormSchema
      security:
        - BearerAuth: []
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
              description: JSON schema defining the form structure
      responses:
        '200':
          description: Form schema updated successfully
          content:
            application/json:
              schema:
                type: object
                description: Updated JSON schema
        '400':
          $ref: '#/components/responses/BadRequest'
        '403':
          $ref: '#/components/responses/Forbidden'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'
  /api/v1/forms/{id}/validation:
    get:
      summary: Get form validation schema
      description: Retrieve client-side validation rules for a form
      operationId: getFormValidation
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
                allOf:
                  - $ref: '#/components/schemas/APIResponse'
                  - type: object
                    properties:
                      data:
                        type: object
                        description: Client-side validation rules
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'
  /api/v1/forms/{id}/submit:
    post:
      summary: Submit form data
      description: Submit form data for processing
      operationId: submitForm
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
      responses:
        '200':
          description: Form submitted successfully
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
                          submission_id:
                            type: string
                            format: uuid
                          status:
                            type: string
                            enum: [pending, processing, completed, failed]
                          submitted_at:
                            type: string
                            format: date-time
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
                type: object
                properties:
                  status:
                    type: string
                    enum: [healthy]
                  timestamp:
                    type: string
                    format: date-time
                  version:
                    type: string
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
  schemas:
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
    Form:
      type: object
      required:
        - id
        - title
        - created_at
        - updated_at
      properties:
        id:
          type: string
          format: uuid
          description: Unique identifier for the form
        title:
          type: string
          minLength: 3
          maxLength: 100
          description: Title of the form
        description:
          type: string
          maxLength: 500
          description: Optional description of the form
        status:
          type: string
          enum: [draft, active, inactive]
          description: Current status of the form
        schema:
          type: object
          description: JSON schema defining the form structure
        created_at:
          type: string
          format: date-time
          description: Timestamp when the form was created
        updated_at:
          type: string
          format: date-time
          description: Timestamp when the form was last updated
    FormListItem:
      type: object
      required:
        - id
        - title
        - created_at
        - updated_at
      properties:
        id:
          type: string
          format: uuid
          description: Unique identifier for the form
        title:
          type: string
          minLength: 3
          maxLength: 100
          description: Title of the form
        description:
          type: string
          maxLength: 500
          description: Optional description of the form
        status:
          type: string
          enum: [draft, active, inactive]
          description: Current status of the form
        created_at:
          type: string
          format: date-time
          description: Timestamp when the form was created
        updated_at:
          type: string
          format: date-time
          description: Timestamp when the form was last updated
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
          type: string
          format: uuid
          description: Unique identifier for the submission
        form_id:
          type: string
          format: uuid
          description: ID of the form this submission belongs to
        data:
          type: object
          description: Submitted form data
        submitted_at:
          type: string
          format: date-time
          description: Timestamp when the form was submitted
        status:
          type: string
          enum: [pending, processing, completed, failed]
          description: Current status of the submission
        metadata:
          type: object
          description: Additional metadata for the submission
    CreateFormRequest:
      type: object
      required:
        - name
      properties:
        name:
          type: string
          minLength: 1
          maxLength: 255
          description: Name of the form
        description:
          type: string
          maxLength: 1000
          description: Optional description of the form
        schema:
          type: object
          description: JSON schema defining the form structure
    UpdateFormRequest:
      type: object
      properties:
        name:
          type: string
          minLength: 1
          maxLength: 255
          description: Name of the form
        description:
          type: string
          maxLength: 1000
          description: Optional description of the form
        schema:
          type: object
          description: JSON schema defining the form structure
    Error:
      type: object
      required:
        - success
        - message
      properties:
        success:
          type: boolean
          example: false
        message:
          type: string
          description: Error message
    ValidationError:
      type: object
      required:
        - success
        - message
        - data
      properties:
        success:
          type: boolean
          example: false
        message:
          type: string
          example: "Validation failed"
        data:
          type: object
          properties:
            errors:
              type: array
              items:
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
  responses:
    BadRequest:
      description: Bad request
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    Unauthorized:
      description: Authentication required
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    Forbidden:
      description: Access forbidden
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
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
            $ref: '#/components/schemas/Error'
`
