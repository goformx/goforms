/**
 * DOM cache for optimizing repeated queries
 */
class DOMCache {
  private cache = new Map<string, Element>();

  get(selector: string, container?: Element): Element | null {
    const cacheKey = `${selector}-${container?.id || "document"}`;

    if (this.cache.has(cacheKey)) {
      const element = this.cache.get(cacheKey)!;
      // Verify element is still in DOM
      if (document.contains(element)) {
        return element;
      }
      this.cache.delete(cacheKey);
    }

    const element = (container || document).querySelector(selector);
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
 * Enhanced DOM utilities with caching
 */
export const dom = {
  getElement<T extends HTMLElement>(id: string): T | null {
    return domCache.get(`#${id}`) as T | null;
  },

  getElementBySelector<T extends HTMLElement>(
    selector: string,
    container?: Element,
  ): T | null {
    return domCache.get(selector, container) as T | null;
  },

  createElement<T extends HTMLElement>(tag: string, className?: string): T {
    const element = document.createElement(tag) as T;
    if (className) element.className = className;
    return element;
  },

  showError(message: string, container?: Element): void {
    const errorContainer = domCache.get(".gf-error-message", container);

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

  hideError(container?: Element): void {
    const errorContainer = domCache.get(".gf-error-message", container);
    if (errorContainer instanceof HTMLElement) {
      errorContainer.style.display = "none";
    }
  },

  showSuccess(message: string, container?: Element): void {
    const successContainer = domCache.get(".gf-success-message", container);

    if (successContainer instanceof HTMLElement) {
      successContainer.textContent = message;
      successContainer.style.display = "block";
      return;
    }

    const successDiv = dom.createElement<HTMLDivElement>(
      "div",
      "gf-success-message",
    );
    successDiv.textContent = message;
    document.body.insertBefore(successDiv, document.body.firstChild);
  },

  hideSuccess(container?: Element): void {
    const successContainer = domCache.get(".gf-success-message", container);
    if (successContainer instanceof HTMLElement) {
      successContainer.style.display = "none";
    }
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
};
