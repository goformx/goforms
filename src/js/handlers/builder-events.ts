import { FormBuilder } from "@formio/js";
import { FormService, FormSchema } from "../services/form-service";
import { debounce } from "lodash";

interface FormBuilderWithSchema extends FormBuilder {
  form: FormSchema;
  saveSchema: () => Promise<boolean>;
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
  console.log(`[${new Date().toISOString()}] ${message}`, data);
};

// Define event handlers map as a constant
const EVENT_MAP: EventHandlerMap = {
  saveComponent: (builder) => {
    logWithTimestamp("Component saved:", builder.form);
  },
  change: debounce((builder: FormBuilderWithSchema) => {
    logWithTimestamp("Form modified:", builder.form);
  }, 300),
};

export const setupBuilderEvents = (
  builder: FormBuilder,
  formId: number,
  formService: FormService,
): void => {
  const typedBuilder = builder as FormBuilderWithSchema;

  // Add saveSchema method to the builder instance
  typedBuilder.saveSchema = async () => {
    try {
      await formService.saveSchema(formId, typedBuilder.form);
      return true;
    } catch (error) {
      console.error("Error saving schema:", error);
      // Re-throw with a more descriptive error
      throw new Error(
        `Failed to save form schema: ${error instanceof Error ? error.message : "Unknown error"}`,
      );
    }
  };

  // Register all event handlers
  Object.entries(EVENT_MAP).forEach(([event, handler]) => {
    builder.on(event, () => handler(typedBuilder));
  });

  // Store builder instance globally
  window.formBuilderInstance = typedBuilder;
};
