import { FormService } from "../services/form-service";

async function deleteForm(formId: string) {
  if (
    !confirm(
      "Are you sure you want to delete this form? This action cannot be undone.",
    )
  ) {
    return;
  }

  try {
    const formService = FormService.getInstance();
    await formService.deleteForm(formId);
    window.location.reload();
  } catch (error: unknown) {
    console.error("Failed to delete form:", error);
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
        deleteForm(formId);
      }
    }
  });
}

// Initialize when the DOM is ready
document.addEventListener("DOMContentLoaded", initDashboard);
