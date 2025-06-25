// ===== src/js/forms/handlers/enhanced-form-handler.ts =====
import { Logger } from "@/core/logger";
import type { FormConfig } from "@/shared/types/form-types";
import { validation } from "@/features/forms/validation/validation";
import { ValidationHandler } from "./validation-handler";
import { ResponseHandler } from "./response-handler";
import { UIManager } from "./ui-manager";
import { isAuthenticationEndpoint } from "@/shared/utils/endpoint-utils";
import { HttpClient } from "@/core/http-client";

/**
 * Class-based form handler for more complex use cases
 */
export class EnhancedFormHandler {
  private form: HTMLFormElement;
  private formId: string;

  constructor(config: FormConfig) {
    Logger.debug("EnhancedFormHandler: Initializing with config:", config);

    const formElement = document.querySelector<HTMLFormElement>(
      `#${config.formId}`,
    );
    if (!formElement) {
      throw new Error(`Form with ID "${config.formId}" not found`);
    }

    Logger.debug("EnhancedFormHandler: Form element found:", formElement);
    Logger.debug("EnhancedFormHandler: Form action:", formElement.action);

    this.form = formElement;
    this.formId = config.formId;

    this.initialize();
  }

  private initialize(): void {
    validation.setupRealTimeValidation(this.form.id, this.formId);

    this.form.addEventListener("submit", (event) =>
      this.handleFormSubmission(event),
    );
  }

  private async sendFormData(formData: FormData): Promise<Response> {
    Logger.group("Form Submission - Enhanced Handler");

    try {
      const csrfToken = validation.getCSRFToken();
      Logger.debug("CSRF Token from meta tag:", csrfToken);
      Logger.debug("Sending request to:", this.form.action);
      Logger.debug("All cookies:", document.cookie);
      Logger.debug("Cookies that will be sent:", document.cookie);

      const isAuthEndpoint = isAuthenticationEndpoint(this.form.action);

      if (isAuthEndpoint) {
        return this.sendAuthRequest(formData, csrfToken);
      } else {
        return this.sendStandardRequest(formData);
      }
    } finally {
      Logger.groupEnd();
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
      Accept: "application/json",
      "X-Requested-With": "XMLHttpRequest",
    };

    if (csrfToken) {
      headers["X-Csrf-Token"] = csrfToken;
    }

    Logger.debug("Cleaned Form Data:", data);

    return await HttpClient.post(this.form.action, JSON.stringify(data), {
      headers,
    });
  }

  private async sendStandardRequest(formData: FormData): Promise<Response> {
    // Remove CSRF token from form data since HttpClient should add it to headers
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
    Logger.debug("EnhancedFormHandler: Form submission intercepted");
    event.preventDefault();

    try {
      Logger.debug("EnhancedFormHandler: Starting form validation");
      const isValid = await ValidationHandler.validateFormSubmission(
        this.form,
        this.formId,
      );

      if (!isValid) {
        Logger.debug("EnhancedFormHandler: Form validation failed");
        this.showError("Please check the form for errors.");
        return;
      }

      Logger.debug("EnhancedFormHandler: Form validation passed, sending data");
      const formData = new FormData(this.form);
      const response = await this.sendFormData(formData);
      await ResponseHandler.handleServerResponse(response, this.form);
    } catch (error) {
      Logger.error("EnhancedFormHandler: Form submission error:", error);
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
