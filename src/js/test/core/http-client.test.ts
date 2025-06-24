import { describe, it, expect, beforeEach, vi } from "vitest";
import { HttpClient } from "@/core/http-client";
import { FormBuilderError } from "@/core/errors/form-builder-error";

describe("HttpClient", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Mock document.querySelector for CSRF token
    document.querySelector = vi.fn().mockReturnValue({
      content: "test-csrf-token",
    });
  });

  describe("CSRF token handling", () => {
    it("should add CSRF token to non-GET requests", async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        json: () => Promise.resolve({ success: true }),
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
        json: () => Promise.resolve({ data: "test" }),
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

    it("should handle missing CSRF token gracefully", async () => {
      document.querySelector = vi.fn().mockReturnValue(null);

      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        json: () => Promise.resolve({ success: true }),
      });

      global.fetch = mockFetch;

      await HttpClient.post("http://api.example.com/test", "test data");

      const headers = mockFetch.mock.calls[0][1].headers;
      expect(headers.get("X-Csrf-Token")).toBeNull();
    });
  });

  describe("HTTP methods", () => {
    it("should make GET requests correctly", async () => {
      const mockResponse = { data: "test" };
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        json: () => Promise.resolve(mockResponse),
      });

      global.fetch = mockFetch;

      const result = await HttpClient.get("http://api.example.com/test");

      expect(mockFetch).toHaveBeenCalledWith(
        "http://api.example.com/test",
        expect.objectContaining({
          method: "GET",
          headers: expect.any(Headers),
        }),
      );
      expect(result).toEqual(mockResponse);
    });

    it("should make POST requests correctly", async () => {
      const mockResponse = { success: true };
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        json: () => Promise.resolve(mockResponse),
      });

      global.fetch = mockFetch;

      const postData = { name: "test", email: "test@example.com" };
      const result = await HttpClient.post(
        "http://api.example.com/test",
        JSON.stringify(postData),
      );

      expect(mockFetch).toHaveBeenCalledWith(
        "http://api.example.com/test",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify(postData),
          headers: expect.any(Headers),
        }),
      );
      expect(result).toEqual(mockResponse);
    });

    it("should make PUT requests correctly", async () => {
      const mockResponse = { updated: true };
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        json: () => Promise.resolve(mockResponse),
      });

      global.fetch = mockFetch;

      const putData = { id: 1, name: "updated" };
      const result = await HttpClient.put(
        "http://api.example.com/test/1",
        JSON.stringify(putData),
      );

      expect(mockFetch).toHaveBeenCalledWith(
        "http://api.example.com/test/1",
        expect.objectContaining({
          method: "PUT",
          body: JSON.stringify(putData),
          headers: expect.any(Headers),
        }),
      );
      expect(result).toEqual(mockResponse);
    });

    it("should make DELETE requests correctly", async () => {
      const mockResponse = { deleted: true };
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        json: () => Promise.resolve(mockResponse),
      });

      global.fetch = mockFetch;

      const result = await HttpClient.delete("http://api.example.com/test/1");

      expect(mockFetch).toHaveBeenCalledWith(
        "http://api.example.com/test/1",
        expect.objectContaining({
          method: "DELETE",
          headers: expect.any(Headers),
        }),
      );
      expect(result).toEqual(mockResponse);
    });
  });

  describe("error handling", () => {
    it("should throw FormBuilderError for network errors", async () => {
      const networkError = new Error("Network error");
      global.fetch = vi.fn().mockRejectedValue(networkError);

      await expect(
        HttpClient.get("http://api.example.com/test"),
      ).rejects.toThrow(FormBuilderError);

      await expect(
        HttpClient.get("http://api.example.com/test"),
      ).rejects.toMatchObject({
        code: "NETWORK_ERROR",
        userMessage:
          "Network connection failed. Please check your internet connection and try again.",
      });
    });

    it("should throw FormBuilderError for HTTP error responses", async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        statusText: "Internal Server Error",
        headers: new Headers(),
        json: () => Promise.resolve({ error: "Server error" }),
      });

      global.fetch = mockFetch;

      await expect(
        HttpClient.get("http://api.example.com/test"),
      ).rejects.toThrow(FormBuilderError);

      await expect(
        HttpClient.get("http://api.example.com/test"),
      ).rejects.toMatchObject({
        code: "NETWORK_ERROR",
        userMessage:
          "Network connection failed. Please check your internet connection and try again.",
      });
    });

    it("should throw FormBuilderError for JSON parsing errors", async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        json: () => Promise.reject(new Error("Invalid JSON")),
      });

      global.fetch = mockFetch;

      await expect(
        HttpClient.get("http://api.example.com/test"),
      ).rejects.toThrow(FormBuilderError);
    });
  });

  describe("headers", () => {
    it("should set correct content type for JSON requests", async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        json: () => Promise.resolve({ success: true }),
      });

      global.fetch = mockFetch;

      await HttpClient.post(
        "http://api.example.com/test",
        JSON.stringify({ data: "test" }),
      );

      const headers = mockFetch.mock.calls[0][1].headers;
      expect(headers.get("Content-Type")).toBe("application/json");
    });

    it("should include custom headers", async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        json: () => Promise.resolve({ success: true }),
      });

      global.fetch = mockFetch;

      const customHeaders = { "X-Custom-Header": "custom-value" };
      await HttpClient.get("http://api.example.com/test", {
        headers: customHeaders,
      });

      const headers = mockFetch.mock.calls[0][1].headers;
      expect(headers.get("X-Custom-Header")).toBe("custom-value");
    });
  });

  describe("FormData handling", () => {
    it("should handle FormData correctly", async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        json: () => Promise.resolve({ success: true }),
      });

      global.fetch = mockFetch;

      const formData = new FormData();
      formData.append("file", new Blob(["test"]), "test.txt");
      formData.append("name", "test");

      await HttpClient.post("http://api.example.com/upload", formData);

      const callArgs = mockFetch.mock.calls[0];
      expect(callArgs[1].body).toBeInstanceOf(FormData);
      expect(callArgs[1].headers.get("Content-Type")).toBeNull(); // Let browser set it
    });
  });
});
