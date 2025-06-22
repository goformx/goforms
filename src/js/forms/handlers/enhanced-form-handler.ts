// ===== src/js/forms/handlers/enhanced-form-handler.ts =====
import type { FormConfig } from "../types/form-types";
import { validation } from "../validation/validation";
import { ValidationHandler } from "./validation-handler";
import { ResponseHandler } from "./response-handler";
import { UIManager } from "./ui-manager";
import { isAuthenticationEndpoint } from "../utils/endpoint-utils";

/**
 * Class-based form handler for more complex use cases
 */
export class EnhancedFormHandler {
  private form: HTMLFormElement;
  private validationType: string;

  constructor(config: FormConfig) {
    const formElement = document.querySelector<HTMLFormElement>(
      `#${config.formId}`,
    );
    if (!formElement) {
      throw new Error(`Form with ID "${config.formId}" not found`);
    }

    this.form = formElement;
    this.validationType = config.validationType;

    this.initialize(config);
  }

  private initialize(config: FormConfig): void {
    validation.setupRealTimeValidation(this.form.id, config.validationType);

    this.form.addEventListener("submit", (event) =>
      this.handleFormSubmission(event),
    );
  }

  private async sendFormData(formData: FormData): Promise<Response> {
    console.group("Form Submission - Enhanced Handler");

    try {
      const csrfToken = validation.getCSRFToken();
      console.log("CSRF Token from meta tag:", csrfToken);
      console.log("Sending request to:", this.form.action);
      console.log("All cookies:", document.cookie);
      console.log("Cookies that will be sent:", document.cookie);

      const isAuthEndpoint = isAuthenticationEndpoint(this.form.action);

      if (isAuthEndpoint) {
        return this.sendAuthRequest(formData, csrfToken);
      } else {
        return this.sendStandardRequest(formData);
      }
    } finally {
      console.groupEnd();
    }
  }

  private async sendAuthRequest(
    formData: FormData,
    csrfToken: string | null,
  ): Promise<Response> {
    const data = Object.fromEntries(formData.entries());
    delete data.csrf_token; // Remove from payload

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      "Accept": "application/json",
      "X-Requested-With": "XMLHttpRequest",
    };

    if (csrfToken) {
      headers["X-Csrf-Token"] = csrfToken;
    }

    console.log("Cleaned Form Data:", data);

    return await fetch(this.form.action, {
      method: "POST",
      body: JSON.stringify(data),
      credentials: "include",
      headers,
    });
  }

  private async sendStandardRequest(formData: FormData): Promise<Response> {
    // Remove CSRF token from form data since fetchWithAuth should add it to headers
    const cleanFormData = new FormData();
    for (const [key, value] of formData.entries()) {
      if (key !== "csrf_token") {
        cleanFormData.append(key, value);
      }
    }

    return validation.fetchWithAuth(this.form.action, {
      method: this.form.method,
      body: cleanFormData,
    });
  }

  private async handleFormSubmission(event: Event): Promise<void> {
    event.preventDefault();

    try {
      const isValid = await ValidationHandler.validateFormSubmission(
        this.form,
        this.validationType,
      );

      if (!isValid) {
        this.showError("Please check the form for errors.");
        return;
      }

      const formData = new FormData(this.form);
      const response = await this.sendFormData(formData);
      await ResponseHandler.handleServerResponse(response, this.form);
    } catch (error) {
      console.error("Form submission error:", error);
      this.showError("An unexpected error occurred. Please try again.");
    }
  }

  private showError(message: string, field?: string): void {
    UIManager.displayFormError(this.form, message);

    if (field) {
      const fieldElement = this.form.querySelector(`[name="${field}"]`);
      if (fieldElement) {
        fieldElement.classList.add("error");
      }
    }
  }
}
