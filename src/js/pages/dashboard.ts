import { Logger } from "@/core/logger";
import { FormService } from "@/features/forms/services/form-service";

async function deleteForm(formId: string) {
  if (
    !confirm(
      "Are you sure you want to delete this form? This action cannot be undone.",
    )
  ) {
    return;
  }

  try {
    Logger.group("Form Deletion");
    Logger.debug("Starting form deletion for ID:", formId);

    const formService = FormService.getInstance();
    await formService.deleteForm(formId);

    Logger.debug("Form deleted successfully, reloading page");
    Logger.groupEnd();
    window.location.reload();
  } catch (error: unknown) {
    Logger.group("Form Deletion Error");
    Logger.error("Failed to delete form:", error);
    Logger.groupEnd();
    alert(
      error instanceof Error
        ? error.message
        : "Failed to delete form. Please try again.",
    );
  }
}

// Initialize dashboard functionality
function initDashboard() {
  // Handle delete button clicks using event delegation
  document.addEventListener("click", (event) => {
    const target = event.target as HTMLElement;
    const deleteButton = target.closest("button[data-form-id]");
    if (deleteButton) {
      const formId = deleteButton.getAttribute("data-form-id");
      if (formId) {
        deleteForm(formId).catch((error) => {
          Logger.error("Failed to delete form:", error);
        });
      }
    }
  });
}

// Initialize when the DOM is ready
document.addEventListener("DOMContentLoaded", initDashboard);
