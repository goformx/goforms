import { Logger } from "@/core/logger";
import { FormBuilderError } from "@/core/errors/form-builder-error";
import { HttpClient } from "@/core/http-client";
import DOMPurify from "dompurify";
import type { FormSchema } from "@/shared/types/form-types";

/**
 * Handles all HTTP operations for forms
 */
export class FormApiService {
  private static instance: FormApiService;
  private baseUrl: string;

  private constructor() {
    this.baseUrl = window.location.origin;
    Logger.debug("FormApiService initialized with base URL:", this.baseUrl);
  }

  public static getInstance(): FormApiService {
    if (!FormApiService.instance) {
      FormApiService.instance = new FormApiService();
    }
    return FormApiService.instance;
  }

  public setBaseUrl(url: string): void {
    this.baseUrl = url;
    Logger.debug("FormApiService base URL updated to:", this.baseUrl);
  }

  async getSchema(formId: string): Promise<FormSchema> {
    const url = `${this.baseUrl}/api/v1/forms/${formId}/schema`;
    Logger.debug("Fetching schema from:", url);

    try {
      // HttpClient.get returns HttpResponse with parsed data
      const response = await HttpClient.get<FormSchema>(url);

      // Validate the schema structure
      if (!response.data || typeof response.data !== "object") {
        throw FormBuilderError.schemaError(
          "Invalid schema format received from server",
        );
      }

      // Ensure it has the expected FormSchema structure
      if (
        !("components" in response.data) ||
        !Array.isArray(response.data.components)
      ) {
        throw FormBuilderError.schemaError(
          "Schema missing required 'components' array",
        );
      }

      return response.data;
    } catch (error) {
      Logger.error("Error in getSchema:", error);

      // If it's already a FormBuilderError, re-throw it
      if (error instanceof FormBuilderError) {
        throw error;
      }

      // For any other error, wrap it appropriately
      throw FormBuilderError.loadFailed(
        "Failed to load form schema",
        formId,
        error instanceof Error ? error : undefined,
      );
    }
  }

  async saveSchema(formId: string, schema: FormSchema): Promise<FormSchema> {
    try {
      Logger.group("Schema Save Operation");
      Logger.debug("Saving schema for form:", formId);
      Logger.debug("Schema to save:", schema);

      // HttpClient.put returns HttpResponse with parsed data
      const response = await HttpClient.put<FormSchema>(
        `${this.baseUrl}/api/v1/forms/${formId}/schema`,
        JSON.stringify(schema),
      );

      Logger.debug("Response data received:", response.data);

      if (!response.data || typeof response.data !== "object") {
        Logger.error("Invalid response format:", response.data);
        Logger.groupEnd();
        throw FormBuilderError.schemaError("Invalid response from server");
      }

      if (
        !response.data.components ||
        !Array.isArray(response.data.components)
      ) {
        Logger.error("Invalid schema structure in response:", response.data);
        Logger.groupEnd();
        throw FormBuilderError.schemaError(
          "Invalid schema structure in response",
        );
      }

      Logger.debug("Schema saved successfully");
      Logger.groupEnd();
      return response.data;
    } catch (error) {
      Logger.groupEnd();

      if (error instanceof FormBuilderError) {
        throw error;
      }

      Logger.error("Unexpected error in saveSchema:", error);
      throw FormBuilderError.saveFailed(
        "Failed to save schema",
        formId,
        error instanceof Error ? error : undefined,
      );
    }
  }

  async updateFormDetails(
    formId: string,
    details: { title: string; description: string },
  ): Promise<void> {
    try {
      // HttpClient.put returns HttpResponse, not Response
      await HttpClient.put(
        `${this.baseUrl}/dashboard/forms/${formId}`,
        JSON.stringify(details),
      );

      // If we get here, the request was successful
      Logger.debug("Form details updated successfully");
    } catch (error) {
      if (error instanceof FormBuilderError) {
        throw error;
      }
      throw FormBuilderError.saveFailed(
        "Failed to update form details",
        formId,
        error instanceof Error ? error : undefined,
      );
    }
  }

  public async deleteForm(formId: string): Promise<void> {
    try {
      // HttpClient.delete returns HttpResponse, not Response
      await HttpClient.delete(`${this.baseUrl}/forms/${formId}`);

      // If we get here, the request was successful
      Logger.debug("Form deleted successfully");
    } catch (error) {
      if (error instanceof FormBuilderError) {
        throw error;
      }
      throw FormBuilderError.saveFailed(
        "Failed to delete form",
        formId,
        error instanceof Error ? error : undefined,
      );
    }
  }

  async submitForm(formId: string, data: FormData): Promise<Response> {
    try {
      // Sanitize the form data before sending
      const sanitizedData = this.sanitizeFormData(data);

      // Use HttpClient for consistency with standardized response format
      const response = await HttpClient.post(
        `${this.baseUrl}/api/v1/forms/${formId}/submit`,
        sanitizedData as object,
      );

      // Convert HttpResponse back to Response for compatibility
      // This is needed because the ResponseHandler expects a Response object
      const responseData = response.data;
      const responseText = JSON.stringify(responseData);

      return new Response(responseText, {
        status: response.status,
        statusText: response.statusText,
        headers: response.headers,
      });
    } catch (error) {
      if (error instanceof FormBuilderError) {
        throw error;
      }
      throw FormBuilderError.saveFailed(
        "Failed to submit form",
        formId,
        error instanceof Error ? error : undefined,
      );
    }
  }

  // =============================================
  // CREATE OPERATIONS
  // =============================================

  /**
   * Create a new form
   */
  public async createForm(data: {
    title: string;
    description?: string;
  }): Promise<{ formId: string }> {
    try {
      // Create FormData for form submission
      const formData = new FormData();
      formData.append("title", data.title);
      if (data.description) {
        formData.append("description", data.description);
      }

      const response = await HttpClient.post<{ form_id: string }>(
        `${this.baseUrl}/forms`,
        formData as object,
      );

      return { formId: response.data.form_id };
    } catch (error) {
      if (error instanceof FormBuilderError) {
        throw error;
      }
      throw FormBuilderError.saveFailed(
        "Failed to create form",
        "new",
        error instanceof Error ? error : undefined,
      );
    }
  }

  // =============================================
  // READ OPERATIONS
  // =============================================

  /**
   * Get complete form data
   */
  public async getForm(formId: string): Promise<any> {
    try {
      const response = await HttpClient.get(
        `${this.baseUrl}/api/v1/forms/${formId}`,
      );
      return response.data;
    } catch (error) {
      if (error instanceof FormBuilderError) {
        throw error;
      }
      throw FormBuilderError.loadFailed(
        "Failed to get form",
        formId,
        error instanceof Error ? error : undefined,
      );
    }
  }

  /**
   * Get user's forms list
   */
  public async listForms(): Promise<any[]> {
    try {
      const response = await HttpClient.get(`${this.baseUrl}/dashboard/forms`);
      return (response.data as any[]) ?? [];
    } catch (error) {
      if (error instanceof FormBuilderError) {
        throw error;
      }
      throw FormBuilderError.loadFailed(
        "Failed to list forms",
        "dashboard",
        error instanceof Error ? error : undefined,
      );
    }
  }

  /**
   * Get form details (metadata only)
   */
  public async getFormDetails(formId: string): Promise<any> {
    try {
      const response = await HttpClient.get(
        `${this.baseUrl}/api/v1/forms/${formId}/details`,
      );
      return response.data;
    } catch (error) {
      if (error instanceof FormBuilderError) {
        throw error;
      }
      throw FormBuilderError.loadFailed(
        "Failed to get form details",
        formId,
        error instanceof Error ? error : undefined,
      );
    }
  }

  // =============================================
  // UPDATE OPERATIONS
  // =============================================

  /**
   * Update form status
   */
  public async updateFormStatus(formId: string, status: string): Promise<void> {
    try {
      await HttpClient.put(
        `${this.baseUrl}/forms/${formId}/status`,
        JSON.stringify({ status }),
      );
    } catch (error) {
      if (error instanceof FormBuilderError) {
        throw error;
      }
      throw FormBuilderError.saveFailed(
        "Failed to update form status",
        formId,
        error instanceof Error ? error : undefined,
      );
    }
  }

  /**
   * Update form CORS settings
   */
  public async updateFormCors(
    formId: string,
    corsOrigins: string,
  ): Promise<void> {
    try {
      await HttpClient.put(
        `${this.baseUrl}/forms/${formId}/cors`,
        JSON.stringify({ cors_origins: corsOrigins }),
      );
    } catch (error) {
      if (error instanceof FormBuilderError) {
        throw error;
      }
      throw FormBuilderError.saveFailed(
        "Failed to update form CORS settings",
        formId,
        error instanceof Error ? error : undefined,
      );
    }
  }

  /**
   * Comprehensive sanitization for all data types
   */
  private sanitizeFormData(data: unknown): unknown {
    // Handle null and undefined
    if (data === null || data === undefined) {
      return data;
    }

    // Handle primitive types
    if (typeof data === "string") {
      return DOMPurify.sanitize(data);
    }

    if (typeof data === "number" || typeof data === "boolean") {
      return data;
    }

    // Handle arrays
    if (Array.isArray(data)) {
      return data.map((item) => this.sanitizeFormData(item));
    }

    // Handle objects
    if (typeof data === "object") {
      const sanitized: Record<string, unknown> = {};

      for (const [key, value] of Object.entries(
        data as Record<string, unknown>,
      )) {
        // Sanitize the key as well
        const sanitizedKey = DOMPurify.sanitize(key);
        sanitized[sanitizedKey] = this.sanitizeFormData(value);
      }

      return sanitized;
    }

    // For any other type, return as is
    return data;
  }
}
