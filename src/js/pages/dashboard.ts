import { Logger } from "@/core/logger";
import { FormService } from "@/features/forms/services/form-service";

// Initialize dashboard functionality
function initDashboard() {
  // Handle delete button clicks using event delegation
  document.addEventListener("click", (event) => {
    const target = event.target as HTMLElement;
    const deleteButton = target.closest("button[data-form-id]");
    if (deleteButton) {
      const formId = deleteButton.getAttribute("data-form-id");
      if (formId) {
        // Delegate to the form service for the actual deletion logic
        const formService = FormService.getInstance();
        formService.deleteFormWithConfirmation(formId).catch((error) => {
          Logger.error("Failed to delete form:", error);
        });
      }
    }
  });
}

// Initialize when the DOM is ready
document.addEventListener("DOMContentLoaded", initDashboard);
