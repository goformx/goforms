// ===== src/js/forms/handlers/validation-handler.ts =====
import { validation } from "../validation/validation";
import { debounce } from "../utils/debounce";
import { UIManager } from "./ui-manager";

export class ValidationHandler {
  /**
   * Sets up real-time validation for form inputs with debouncing
   */
  static setupRealTimeValidation(
    form: HTMLFormElement,
    validationType: string,
    delay = 300,
  ): void {
    const inputs = form.querySelectorAll<HTMLInputElement>("input[id]");

    inputs.forEach((input) => {
      input.addEventListener(
        "input",
        debounce(
          () => this.handleInputValidation(input, form, validationType),
          delay,
        ),
      );
    });
  }

  /**
   * Handles real-time validation for individual input fields
   */
  private static async handleInputValidation(
    input: HTMLInputElement,
    form: HTMLFormElement,
    validationType: string,
  ): Promise<void> {
    try {
      validation.clearError(input.id);
      UIManager.setAriaInvalid(input.id, false);

      const result = await validation.validateForm(form, validationType);

      if (!result.success && result.error?.errors) {
        result.error.errors.forEach((err) => {
          if (err.path[0] === input.id) {
            validation.showError(input.id, err.message);
            UIManager.setAriaInvalid(input.id, true);
          }
        });
      }
    } catch (error) {
      console.error("Validation error:", error);
      UIManager.displayFormError(form, "Validation error occurred");
    }
  }

  /**
   * Validates the entire form before submission
   */
  static async validateFormSubmission(
    form: HTMLFormElement,
    validationType: string,
  ): Promise<boolean> {
    validation.clearAllErrors();
    UIManager.resetAriaInvalid(form);

    try {
      const result = await validation.validateForm(form, validationType);
      if (!result.success) {
        throw result.error;
      }
      return true;
    } catch (error) {
      console.error("Form validation failed:", error);
      return false;
    }
  }
}
