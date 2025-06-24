/**
 * Test file demonstrating BuilderEventManager usage
 * This file shows how to properly use the event manager to prevent memory leaks
 */

import {
  BuilderEventManager,
  setupBuilderEvents,
  cleanupBuilderEvents,
} from "./builder-events";
import { FormService } from "@/features/forms/services/form-service";
import type { FormBuilderWithSchema } from "./builder-events";

// Mock Form.io builder for testing
const createMockBuilder = (): FormBuilderWithSchema => {
  return {
    form: {
      display: "form",
      components: [],
    },
    saveSchema: async () => ({
      display: "form",
      components: [],
    }),
    element: document.createElement("div"),
  } as FormBuilderWithSchema;
};

/**
 * Example usage of BuilderEventManager
 */
export class FormBuilderExample {
  private eventManager: BuilderEventManager | null = null;
  private formService: FormService;

  constructor() {
    this.formService = FormService.getInstance();
  }

  /**
   * Initialize form builder with proper event management
   */
  initializeFormBuilder(formId: string): void {
    const builder = createMockBuilder();

    // Set up events with automatic cleanup
    this.eventManager = setupBuilderEvents(builder, formId, this.formService);

    console.log("Form builder initialized with event management");
  }

  /**
   * Clean up form builder when component is destroyed
   */
  destroyFormBuilder(): void {
    if (this.eventManager) {
      // Get cleanup statistics before cleanup
      const stats = this.eventManager.getCleanupStats();
      console.log("Cleanup stats before destruction:", stats);

      // Perform cleanup
      cleanupBuilderEvents(this.eventManager);
      this.eventManager = null;

      console.log("Form builder destroyed and cleaned up");
    }
  }

  /**
   * Get event manager statistics for monitoring
   */
  getEventManagerStats(): { eventHandlers: number; timers: number } | null {
    return this.eventManager?.getCleanupStats() || null;
  }

  /**
   * Example of adding custom event listeners
   */
  addCustomEvents(): void {
    if (!this.eventManager) return;

    // Add custom event listener with cleanup
    this.eventManager.addEventListener("custom-event", (event) => {
      console.log("Custom event triggered:", event);
    });

    // Add debounced event listener
    this.eventManager.addDebouncedEventListener(
      "resize",
      () => {
        console.log("Window resized - debounced");
      },
      500,
    );
  }

  /**
   * Example of removing specific event listeners
   */
  removeSpecificEvents(): void {
    if (!this.eventManager) return;

    // Remove specific event type
    this.eventManager.removeEventListener("custom-event");
    console.log("Custom event listener removed");
  }
}

/**
 * Memory leak prevention example
 */
export const demonstrateMemoryLeakPrevention = (): void => {
  const example = new FormBuilderExample();

  // Initialize form builder
  example.initializeFormBuilder("test-form");

  // Add custom events
  example.addCustomEvents();

  // Simulate some time passing
  setTimeout(() => {
    // Remove specific events
    example.removeSpecificEvents();

    // Clean up everything
    example.destroyFormBuilder();

    console.log("Memory leak prevention demonstration completed");
  }, 5000);
};

/**
 * Performance monitoring example
 */
export const monitorEventManagerPerformance = (): void => {
  const example = new FormBuilderExample();
  example.initializeFormBuilder("performance-test");

  // Add many event listeners to test performance
  for (let i = 0; i < 10; i++) {
    example.addCustomEvents();
  }

  // Monitor memory usage
  const stats = example.getEventManagerStats();
  console.log("Event manager stats:", stats);

  // Clean up
  example.destroyFormBuilder();
};

// Export for use in other modules
export { FormBuilderExample as FormBuilderWithEventManagement };
