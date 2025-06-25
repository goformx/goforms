import { Logger } from "@/core/logger";
import { FormService } from "@/features/forms/services/form-service";
import { formState } from "@/features/forms/state/form-state";
import { FormBuilderError, ErrorCode } from "@/core/errors/form-builder-error";

/**
 * Set up the Save Fields button event handler
 */
export function setupSaveFieldsButton(formId: string): void {
  const saveFieldsBtn = document.getElementById(
    "save-fields-btn",
  ) as HTMLButtonElement;
  const feedbackSpan = document.getElementById(
    "schema-save-feedback",
  ) as HTMLSpanElement;

  if (!saveFieldsBtn) {
    Logger.warn("Save Fields button not found");
    return;
  }

  if (!feedbackSpan) {
    Logger.warn("Schema save feedback element not found");
    return;
  }

  saveFieldsBtn.addEventListener("click", async () => {
    try {
      // Get the current builder instance
      const builder = formState.get("formBuilder") as any;
      if (!builder) {
        throw new FormBuilderError(
          "Form builder not found",
          ErrorCode.FORM_NOT_FOUND,
          "Form builder not found. Please refresh the page.",
        );
      }

      // Show loading state
      const spinner = saveFieldsBtn.querySelector(".spinner") as HTMLElement;
      const buttonText = saveFieldsBtn.querySelector(
        "span:not(.spinner)",
      ) as HTMLElement;

      if (spinner) spinner.style.display = "inline-block";
      if (buttonText) buttonText.textContent = "Saving...";
      saveFieldsBtn.disabled = true;

      // Clear previous feedback
      feedbackSpan.textContent = "";
      feedbackSpan.className = "";

      // Get current schema from builder
      const currentSchema = builder.form;

      // Save schema using form service
      const formService = FormService.getInstance();
      await formService.saveSchema(formId, currentSchema);

      // Show success feedback
      feedbackSpan.textContent = "Fields saved successfully!";
      feedbackSpan.className = "success";
      feedbackSpan.style.color = "#28a745";

      Logger.debug("Form schema saved successfully:", formId);

      // Clear success message after 3 seconds
      setTimeout(() => {
        feedbackSpan.textContent = "";
        feedbackSpan.className = "";
      }, 3000);
    } catch (error) {
      Logger.error("Error saving form schema:", error);

      // Show error feedback
      feedbackSpan.textContent =
        error instanceof FormBuilderError
          ? error.userMessage
          : "Failed to save fields. Please try again.";
      feedbackSpan.className = "error";
      feedbackSpan.style.color = "#dc3545";

      // Clear error message after 5 seconds
      setTimeout(() => {
        feedbackSpan.textContent = "";
        feedbackSpan.className = "";
      }, 5000);
    } finally {
      // Reset button state
      const spinner = saveFieldsBtn.querySelector(".spinner") as HTMLElement;
      const buttonText = saveFieldsBtn.querySelector(
        "span:not(.spinner)",
      ) as HTMLElement;

      if (spinner) spinner.style.display = "none";
      if (buttonText) buttonText.textContent = "Save Fields";
      saveFieldsBtn.disabled = false;
    }
  });

  Logger.debug("Save Fields button event handler set up");
}
