import { describe, it, expect, beforeEach, vi } from "vitest";
import { HttpClient } from "@/core/http-client";

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
        text: () => Promise.resolve('{"success": true}'),
        json: () => Promise.resolve({ success: true }),
      });

      global.fetch = mockFetch;

      await HttpClient.post("http://api.example.com/test", { data: "test" });

      const headers = mockFetch.mock.calls[0][1].headers;
      expect(headers.get("X-Csrf-Token")).toBe("test-csrf-token");
    });

    it("should not add CSRF token to GET requests", async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        text: () => Promise.resolve('{"success": true}'),
        json: () => Promise.resolve({ success: true }),
      });

      global.fetch = mockFetch;

      await HttpClient.get("http://api.example.com/test");

      const headers = mockFetch.mock.calls[0][1].headers;
      expect(headers.get("X-Csrf-Token")).toBeNull();
    });

    it("should handle missing CSRF token gracefully", async () => {
      // Mock document.querySelector to return null for CSRF token
      document.querySelector = vi.fn().mockReturnValue(null);

      await expect(
        HttpClient.post("http://api.example.com/test", { data: "test" }),
      ).rejects.toThrow("CSRF token not found");

      // Restore original mock
      document.querySelector = vi.fn().mockReturnValue({
        content: "test-csrf-token",
      });
    });
  });

  describe("HTTP methods", () => {
    it("should make GET requests correctly", async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        text: () => Promise.resolve('{"data": "test"}'),
        json: () => Promise.resolve({ data: "test" }),
      });

      global.fetch = mockFetch;

      const result = await HttpClient.get("http://api.example.com/test");

      expect(mockFetch).toHaveBeenCalledWith(
        "http://api.example.com/test",
        expect.objectContaining({
          method: "GET",
        }),
      );
      expect(result).toEqual({ data: "test" });
    });

    it("should make POST requests correctly", async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        text: () => Promise.resolve('{"success": true}'),
        json: () => Promise.resolve({ success: true }),
      });

      global.fetch = mockFetch;

      const result = await HttpClient.post("http://api.example.com/test", {
        name: "test",
        email: "test@example.com",
      });

      expect(mockFetch).toHaveBeenCalledWith(
        "http://api.example.com/test",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({ name: "test", email: "test@example.com" }),
        }),
      );
      expect(result).toEqual({ success: true });
    });

    it("should make PUT requests correctly", async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        text: () => Promise.resolve('{"updated": true}'),
        json: () => Promise.resolve({ updated: true }),
      });

      global.fetch = mockFetch;

      const result = await HttpClient.put("http://api.example.com/test", {
        id: 1,
        name: "updated",
      });

      expect(mockFetch).toHaveBeenCalledWith(
        "http://api.example.com/test",
        expect.objectContaining({
          method: "PUT",
          body: JSON.stringify({ id: 1, name: "updated" }),
        }),
      );
      expect(result).toEqual({ updated: true });
    });

    it("should make DELETE requests correctly", async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        text: () => Promise.resolve(""),
        json: () => Promise.resolve(null),
      });

      global.fetch = mockFetch;

      const result = await HttpClient.delete("http://api.example.com/test");

      expect(mockFetch).toHaveBeenCalledWith(
        "http://api.example.com/test",
        expect.objectContaining({
          method: "DELETE",
        }),
      );
      expect(result).toBeNull();
    });
  });

  describe("error handling", () => {
    it("should throw FormBuilderError for network errors", async () => {
      const networkError = new Error("Network error");
      global.fetch = vi.fn().mockRejectedValue(networkError);

      await expect(
        HttpClient.get("http://api.example.com/test"),
      ).rejects.toThrow("Network error: Network error");
    });

    it("should throw FormBuilderError for HTTP error responses", async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
        statusText: "Not Found",
        url: "http://api.example.com/test",
        headers: new Headers(),
        text: () => Promise.resolve("Not found"),
      });

      global.fetch = mockFetch;

      await expect(
        HttpClient.get("http://api.example.com/test"),
      ).rejects.toThrow("HTTP 404: Not Found");
    });

    it("should throw FormBuilderError for JSON parsing errors", async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        text: () => Promise.resolve("invalid json"),
      });

      global.fetch = mockFetch;

      const result = await HttpClient.get("http://api.example.com/test");
      expect(result).toBe("invalid json");
    });
  });

  describe("headers", () => {
    it("should set correct content type for JSON requests", async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        text: () => Promise.resolve('{"success": true}'),
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
        text: () => Promise.resolve('{"success": true}'),
        json: () => Promise.resolve({ success: true }),
      });

      global.fetch = mockFetch;

      await HttpClient.get("http://api.example.com/test", {
        headers: { "X-Custom-Header": "test" },
      });

      const headers = mockFetch.mock.calls[0][1].headers;
      expect(headers.get("X-Custom-Header")).toBe("test");
    });
  });

  describe("FormData handling", () => {
    it("should handle FormData correctly", async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers(),
        text: () => Promise.resolve('{"success": true}'),
        json: () => Promise.resolve({ success: true }),
      });

      global.fetch = mockFetch;

      const formData = new FormData();
      formData.append("file", new Blob(["test"]), "test.txt");

      await HttpClient.post("http://api.example.com/upload", formData);

      const headers = mockFetch.mock.calls[0][1].headers;
      expect(headers.get("Content-Type")).not.toBe("application/json");
    });
  });
});
