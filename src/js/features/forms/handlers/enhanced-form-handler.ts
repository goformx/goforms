// ===== src/js/forms/handlers/enhanced-form-handler.ts =====
import { Logger } from "@/core/logger";
import type { FormConfig } from "@/shared/types/form-types";
import { validation } from "@/features/forms/validation/validation";
import { ValidationHandler } from "./validation-handler";
import { ResponseHandler } from "./response-handler";
import { RequestHandler } from "./request-handler";
import { UIManager } from "./ui-manager";

/**
 * Class-based form handler for more complex use cases
 */
export class EnhancedFormHandler {
  private readonly form: HTMLFormElement;
  private readonly formId: string;
  private readonly config: FormConfig;

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
    this.config = config;

    this.initialize();
  }

  private initialize(): void {
    // Set up real-time validation based on validation type
    if (
      this.config.validationType === "realtime" ||
      this.config.validationType === "hybrid"
    ) {
      ValidationHandler.setupRealTimeValidation(
        this.form,
        this.config.validationDelay,
      );
    }

    // Set up schema-based validation if needed
    validation.setupRealTimeValidation(this.form.id, this.formId);

    this.form.addEventListener("submit", (event) =>
      this.handleFormSubmission(event),
    );
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
      const response = await RequestHandler.sendFormData(this.form);
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
