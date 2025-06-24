import { Logger } from "@/core/logger";

/**
 * Event manager for form builder interactions
 */
export class BuilderEventManager {
  private handlers = new Map<string, Array<(event: Event) => void>>();
  private debounceTimers = new Map<string, ReturnType<typeof setTimeout>>();
  private element: HTMLElement;

  constructor(element: HTMLElement) {
    this.element = element;
  }

  /**
   * Add event listener with optional debouncing
   */
  addEventListener(
    eventType: string,
    handler: (event: Event) => void,
    options: { debounce?: number } = {},
  ): void {
    if (!this.handlers.has(eventType)) {
      this.handlers.set(eventType, []);
    }

    const handlers = this.handlers.get(eventType)!;

    if (options.debounce) {
      // Create debounced handler
      const debouncedHandler = (event: Event) => {
        const timerKey = `${eventType}-${handler.toString()}`;

        // Clear existing timer
        if (this.debounceTimers.has(timerKey)) {
          clearTimeout(this.debounceTimers.get(timerKey)!);
        }

        // Set new timer
        const timerId = setTimeout(() => {
          try {
            handler(event);
          } catch (error) {
            Logger.error("Event handler error:", error);
          }
          this.debounceTimers.delete(timerKey);
        }, options.debounce);

        this.debounceTimers.set(timerKey, timerId);
      };

      handlers.push(debouncedHandler);
      this.element.addEventListener(eventType, debouncedHandler);
    } else {
      // Regular handler
      const wrappedHandler = (event: Event) => {
        try {
          handler(event);
        } catch (error) {
          Logger.error("Event handler error:", error);
        }
      };

      handlers.push(wrappedHandler);
      this.element.addEventListener(eventType, wrappedHandler);
    }
  }

  /**
   * Remove event listener
   */
  removeEventListener(
    eventType: string,
    handler: (event: Event) => void,
  ): void {
    const handlers = this.handlers.get(eventType);
    if (!handlers) return;

    const index = handlers.indexOf(handler);
    if (index > -1) {
      const removedHandler = handlers.splice(index, 1)[0];
      this.element.removeEventListener(eventType, removedHandler);
    }
  }

  /**
   * Remove all event listeners
   */
  removeAllEventListeners(): void {
    for (const [eventType, handlers] of this.handlers) {
      for (const handler of handlers) {
        this.element.removeEventListener(eventType, handler);
      }
    }
    this.handlers.clear();
    this.clearDebounceTimers();
  }

  /**
   * Clear all debounce timers
   */
  private clearDebounceTimers(): void {
    for (const timerId of this.debounceTimers.values()) {
      clearTimeout(timerId);
    }
    this.debounceTimers.clear();
  }

  /**
   * Get number of handlers for an event type
   */
  getHandlerCount(eventType: string): number {
    return this.handlers.get(eventType)?.length || 0;
  }

  /**
   * Check if event type has handlers
   */
  hasHandlers(eventType: string): boolean {
    return this.getHandlerCount(eventType) > 0;
  }

  /**
   * Clean up resources
   */
  cleanup(): void {
    this.removeAllEventListeners();
  }
}

/**
 * Create a new event manager for an element
 */
export function createEventManager(element: HTMLElement): BuilderEventManager {
  return new BuilderEventManager(element);
}

/**
 * Set up event handlers for the form builder
 */
export function setupBuilderEvents(
  builder: any,
  formId: string,
  formService: any,
): void {
  // Create event manager for the builder element
  const eventManager = createEventManager(builder.element);

  // Set up form change events
  eventManager.addEventListener(
    "change",
    async (_event) => {
      try {
        // Handle form changes - save schema when form changes
        const schema = builder.form;
        await formService.saveSchema(formId, schema);
        Logger.debug("Form schema saved:", formId);
      } catch (error) {
        Logger.error("Error handling form change:", error);
      }
    },
    { debounce: 500 },
  );

  // Set up form submission events - fix the unused parameter
  eventManager.addEventListener("submit", async (_event) => {
    try {
      // Handle form submission
      Logger.debug("Form submitted");
    } catch (error) {
      Logger.error("Error handling form submission:", error);
    }
  });

  // Store the event manager on the builder for cleanup
  (builder as any).eventManager = eventManager;
}
