import { Formio } from "@formio/js";
import goforms from "@goformx/formio";
import { FormService } from "@/features/forms/services/form-service";
import { setupBuilderEvents } from "@/features/forms/handlers/builder-events";
import { setupViewSchemaButton } from "@/features/forms/components/form-builder/view-schema-button";
import { setupSaveFieldsButton } from "@/features/forms/components/form-builder/save-fields-button";
import { initializeSidebar } from "@/features/forms/components/form-builder/sidebar";

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
    Logger.group("Form Builder Initialization");
    Logger.debug("Starting form builder initialization...");

    // Validate and get required elements
    Logger.group("Element Validation");
    const { builder: container, formId } = validateFormBuilder();
    Logger.debug("Form builder validation passed:", {
      formId,
      containerExists: !!container,
    });
    Logger.groupEnd();

    // Get schema and create builder
    Logger.group("Schema & Builder Setup");
    const schema = await getFormSchema(formId);
    Logger.debug("Schema fetched successfully");

    const builder = await createFormBuilder(container, schema);
    Logger.debug("Form.io builder created successfully");
    Logger.groupEnd();

    // Set up event handlers
    Logger.group("Event Handler Setup");
    setupBuilderEvents(builder, formId, FormService.getInstance());

    // Set up View Schema button
    setupViewSchemaButton(builder);

    // Set up Save Fields button
    setupSaveFieldsButton(formId);

    // Initialize sidebar
    const sidebar = initializeSidebar({
      enableSwipeGestures: true,
      enableKeyboardShortcuts: true,
      persistState: true,
      onToggle: (isOpen) => {
        Logger.debug(`Sidebar ${isOpen ? "opened" : "closed"}`);
      },
      onResize: (viewport) => {
        Logger.debug(`Viewport changed to: ${viewport}`);
      },
    });

    if (sidebar) {
      Logger.debug("Sidebar initialized successfully");
    }
    Logger.groupEnd();

    Logger.debug("Form builder initialization completed successfully");
    Logger.groupEnd();
  } catch (error) {
    Logger.error("Form builder initialization failed:", error);
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

// Also try immediate initialization if DOM is already ready
if (document.readyState === "loading") {
  // DOM still loading, waiting for DOMContentLoaded
} else {
  initializeFormBuilder().catch((error) => {
    Logger.error("Failed to initialize form builder (immediate):", error);
  });
}
