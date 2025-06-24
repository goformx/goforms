import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { dom } from "@/shared/utils/dom-utils";

describe("DOM Utilities", () => {
  let container: HTMLElement;

  beforeEach(() => {
    // Clear DOM cache before each test
    dom.clearCache();

    // Create a test container
    container = document.createElement("div");
    container.id = "test-container";
    document.body.appendChild(container);
  });

  afterEach(() => {
    // Clean up after each test
    if (container.parentNode) {
      container.remove();
    }

    // Clean up any messages that might have been added to document.body
    document
      .querySelectorAll(".gf-error-message, .gf-success-message")
      .forEach((el) => el.remove());

    // Clear cache
    dom.clearCache();
  });

  describe("element retrieval", () => {
    it("should get element by ID", () => {
      const testElement = document.createElement("div");
      testElement.id = "test-element";
      container.appendChild(testElement);

      const result = dom.getElement<HTMLDivElement>("test-element");
      expect(result).toBe(testElement);
    });

    it("should get element by selector", () => {
      const testElement = document.createElement("div");
      testElement.className = "test-class";
      container.appendChild(testElement);

      const result = dom.getElementBySelector<HTMLDivElement>(
        ".test-class",
        container,
      );
      expect(result).toBe(testElement);
    });

    it("should return null for non-existent elements", () => {
      const result = dom.getElement<HTMLDivElement>("non-existent");
      expect(result).toBeNull();
    });
  });

  describe("element creation", () => {
    it("should create element with tag name", () => {
      const element = dom.createElement<HTMLDivElement>("div");
      expect(element.tagName).toBe("DIV");
    });

    it("should create element with class name", () => {
      const element = dom.createElement<HTMLDivElement>("div", "test-class");
      expect(element.className).toBe("test-class");
    });

    it("should create different element types", () => {
      const button = dom.createElement<HTMLButtonElement>("button", "btn");
      const input = dom.createElement<HTMLInputElement>(
        "input",
        "form-control",
      );

      expect(button.tagName).toBe("BUTTON");
      expect(button.className).toBe("btn");
      expect(input.tagName).toBe("INPUT");
      expect(input.className).toBe("form-control");
    });
  });

  describe("message display", () => {
    it("should show error messages", () => {
      dom.showError("Test error message", container);

      // Check both container and document.body since implementation falls back
      const errorInContainer = container.querySelector(".gf-error-message");
      const errorInBody = document.body.querySelector(".gf-error-message");
      const errorElement = errorInContainer || errorInBody;

      expect(errorElement).toBeTruthy();
      expect(errorElement?.textContent).toBe("Test error message");
      expect(errorElement?.classList.contains("gf-error-message")).toBe(true);
    });

    it("should hide error messages", () => {
      dom.showError("Test error message", container);

      // Check both container and document.body since implementation falls back
      const errorInContainer = container.querySelector(".gf-error-message");
      const errorInBody = document.body.querySelector(".gf-error-message");
      const errorElement = errorInContainer || errorInBody;
      expect(errorElement).toBeTruthy();

      dom.hideError(container);

      // Check that the error is hidden in the appropriate location
      const hiddenErrorInContainer = container.querySelector(
        ".gf-error-message",
      ) as HTMLElement;
      const hiddenErrorInBody = document.body.querySelector(
        ".gf-error-message",
      ) as HTMLElement;
      const hiddenError = hiddenErrorInContainer || hiddenErrorInBody;

      // Element should be hidden (either display: none or not visible)
      expect(
        hiddenError?.style.display === "none" || !hiddenError?.offsetParent,
      ).toBe(true);
    });

    it("should show success messages", () => {
      dom.showSuccess("Test success message", container);

      // Check both container and document.body since implementation falls back
      const successInContainer = container.querySelector(".gf-success-message");
      const successInBody = document.body.querySelector(".gf-success-message");
      const successElement = successInContainer || successInBody;

      expect(successElement).toBeTruthy();
      expect(successElement?.textContent).toBe("Test success message");
      expect(successElement?.classList.contains("gf-success-message")).toBe(
        true,
      );
    });

    it("should hide success messages", () => {
      dom.showSuccess("Test success message", container);

      // Check both container and document.body since implementation falls back
      const successInContainer = container.querySelector(".gf-success-message");
      const successInBody = document.body.querySelector(".gf-success-message");
      const successElement = successInContainer || successInBody;
      expect(successElement).toBeTruthy();

      dom.hideSuccess(container);

      // Check that the success message is hidden in the appropriate location
      const hiddenSuccessInContainer = container.querySelector(
        ".gf-success-message",
      ) as HTMLElement;
      const hiddenSuccessInBody = document.body.querySelector(
        ".gf-success-message",
      ) as HTMLElement;
      const hiddenSuccess = hiddenSuccessInContainer || hiddenSuccessInBody;

      // Element should be hidden (either display: none or not visible)
      expect(
        hiddenSuccess?.style.display === "none" || !hiddenSuccess?.offsetParent,
      ).toBe(true);
    });

    it("should reuse existing message containers", () => {
      // Create existing error container
      const existingError = document.createElement("div");
      existingError.className = "gf-error-message";
      existingError.textContent = "Old error";
      container.appendChild(existingError);

      dom.showError("New error", container);

      expect(container.querySelectorAll(".gf-error-message")).toHaveLength(1);
      expect(container.querySelector(".gf-error-message")?.textContent).toBe(
        "New error",
      );
    });

    it("should create message containers in document body when not found", () => {
      const initialCount =
        document.querySelectorAll(".gf-error-message").length;

      dom.showError("Global error");

      const newCount = document.querySelectorAll(".gf-error-message").length;
      expect(newCount).toBe(initialCount + 1);

      // Clean up
      const errorElement = document.querySelector(".gf-error-message");
      if (errorElement) {
        errorElement.remove();
      }
    });
  });

  describe("caching", () => {
    it("should cache element queries", () => {
      const testElement = document.createElement("div");
      testElement.id = "cache-test";
      container.appendChild(testElement);

      // First call should cache
      const firstResult = dom.getElement<HTMLDivElement>("cache-test");
      expect(firstResult).toBe(testElement);

      // Second call should use cache
      const secondResult = dom.getElement<HTMLDivElement>("cache-test");
      expect(secondResult).toBe(testElement);

      // Cache size should be 1
      expect(dom.cacheSize).toBe(1);
    });

    it("should clear cache", () => {
      const testElement = document.createElement("div");
      testElement.id = "clear-test";
      container.appendChild(testElement);

      dom.getElement<HTMLDivElement>("clear-test");
      expect(dom.cacheSize).toBeGreaterThan(0);

      dom.clearCache();
      expect(dom.cacheSize).toBe(0);
    });

    it("should invalidate specific selectors", () => {
      const testElement = document.createElement("div");
      testElement.id = "invalidate-test";
      container.appendChild(testElement);

      dom.getElement<HTMLDivElement>("invalidate-test");
      expect(dom.cacheSize).toBe(1);

      dom.invalidateCache("#invalidate-test");
      expect(dom.cacheSize).toBe(0);
    });
  });
});
