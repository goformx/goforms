import { HttpClient } from "../utils/http-client";
import { ErrorManager } from "../utils/error-manager";
import { ValidationManager } from "../utils/validation-manager";
import { Logger } from "../utils/logger";

export interface FormConfig {
  formId: string;
  validationType: string;
  validationDelay?: number;
}

/**
 * Enhanced form handler that consolidates all form handling logic
 */
export class EnhancedFormHandler {
  private form: HTMLFormElement;
  private validationType: string;
  private validationDelay: number;
  private isSubmitting = false;

  constructor(config: FormConfig) {
    this.form = this.findForm(config.formId);
    this.validationType = config.validationType;
    this.validationDelay = config.validationDelay ?? 300;

    this.initialize();
  }

  private findForm(formId: string): HTMLFormElement {
    const form = document.querySelector<HTMLFormElement>(`#${formId}`);
    if (!form) {
      throw new Error(`Form with ID "${formId}" not found`);
    }
    return form;
  }

  private initialize(): void {
    this.setupRealTimeValidation();
    this.setupFormSubmission();
    Logger.debug(`Form handler initialized for: ${this.form.id}`);
  }

  private setupRealTimeValidation(): void {
    const debouncedValidation = this.debounce(
      (input: HTMLInputElement) => this.validateField(input),
      this.validationDelay,
    );

    this.form
      .querySelectorAll<HTMLInputElement>("input[id]")
      .forEach((input) => {
        input.addEventListener("input", () => debouncedValidation(input));

        // Special handling for password field to validate confirm_password
        if (input.name === "password") {
          input.addEventListener("input", () => {
            const confirmInput = this.form.querySelector<HTMLInputElement>(
              "input[name='confirm_password']",
            );
            if (confirmInput && confirmInput.value) {
              debouncedValidation(confirmInput);
            }
          });
        }
      });
  }

  private async validateField(input: HTMLInputElement): Promise<void> {
    ErrorManager.clearFieldError(input.id);

    const result = await ValidationManager.validateField(
      input.name || input.id,
      input.value,
      this.validationType,
      this.form,
    );

    if (!result.valid && result.error) {
      ErrorManager.showFieldError(input.id, result.error);
    }
  }

  private setupFormSubmission(): void {
    this.form.addEventListener("submit", (event) => {
      event.preventDefault();
      if (!this.isSubmitting) {
        this.handleSubmission();
      }
    });
  }

  private async handleSubmission(): Promise<void> {
    if (this.isSubmitting) return;

    this.isSubmitting = true;
    this.setSubmitButtonState(true);
    ErrorManager.clearAllErrors(this.form);

    try {
      // Validate entire form
      const validationResult = await ValidationManager.validateForm(
        this.form,
        this.validationType,
      );

      if (!validationResult.success) {
        this.displayValidationErrors(validationResult.error?.errors || []);
        return;
      }

      // Submit form
      const response = await this.submitForm();
      await this.handleResponse(response);
    } catch (error) {
      Logger.error("Form submission failed:", error);
      ErrorManager.showFormError(
        this.form,
        "An unexpected error occurred. Please try again.",
      );
    } finally {
      this.isSubmitting = false;
      this.setSubmitButtonState(false);
    }
  }

  private setSubmitButtonState(disabled: boolean): void {
    const submitButton = this.form.querySelector<HTMLButtonElement>(
      "button[type='submit']",
    );
    if (submitButton) {
      submitButton.disabled = disabled;
      submitButton.textContent = disabled
        ? "Submitting..."
        : submitButton.dataset.originalText || "Submit";

      if (!submitButton.dataset.originalText) {
        submitButton.dataset.originalText = submitButton.textContent;
      }
    }
  }

  private displayValidationErrors(
    errors: Array<{ path: string[]; message: string }>,
  ): void {
    errors.forEach((error) => {
      const fieldId = error.path[0];
      if (fieldId) {
        // Try to find by ID first, then by name
        let input = document.getElementById(fieldId) as HTMLInputElement | null;
        if (!input) {
          input = this.form.querySelector<HTMLInputElement>(
            `input[name="${fieldId}"]`,
          );
        }

        if (input && input.id) {
          ErrorManager.showFieldError(input.id, error.message);
        }
      }
    });
  }

  private async submitForm(): Promise<Response> {
    const formData = new FormData(this.form);
    Logger.debug(
      "Submitting form data:",
      Object.fromEntries(formData.entries()),
    );

    return HttpClient.post(this.form.action, formData);
  }

  private async handleResponse(response: Response): Promise<void> {
    try {
      const data = await response.json();
      Logger.debug("Response data:", data);

      if (response.redirected || data.redirect) {
        const redirectUrl = response.redirected ? response.url : data.redirect;
        Logger.log("Redirecting to:", redirectUrl);
        window.location.href = redirectUrl;
        return;
      }

      if (!response.ok) {
        ErrorManager.showFormError(
          this.form,
          data.message || "Submission failed",
        );
        return;
      }

      // Handle success
      if (data.message) {
        ErrorManager.showFormSuccess(this.form, data.message);
      }
    } catch (error) {
      Logger.error("Error parsing response:", error);
      ErrorManager.showFormError(this.form, "Error processing server response");
    }
  }

  private debounce<T extends (...args: any[]) => any>(
    fn: T,
    delay: number,
  ): (...args: Parameters<T>) => void {
    let timer: NodeJS.Timeout;
    return (...args: Parameters<T>) => {
      clearTimeout(timer);
      timer = setTimeout(() => fn(...args), delay);
    };
  }

  // Public methods for external use
  public reset(): void {
    this.form.reset();
    ErrorManager.clearAllErrors(this.form);
  }

  public getFormData(): FormData {
    return new FormData(this.form);
  }

  public isValid(): Promise<boolean> {
    return ValidationManager.validateForm(this.form, this.validationType).then(
      (result) => result.success,
    );
  }
}
