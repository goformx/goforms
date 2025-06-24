import { Logger } from "@/core/logger";

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

  static async request(
    url: string,
    options: RequestInit = {},
  ): Promise<Response> {
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
          throw error;
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

      Logger.log("Response status:", response.status);
      Logger.log(
        "Response headers:",
        Object.fromEntries(response.headers.entries()),
      );

      return response;
    } catch (error) {
      Logger.error("Request failed:", error);
      throw error;
    } finally {
      Logger.groupEnd();
    }
  }

  static async get(url: string, options: RequestInit = {}): Promise<Response> {
    return this.request(url, { ...options, method: "GET" });
  }

  static async post(
    url: string,
    body?: FormData | string,
    options: RequestInit = {},
  ): Promise<Response> {
    return this.request(url, {
      ...options,
      method: "POST",
      body,
    });
  }

  static async put(
    url: string,
    body?: FormData | string,
    options: RequestInit = {},
  ): Promise<Response> {
    return this.request(url, {
      ...options,
      method: "PUT",
      body,
    });
  }

  static async delete(
    url: string,
    options: RequestInit = {},
  ): Promise<Response> {
    return this.request(url, { ...options, method: "DELETE" });
  }
}
