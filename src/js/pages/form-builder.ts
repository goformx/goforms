import { Formio } from "@formio/js";
import goforms from "@goformx/formio";
import { FormService } from "@/features/forms/services/form-service";
import { setupBuilderEvents } from "@/features/forms/handlers/builder-events";
import { setupViewSchemaButton } from "@/features/forms/components/form-builder/view-schema-button";
import { setupSaveFieldsButton } from "@/features/forms/components/form-builder/save-fields-button";

// Import Form.io styles
import "@formio/js/dist/formio.full.min.css";

// Import our modules
import { FormBuilderError } from "@/core/errors/form-builder-error";
import { Logger } from "@/core/logger";
import { dom } from "@/shared/utils/dom-utils";
import {
  validateFormBuilder,
  getFormSchema,
  createFormBuilder,
} from "@/features/forms/components/form-builder/core";

// Register templates
Formio.use(goforms);

/**
 * Main initialization function
 */
async function initializeFormBuilder(): Promise<void> {
  try {
    console.log("Starting form builder initialization...");

    // Validate and get required elements
    const { builder: container, formId } = validateFormBuilder();
    console.log("Form builder validation passed:", {
      formId,
      containerExists: !!container,
    });

    // Get schema and create builder
    console.log("Fetching form schema...");
    const schema = await getFormSchema(formId);
    console.log("Schema fetched:", schema);

    console.log("Creating Form.io builder...");
    const builder = await createFormBuilder(container, schema);
    console.log("Form.io builder created successfully");

    // Set up event handlers
    console.log("Setting up event handlers...");
    setupBuilderEvents(builder, formId, FormService.getInstance());

    // Set up View Schema button
    console.log("Setting up View Schema button...");
    setupViewSchemaButton(builder);

    // Set up Save Fields button
    console.log("Setting up Save Fields button...");
    setupSaveFieldsButton(formId);

    console.log("Form builder initialization completed successfully");
  } catch (error) {
    console.error("Form builder initialization failed:", error);
    if (error instanceof FormBuilderError) {
      dom.showError(error.userMessage);
    } else {
      dom.showError("An unexpected error occurred. Please refresh the page.");
    }
    throw error;
  }
}

// Initialize when DOM is ready
document.addEventListener("DOMContentLoaded", () => {
  initializeFormBuilder().catch((error) => {
    Logger.error("Failed to initialize form builder:", error);
  });
});
