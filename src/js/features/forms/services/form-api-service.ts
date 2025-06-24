import { FormBuilderError } from "@/core/errors/form-builder-error";
import { HttpClient } from "@/core/http-client";
import DOMPurify from "dompurify";

export interface FormSchema {
  display: string;
  components: any[];
  [key: string]: any;
}

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

    const response = await HttpClient.get(url);
    if (!response.ok) {
      console.error(
        "Failed to fetch schema:",
        response.status,
        response.statusText,
      );
      throw new Error("Failed to load form schema");
    }
    const data = await response.json();
    return data as FormSchema;
  }

  async saveSchema(formId: string, schema: any): Promise<any> {
    try {
      const response = await HttpClient.put(
        `${this.baseUrl}/api/v1/forms/${formId}/schema`,
        JSON.stringify(schema),
      );

      if (!response.ok) {
        throw new Error(`Failed to save schema: ${response.statusText}`);
      }

      const data = await response.json();
      if (!data || typeof data !== "object") {
        throw new Error("Invalid response from server");
      }

      if (!data.components || !Array.isArray(data.components)) {
        throw new Error("Invalid schema structure in response");
      }

      return data;
    } catch (error) {
      console.error("Error saving schema:", error);
      throw new FormBuilderError(
        "Failed to save schema",
        error instanceof Error ? error.message : String(error),
      );
    }
  }

  async updateFormDetails(
    formId: string,
    details: { title: string; description: string },
  ): Promise<void> {
    const response = await HttpClient.put(
      `${this.baseUrl}/dashboard/forms/${formId}`,
      JSON.stringify(details),
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || "Failed to update form details");
    }
  }

  public async deleteForm(formId: string): Promise<void> {
    const response = await HttpClient.delete(`${this.baseUrl}/forms/${formId}`);

    if (!response.ok) {
      throw new Error("Failed to delete form");
    }
  }

  async submitForm(formId: string, data: FormData): Promise<Response> {
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
      throw new Error(error.message || "Failed to submit form");
    }

    return response;
  }

  // Helper function to sanitize form data
  private sanitizeFormData(data: any): any {
    if (typeof data !== "object" || data === null) {
      return data;
    }

    const sanitized: any = Array.isArray(data) ? [] : {};

    for (const [key, value] of Object.entries(data)) {
      if (typeof value === "string") {
        // Use DOMPurify for string sanitization
        sanitized[key] = DOMPurify.sanitize(value);
      } else if (typeof value === "object" && value !== null) {
        sanitized[key] = this.sanitizeFormData(value);
      } else {
        sanitized[key] = value;
      }
    }

    return sanitized;
  }
}
