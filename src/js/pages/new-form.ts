/**
 * New Form Handler
 *
 * Handles form creation with proper validation and redirect after success.
 */

import { Logger } from "@/core/logger";
import { FormService } from "@/features/forms/services/form-service";

Logger.debug("new-form.ts: Script loaded and executing");

// Initialize the new form handler
document.addEventListener("DOMContentLoaded", () => {
  Logger.debug("new-form.ts: DOMContentLoaded event fired");

  const form = document.getElementById("new-form") as HTMLFormElement;
  if (!form) {
    Logger.error("new-form.ts: Form element not found");
    return;
  }

  // Handle form submission
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    Logger.debug("new-form.ts: Form submission intercepted");

    const submitButton = form.querySelector(
      'button[type="submit"]',
    ) as HTMLButtonElement;
    const originalText = submitButton.innerHTML;

    try {
      // Disable submit button and show loading state
      submitButton.disabled = true;
      submitButton.innerHTML = "Creating...";

      // Get form data
      const formData = new FormData(form);
      const title = formData.get("title") as string;
      const description = (formData.get("description") as string) || "";

      // Validate required fields
      if (!title || title.trim() === "") {
        alert("Form title is required");
        return;
      }

      Logger.debug("new-form.ts: Creating form with title:", title);

      // Use the form service to create and redirect
      const formService = FormService.getInstance();
      await formService.createFormWithRedirect(title, description);

      Logger.debug("new-form.ts: Form creation and redirect completed");
    } catch (error) {
      Logger.error("new-form.ts: Form creation error:", error);

      // Show error message
      alert(
        error instanceof Error
          ? error.message
          : "Failed to create form. Please try again.",
      );
    } finally {
      // Restore submit button
      submitButton.disabled = false;
      submitButton.innerHTML = originalText;
    }
  });

  Logger.debug("new-form.ts: Form handler initialized successfully");
});

Logger.debug("new-form.ts: Script execution completed");
