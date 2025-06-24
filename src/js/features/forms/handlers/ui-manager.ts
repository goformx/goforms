// ===== src/js/forms/handlers/ui-manager.ts =====
import { Logger } from "@/core/logger";

export class UIManager {
  /**
   * Displays an error message in the form's error container
   */
  static displayFormError(form: HTMLFormElement, message: string): void {
    Logger.debug("Displaying error message:", message);
    const formError = form.querySelector(".form-error");

    if (formError) {
      formError.textContent = message;
      formError.classList.remove("hidden");
    } else {
      Logger.warn("Form error container not found:", form.id);
    }
  }

  /**
   * Displays a success message in the form's success container
   */
  static displayFormSuccess(form: HTMLFormElement, message: string): void {
    Logger.debug("Displaying success message:", message);
    const formSuccess = form.querySelector(".form-success");

    if (formSuccess) {
      formSuccess.textContent = message;
      formSuccess.classList.remove("hidden");
    } else {
      Logger.warn("Form success container not found:", form.id);
    }
  }

  /**
   * Resets aria-invalid attributes on form inputs
   */
  static resetAriaInvalid(form: HTMLFormElement): void {
    const inputs = form.querySelectorAll<HTMLInputElement>("input[id]");
    inputs.forEach((input) => input.setAttribute("aria-invalid", "false"));
  }

  /**
   * Sets aria-invalid attribute for a specific input
   */
  static setAriaInvalid(inputId: string, invalid: boolean): void {
    const input = document.getElementById(inputId) as HTMLInputElement;
    if (input) {
      input.setAttribute("aria-invalid", invalid.toString());
    }
  }
}
