import { Formio } from "@formio/js";
import goforms from "@goformx/formio";
import { FormService } from "@/features/forms/services/form-service";

// Import Form.io styles
import "@formio/js/dist/formio.full.min.css";

// Import our modules
import { FormBuilderError } from "@/core/errors/form-builder-error";
import { dom } from "@/shared/utils/dom-utils";
import { Logger } from "@/core/logger";

// Register templates
Formio.use(goforms);

// Types
interface FormPreviewData {
  formSchema: any;
  formId: string;
}

/**
 * Initialize form preview
 */
async function initializeFormPreview(): Promise<void> {
  try {
    Logger.group("Form Preview Initialization");
    Logger.debug("Initializing form preview...");

    // Get the form renderer container
    Logger.group("Container & Data Setup");
    const container = document.getElementById("form-renderer");
    if (!container) {
      throw new Error("Form renderer container not found");
    }

    // Get schema and form ID from data attributes
    const schemaAttr = container.getAttribute("data-form-schema");
    const formIdAttr = container.getAttribute("data-form-id");

    if (!schemaAttr || !formIdAttr) {
      throw new Error("Form schema or form ID not found in data attributes");
    }

    // Parse the schema
    let formSchema: any;
    try {
      formSchema = JSON.parse(schemaAttr);
    } catch (_error) {
      throw new Error("Invalid form schema format");
    }

    const data: FormPreviewData = {
      formSchema,
      formId: formIdAttr,
    };

    Logger.debug("Form preview data:", data);
    Logger.groupEnd();

    // Clear loading state
    container.innerHTML = "";

    // Create form instance
    Logger.group("Form.io Form Creation");
    const form = await Formio.createForm(container, data.formSchema, {
      readOnly: false,
      noAlerts: true,
      noDefaultSubmitButton: false,
      submitButton: {
        label: "Submit Form",
        className: "btn btn-primary",
      },
    });
    Logger.groupEnd();

    // Handle form submission
    Logger.group("Event Handler Setup");
    form.on("submit", async (submission: any) => {
      try {
        Logger.group("Form Submission");
        Logger.debug("Form submission:", submission);

        // Submit to the API
        await FormService.getInstance().submitForm(
          data.formId,
          submission.data,
        );

        // Show success message
        dom.showSuccess("Form submitted successfully!");

        // Reset form
        form.reset();
        Logger.groupEnd();
      } catch (error) {
        Logger.group("Submission Error");
        Logger.error("Form submission error:", error);
        Logger.groupEnd();
        dom.showError("Failed to submit form. Please try again.");
      }
    });

    // Handle form errors
    form.on("error", (error: any) => {
      Logger.error("Form error:", error);
      dom.showError("An error occurred while loading the form.");
    });
    Logger.groupEnd();

    Logger.debug("Form preview initialized successfully");
    Logger.groupEnd();
  } catch (error) {
    Logger.group("Form Preview Error");
    Logger.error("Form preview initialization error:", error);
    Logger.groupEnd();

    if (error instanceof FormBuilderError) {
      dom.showError(error.userMessage);
    } else {
      dom.showError(
        "An unexpected error occurred while loading the form preview.",
      );
    }

    // Show error state in the container
    const container = document.getElementById("form-renderer");
    if (container) {
      container.innerHTML = `
        <div class="form-error-state">
          <i class="bi bi-exclamation-triangle"></i>
          <h3>Error Loading Form</h3>
          <p>Unable to load the form preview. Please try refreshing the page.</p>
        </div>
      `;
    }
  }
}

// Initialize when DOM is ready
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", initializeFormPreview);
} else {
  initializeFormPreview();
}
