import { describe, it, expect } from "vitest";
import { FormBuilderError, ErrorCode } from "@/core/errors/form-builder-error";

describe("FormBuilderError", () => {
  describe("static factory methods", () => {
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
      expect(error.name).toBe("FormBuilderError");
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
      expect(error.getContextValue<string>("url")).toBe(
        "http://api.example.com",
      );
      expect(error.getContextValue<number>("status")).toBe(500);
    });

    it("should create schema errors correctly", () => {
      const error = FormBuilderError.schemaError(
        "Invalid schema format",
        "components",
      );

      expect(error.code).toBe(ErrorCode.SCHEMA_ERROR);
      expect(error.userMessage).toBe(
        "Form configuration error. Please contact support.",
      );
      expect(error.getContextValue<string>("schemaId")).toBe("components");
    });

    it("should create CSRF errors correctly", () => {
      const error = FormBuilderError.csrfError("Token expired");

      expect(error.code).toBe(ErrorCode.CSRF_ERROR);
      expect(error.userMessage).toBe(
        "Security token expired. Please refresh the page and try again.",
      );
    });

    it("should create sanitization errors correctly", () => {
      const error = FormBuilderError.sanitizationError(
        "Invalid input detected",
        "email",
        "<script>alert('xss')</script>",
      );

      expect(error.code).toBe(ErrorCode.SANITIZATION_ERROR);
      expect(error.userMessage).toBe(
        "Invalid input detected. Please check your input and try again.",
      );
      expect(error.getContextValue<string>("field")).toBe("email");
      expect(error.getContextValue<unknown>("value")).toBe(
        "<script>alert('xss')</script>",
      );
    });

    it("should create load failed errors correctly", () => {
      const originalError = new Error("Original error");
      const error = FormBuilderError.loadFailed(
        "Failed to load form",
        "form-123",
        originalError,
      );

      expect(error.code).toBe(ErrorCode.LOAD_FAILED);
      expect(error.userMessage).toBe(
        "Failed to load resource. Please try again.",
      );
      expect(error.getContextValue<string>("resourceId")).toBe("form-123");
      expect(error.originalError).toBe(originalError);
    });

    it("should create save failed errors correctly", () => {
      const originalError = new Error("Original error");
      const error = FormBuilderError.saveFailed(
        "Failed to save form",
        "form-123",
        originalError,
      );

      expect(error.code).toBe(ErrorCode.SAVE_FAILED);
      expect(error.userMessage).toBe(
        "Failed to save changes. Please try again.",
      );
      expect(error.getContextValue<string>("resourceId")).toBe("form-123");
      expect(error.originalError).toBe(originalError);
    });
  });

  describe("error code checking", () => {
    it("should check error codes correctly", () => {
      const validationError = FormBuilderError.validationError("Test error");
      const networkError = FormBuilderError.networkError("Test error");

      expect(validationError.isCode(ErrorCode.VALIDATION_FAILED)).toBe(true);
      expect(validationError.isCode(ErrorCode.NETWORK_ERROR)).toBe(false);
      expect(networkError.isCode(ErrorCode.NETWORK_ERROR)).toBe(true);
      expect(networkError.isCode(ErrorCode.VALIDATION_FAILED)).toBe(false);
    });
  });

  describe("context management", () => {
    it("should store and retrieve context values from factory methods", () => {
      const error = FormBuilderError.validationError(
        "Test error",
        "field",
        "value",
      );

      expect(error.getContextValue<string>("field")).toBe("field");
      expect(error.getContextValue<unknown>("value")).toBe("value");
    });

    it("should return undefined for missing context keys", () => {
      const error = FormBuilderError.validationError("Test error");

      expect(error.getContextValue<string>("missingKey")).toBeUndefined();
    });

    it("should handle context with constructor", () => {
      const context = { userId: "123", timestamp: Date.now() };
      const error = new FormBuilderError(
        "Test error",
        ErrorCode.VALIDATION_FAILED,
        "User message",
        context,
      );

      expect(error.getContextValue<string>("userId")).toBe("123");
      expect(error.getContextValue<number>("timestamp")).toBeTypeOf("number");
    });
  });

  describe("serialization", () => {
    it("should serialize to JSON correctly", () => {
      const error = FormBuilderError.validationError("Test error", "field");

      const json = error.toJSON();

      expect(json.name).toBe("FormBuilderError");
      expect(json.code).toBe(ErrorCode.VALIDATION_FAILED);
      expect(json.userMessage).toBe("Test error");
      expect(json.context).toEqual({ field: "field" });
      expect(json.stack).toBeDefined();
    });

    it("should include stack trace in serialization", () => {
      const error = FormBuilderError.validationError("Test error");
      const json = error.toJSON();

      expect(json.stack).toBeDefined();
      expect(typeof json.stack).toBe("string");
      expect(json.stack).toContain("FormBuilderError");
    });

    it("should include original error in serialization", () => {
      const originalError = new Error("Original error");
      const error = FormBuilderError.loadFailed(
        "Test error",
        "resource-123",
        originalError,
      );
      const json = error.toJSON();

      expect(json.originalError).toBe("Original error");
    });
  });

  describe("error inheritance", () => {
    it("should be an instance of Error", () => {
      const error = FormBuilderError.validationError("Test error");

      expect(error).toBeInstanceOf(Error);
      expect(error).toBeInstanceOf(FormBuilderError);
    });

    it("should have proper error properties", () => {
      const error = FormBuilderError.validationError("Test error");

      expect(error.message).toBe("Validation failed: Test error");
      expect(error.name).toBe("FormBuilderError");
      expect(error.stack).toBeDefined();
    });
  });

  describe("error code constants", () => {
    it("should have all required error codes", () => {
      expect(ErrorCode.VALIDATION_FAILED).toBe("VALIDATION_FAILED");
      expect(ErrorCode.NETWORK_ERROR).toBe("NETWORK_ERROR");
      expect(ErrorCode.SCHEMA_ERROR).toBe("SCHEMA_ERROR");
      expect(ErrorCode.PERMISSION_DENIED).toBe("PERMISSION_DENIED");
      expect(ErrorCode.SANITIZATION_ERROR).toBe("SANITIZATION_ERROR");
      expect(ErrorCode.CSRF_ERROR).toBe("CSRF_ERROR");
      expect(ErrorCode.FORM_NOT_FOUND).toBe("FORM_NOT_FOUND");
      expect(ErrorCode.INVALID_INPUT).toBe("INVALID_INPUT");
      expect(ErrorCode.SAVE_FAILED).toBe("SAVE_FAILED");
      expect(ErrorCode.LOAD_FAILED).toBe("LOAD_FAILED");
    });
  });
});
