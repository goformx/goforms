import { Formio } from "@formio/js";
import goforms from "@goformx/formio";
import { FormService } from "../forms/services/form-service";
import type { FormSchema } from "../forms/services/form-service";
import { builderOptions } from "../utils/constants/builder-config";
import { setupBuilderEvents } from "../forms/handlers/builder-events";

// Import Form.io styles
import "@formio/js/dist/formio.full.min.css";

// Register templates
Formio.use(goforms);

/**
 * Form builder error handling
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
 * DOM utilities
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
async function getFormSchema(formId: string): Promise<FormSchema> {
  // For new form creation, return a default schema
  if (formId === "new") {
    return {
      display: "form",
      components: [],
    };
  }

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
  schema: FormSchema,
): Promise<any> {
  try {
    // Initialize Formio with project settings
    Formio.setProjectUrl("https://goforms.io");
    Formio.setBaseUrl("https://goforms.io");

    // Ensure schema has required properties
    const formSchema = {
      ...schema,
      projectId: "goforms",
      display: "form",
      components: schema.components || [],
    };

    // Create builder with options
    const builder = await Formio.builder(container, formSchema, {
      ...builderOptions,
      noDefaultSubmitButton: true,
      builder: {
        ...builderOptions.builder,
        basic: {
          components: {
            textfield: true,
            textarea: true,
            email: true,
            phoneNumber: true,
            number: true,
            password: true,
            checkbox: true,
            selectboxes: true,
            select: true,
            radio: true,
            button: true,
          },
        },
      },
    });

    // Store builder instance globally
    window.formBuilder = builder;
    return builder;
  } catch (error) {
    console.error("Form builder initialization error:", error);
    throw new FormBuilderError(
      "Failed to initialize builder",
      "Failed to initialize form builder. Please refresh the page.",
    );
  }
}

/**
 * Event handlers setup
 */
function setupEventHandlers(builder: any, formId: string): void {
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

        // Get the current schema using saveSchema
        const schema = await builder.saveSchema();
        if (!schema) {
          throw new Error("Failed to get form schema");
        }

        // Save using form service
        const formService = FormService.getInstance();
        await formService.saveSchema(formId, schema);

        feedback.textContent = "Schema saved successfully.";
        feedback.className = "schema-save-feedback success";
      } catch (error) {
        console.error("Failed to save form fields:", error);
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

  // For new form creation, update the hidden schema field before form submission
  if (formId === "new") {
    const form = dom.getElement<HTMLFormElement>("new-form");
    if (form) {
      form.addEventListener("submit", async (e) => {
        e.preventDefault();
        try {
          const schema = await builder.saveSchema();
          const schemaInput = dom.getElement<HTMLInputElement>("schema");
          if (schemaInput) {
            schemaInput.value = JSON.stringify(schema);
          }
          form.submit();
        } catch (_error) {
          dom.showError("Failed to save form schema. Please try again.");
        }
      });
    }
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
