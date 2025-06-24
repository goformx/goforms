import { FormBuilderError } from "../../../core/errors/form-builder-error";
import DOMPurify from "dompurify";

export interface FormSchema {
  display: string;
  components: any[];
  [key: string]: any;
}

export class FormService {
  private static instance: FormService;
  private baseUrl: string;

  private constructor() {
    this.baseUrl = window.location.origin;
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
            "X-Requested-With": "XMLHttpRequest",
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
    const response = await fetch(`${this.baseUrl}/dashboard/forms/${formId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        "X-Requested-With": "XMLHttpRequest",
      },
      body: JSON.stringify(details),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || "Failed to update form details");
    }
  }

  public async deleteForm(formId: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/forms/${formId}`, {
      method: "DELETE",
      credentials: "include",
    });

    if (!response.ok) {
      throw new Error("Failed to delete form");
    }
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

  async submitForm(formId: string, data: FormData): Promise<Response> {
    // Sanitize the form data before sending
    const sanitizedData = this.sanitizeFormData(data);

    const response = await fetch(
      `${this.baseUrl}/api/v1/forms/${formId}/submit`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify(sanitizedData),
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

// Initialize form deletion handlers
document.addEventListener("DOMContentLoaded", () => {
  const formService = FormService.getInstance();

  document.querySelectorAll(".delete-form").forEach((button) => {
    button.addEventListener("click", async (e) => {
      e.preventDefault();
      const formId = button.getAttribute("data-form-id");
      if (!formId) return;

      if (
        !confirm(
          "Are you sure you want to delete this form? This action cannot be undone.",
        )
      ) {
        return;
      }

      try {
        await formService.deleteForm(formId);
        const formCard = button.closest(".form-card");
        if (formCard) {
          formCard.remove();
        }

        // If no forms left, show empty state
        const formsGrid = document.querySelector(".forms-grid");
        if (formsGrid && !formsGrid.querySelector(".form-card")) {
          formsGrid.innerHTML = `
            <div class="empty-state">
              <i class="bi bi-file-earmark-text"></i>
              <p>You haven't created any forms yet.</p>
              <a href="/forms/new" class="btn btn-primary">Create Your First Form</a>
            </div>
          `;
        }
      } catch (error) {
        console.error("Failed to delete form:", error);
        alert(error instanceof Error ? error.message : "Failed to delete form");
      }
    });
  });
});
