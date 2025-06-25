import { Logger } from "@/core/logger";

export class FormUIService {
  private readonly form: HTMLFormElement;
  private readonly errorContainer: HTMLElement | null;
  private readonly successContainer: HTMLElement | null;

  constructor(form: HTMLFormElement) {
    this.form = form;
    this.errorContainer = form.querySelector(".form-error");
    this.successContainer = form.querySelector(".form-success");
  }

  showError(message: string, fieldName?: string): void {
    Logger.debug(`Showing error: ${message}`);

    if (this.errorContainer) {
      this.errorContainer.textContent = message;
      this.errorContainer.classList.remove("hidden");
    } else {
      Logger.warn(`Error container not found in form: ${this.form.id}`);
    }

    if (fieldName) {
      this.highlightFieldError(fieldName);
    }
  }

  showSuccess(message: string): void {
    Logger.debug(`Showing success: ${message}`);

    if (this.successContainer) {
      this.successContainer.textContent = message;
      this.successContainer.classList.remove("hidden");
    } else {
      Logger.warn(`Success container not found in form: ${this.form.id}`);
    }

    this.clearErrors();
  }

  clearMessages(): void {
    this.clearErrors();
    this.clearSuccess();
  }

  private clearErrors(): void {
    if (this.errorContainer) {
      this.errorContainer.classList.add("hidden");
    }
    this.resetFieldErrors();
  }

  private clearSuccess(): void {
    if (this.successContainer) {
      this.successContainer.classList.add("hidden");
    }
  }

  private highlightFieldError(fieldName: string): void {
    const field = this.form.querySelector(
      `[name="${fieldName}"]`,
    ) as HTMLElement;
    if (field) {
      field.classList.add("error");
      field.setAttribute("aria-invalid", "true");
    }
  }

  private resetFieldErrors(): void {
    const fields = this.form.querySelectorAll<HTMLElement>(
      "input[id], textarea[id], select[id]",
    );
    fields.forEach((field) => {
      field.classList.remove("error");
      field.setAttribute("aria-invalid", "false");
    });
  }

  setLoading(loading: boolean): void {
    const submitButton = this.form.querySelector<HTMLButtonElement>(
      'button[type="submit"]',
    );
    if (submitButton) {
      submitButton.disabled = loading;
      submitButton.textContent = loading ? "Submitting..." : "Submit";
    }
  }
}
