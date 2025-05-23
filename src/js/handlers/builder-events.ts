import { FormBuilder } from "@formio/js";
import { FormService } from "../services/form-service";
import type { FormSchema } from "../services/form-service";
import { debounce } from "lodash";

interface FormBuilderWithSchema extends FormBuilder {
  form: FormSchema;
  saveSchema: () => Promise<FormSchema>;
}

// Extend Window interface globally
declare global {
  interface Window {
    formBuilderInstance?: FormBuilderWithSchema;
  }
}

// Define event handlers map type
type EventHandlerMap = {
  [key: string]: (builder: FormBuilderWithSchema) => void;
};

// Helper function for timestamped logging
const logWithTimestamp = (message: string, data?: unknown) => {
  if (data !== undefined) {
    console.log(`[${new Date().toISOString()}] ${message}`, data);
  } else {
    console.log(`[${new Date().toISOString()}] ${message}`);
  }
};

// Define event handlers map as a constant
const EVENT_MAP: EventHandlerMap = {
  change: debounce((builder: FormBuilderWithSchema) => {
    logWithTimestamp("Form modified:", builder.form);
    // Only log changes, no automatic saving
  }, 300),
  saveComponent: (builder: FormBuilderWithSchema) => {
    logWithTimestamp("Component saved:", builder.form);
    // Only log component saves, no automatic saving
  },
};

export const setupBuilderEvents = (
  builder: FormBuilder,
  formId: string,
  formService: FormService,
): void => {
  const typedBuilder = builder as FormBuilderWithSchema;

  // Create a promise that will be resolved with the save result
  let savePromise: Promise<FormSchema> | null = null;

  // Create the save function
  const saveFunction = async (): Promise<FormSchema> => {
    try {
      console.log("Builder-events: Starting schema save...");
      logWithTimestamp("Explicitly saving schema...");
      const savedSchema = await formService.saveSchema(
        formId,
        typedBuilder.form,
      );
      console.log("Builder-events: Schema saved successfully:", savedSchema);
      logWithTimestamp("Schema saved successfully");
      if (!savedSchema) {
        console.error("Builder-events: No schema returned from server");
        throw new Error("No schema returned from server");
      }
      return savedSchema;
    } catch (error) {
      console.error("Builder-events: Error saving schema:", error);
      // Re-throw with a more descriptive error
      throw new Error(
        `Failed to save form schema: ${error instanceof Error ? error.message : "Unknown error"}`,
      );
    }
  };

  // Add saveSchema method to the builder instance
  typedBuilder.saveSchema = async (): Promise<FormSchema> => {
    console.log("Builder-events: saveSchema called");
    // If there's an ongoing save, return its promise
    if (savePromise) {
      console.log("Builder-events: Returning existing save promise");
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

  // Store builder instance globally
  window.formBuilderInstance = typedBuilder;
};
