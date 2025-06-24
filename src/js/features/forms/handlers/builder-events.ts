import { Formio } from "@formio/js";
import { FormService } from "../services/form-service";
import type { FormSchema } from "../services/form-service";
import { debounce } from "lodash";
import { showSchemaModal } from "../components/form-builder/schema-modal";
import { dom } from "@/shared/utils/dom-utils";
import { formState } from "../state/form-state";

interface FormBuilderWithSchema extends Formio {
  form: FormSchema;
  saveSchema: () => Promise<FormSchema>;
}

// Define event handlers map type
type EventHandlerMap = {
  [key: string]: (builder: FormBuilderWithSchema) => void;
};

// Define event handlers map as a constant
const EVENT_MAP: EventHandlerMap = {
  change: debounce((_builder: FormBuilderWithSchema) => {
    // Only log changes, no automatic saving
  }, 300),
  saveComponent: (_builder: FormBuilderWithSchema) => {
    // Only log component saves, no automatic saving
  },
};

export const setupBuilderEvents = (
  builder: Formio,
  formId: string,
  formService: FormService,
): void => {
  const typedBuilder = builder as FormBuilderWithSchema;

  // Create a promise that will be resolved with the save result
  let savePromise: Promise<FormSchema> | null = null;

  // Create the save function
  const saveFunction = async (): Promise<FormSchema> => {
    try {
      const savedSchema = await formService.saveSchema(
        formId,
        typedBuilder.form,
      );
      if (!savedSchema) {
        throw new Error("No schema returned from server");
      }

      // Update the builder's form with the saved schema
      typedBuilder.form = savedSchema;

      return savedSchema;
    } catch (error) {
      // Re-throw with a more descriptive error
      throw new Error(
        `Failed to save form schema: ${error instanceof Error ? error.message : "Unknown error"}`,
      );
    }
  };

  // Add saveSchema method to the builder instance
  typedBuilder.saveSchema = async (): Promise<FormSchema> => {
    // If there's an ongoing save, return its promise
    if (savePromise) {
      return savePromise;
    }

    // Start a new save
    savePromise = saveFunction();
    try {
      const result = await savePromise;
      if (!result) {
        throw new Error("No schema returned from save operation");
      }
      return result;
    } finally {
      // Clear the promise after it's done
      savePromise = null;
    }
  };

  // Register all event handlers
  Object.entries(EVENT_MAP).forEach(([event, handler]) => {
    builder.on(event, () => handler(typedBuilder));
  });

  // Store builder instance in state management instead of global window
  formState.set("formBuilderInstance", typedBuilder);

  // Set up View Schema button handler
  setupViewSchemaHandler(typedBuilder);
};

/**
 * Set up the View Schema button functionality
 */
function setupViewSchemaHandler(builder: FormBuilderWithSchema): void {
  const viewSchemaBtn = dom.getElement<HTMLButtonElement>("view-schema-btn");
  if (!viewSchemaBtn) return;

  viewSchemaBtn.addEventListener("click", async () => {
    try {
      // Get the current schema
      const schema = await builder.saveSchema();
      if (!schema) {
        throw new Error("Failed to get form schema");
      }

      // Show the schema in a formatted way
      const schemaString = JSON.stringify(schema, null, 2);
      showSchemaModal(schemaString);
    } catch (error) {
      console.error("Failed to get schema:", error);
      dom.showError("Failed to get form schema. Please try again.");
    }
  });
}
