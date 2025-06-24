/**
 * DOM utilities
 */
export const dom = {
  getElement<T extends HTMLElement>(id: string): T | null {
    return document.getElementById(id) as T | null;
  },

  createElement<T extends HTMLElement>(tag: string, className?: string): T {
    const element = document.createElement(tag) as T;
    if (className) element.className = className;
    return element;
  },

  showError(message: string, container?: HTMLElement): void {
    const errorContainer =
      container?.querySelector(".gf-error-message") ||
      document.querySelector(".gf-error-message");

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
};
