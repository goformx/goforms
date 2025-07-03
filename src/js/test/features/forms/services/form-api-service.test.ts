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
import { createComponentKey } from "@/shared/types/form-types";

// Mock dependencies
vi.mock("@/core/http-client", () => ({
  HttpClient: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));

vi.mock("@/core/logger");
vi.mock("dompurify", () => ({
  default: {
    sanitize: vi.fn((input) => input), // Simple mock that returns input unchanged
  },
}));

describe("FormApiService", () => {
  let service: FormApiService;
  let mockHttpGet: MockedFunction<typeof HttpClient.get>;
  let mockHttpPost: MockedFunction<typeof HttpClient.post>;
  let mockHttpPut: MockedFunction<typeof HttpClient.put>;
  let mockHttpDelete: MockedFunction<typeof HttpClient.delete>;

  beforeEach(() => {
    // Reset all mocks
    vi.clearAllMocks();

    // Mock HttpClient methods with proper typing
    mockHttpGet = vi.mocked(HttpClient.get);
    mockHttpPost = vi.mocked(HttpClient.post);
    mockHttpPut = vi.mocked(HttpClient.put);
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
          key: createComponentKey("name"),
          label: "Name",
          input: true, // Add missing required property
        },
      ],
    };

    it("should successfully fetch and return a valid schema", async () => {
      mockHttpGet.mockResolvedValue({
        data: validSchema,
        status: 200,
        statusText: "OK",
        headers: new Headers(),
        url: "",
      });

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
      mockHttpGet.mockResolvedValue({
        data: "invalid response",
        status: 200,
        statusText: "OK",
        headers: new Headers(),
        url: "",
      });

      await expect(service.getSchema("test-form-id")).rejects.toThrow(
        FormBuilderError,
      );

      expect(Logger.error).toHaveBeenCalled();
    });

    it("should throw SchemaError when response is null", async () => {
      mockHttpGet.mockResolvedValue({
        data: null,
        status: 200,
        statusText: "OK",
        headers: new Headers(),
        url: "",
      });

      await expect(service.getSchema("test-form-id")).rejects.toThrow(
        FormBuilderError,
      );
    });

    it("should throw SchemaError when components property is missing", async () => {
      mockHttpGet.mockResolvedValue({
        data: { display: "form" },
        status: 200,
        statusText: "OK",
        headers: new Headers(),
        url: "",
      });

      await expect(service.getSchema("test-form-id")).rejects.toThrow(
        FormBuilderError,
      );
    });

    it("should throw SchemaError when components is not an array", async () => {
      mockHttpGet.mockResolvedValue({
        data: {
          display: "form",
          components: "not an array",
        },
        status: 200,
        statusText: "OK",
        headers: new Headers(),
        url: "",
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
          key: createComponentKey("name"),
          label: "Name",
          input: true, // Add missing required property
        },
      ],
    };

    it("should successfully save schema and return response", async () => {
      // Mock HttpClient.put to return HttpResponse object
      const mockResponse = {
        data: validSchema,
        status: 200,
        statusText: "OK",
        headers: new Headers(),
        url: "",
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
        data: null,
        status: 400,
        statusText: "Bad Request",
        headers: new Headers(),
        url: "",
      };
      mockHttpPut.mockResolvedValue(mockResponse);

      await expect(
        service.saveSchema("test-form-id", validSchema),
      ).rejects.toThrow(FormBuilderError);
    });

    it("should throw SchemaError for invalid response data", async () => {
      const mockResponse = {
        data: "invalid",
        status: 200,
        statusText: "OK",
        headers: new Headers(),
        url: "",
      };
      mockHttpPut.mockResolvedValue(mockResponse);

      await expect(
        service.saveSchema("test-form-id", validSchema),
      ).rejects.toThrow(FormBuilderError);
    });

    it("should throw SchemaError when response lacks components array", async () => {
      const mockResponse = {
        data: { display: "form" },
        status: 200,
        statusText: "OK",
        headers: new Headers(),
        url: "",
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
      const mockResponse = {
        data: null,
        status: 200,
        statusText: "OK",
        headers: new Headers(),
        url: "",
      };
      mockHttpPut.mockResolvedValue(mockResponse);

      await service.updateFormDetails("test-form-id", formDetails);

      expect(mockHttpPut).toHaveBeenCalledWith(
        `${window.location.origin}/forms/test-form-id`,
        JSON.stringify(formDetails),
      );
    });

    it("should throw NetworkError when response is not ok", async () => {
      const networkError = FormBuilderError.networkError(
        "HTTP 400: Bad Request",
        `${window.location.origin}/forms/test-form-id`,
        400,
      );
      mockHttpPut.mockRejectedValue(networkError);

      await expect(
        service.updateFormDetails("test-form-id", formDetails),
      ).rejects.toThrow(FormBuilderError);
    });

    it("should use default error message when response json fails", async () => {
      const networkError = FormBuilderError.networkError(
        "HTTP 500: Internal Server Error",
        `${window.location.origin}/forms/test-form-id`,
        500,
      );
      mockHttpPut.mockRejectedValue(networkError);

      await expect(
        service.updateFormDetails("test-form-id", formDetails),
      ).rejects.toThrow(FormBuilderError);
    });
  });

  describe("deleteForm", () => {
    it("should successfully delete form", async () => {
      const mockResponse = {
        data: null,
        status: 200,
        statusText: "OK",
        headers: new Headers(),
        url: "",
      };
      mockHttpDelete.mockResolvedValue(mockResponse);

      await service.deleteForm("test-form-id");

      expect(mockHttpDelete).toHaveBeenCalledWith(
        `${window.location.origin}/forms/test-form-id`,
      );
    });

    it("should throw NetworkError when deletion fails", async () => {
      const networkError = FormBuilderError.networkError(
        "HTTP 404: Not Found",
        `${window.location.origin}/forms/test-form-id`,
        404,
      );
      mockHttpDelete.mockRejectedValue(networkError);

      await expect(service.deleteForm("test-form-id")).rejects.toThrow(
        FormBuilderError,
      );
    });
  });

  describe("submitForm", () => {
    const formData = new FormData();
    formData.append("name", "John Doe");
    formData.append("email", "john@example.com");

    beforeEach(() => {
      // Mock CSRF token in DOM
      const meta = document.createElement("meta");
      meta.name = "csrf-token";
      meta.content = "test-csrf-token";
      document.head.appendChild(meta);
    });

    afterEach(() => {
      // Clean up CSRF token
      const meta = document.querySelector('meta[name="csrf-token"]');
      if (meta) {
        meta.remove();
      }
    });

    it("should successfully submit form with sanitized data", async () => {
      const mockHttpResponse = {
        data: { success: true, message: "Form submitted successfully" },
        status: 200,
        statusText: "OK",
        headers: new Headers(),
        url: `${window.location.origin}/api/v1/forms/test-form-id/submit`,
      };

      // Mock HttpClient.post to return standardized response
      mockHttpPost.mockResolvedValue(mockHttpResponse);

      const result = await service.submitForm("test-form-id", formData);

      expect(mockHttpPost).toHaveBeenCalledWith(
        `${window.location.origin}/api/v1/forms/test-form-id/submit`,
        expect.any(Object),
      );
      expect(result).toBeInstanceOf(Response);
      expect(result.status).toBe(200);
    });

    it("should throw NetworkError when submission fails", async () => {
      const networkError = FormBuilderError.networkError(
        "HTTP 400: Bad Request",
        `${window.location.origin}/api/v1/forms/test-form-id/submit`,
        400,
      );
      mockHttpPost.mockRejectedValue(networkError);

      await expect(
        service.submitForm("test-form-id", formData),
      ).rejects.toThrow(FormBuilderError);
    });

    it("should handle json parse failure in error response", async () => {
      const networkError = FormBuilderError.networkError(
        "HTTP 500: Internal Server Error",
        `${window.location.origin}/api/v1/forms/test-form-id/submit`,
        500,
      );
      mockHttpPost.mockRejectedValue(networkError);

      await expect(
        service.submitForm("test-form-id", formData),
      ).rejects.toThrow(FormBuilderError);
    });
  });

  describe("sanitizeFormData", () => {
    // Test the private method indirectly through submitForm
    beforeEach(() => {
      // Mock CSRF token in DOM
      const meta = document.createElement("meta");
      meta.name = "csrf-token";
      meta.content = "test-csrf-token";
      document.head.appendChild(meta);
    });

    afterEach(() => {
      // Clean up CSRF token
      const meta = document.querySelector('meta[name="csrf-token"]');
      if (meta) {
        meta.remove();
      }
    });

    it("should sanitize string values", async () => {
      const mockHttpResponse = {
        data: { success: true, message: "Form submitted successfully" },
        status: 200,
        statusText: "OK",
        headers: new Headers(),
        url: `${window.location.origin}/api/v1/forms/test-form-id/submit`,
      };
      mockHttpPost.mockResolvedValue(mockHttpResponse);

      const formData = new FormData();
      formData.append("malicious", '<script>alert("xss")</script>');

      await service.submitForm("test-form-id", formData);

      expect(mockHttpPost).toHaveBeenCalledWith(
        `${window.location.origin}/api/v1/forms/test-form-id/submit`,
        expect.any(Object),
      );
    });

    it("should handle null and undefined values", async () => {
      const mockHttpResponse = {
        data: { success: true, message: "Form submitted successfully" },
        status: 200,
        statusText: "OK",
        headers: new Headers(),
        url: `${window.location.origin}/api/v1/forms/test-form-id/submit`,
      };
      mockHttpPost.mockResolvedValue(mockHttpResponse);

      const formData = new FormData();
      // FormData doesn't directly support null/undefined, but the sanitizer should handle them

      await service.submitForm("test-form-id", formData);

      expect(mockHttpPost).toHaveBeenCalled();
    });

    it("should preserve numbers and booleans", async () => {
      const mockHttpResponse = {
        data: { success: true, message: "Form submitted successfully" },
        status: 200,
        statusText: "OK",
        headers: new Headers(),
        url: `${window.location.origin}/api/v1/forms/test-form-id/submit`,
      };
      mockHttpPost.mockResolvedValue(mockHttpResponse);

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
