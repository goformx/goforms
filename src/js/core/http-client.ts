import { Logger } from "@/core/logger";
import { FormBuilderError } from "@/core/errors/form-builder-error";

/**
 * Unified HTTP client for making authenticated requests
 */
export class HttpClient {
  private static getCSRFToken(): string {
    const meta = document.querySelector<HTMLMetaElement>(
      "meta[name='csrf-token']",
    );
    if (!meta?.content) {
      throw new Error("Catastrophic. CSRF token not found.");
    }
    return meta.content;
  }

  private static async handleResponse(response: Response): Promise<any> {
    Logger.log("Response status:", response.status);
    Logger.log(
      "Response headers:",
      Object.fromEntries(response.headers.entries()),
    );

    if (!response.ok) {
      throw FormBuilderError.networkError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.url,
        response.status,
      );
    }

    // Try to parse JSON, fallback to text for empty responses
    const text = await response.text();
    if (!text) {
      return null;
    }

    try {
      return JSON.parse(text);
    } catch (parseError) {
      // If JSON parsing fails, return the text
      Logger.warn("Failed to parse JSON response, returning text:", text);
      Logger.debug("Parse error details:", parseError);
      return text;
    }
  }

  private static async makeRequest(
    url: string,
    options: RequestInit = {},
  ): Promise<any> {
    Logger.group("HTTP Request");
    Logger.log("URL:", url);
    Logger.log("Method:", options.method || "GET");

    try {
      const headers = new Headers(options.headers);

      // Add CSRF token for non-GET requests
      if (options.method !== "GET") {
        try {
          const csrfToken = this.getCSRFToken();
          headers.set("X-Csrf-Token", csrfToken);
          Logger.log("CSRF Token status verified");
        } catch (error) {
          Logger.error("CSRF token error:", error);
          throw FormBuilderError.networkError("CSRF token not found", url);
        }
      }

      // Set content type if not FormData
      if (!(options.body instanceof FormData)) {
        headers.set("Content-Type", "application/json");
      }

      // Add common headers
      headers.set("Accept", "application/json");
      headers.set("X-Requested-With", "XMLHttpRequest");

      Logger.log("Headers:", Object.fromEntries(headers.entries()));
      Logger.log("Cookies:", document.cookie);

      const response = await fetch(url, {
        ...options,
        headers,
        credentials: "include",
      });

      return await this.handleResponse(response);
    } catch (error) {
      Logger.error("Request failed:", error);

      // If it's already a FormBuilderError, re-throw it
      if (error instanceof FormBuilderError) {
        throw error;
      }

      // Convert other errors to FormBuilderError
      throw FormBuilderError.networkError(
        `Network error: ${error instanceof Error ? error.message : String(error)}`,
        url,
      );
    } finally {
      Logger.groupEnd();
    }
  }

  static async get(url: string, options: RequestInit = {}): Promise<any> {
    return this.makeRequest(url, { ...options, method: "GET" });
  }

  static async post(
    url: string,
    body?: FormData | string | object,
    options: RequestInit = {},
  ): Promise<any> {
    let requestBody: FormData | string | undefined = body as
      | FormData
      | string
      | undefined;

    // Convert objects to JSON string
    if (body && typeof body === "object" && !(body instanceof FormData)) {
      requestBody = JSON.stringify(body);
    }

    return this.makeRequest(url, {
      ...options,
      method: "POST",
      body: requestBody,
    });
  }

  static async put(
    url: string,
    body?: FormData | string | object,
    options: RequestInit = {},
  ): Promise<any> {
    let requestBody: FormData | string | undefined = body as
      | FormData
      | string
      | undefined;

    // Convert objects to JSON string
    if (body && typeof body === "object" && !(body instanceof FormData)) {
      requestBody = JSON.stringify(body);
    }

    return this.makeRequest(url, {
      ...options,
      method: "PUT",
      body: requestBody,
    });
  }

  static async delete(url: string, options: RequestInit = {}): Promise<any> {
    return this.makeRequest(url, { ...options, method: "DELETE" });
  }
}
