import { HttpClient } from "@/core/http-client";
import { Logger } from "@/core/logger";

/**
 * Initialize the new form page
 */
function initializeNewForm(): void {
  Logger.debug("Initializing new form page");

  const form = document.getElementById("new-form") as HTMLFormElement;
  if (!form) {
    Logger.error("New form not found");
    return;
  }

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    await handleFormSubmission(form);
  });

  Logger.debug("New form handler initialized successfully");
}

/**
 * Handle form submission with CSRF token support
 */
async function handleFormSubmission(form: HTMLFormElement): Promise<void> {
  try {
    Logger.debug("Submitting new form");

    const formData = new FormData(form);

    // Validate required fields
    const title = formData.get("title") as string;
    if (!title || title.trim() === "") {
      showError("Form title is required");
      return;
    }

    const result = await HttpClient.post("/forms", formData);

    if (result && result.form_id) {
      // Redirect to the form edit page
      window.location.href = `/forms/${result.form_id}/edit`;
    } else {
      window.location.href = "/forms";
    }
  } catch (error) {
    Logger.error("Form submission error:", error);
    showError("An unexpected error occurred. Please try again.");
  }
}

/**
 * Show error message
 */
function showError(message: string): void {
  const errorContainer = document.querySelector(".form-error") as HTMLElement;
  if (errorContainer) {
    errorContainer.textContent = message;
    errorContainer.style.display = "block";
  } else {
    alert(message);
  }
}

// Initialize when DOM is ready
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initializeNewForm);
} else {
  initializeNewForm();
}
