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
    console.debug("FormApiService initialized with base URL:", this.baseUrl);
  }

  public static getInstance(): FormApiService {
    if (!FormApiService.instance) {
      FormApiService.instance = new FormApiService();
    }
    return FormApiService.instance;
  }

  public setBaseUrl(url: string): void {
    this.baseUrl = url;
    console.debug("FormApiService base URL updated to:", this.baseUrl);
  }

  async getSchema(formId: string): Promise<FormSchema> {
    const url = `${this.baseUrl}/api/v1/forms/${formId}/schema`;
    console.debug("Fetching schema from:", url);

    try {
      const response = await HttpClient.get(url);
      if (!response.ok) {
        throw FormBuilderError.networkError(
          `Failed to fetch schema: ${response.status} ${response.statusText}`,
          url,
          response.status,
        );
      }
      const data = await response.json();
      return data as FormSchema;
    } catch (error) {
      if (error instanceof FormBuilderError) {
        throw error;
      }
      throw FormBuilderError.loadFailed(
        "Failed to load form schema",
        formId,
        error instanceof Error ? error : undefined,
      );
    }
  }

  async saveSchema(formId: string, schema: FormSchema): Promise<FormSchema> {
    try {
      const response = await HttpClient.put(
        `${this.baseUrl}/api/v1/forms/${formId}/schema`,
        JSON.stringify(schema),
      );

      if (!response.ok) {
        throw FormBuilderError.networkError(
          `Failed to save schema: ${response.statusText}`,
          `${this.baseUrl}/api/v1/forms/${formId}/schema`,
          response.status,
        );
      }

      const data = await response.json();
      if (!data || typeof data !== "object") {
        throw FormBuilderError.schemaError("Invalid response from server");
      }

      if (!data.components || !Array.isArray(data.components)) {
        throw FormBuilderError.schemaError(
          "Invalid schema structure in response",
        );
      }

      return data as FormSchema;
    } catch (error) {
      if (error instanceof FormBuilderError) {
        throw error;
      }
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
