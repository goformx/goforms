import { Formio } from "@formio/js";
import goforms from "@goforms/formio";
import { FormService } from "./services/form-service";
import { builderOptions } from "./constants/builder-config";
import { setupBuilderEvents } from "./handlers/builder-events";

// Import Form.io styles
// import "@formio/js/dist/formio.full.min.css";

// Register templates
Formio.use(goforms);

/**
 * Error handling utility
 */
class FormBuilderError extends Error {
  constructor(
    message: string,
    public readonly userMessage: string,
  ) {
    super(message);
    this.name = "FormBuilderError";
  }
}

/**
 * DOM utility functions
 */
const dom = {
  getElement<T extends HTMLElement>(id: string): T | null {
    return document.getElementById(id) as T | null;
  },

  createElement<T extends HTMLElement>(tag: string, className?: string): T {
    const element = document.createElement(tag) as T;
    if (className) element.className = className;
    return element;
  },

  showError(message: string, container?: HTMLElement): void {
    const errorContainer =
      container?.querySelector(".gf-error-message") ||
      document.querySelector(".gf-error-message");

    if (errorContainer instanceof HTMLElement) {
      errorContainer.textContent = message;
      errorContainer.style.display = "block";
      return;
    }

    const errorDiv = dom.createElement<HTMLDivElement>(
      "div",
      "gf-error-message",
    );
    errorDiv.textContent = message;
    document.body.insertBefore(errorDiv, document.body.firstChild);
  },
};

/**
 * Form builder validation
 */
function validateFormBuilder(): { builder: HTMLElement; formId: string } {
  const builder = dom.getElement<HTMLElement>("form-schema-builder");
  if (!builder) {
    throw new FormBuilderError(
      "Form builder element not found",
      "Form builder element not found. Please refresh the page.",
    );
  }

  const formId = builder.getAttribute("data-form-id");
  if (!formId) {
    throw new FormBuilderError(
      "Form ID not found",
      "Form ID not found. Please refresh the page.",
    );
  }

  return { builder, formId };
}

/**
 * Schema management
 */
async function getFormSchema(formId: string): Promise<any> {
  const formService = FormService.getInstance();
  try {
    return await formService.getSchema(formId);
  } catch {
    throw new FormBuilderError(
      "Failed to fetch schema",
      "Failed to load form schema. Please try again later.",
    );
  }
}

/**
 * Builder initialization
 */
async function createFormBuilder(
  container: HTMLElement,
  schema: any,
): Promise<any> {
  try {
    return await Formio.builder(container, schema, builderOptions);
  } catch {
    throw new FormBuilderError(
      "Failed to initialize builder",
      "Failed to initialize form builder. Please refresh the page.",
    );
  }
}

/**
 * Event handlers setup
 */
function setupEventHandlers(builder: any): void {
  const viewSchemaBtn = dom.getElement<HTMLButtonElement>("view-schema-btn");
  if (viewSchemaBtn) {
    viewSchemaBtn.addEventListener("click", () => {
      // Show the JSON editor with the current schema
      builder.showJSON();
    });
  }

  const saveBtn = dom.getElement<HTMLButtonElement>("save-fields-btn");
  const feedback = dom.getElement<HTMLElement>("schema-save-feedback");
  if (saveBtn && feedback) {
    saveBtn.addEventListener("click", async () => {
      const spinner = saveBtn.querySelector(".spinner") as HTMLElement;
      try {
        feedback.textContent = "Saving...";
        feedback.className = "schema-save-feedback";
        saveBtn.disabled = true;
        if (spinner) spinner.style.display = "inline-block";

        // Save the schema and get the response
        const savedSchema = await builder.saveSchema();

        // Check if we got a valid schema response
        if (
          savedSchema &&
          savedSchema.schema &&
          savedSchema.schema.components
        ) {
          feedback.textContent = "Schema saved successfully.";
          feedback.className = "schema-save-feedback success";
        } else {
          feedback.textContent = "Failed to save schema - invalid response.";
          feedback.className = "schema-save-feedback error";
        }
      } catch (error) {
        feedback.textContent =
          error instanceof Error ? error.message : "Error saving schema.";
        feedback.className = "schema-save-feedback error";
      } finally {
        saveBtn.disabled = false;
        if (spinner) spinner.style.display = "none";
        setTimeout(() => {
          feedback.textContent = "";
          feedback.className = "schema-save-feedback";
        }, 3000);
      }
    });
  }
}

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
    setupEventHandlers(builder);
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
