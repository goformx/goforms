import { Formio } from "@formio/js";
import { FormService } from "@/features/forms/services/form-service";
import type { FormSchema } from "@/features/forms/services/form-service";
import { debounce } from "lodash";
import { showSchemaModal } from "@/features/forms/components/form-builder/schema-modal";
import { dom } from "@/shared/utils/dom-utils";
import { formState } from "@/features/forms/state/form-state";

export interface FormBuilderWithSchema extends Formio {
  form: FormSchema;
  saveSchema: () => Promise<FormSchema>;
  element?: HTMLElement;
}

// Define event handlers map type
type EventHandlerMap = {
  [key: string]: (builder: FormBuilderWithSchema) => void;
};

/**
 * Manages event listeners and timers to prevent memory leaks
 */
export class BuilderEventManager {
  private eventHandlers = new Map<string, AbortController>();
  private debounceTimers = new Map<string, NodeJS.Timeout>();
  private builder: FormBuilderWithSchema;
  private formId: string;

  constructor(
    builder: FormBuilderWithSchema,
    formId: string,
    _formService: FormService,
  ) {
    this.builder = builder;
    this.formId = formId;
  }

  /**
   * Add event listener with automatic cleanup
   */
  addEventListener(
    eventType: string,
    handler: EventListener,
    element?: Element,
  ): void {
    const controller = new AbortController();
    const target = element || this.builder.element || document;
    const handlerId = `${eventType}-${Date.now()}`;

    target.addEventListener(eventType, handler, {
      signal: controller.signal,
    });

    this.eventHandlers.set(handlerId, controller);
  }

  /**
   * Add debounced event listener with cleanup
   */
  addDebouncedEventListener(
    eventType: string,
    handler: (builder: FormBuilderWithSchema) => void,
    delay: number = 300,
  ): void {
    const debouncedHandler = debounce(() => handler(this.builder), delay);
    const handlerId = `${eventType}-debounced`;

    // Store the timer for cleanup
    this.debounceTimers.set(
      handlerId,
      debouncedHandler.flush as unknown as NodeJS.Timeout,
    );

    this.addEventListener(eventType, debouncedHandler);
  }

  /**
   * Remove specific event listener
   */
  removeEventListener(eventType: string): void {
    const handlerId = Array.from(this.eventHandlers.keys()).find((key) =>
      key.startsWith(eventType),
    );

    if (handlerId) {
      const controller = this.eventHandlers.get(handlerId);
      if (controller) {
        controller.abort();
        this.eventHandlers.delete(handlerId);
      }
    }
  }

  /**
   * Clear all debounce timers
   */
  clearDebounceTimers(): void {
    this.debounceTimers.forEach((timer) => {
      if (timer && typeof timer === "function") {
        clearTimeout(timer as unknown as NodeJS.Timeout);
      }
    });
    this.debounceTimers.clear();
  }

  /**
   * Comprehensive cleanup of all event listeners and timers
   */
  cleanup(): void {
    // Clean up all event listeners
    this.eventHandlers.forEach((controller) => controller.abort());
    this.eventHandlers.clear();

    // Clear all timers
    this.clearDebounceTimers();

    // Remove from state management
    formState.delete("formBuilderInstance");

    console.debug(
      `BuilderEventManager: Cleaned up ${this.eventHandlers.size} event handlers and ${this.debounceTimers.size} timers for form ${this.formId}`,
    );
  }

  /**
   * Get cleanup statistics for debugging
   */
  getCleanupStats(): { eventHandlers: number; timers: number } {
    return {
      eventHandlers: this.eventHandlers.size,
      timers: this.debounceTimers.size,
    };
  }
}

// Define event handlers map as a constant
const EVENT_MAP: EventHandlerMap = {
  change: (_builder: FormBuilderWithSchema) => {
    // Only log changes, no automatic saving
    console.debug("Form builder change detected");
  },
  saveComponent: (_builder: FormBuilderWithSchema) => {
    // Only log component saves, no automatic saving
    console.debug("Form component saved");
  },
};

export const setupBuilderEvents = (
  builder: Formio,
  formId: string,
  formService: FormService,
): BuilderEventManager => {
  const typedBuilder = builder as FormBuilderWithSchema;

  // Create event manager instance
  const eventManager = new BuilderEventManager(
    typedBuilder,
    formId,
    formService,
  );

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

  // Register all event handlers with debouncing
  Object.entries(EVENT_MAP).forEach(([event, handler]) => {
    eventManager.addDebouncedEventListener(event, handler);
  });

  // Store builder instance in state management instead of global window
  formState.set("formBuilderInstance", typedBuilder);

  // Set up View Schema button handler
  setupViewSchemaHandler(typedBuilder, eventManager);

  // Set up cleanup on page unload
  window.addEventListener("beforeunload", () => {
    eventManager.cleanup();
  });

  return eventManager;
};

/**
 * Set up the View Schema button functionality
 */
function setupViewSchemaHandler(
  builder: FormBuilderWithSchema,
  eventManager: BuilderEventManager,
): void {
  const viewSchemaBtn = dom.getElement<HTMLButtonElement>("view-schema-btn");
  if (!viewSchemaBtn) return;

  const handleViewSchema = async () => {
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
  };

  // Add event listener with cleanup
  eventManager.addEventListener("click", handleViewSchema, viewSchemaBtn);
}

/**
 * Cleanup function to be called when form builder is destroyed
 */
export const cleanupBuilderEvents = (
  eventManager: BuilderEventManager,
): void => {
  if (eventManager) {
    eventManager.cleanup();
  }
};
