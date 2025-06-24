import { Logger } from "@/core/logger";

interface EventHandler {
  original: (event: Event) => void;
  wrapped: (event: Event) => void;
}

/**
 * Event manager for form builder interactions
 */
export class BuilderEventManager {
  private handlers = new Map<string, EventHandler[]>();
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

    const index = handlers.findIndex((h) => h.original === handler);
    if (index > -1) {
      const removedHandler = handlers.splice(index, 1)[0];
      this.element.removeEventListener(eventType, removedHandler.wrapped);
    }
  }

  /**
   * Remove all event listeners
   */
  removeAllEventListeners(): void {
    for (const [eventType, handlers] of this.handlers) {
      for (const handler of handlers) {
        this.element.removeEventListener(eventType, handler.wrapped);
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
