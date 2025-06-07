import { FormBuilderError } from "../utils/errors";

// Add global type declaration
declare global {
  interface Window {
    CSRF_TOKEN?: string;
  }
}

export interface FormSchema {
  display: string;
  components: any[];
  [key: string]: any;
}

export class FormService {
  private static instance: FormService;
  private baseUrl: string;

  private constructor() {
    this.baseUrl =
      process.env.NODE_ENV === "production"
        ? "https://goformx.com"
        : window.location.origin;
    console.debug("FormService initialized with base URL:", this.baseUrl);
  }

  public static getInstance(): FormService {
    if (!FormService.instance) {
      FormService.instance = new FormService();
    }
    return FormService.instance;
  }

  public setBaseUrl(url: string): void {
    this.baseUrl = url;
    console.debug("FormService base URL updated to:", this.baseUrl);
  }

  private getCSRFToken(): string {
    // Try to get token from form builder element
    const formBuilder = document.getElementById("form-schema-builder");
    if (formBuilder) {
      const token = formBuilder.getAttribute("data-csrf-token");
      if (token) {
        console.debug("CSRF token from form builder:", token);
        return token;
      }
    }

    // Try to get token from meta tag
    const metaTag = document.querySelector('meta[name="csrf-token"]');
    if (metaTag) {
      const token = metaTag.getAttribute("content");
      if (token) {
        console.debug("CSRF token from meta tag:", token);
        return token;
      }
    }

    // Try to get token from window object
    if (window.CSRF_TOKEN) {
      console.debug("CSRF token from window object:", window.CSRF_TOKEN);
      return window.CSRF_TOKEN;
    }

    console.error("CSRF token not found in any source");
    return "";
  }

  async getSchema(formId: string): Promise<FormSchema> {
    const url = `${this.baseUrl}/api/v1/forms/${formId}/schema`;
    console.debug("Fetching schema from:", url);

    const response = await fetch(url);
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
      const response = await fetch(
        `${this.baseUrl}/api/v1/forms/${formId}/schema`,
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            "X-CSRF-Token": this.getCSRFToken(),
          },
          body: JSON.stringify(schema),
        },
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
      throw new FormBuilderError("Failed to save schema", error);
    }
  }

  async updateFormDetails(
    formId: string,
    details: { title: string; description: string },
  ): Promise<void> {
    const response = await fetch(`${this.baseUrl}/dashboard/forms/${formId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": this.getCSRFToken(),
      },
      body: JSON.stringify(details),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || "Failed to update form details");
    }
  }

  async deleteForm(formId: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/dashboard/forms/${formId}`, {
      method: "DELETE",
      headers: {
        "X-CSRF-Token": this.getCSRFToken(),
      },
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || "Failed to delete form");
    }
  }

  async submitForm(formId: string, data: FormData): Promise<Response> {
    const response = await fetch(
      `${this.baseUrl}/api/v1/forms/${formId}/submit`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify(data),
      },
    );

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ message: "Failed to submit form" }));
      throw new Error(error.message || "Failed to submit form");
    }

    return response;
  }
}
