import { Formio } from "@formio/js";
import goforms from "goforms";
import { FormService } from "./services/form-service";
import { builderOptions } from "./constants/builder-config";
import { setupBuilderEvents } from "./handlers/builder-events";

// Import Form.io styles
import "@formio/js/dist/formio.full.min.css";

// Register templates
Formio.use(goforms);

/**
 * Displays an error message in a consistent way across the application
 * @param message The error message to display
 * @param container Optional container element to show the error in
 */
function displayErrorMessage(message: string, container?: HTMLElement): void {
  // Try to find existing error container first
  const errorContainer =
    container?.querySelector(".gf-error-message") ||
    document.querySelector(".gf-error-message");

  if (errorContainer instanceof HTMLElement) {
    errorContainer.textContent = message;
    errorContainer.style.display = "block";
    return;
  }

  // Create new error element if none exists
  const errorDiv = document.createElement("div");
  errorDiv.className = "gf-error-message";
  errorDiv.textContent = message;

  // Add to the top of the page for better visibility
  document.body.insertBefore(errorDiv, document.body.firstChild);
}

// Initialize form builder when the module is loaded
const formSchemaBuilder = document.getElementById("form-schema-builder");

if (!formSchemaBuilder) {
  displayErrorMessage(
    "Form builder element not found. Please refresh the page.",
  );
  throw new Error("Form schema builder element not found");
}

const formIdAttr = formSchemaBuilder.getAttribute("data-form-id");
if (!formIdAttr) {
  displayErrorMessage("Form ID not found. Please refresh the page.");
  throw new Error("Form ID not found");
}

// More robust formId parsing
const formId = parseInt(formIdAttr, 10);
if (!Number.isInteger(formId) || formId <= 0) {
  displayErrorMessage("Invalid form ID. Please refresh the page.");
  throw new Error(`Invalid form ID: ${formIdAttr}`);
}

let builder: any;

async function initializeFormBuilder(formId: number): Promise<void> {
  const formService = FormService.getInstance();

  try {
    const schema = await formService.getSchema(formId);
    builder = await Formio.builder(formSchemaBuilder, schema, builderOptions);
    setupBuilderEvents(builder, formId, formService);

    // Add schema view button handler
    const viewSchemaBtn = document.getElementById("view-schema-btn");
    if (viewSchemaBtn) {
      viewSchemaBtn.addEventListener("click", () => {
        if (builder) {
          builder.showSchema();
        }
      });
    }
  } catch (error) {
    console.error("Error initializing form builder:", error);
    // Display user-friendly error message
    if (formSchemaBuilder instanceof HTMLElement) {
      displayErrorMessage(
        "Failed to load form builder. Please refresh the page or try again later.",
        formSchemaBuilder,
      );
    } else {
      displayErrorMessage(
        "Failed to load form builder. Please refresh the page or try again later.",
      );
    }
    throw error; // Re-throw for logging purposes
  }
}

// Initialize the form builder
initializeFormBuilder(formId);
