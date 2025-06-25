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
      // HttpClient.get returns parsed JSON data directly, not a Response object
      const data = await HttpClient.get(url);

      // Validate the schema structure
      if (!data || typeof data !== "object") {
        throw FormBuilderError.schemaError(
          "Invalid schema format received from server",
        );
      }

      // Ensure it has the expected FormSchema structure
      if (!("components" in data) || !Array.isArray(data.components)) {
        throw FormBuilderError.schemaError(
          "Schema missing required 'components' array",
        );
      }

      return data as FormSchema;
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

      // HttpClient.put returns the parsed response data directly
      // If the request fails, it will throw a FormBuilderError
      const data = await HttpClient.put(
        `${this.baseUrl}/api/v1/forms/${formId}/schema`,
        JSON.stringify(schema),
      );

      Logger.debug("Response data received:", data);

      if (!data || typeof data !== "object") {
        Logger.error("Invalid response format:", data);
        Logger.groupEnd();
        throw FormBuilderError.schemaError("Invalid response from server");
      }

      if (!data.components || !Array.isArray(data.components)) {
        Logger.error("Invalid schema structure in response:", data);
        Logger.groupEnd();
        throw FormBuilderError.schemaError(
          "Invalid schema structure in response",
        );
      }

      Logger.debug("Schema saved successfully");
      Logger.groupEnd();
      return data as FormSchema;
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
      const response = await HttpClient.put(
        `${this.baseUrl}/dashboard/forms/${formId}`,
        JSON.stringify(details),
      );

      if (!response.ok) {
        const error = await response.json();
        throw FormBuilderError.networkError(
          error.message || "Failed to update form details",
          `${this.baseUrl}/dashboard/forms/${formId}`,
          response.status,
        );
      }
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
      const response = await HttpClient.delete(
        `${this.baseUrl}/forms/${formId}`,
      );

      if (!response.ok) {
        throw FormBuilderError.networkError(
          "Failed to delete form",
          `${this.baseUrl}/forms/${formId}`,
          response.status,
        );
      }
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

      const response = await HttpClient.post(
        `${this.baseUrl}/api/v1/forms/${formId}/submit`,
        JSON.stringify(sanitizedData),
      );

      if (!response.ok) {
        const error = await response
          .json()
          .catch(() => ({ message: "Failed to submit form" }));
        throw FormBuilderError.networkError(
          error.message || "Failed to submit form",
          `${this.baseUrl}/api/v1/forms/${formId}/submit`,
          response.status,
        );
      }

      return response;
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
