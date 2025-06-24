import { Formio } from "@formio/js";
import goforms from "@goformx/formio";
import { FormService } from "../forms/services/form-service";
import { setupBuilderEvents } from "../forms/handlers/builder-events";

// Import Form.io styles
import "@formio/js/dist/formio.full.min.css";

// Import our modules
import { FormBuilderError } from "../errors/form-builder-error";
import { dom } from "../utils/dom-utils";
import {
  validateFormBuilder,
  getFormSchema,
  createFormBuilder,
} from "../core/form-builder-core";
import { setupEventHandlers } from "../handlers/form-builder-events";

// Register templates
Formio.use(goforms);

/**
 * Main initialization function
 */
async function initializeFormBuilder(): Promise<void> {
  try {
    // Validate and get required elements
    const { builder: container, formId } = validateFormBuilder();

    // Get schema and create builder
    const schema = await getFormSchema(formId);
    const builder = await createFormBuilder(container, schema);

    // Set up event handlers
    setupEventHandlers(builder, formId);
    setupBuilderEvents(builder, formId, FormService.getInstance());
  } catch (error) {
    if (error instanceof FormBuilderError) {
      dom.showError(error.userMessage);
    } else {
      dom.showError("An unexpected error occurred. Please refresh the page.");
    }
    throw error;
  }
}

// Initialize the form builder
initializeFormBuilder();
