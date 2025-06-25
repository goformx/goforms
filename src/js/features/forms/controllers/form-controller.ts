import { Logger } from "@/core/logger";
import type { FormConfig } from "@/shared/types";
import { validation } from "@/features/forms/validation/validation";
import { ValidationHandler } from "../handlers/validation-handler";
import { RequestHandler } from "../handlers/request-handler";
import { ResponseHandler } from "../handlers/response-handler";
import { FormUIService } from "../services/form-ui-service";

export class FormController {
  private readonly form: HTMLFormElement;
  private readonly config: FormConfig;
  private readonly uiService: FormUIService;

  constructor(config: FormConfig) {
    this.config = config;
    this.form = this.getFormElement(config.formId);
    this.uiService = new FormUIService(this.form);

    this.initialize();
  }

  private getFormElement(formId: string): HTMLFormElement {
    const form = document.querySelector<HTMLFormElement>(`#${formId}`);
    if (!form) {
      throw new Error(`Form with ID "${formId}" not found`);
    }
    return form;
  }

  private initialize(): void {
    this.setupValidation();
    this.setupSubmissionHandler();

    Logger.debug(`FormController initialized for form: ${this.config.formId}`);
  }

  private setupValidation(): void {
    // Setup real-time validation if configured
    if (this.shouldEnableRealTimeValidation()) {
      ValidationHandler.setupRealTimeValidation(
        this.form,
        this.config.validationDelay,
      );
    }

    // Setup schema-based validation
    validation.setupRealTimeValidation(this.form.id, this.config.formId);
  }

  private shouldEnableRealTimeValidation(): boolean {
    return (
      this.config.validationType === "realtime" ||
      this.config.validationType === "hybrid"
    );
  }

  private setupSubmissionHandler(): void {
    this.form.addEventListener("submit", (event) =>
      this.handleSubmission(event),
    );
  }

  private async handleSubmission(event: Event): Promise<void> {
    event.preventDefault();

    try {
      await this.processSubmission();
    } catch (error) {
      this.handleSubmissionError(error);
    }
  }

  private async processSubmission(): Promise<void> {
    Logger.debug(`Processing submission for form: ${this.config.formId}`);

    // Set loading state
    this.uiService.setLoading(true);

    try {
      // Validate form
      const isValid = await ValidationHandler.validateFormSubmission(
        this.form,
        this.config.formId,
      );

      if (!isValid) {
        this.uiService.showError("Please check the form for errors.");
        return;
      }

      // Submit form data
      const response = await RequestHandler.sendFormData(this.form);
      await ResponseHandler.handleServerResponse(response, this.form);
    } finally {
      this.uiService.setLoading(false);
    }
  }

  private handleSubmissionError(error: unknown): void {
    Logger.error(`Form submission error for ${this.config.formId}:`, error);
    this.uiService.showError("An unexpected error occurred. Please try again.");
  }

  // Public API for external control
  public async submitForm(): Promise<void> {
    await this.processSubmission();
  }

  public reset(): void {
    this.form.reset();
    this.uiService.clearMessages();
  }

  public destroy(): void {
    // Cleanup event listeners if needed
    this.form.removeEventListener("submit", this.handleSubmission);
  }
}
