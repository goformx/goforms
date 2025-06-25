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
      let data: Partial<ServerResponse> = {};

      // Only try to parse JSON if the response has JSON content
      if (contentType?.includes("application/json")) {
        data = await response.json();
        Logger.debug("Response data:", data);
      } else {
        Logger.debug("Response has no JSON content, status:", response.status);
      }

      if (response.redirected) {
        const redirectUrl = response.url;
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

      // Check for standardized response format
      if (data.success === true) {
        this.handleSuccess(data, form);
      } else if (data.success === false) {
        this.handleError(data, form);
      } else {
        // Handle legacy response format (no success field)
        if (data.message || data.errors) {
          this.handleError(data, form);
        } else {
          this.handleSuccess(data, form);
        }
      }
    } catch (error) {
      Logger.error("Failed to parse response:", error);
      this.handleError(
        { success: false, message: "Failed to process server response" },
        form,
      );
    } finally {
      Logger.groupEnd();
    }
  }

  private static handleSuccess(
    data: Partial<ServerResponse>,
    form: HTMLFormElement,
  ): void {
    Logger.debug("Form submission successful:", data);

    // Get message from data field if available, fallback to message field
    const message =
      (data.data as any)?.message ??
      data.message ??
      "Form submitted successfully!";
    UIManager.displayFormSuccess(form, message);

    // Handle redirect from data field
    const redirectUrl = (data.data as any)?.redirect;
    if (redirectUrl) {
      setTimeout(() => {
        window.location.href = redirectUrl;
      }, 1500);
    }

    // Reset form if no redirect
    if (!redirectUrl) {
      form.reset();
    }
  }

  private static handleError(
    data: Partial<ServerResponse>,
    form: HTMLFormElement,
  ): void {
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
