// ===== src/js/forms/handlers/form-handler.ts =====
import { Logger } from "@/core/logger";
import type { FormConfig } from "@/shared/types/form-types";
import { validation } from "@/features/forms/validation/validation";
import { ValidationHandler } from "@/features/forms/handlers/validation-handler";
import { RequestHandler } from "@/features/forms/handlers/request-handler";
import { ResponseHandler } from "@/features/forms/handlers/response-handler";
import { UIManager } from "@/features/forms/handlers/ui-manager";

/**
 * Sets up a form with validation and submission handling
 */
export function setupForm(config: FormConfig): void {
  const form = document.querySelector<HTMLFormElement>(`#${config.formId}`);
  if (!form) {
    Logger.error(`Form with ID "${config.formId}" not found`);
    return;
  }

  validation.setupRealTimeValidation(form.id, config.formId);
  ValidationHandler.setupRealTimeValidation(form, config.validationDelay);

  form.addEventListener("submit", (event) =>
    handleFormSubmission(event, form, config.formId),
  );
}

/**
 * Handles form submission including validation and server communication
 */
async function handleFormSubmission(
  event: Event,
  form: HTMLFormElement,
  schemaName: string,
): Promise<void> {
  event.preventDefault();

  try {
    const isValid = await ValidationHandler.validateFormSubmission(
      form,
      schemaName,
    );

    if (!isValid) {
      UIManager.displayFormError(form, "Please check the form for errors.");
      return;
    }

    const response = await RequestHandler.sendFormData(form);
    await ResponseHandler.handleServerResponse(response, form);
  } catch (error) {
    Logger.error("Form submission error:", error);
    UIManager.displayFormError(form, "An error occurred. Please try again.");
  }
}
