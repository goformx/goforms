import { Logger } from "@/core/logger";

interface EventHandler {
  original: (event: Event) => void;
  wrapped: (event: Event) => void;
}

/**
 * Event manager for form builder interactions
 */
export class BuilderEventManager {
  private readonly handlers = new Map<string, EventHandler[]>();
  private readonly debounceTimers = new Map<
    string,
    ReturnType<typeof setTimeout>
  >();
  private readonly element: HTMLElement;

  constructor(element: HTMLElement) {
    this.element = element;
  }

  /**
   * Add event listener with optional debouncing
   * @returns true if the handler was successfully added
   */
  addEventListener(
    eventType: string,
    handler: (event: Event) => void,
    options: { debounce?: number } = {},
  ): boolean {
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

      // Store both original handler and wrapped handler
      handlers.push({ original: handler, wrapped: debouncedHandler });
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

      // Store both original handler and wrapped handler
      handlers.push({ original: handler, wrapped: wrappedHandler });
      this.element.addEventListener(eventType, wrappedHandler);
    }

    return true;
  }

  /**
   * Remove event listener
   * @returns true if the handler was successfully removed, false if not found
   */
  removeEventListener(
    eventType: string,
    handler: (event: Event) => void,
  ): boolean {
    const handlers = this.handlers.get(eventType);
    if (!handlers) return false;

    const index = handlers.findIndex((h) => h.original === handler);
    if (index === -1) return false;

    // Remove the wrapped handler from DOM
    const removedHandler = handlers.splice(index, 1)[0];
    this.element.removeEventListener(eventType, removedHandler.wrapped);

    // Clean up empty handler arrays to prevent memory leaks
    if (handlers.length === 0) {
      this.handlers.delete(eventType);
    }

    return true;
  }

  /**
   * Remove all event listeners
   * @returns the number of event types that were cleaned up
   */
  removeAllEventListeners(): number {
    let cleanupCount = 0;

    for (const [eventType, handlers] of this.handlers) {
      for (const handler of handlers) {
        this.element.removeEventListener(eventType, handler.wrapped);
      }
      cleanupCount++;
    }

    this.handlers.clear();
    this.clearDebounceTimers();

    return cleanupCount;
  }

  /**
   * Clear all debounce timers
   * @returns the number of timers that were cleared
   */
  private clearDebounceTimers(): number {
    let timerCount = 0;

    for (const timerId of this.debounceTimers.values()) {
      clearTimeout(timerId);
      timerCount++;
    }

    this.debounceTimers.clear();
    return timerCount;
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
   * Get all registered event types
   */
  getRegisteredEventTypes(): string[] {
    return Array.from(this.handlers.keys());
  }

  /**
   * Check if a specific handler is registered for an event type
   */
  hasHandler(eventType: string, handler: (event: Event) => void): boolean {
    const handlers = this.handlers.get(eventType);
    if (!handlers) return false;

    return handlers.some((h) => h.original === handler);
  }

  /**
   * Clean up resources
   * @returns object with cleanup statistics
   */
  cleanup(): { eventTypesCleaned: number; timersCleared: number } {
    const eventTypesCleaned = this.removeAllEventListeners();
    const timersCleared = this.clearDebounceTimers();

    return { eventTypesCleaned, timersCleared };
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
