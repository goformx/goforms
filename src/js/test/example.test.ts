/**
 * Example test file demonstrating testing patterns for GoForms
 */

import { describe, it, expect, beforeEach, vi } from "vitest";
import { FormBuilderError, ErrorCode } from "@/core/errors/form-builder-error";
import { HttpClient } from "@/core/http-client";
import { dom } from "@/shared/utils/dom-utils";

// Mock DOM elements
const createMockElement = (id: string, className?: string): HTMLElement => {
  const element = document.createElement("div");
  element.id = id;
  if (className) element.className = className;
  return element;
};

describe("FormBuilderError", () => {
  it("should create validation errors correctly", () => {
    const error = FormBuilderError.validationError(
      "Field is required",
      "email",
      "test@example.com",
    );

    expect(error.code).toBe(ErrorCode.VALIDATION_FAILED);
    expect(error.userMessage).toBe("Field is required");
    expect(error.getContextValue<string>("field")).toBe("email");
    expect(error.getContextValue<unknown>("value")).toBe("test@example.com");
  });

  it("should create network errors correctly", () => {
    const error = FormBuilderError.networkError(
      "Connection failed",
      "http://api.example.com",
      500,
    );

    expect(error.code).toBe(ErrorCode.NETWORK_ERROR);
    expect(error.userMessage).toBe(
      "Network connection failed. Please check your internet connection and try again.",
    );
    expect(error.getContextValue<string>("url")).toBe("http://api.example.com");
    expect(error.getContextValue<number>("status")).toBe(500);
  });

  it("should check error codes correctly", () => {
    const error = FormBuilderError.validationError("Test error");
    expect(error.isCode(ErrorCode.VALIDATION_FAILED)).toBe(true);
    expect(error.isCode(ErrorCode.NETWORK_ERROR)).toBe(false);
  });

  it("should serialize to JSON correctly", () => {
    const error = FormBuilderError.validationError("Test error", "field");
    const json = error.toJSON();

    expect(json.name).toBe("FormBuilderError");
    expect(json.code).toBe(ErrorCode.VALIDATION_FAILED);
    expect(json.userMessage).toBe("Test error");
    expect(json.context).toEqual({ field: "field" });
  });
});

describe("HttpClient", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Mock document.querySelector for CSRF token
    document.querySelector = vi.fn().mockReturnValue({
      content: "test-csrf-token",
    });
  });

  it("should add CSRF token to non-GET requests", async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      headers: new Headers(),
    });

    global.fetch = mockFetch;

    await HttpClient.post("http://api.example.com/test", "test data");

    expect(mockFetch).toHaveBeenCalledWith(
      "http://api.example.com/test",
      expect.objectContaining({
        headers: expect.any(Headers),
        method: "POST",
      }),
    );

    const headers = mockFetch.mock.calls[0][1].headers;
    expect(headers.get("X-Csrf-Token")).toBe("test-csrf-token");
  });

  it("should not add CSRF token to GET requests", async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      headers: new Headers(),
    });

    global.fetch = mockFetch;

    await HttpClient.get("http://api.example.com/test");

    expect(mockFetch).toHaveBeenCalledWith(
      "http://api.example.com/test",
      expect.objectContaining({
        headers: expect.any(Headers),
        method: "GET",
      }),
    );

    const headers = mockFetch.mock.calls[0][1].headers;
    expect(headers.get("X-Csrf-Token")).toBeNull();
  });
});

describe("DOM Utilities", () => {
  beforeEach(() => {
    // Clear DOM cache before each test
    dom.clearCache();
  });

  it("should cache DOM elements", () => {
    const element = createMockElement("test-element");
    document.body.appendChild(element);

    const found1 = dom.getElement("test-element");
    const found2 = dom.getElement("test-element");

    expect(found1).toBe(element);
    expect(found2).toBe(element);
    expect(dom.cacheSize).toBeGreaterThan(0);
  });

  it("should show error messages", () => {
    const container = createMockElement("container");
    document.body.appendChild(container);

    dom.showError("Test error message", container);

    const errorElement = container.querySelector(".gf-error-message");
    expect(errorElement).toBeTruthy();
    expect(errorElement?.textContent).toBe("Test error message");
  });

  it("should show success messages", () => {
    const container = createMockElement("container");
    document.body.appendChild(container);

    dom.showSuccess("Test success message", container);

    const successElement = container.querySelector(".gf-success-message");
    expect(successElement).toBeTruthy();
    expect(successElement?.textContent).toBe("Test success message");
  });

  it("should clear cache", () => {
    const element = createMockElement("test-element");
    document.body.appendChild(element);

    dom.getElement("test-element");
    expect(dom.cacheSize).toBeGreaterThan(0);

    dom.clearCache();
    expect(dom.cacheSize).toBe(0);
  });
});

describe("Form Types", () => {
  it("should validate form configuration", () => {
    const validConfig = {
      formId: "test-form",
      validationType: "realtime" as const,
      validationDelay: 300,
    };

    expect(validConfig.formId).toBe("test-form");
    expect(validConfig.validationType).toBe("realtime");
    expect(validConfig.validationDelay).toBe(300);
  });

  it("should handle form schema structure", () => {
    const schema = {
      display: "form",
      components: [
        {
          type: "textfield" as const,
          key: "name",
          label: "Name",
          input: true,
        },
      ],
    };

    expect(schema.display).toBe("form");
    expect(schema.components).toHaveLength(1);
    expect(schema.components[0].type).toBe("textfield");
    expect(schema.components[0].key).toBe("name");
  });
});
