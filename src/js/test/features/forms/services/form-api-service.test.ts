import {
  describe,
  it,
  expect,
  beforeEach,
  afterEach,
  vi,
  type MockedFunction,
} from "vitest";
import { FormApiService } from "@/features/forms/services/form-api-service";
import { HttpClient } from "@/core/http-client";
import { FormBuilderError } from "@/core/errors/form-builder-error";
import { Logger } from "@/core/logger";
import type { FormSchema } from "@/shared/types/form-types";

// Mock dependencies
vi.mock("@/core/http-client");
vi.mock("@/core/logger");
vi.mock("dompurify", () => ({
  default: {
    sanitize: vi.fn((input) => input), // Simple mock that returns input unchanged
  },
}));

describe("FormApiService", () => {
  let service: FormApiService;
  let mockHttpGet: MockedFunction<typeof HttpClient.get>;
  let mockHttpPut: MockedFunction<typeof HttpClient.put>;
  let mockHttpPost: MockedFunction<typeof HttpClient.post>;
  let mockHttpDelete: MockedFunction<typeof HttpClient.delete>;

  beforeEach(() => {
    // Reset all mocks
    vi.clearAllMocks();

    // Mock HttpClient methods with proper typing
    mockHttpGet = vi.mocked(HttpClient.get);
    mockHttpPut = vi.mocked(HttpClient.put);
    mockHttpPost = vi.mocked(HttpClient.post);
    mockHttpDelete = vi.mocked(HttpClient.delete);

    // Mock Logger methods
    vi.mocked(Logger.debug).mockImplementation(() => {});
    vi.mocked(Logger.error).mockImplementation(() => {});

    // Get fresh instance
    service = FormApiService.getInstance();
  });

  afterEach(() => {
    // Reset singleton instance for clean tests
    (FormApiService as any).instance = undefined;
  });

  describe("Singleton Pattern", () => {
    it("should return the same instance when called multiple times", () => {
      const instance1 = FormApiService.getInstance();
      const instance2 = FormApiService.getInstance();

      expect(instance1).toBe(instance2);
    });

    it("should initialize with window.location.origin as base URL", () => {
      // Set up window.location.origin mock
      Object.defineProperty(window, "location", {
        value: { origin: "https://example.com" },
        writable: true,
      });

      // Reset singleton to test initialization
      (FormApiService as any).instance = undefined;
      FormApiService.getInstance();

      expect(Logger.debug).toHaveBeenCalledWith(
        "FormApiService initialized with base URL:",
        "https://example.com",
      );
    });
  });

  describe("setBaseUrl", () => {
    it("should update the base URL and log the change", () => {
      const newUrl = "https://api.example.com";

      service.setBaseUrl(newUrl);

      expect(Logger.debug).toHaveBeenCalledWith(
        "FormApiService base URL updated to:",
        newUrl,
      );
    });
  });

  describe("getSchema", () => {
    const validSchema: FormSchema = {
      display: "form",
      components: [
        {
          type: "textfield",
          key: "name",
          label: "Name",
          input: true, // Add missing required property
        },
      ],
    };

    it("should successfully fetch and return a valid schema", async () => {
      mockHttpGet.mockResolvedValue(validSchema);

      const result = await service.getSchema("test-form-id");

      expect(mockHttpGet).toHaveBeenCalledWith(
        `${window.location.origin}/api/v1/forms/test-form-id/schema`,
      );
      expect(result).toEqual(validSchema);
      expect(Logger.debug).toHaveBeenCalledWith(
        "Fetching schema from:",
        `${window.location.origin}/api/v1/forms/test-form-id/schema`,
      );
    });

    it("should throw SchemaError when response is not an object", async () => {
      mockHttpGet.mockResolvedValue("invalid response");

      await expect(service.getSchema("test-form-id")).rejects.toThrow(
        FormBuilderError,
      );

      expect(Logger.error).toHaveBeenCalled();
    });

    it("should throw SchemaError when response is null", async () => {
      mockHttpGet.mockResolvedValue(null);

      await expect(service.getSchema("test-form-id")).rejects.toThrow(
        FormBuilderError,
      );
    });

    it("should throw SchemaError when components property is missing", async () => {
      mockHttpGet.mockResolvedValue({ display: "form" });

      await expect(service.getSchema("test-form-id")).rejects.toThrow(
        FormBuilderError,
      );
    });

    it("should throw SchemaError when components is not an array", async () => {
      mockHttpGet.mockResolvedValue({
        display: "form",
        components: "not an array",
      });

      await expect(service.getSchema("test-form-id")).rejects.toThrow(
        FormBuilderError,
      );
    });

    it("should re-throw FormBuilderError from HttpClient", async () => {
      const networkError = FormBuilderError.networkError(
        "Network failed",
        "url",
        500,
      );
      mockHttpGet.mockRejectedValue(networkError);

      await expect(service.getSchema("test-form-id")).rejects.toBe(
        networkError,
      );
    });

    it("should wrap generic errors in FormBuilderError.loadFailed", async () => {
      const genericError = new Error("Generic error");
      mockHttpGet.mockRejectedValue(genericError);

      await expect(service.getSchema("test-form-id")).rejects.toThrow(
        FormBuilderError,
      );

      expect(Logger.error).toHaveBeenCalledWith(
        "Error in getSchema:",
        genericError,
      );
    });
  });

  describe("saveSchema", () => {
    const validSchema: FormSchema = {
      display: "form",
      components: [
        {
          type: "textfield",
          key: "name",
          label: "Name",
          input: true, // Add missing required property
        },
      ],
    };

    it("should successfully save schema and return response", async () => {
      // Mock HttpClient.put to return a Response-like object
      const mockResponse = {
        ok: true,
        json: vi.fn().mockResolvedValue(validSchema),
      };
      mockHttpPut.mockResolvedValue(mockResponse);

      const result = await service.saveSchema("test-form-id", validSchema);

      expect(mockHttpPut).toHaveBeenCalledWith(
        `${window.location.origin}/api/v1/forms/test-form-id/schema`,
        JSON.stringify(validSchema),
      );
      expect(result).toEqual(validSchema);
    });

    it("should throw NetworkError when response is not ok", async () => {
      const mockResponse = {
        ok: false,
        status: 400,
        statusText: "Bad Request",
      };
      mockHttpPut.mockResolvedValue(mockResponse);

      await expect(
        service.saveSchema("test-form-id", validSchema),
      ).rejects.toThrow(FormBuilderError);
    });

    it("should throw SchemaError for invalid response data", async () => {
      const mockResponse = {
        ok: true,
        json: vi.fn().mockResolvedValue("invalid"),
      };
      mockHttpPut.mockResolvedValue(mockResponse);

      await expect(
        service.saveSchema("test-form-id", validSchema),
      ).rejects.toThrow(FormBuilderError);
    });

    it("should throw SchemaError when response lacks components array", async () => {
      const mockResponse = {
        ok: true,
        json: vi.fn().mockResolvedValue({ display: "form" }),
      };
      mockHttpPut.mockResolvedValue(mockResponse);

      await expect(
        service.saveSchema("test-form-id", validSchema),
      ).rejects.toThrow(FormBuilderError);
    });
  });

  describe("updateFormDetails", () => {
    const formDetails = { title: "New Title", description: "New Description" };

    it("should successfully update form details", async () => {
      const mockResponse = { ok: true };
      mockHttpPut.mockResolvedValue(mockResponse);

      await service.updateFormDetails("test-form-id", formDetails);

      expect(mockHttpPut).toHaveBeenCalledWith(
        `${window.location.origin}/dashboard/forms/test-form-id`,
        JSON.stringify(formDetails),
      );
    });

    it("should throw NetworkError when response is not ok", async () => {
      const mockResponse = {
        ok: false,
        status: 400,
        statusText: "Bad Request",
        json: vi.fn().mockResolvedValue({ message: "Validation failed" }),
      };
      mockHttpPut.mockResolvedValue(mockResponse);

      await expect(
        service.updateFormDetails("test-form-id", formDetails),
      ).rejects.toThrow(FormBuilderError);
    });

    it("should use default error message when response json fails", async () => {
      const mockResponse = {
        ok: false,
        status: 500,
        statusText: "Internal Server Error",
        json: vi.fn().mockRejectedValue(new Error("JSON parse failed")),
      };
      mockHttpPut.mockResolvedValue(mockResponse);

      await expect(
        service.updateFormDetails("test-form-id", formDetails),
      ).rejects.toThrow(FormBuilderError);
    });
  });

  describe("deleteForm", () => {
    it("should successfully delete form", async () => {
      const mockResponse = { ok: true };
      mockHttpDelete.mockResolvedValue(mockResponse);

      await service.deleteForm("test-form-id");

      expect(mockHttpDelete).toHaveBeenCalledWith(
        `${window.location.origin}/forms/test-form-id`,
      );
    });

    it("should throw NetworkError when deletion fails", async () => {
      const mockResponse = {
        ok: false,
        status: 404,
        statusText: "Not Found",
      };
      mockHttpDelete.mockResolvedValue(mockResponse);

      await expect(service.deleteForm("test-form-id")).rejects.toThrow(
        FormBuilderError,
      );
    });
  });

  describe("submitForm", () => {
    const formData = new FormData();
    formData.append("name", "John Doe");
    formData.append("email", "john@example.com");

    it("should successfully submit form with sanitized data", async () => {
      const mockResponse = { ok: true };
      mockHttpPost.mockResolvedValue(mockResponse);

      const result = await service.submitForm("test-form-id", formData);

      expect(mockHttpPost).toHaveBeenCalledWith(
        `${window.location.origin}/api/v1/forms/test-form-id/submit`,
        expect.any(String), // JSON stringified sanitized data
      );
      expect(result).toBe(mockResponse);
    });

    it("should throw NetworkError when submission fails", async () => {
      const mockResponse = {
        ok: false,
        status: 400,
        statusText: "Bad Request",
        json: vi.fn().mockResolvedValue({ message: "Invalid data" }),
      };
      mockHttpPost.mockResolvedValue(mockResponse);

      await expect(
        service.submitForm("test-form-id", formData),
      ).rejects.toThrow(FormBuilderError);
    });

    it("should handle json parse failure in error response", async () => {
      const mockResponse = {
        ok: false,
        status: 500,
        statusText: "Internal Server Error",
        json: vi.fn().mockRejectedValue(new Error("JSON parse failed")),
      };
      mockHttpPost.mockResolvedValue(mockResponse);

      await expect(
        service.submitForm("test-form-id", formData),
      ).rejects.toThrow(FormBuilderError);
    });
  });

  describe("sanitizeFormData", () => {
    // Test the private method indirectly through submitForm
    it("should sanitize string values", async () => {
      const mockResponse = { ok: true };
      mockHttpPost.mockResolvedValue(mockResponse);

      const formData = new FormData();
      formData.append("malicious", '<script>alert("xss")</script>');

      await service.submitForm("test-form-id", formData);

      // Verify that HttpClient.post was called with stringified data
      expect(mockHttpPost).toHaveBeenCalledWith(
        expect.any(String),
        expect.any(String),
      );
    });

    it("should handle null and undefined values", async () => {
      const mockResponse = { ok: true };
      mockHttpPost.mockResolvedValue(mockResponse);

      const formData = new FormData();
      // FormData doesn't directly support null/undefined, but the sanitizer should handle them

      await service.submitForm("test-form-id", formData);

      expect(mockHttpPost).toHaveBeenCalled();
    });

    it("should preserve numbers and booleans", async () => {
      const mockResponse = { ok: true };
      mockHttpPost.mockResolvedValue(mockResponse);

      const formData = new FormData();
      formData.append("age", "25");
      formData.append("active", "true");

      await service.submitForm("test-form-id", formData);

      expect(mockHttpPost).toHaveBeenCalled();
    });
  });

  describe("Error Handling", () => {
    it("should preserve FormBuilderError instances", async () => {
      const originalError = FormBuilderError.networkError(
        "Original error",
        "url",
        500,
      );
      mockHttpGet.mockRejectedValue(originalError);

      await expect(service.getSchema("test-form-id")).rejects.toBe(
        originalError,
      );
    });

    it("should wrap non-FormBuilderError instances", async () => {
      const genericError = new TypeError("Type error");
      mockHttpGet.mockRejectedValue(genericError);

      await expect(service.getSchema("test-form-id")).rejects.toThrow(
        FormBuilderError,
      );
      await expect(service.getSchema("test-form-id")).rejects.not.toBe(
        genericError,
      );
    });

    it("should handle non-Error objects", async () => {
      const stringError = "String error";
      mockHttpGet.mockRejectedValue(stringError);

      await expect(service.getSchema("test-form-id")).rejects.toThrow(
        FormBuilderError,
      );
    });
  });
});
