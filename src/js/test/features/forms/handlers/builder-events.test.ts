import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import {
  BuilderEventManager,
  createEventManager,
} from "@/features/forms/handlers/builder-events";
import { Logger } from "@/core/logger";

describe("BuilderEventManager", () => {
  let eventManager: BuilderEventManager;
  let mockBuilder: { element: HTMLElement };

  beforeEach(() => {
    // Create a mock builder element
    const element = document.createElement("div");
    element.id = "form-builder";
    document.body.appendChild(element);

    mockBuilder = { element };
    eventManager = createEventManager(element);
  });

  afterEach(() => {
    // Clean up
    eventManager.cleanup();
    if (mockBuilder.element.parentNode) {
      mockBuilder.element.remove();
    }
    vi.clearAllMocks();
  });

  describe("event listener management", () => {
    it("should add and remove event listeners", () => {
      const handler = vi.fn();
      eventManager.addEventListener("click", handler);

      expect(eventManager.hasHandlers("click")).toBe(true);
      expect(eventManager.getHandlerCount("click")).toBe(1);

      mockBuilder.element.click();
      expect(handler).toHaveBeenCalledTimes(1);

      eventManager.removeEventListener("click", handler);
      expect(eventManager.hasHandlers("click")).toBe(false);
      expect(eventManager.getHandlerCount("click")).toBe(0);
    });

    it("should handle multiple handlers for same event", () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();

      eventManager.addEventListener("click", handler1);
      eventManager.addEventListener("click", handler2);

      expect(eventManager.getHandlerCount("click")).toBe(2);

      mockBuilder.element.click();
      expect(handler1).toHaveBeenCalledTimes(1);
      expect(handler2).toHaveBeenCalledTimes(1);
    });

    it("should handle different event types", () => {
      const clickHandler = vi.fn();
      const changeHandler = vi.fn();

      eventManager.addEventListener("click", clickHandler);
      eventManager.addEventListener("change", changeHandler);

      mockBuilder.element.click();
      expect(clickHandler).toHaveBeenCalledTimes(1);
      expect(changeHandler).not.toHaveBeenCalled();

      mockBuilder.element.dispatchEvent(new Event("change"));
      expect(changeHandler).toHaveBeenCalledTimes(1);
    });
  });

  describe("debounced events", () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("should debounce rapid events", () => {
      const handler = vi.fn();
      eventManager.addEventListener("input", handler, { debounce: 100 });

      // Trigger multiple events rapidly
      mockBuilder.element.dispatchEvent(new Event("input"));
      mockBuilder.element.dispatchEvent(new Event("input"));
      mockBuilder.element.dispatchEvent(new Event("input"));

      // Handler should not be called immediately
      expect(handler).not.toHaveBeenCalled();

      // Fast-forward time
      vi.advanceTimersByTime(100);

      // Handler should be called once
      expect(handler).toHaveBeenCalledTimes(1);
    });

    it("should handle multiple debounced events", () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();

      eventManager.addEventListener("input", handler1, { debounce: 100 });
      eventManager.addEventListener("change", handler2, { debounce: 200 });

      mockBuilder.element.dispatchEvent(new Event("input"));
      mockBuilder.element.dispatchEvent(new Event("change"));

      vi.advanceTimersByTime(100);
      expect(handler1).toHaveBeenCalledTimes(1);
      expect(handler2).not.toHaveBeenCalled();

      vi.advanceTimersByTime(100);
      expect(handler2).toHaveBeenCalledTimes(1);
    });
  });

  describe("cleanup", () => {
    it("should remove all event listeners on cleanup", () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();

      eventManager.addEventListener("click", handler1);
      eventManager.addEventListener("change", handler2);

      expect(eventManager.hasHandlers("click")).toBe(true);
      expect(eventManager.hasHandlers("change")).toBe(true);

      eventManager.cleanup();

      expect(eventManager.hasHandlers("click")).toBe(false);
      expect(eventManager.hasHandlers("change")).toBe(false);
    });

    it("should clear debounce timers on cleanup", () => {
      vi.useFakeTimers();

      const handler = vi.fn();
      eventManager.addEventListener("input", handler, { debounce: 100 });

      mockBuilder.element.dispatchEvent(new Event("input"));
      eventManager.cleanup();

      vi.advanceTimersByTime(100);
      expect(handler).not.toHaveBeenCalled();

      vi.useRealTimers();
    });
  });

  describe("error handling", () => {
    it("should handle errors in event handlers gracefully", () => {
      const errorHandler = vi.fn();
      const throwingHandler = () => {
        throw new Error("Test error");
      };

      // Mock Logger.error instead of console.error
      const loggerSpy = vi.spyOn(Logger, "error").mockImplementation(() => {});

      eventManager.addEventListener("click", throwingHandler);
      eventManager.addEventListener("click", errorHandler);

      mockBuilder.element.click();

      expect(errorHandler).toHaveBeenCalled();
      expect(loggerSpy).toHaveBeenCalledWith(
        "Event handler error:",
        expect.any(Error),
      );

      // Restore Logger.error
      loggerSpy.mockRestore();
    });

    it("should handle errors in debounced handlers gracefully", () => {
      vi.useFakeTimers();

      const errorHandler = vi.fn();
      const throwingHandler = () => {
        throw new Error("Test error");
      };

      // Mock Logger.error instead of console.error
      const loggerSpy = vi.spyOn(Logger, "error").mockImplementation(() => {});

      eventManager.addEventListener("input", throwingHandler, {
        debounce: 100,
      });
      eventManager.addEventListener("input", errorHandler, { debounce: 100 });

      mockBuilder.element.dispatchEvent(new Event("input"));
      vi.advanceTimersByTime(100);

      expect(errorHandler).toHaveBeenCalled();
      expect(loggerSpy).toHaveBeenCalledWith(
        "Event handler error:",
        expect.any(Error),
      );

      // Restore Logger.error and timers
      loggerSpy.mockRestore();
      vi.useRealTimers();
    });

    it("should continue processing other handlers after error", () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();
      const throwingHandler = () => {
        throw new Error("Test error");
      };

      // Mock Logger.error instead of console.error
      const loggerSpy = vi.spyOn(Logger, "error").mockImplementation(() => {});

      eventManager.addEventListener("click", handler1);
      eventManager.addEventListener("click", throwingHandler);
      eventManager.addEventListener("click", handler2);

      mockBuilder.element.click();

      expect(handler1).toHaveBeenCalled();
      expect(handler2).toHaveBeenCalled();
      expect(loggerSpy).toHaveBeenCalledWith(
        "Event handler error:",
        expect.any(Error),
      );

      // Restore Logger.error
      loggerSpy.mockRestore();
    });
  });
});
