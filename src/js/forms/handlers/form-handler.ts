// ===== src/js/forms/handlers/form-handler.ts =====
import type { FormConfig } from "../types/form-types";
import { validation } from "../validation/validation";
import { ValidationHandler } from "../handlers/validation-handler";
import { RequestHandler } from "../handlers/request-handler";
import { ResponseHandler } from "../handlers/response-handler";
import { UIManager } from "../handlers/ui-manager";

/**
 * Sets up a form with validation and submission handling
 */
export function setupForm(config: FormConfig): void {
  const form = document.querySelector<HTMLFormElement>(`#${config.formId}`);
  if (!form) {
    console.error(`Form with ID "${config.formId}" not found`);
    return;
  }

  validation.setupRealTimeValidation(form.id, config.validationType);
  ValidationHandler.setupRealTimeValidation(
    form,
    config.validationType,
    config.validationDelay,
  );

  form.addEventListener("submit", (event) =>
    handleFormSubmission(event, form, config.validationType),
  );
}

/**
 * Handles form submission including validation and server communication
 */
async function handleFormSubmission(
  event: Event,
  form: HTMLFormElement,
  validationType: string,
): Promise<void> {
  event.preventDefault();

  try {
    const isValid = await ValidationHandler.validateFormSubmission(
      form,
      validationType,
    );

    if (!isValid) {
      UIManager.displayFormError(form, "Please check the form for errors.");
      return;
    }

    const response = await RequestHandler.sendFormData(form);
    await ResponseHandler.handleServerResponse(response, form);
  } catch (error) {
    console.error("Form submission error:", error);
    UIManager.displayFormError(form, "An error occurred. Please try again.");
  }
}
