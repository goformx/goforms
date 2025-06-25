import { Logger } from "@/core/logger";
import type { ServerResponse } from "@/shared/types/form-types";
import { UIManager } from "./ui-manager";

export class ResponseHandler {
  /**
   * Handles the server's response to the form submission
   */
  static async handleServerResponse(
    response: Response,
    form: HTMLFormElement,
  ): Promise<void> {
    Logger.group("Response Handling");

    try {
      const contentType = response.headers.get("content-type");
      let data: ServerResponse = {};

      // Only try to parse JSON if the response has JSON content
      if (contentType?.includes("application/json")) {
        data = await response.json();
        Logger.debug("Response data:", data);
      } else {
        Logger.debug("Response has no JSON content, status:", response.status);
      }

      if (response.redirected || data.redirect) {
        const redirectUrl = response.redirected ? response.url : data.redirect!;
        Logger.debug("Redirecting to:", redirectUrl);
        window.location.href = redirectUrl;
        return;
      }

      if (!response.ok) {
        const message = data.message ?? "An error occurred. Please try again.";
        Logger.warn("Request failed:", message);
        UIManager.displayFormError(form, message);
        return;
      }

      if (data.success) {
        this.handleSuccess(data, form);
      } else {
        this.handleError(data, form);
      }
    } catch (error) {
      Logger.error("Failed to parse response:", error);
      this.handleError({ message: "Failed to process server response" }, form);
    } finally {
      Logger.groupEnd();
    }
  }

  private static handleSuccess(data: any, form: HTMLFormElement): void {
    Logger.debug("Form submission successful:", data);

    // Show success message
    const message = data.message ?? "Form submitted successfully!";
    UIManager.displayFormSuccess(form, message);

    // Handle redirect if specified
    if (data.redirect) {
      setTimeout(() => {
        window.location.href = data.redirect;
      }, 1500);
    }

    // Reset form if no redirect
    if (!data.redirect) {
      form.reset();
    }
  }

  private static handleError(data: any, form: HTMLFormElement): void {
    Logger.error("Form submission failed:", data);

    // Show error message
    const message = data.message ?? "An error occurred. Please try again.";
    UIManager.displayFormError(form, message);

    // Handle field-specific errors
    if (data.errors && typeof data.errors === "object") {
      Object.entries(data.errors).forEach(([fieldName, errorMessage]) => {
        this.showFieldError(form, fieldName, String(errorMessage));
      });
    }
  }

  private static showFieldError(
    form: HTMLFormElement,
    fieldName: string,
    _message: string,
  ): void {
    const field = form.querySelector(`[name="${fieldName}"]`) as HTMLElement;
    if (field) {
      field.classList.add("error");
      field.setAttribute("aria-invalid", "true");
    }
  }
}
