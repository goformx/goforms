/**
 * DOM cache for optimizing repeated queries
 */
class DOMCache {
  private readonly cache = new Map<string, Element>();

  get(selector: string, container?: Element): Element | null {
    const cacheKey = `${selector}-${container?.id ?? "document"}`;

    if (this.cache.has(cacheKey)) {
      const element = this.cache.get(cacheKey)!;
      // Verify element is still in DOM
      if (document.contains(element)) {
        return element;
      }
      this.cache.delete(cacheKey);
    }

    const element = (container ?? document).querySelector(selector);
    if (element) {
      this.cache.set(cacheKey, element);
    }

    return element;
  }

  clear(): void {
    this.cache.clear();
  }

  invalidate(selector?: string): void {
    if (selector) {
      // Remove specific selector from cache
      for (const key of this.cache.keys()) {
        if (key.startsWith(selector)) {
          this.cache.delete(key);
        }
      }
    } else {
      this.clear();
    }
  }

  get size(): number {
    return this.cache.size;
  }
}

// Global DOM cache instance
export const domCache = new DOMCache();

/**
 * DOM element type guard
 */
export const isHTMLElement = (element: Element): element is HTMLElement => {
  return element instanceof HTMLElement;
};

/**
 * DOM element type guard for specific types
 */
export const isHTMLInputElement = (
  element: Element,
): element is HTMLInputElement => {
  return element instanceof HTMLInputElement;
};

export const isHTMLFormElement = (
  element: Element,
): element is HTMLFormElement => {
  return element instanceof HTMLFormElement;
};

export const isHTMLButtonElement = (
  element: Element,
): element is HTMLButtonElement => {
  return element instanceof HTMLButtonElement;
};

/**
 * Enhanced DOM utilities with caching and type safety
 */
export const dom = {
  /**
   * Get element by ID with type safety
   */
  getElement<T extends HTMLElement>(id: string): T | null {
    return domCache.get(`#${id}`) as T | null;
  },

  /**
   * Get element by selector with type safety
   */
  getElementBySelector<T extends HTMLElement>(
    selector: string,
    container?: Element,
  ): T | null {
    return domCache.get(selector, container) as T | null;
  },

  /**
   * Get multiple elements by selector
   */
  getElementsBySelector<T extends HTMLElement>(
    selector: string,
    container?: Element,
  ): T[] {
    const elements = (container ?? document).querySelectorAll(selector);
    return Array.from(elements) as T[];
  },

  /**
   * Create element with type safety
   */
  createElement<T extends HTMLElement>(tag: string, className?: string): T {
    const element = document.createElement(tag) as T;
    if (className) element.className = className;
    return element;
  },

  /**
   * Create element with attributes
   */
  createElementWithAttributes<T extends HTMLElement>(
    tag: string,
    attributes: Record<string, string>,
  ): T {
    const element = this.createElement<T>(tag);
    Object.entries(attributes).forEach(([key, value]) => {
      element.setAttribute(key, value);
    });
    return element;
  },

  /**
   * Show error message with type safety
   */
  showError(message: string, container?: Element): void {
    const errorContainer = domCache.get(".gf-error-message", container);

    if (errorContainer && isHTMLElement(errorContainer)) {
      errorContainer.textContent = message;
      errorContainer.style.display = "block";
      return;
    }

    const errorDiv = this.createElement<HTMLDivElement>(
      "div",
      "gf-error-message",
    );
    errorDiv.textContent = message;
    document.body.insertBefore(errorDiv, document.body.firstChild);
  },

  /**
   * Hide error message
   */
  hideError(container?: Element): void {
    const errorContainer = domCache.get(".gf-error-message", container);
    if (errorContainer && isHTMLElement(errorContainer)) {
      errorContainer.style.display = "none";
    }
  },

  /**
   * Show success message with type safety
   */
  showSuccess(message: string, container?: Element): void {
    const successContainer = domCache.get(".gf-success-message", container);

    if (successContainer && isHTMLElement(successContainer)) {
      successContainer.textContent = message;
      successContainer.style.display = "block";
      return;
    }

    const successDiv = this.createElement<HTMLDivElement>(
      "div",
      "gf-success-message",
    );
    successDiv.textContent = message;
    document.body.insertBefore(successDiv, document.body.firstChild);
  },

  /**
   * Hide success message
   */
  hideSuccess(container?: Element): void {
    const successContainer = domCache.get(".gf-success-message", container);
    if (successContainer && isHTMLElement(successContainer)) {
      successContainer.style.display = "none";
    }
  },

  /**
   * Add event listener with type safety
   */
  addEventListener<T extends Event>(
    element: Element,
    event: string,
    listener: (event: T) => void,
    options?: AddEventListenerOptions,
  ): () => void {
    element.addEventListener(event, listener as EventListener, options);
    return () => {
      element.removeEventListener(event, listener as EventListener, options);
    };
  },

  /**
   * Remove event listener
   */
  removeEventListener<T extends Event>(
    element: Element,
    event: string,
    listener: (event: T) => void,
    options?: EventListenerOptions,
  ): void {
    element.removeEventListener(event, listener as EventListener, options);
  },

  /**
   * Toggle element visibility
   */
  toggleElement(element: Element, show: boolean): void {
    if (isHTMLElement(element)) {
      element.style.display = show ? "block" : "none";
    }
  },

  /**
   * Set element attributes
   */
  setAttributes(element: Element, attributes: Record<string, string>): void {
    Object.entries(attributes).forEach(([key, value]) => {
      element.setAttribute(key, value);
    });
  },

  /**
   * Remove element attributes
   */
  removeAttributes(element: Element, attributes: readonly string[]): void {
    attributes.forEach((attr) => {
      element.removeAttribute(attr);
    });
  },

  /**
   * Clear DOM cache - useful when DOM structure changes significantly
   */
  clearCache(): void {
    domCache.clear();
  },

  /**
   * Invalidate specific selectors from cache
   */
  invalidateCache(selector?: string): void {
    domCache.invalidate(selector);
  },

  /**
   * Get cache size for debugging
   */
  get cacheSize(): number {
    return domCache.size;
  },
} as const;
