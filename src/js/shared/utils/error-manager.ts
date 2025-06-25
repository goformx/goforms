import { Logger } from "@/core/logger";

/**
 * Centralized error management for form validation and display
 */
export class ErrorManager {
  static showFieldError(fieldId: string, message: string): void {
    Logger.debug("Showing field error", { fieldId, message });
    this.clearFieldError(fieldId);

    const input = document.getElementById(fieldId) as HTMLInputElement;
    const errorElement = document.getElementById(`${fieldId}_error`);

    if (input) {
      input.classList.add("error");
      input.setAttribute("aria-invalid", "true");
      Logger.debug("Added error class to input", { fieldId });
    } else {
      Logger.warn("Input element not found", { fieldId });
    }

    if (errorElement) {
      errorElement.textContent = message;
      Logger.debug("Set error message", { fieldId, message });
    } else {
      Logger.warn("Error element not found", {
        fieldId,
        errorElementId: `${fieldId}_error`,
      });
    }
  }

  static clearFieldError(fieldId: string): void {
    Logger.debug("Clearing field error", { fieldId });
    const input = document.getElementById(fieldId) as HTMLInputElement;
    const errorElement = document.getElementById(`${fieldId}_error`);

    if (input) {
      input.classList.remove("error");
      input.setAttribute("aria-invalid", "false");
    }

    if (errorElement) {
      errorElement.textContent = "";
    }
  }

  static showFormError(form: HTMLFormElement, message: string): void {
    const errorContainer = form.querySelector(".form-error") as HTMLElement;
    if (errorContainer) {
      errorContainer.textContent = message;
      errorContainer.classList.remove("hidden");
    }
  }

  static showFormSuccess(form: HTMLFormElement, message: string): void {
    const successContainer = form.querySelector(".form-success") as HTMLElement;
    if (successContainer) {
      successContainer.textContent = message;
      successContainer.classList.remove("hidden");
    }
  }

  static clearAllErrors(form: HTMLFormElement): void {
    // Clear all field errors
    form.querySelectorAll<HTMLInputElement>("input[id]").forEach((input) => {
      this.clearFieldError(input.id);
    });

    // Clear form-level errors
    const errorContainer = form.querySelector(".form-error") as HTMLElement;
    if (errorContainer) {
      errorContainer.textContent = "";
      errorContainer.classList.add("hidden");
    }

    // Clear form-level success messages
    const successContainer = form.querySelector(".form-success") as HTMLElement;
    if (successContainer) {
      successContainer.textContent = "";
      successContainer.classList.add("hidden");
    }
  }

  static showErrors(
    form: HTMLFormElement,
    errors: Record<string, string>,
  ): void {
    Object.entries(errors).forEach(([field, message]) => {
      const input = form.querySelector(`[name="${field}"]`) as HTMLInputElement;
      if (input ?.input.id) {
        this.showFieldError(input.id, message);
      }
    });
  }

  static clearErrors(form: HTMLFormElement): void {
    form.querySelectorAll(".error-message").forEach((el) => el.remove());
    form
      .querySelectorAll(".error")
      .forEach((el) => el.classList.remove("error"));
  }
}
